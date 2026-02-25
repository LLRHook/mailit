package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)


// Task type constant for webhook delivery (mirrors worker.TaskWebhookDeliver).
const taskWebhookDeliver = "webhook:deliver"

// Webhook event status constants.
const (
	EventStatusPending   = "pending"
	EventStatusDelivered = "delivered"
	EventStatusFailed    = "failed"
)

// Default configuration values.
const (
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 5
	maxResponseBody   = 4096 // max bytes to store from the response body
)

// DispatcherConfig holds configuration for the webhook dispatcher.
type DispatcherConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// Dispatcher manages the lifecycle of webhook events: finding matching webhooks,
// creating event records, enqueuing delivery tasks, and performing HTTP delivery.
type Dispatcher struct {
	webhookRepo      postgres.WebhookRepository
	webhookEventRepo postgres.WebhookEventRepository
	asynqClient      *asynq.Client
	httpClient       *http.Client
	maxRetries       int
	logger           *slog.Logger
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(
	webhookRepo postgres.WebhookRepository,
	webhookEventRepo postgres.WebhookEventRepository,
	asynqClient *asynq.Client,
	cfg DispatcherConfig,
	logger *slog.Logger,
) *Dispatcher {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}

	return &Dispatcher{
		webhookRepo:      webhookRepo,
		webhookEventRepo: webhookEventRepo,
		asynqClient:      asynqClient,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		logger:     logger,
	}
}

// Dispatch finds all active webhooks for a team that subscribe to the given event type,
// creates a webhook_event record for each, and enqueues delivery tasks.
func (d *Dispatcher) Dispatch(ctx context.Context, teamID uuid.UUID, eventType string, payload interface{}) error {
	// 1. Find all webhooks for the team.
	webhooks, err := d.webhookRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("listing webhooks for team %s: %w", teamID, err)
	}

	// 2. Filter to active webhooks that subscribe to this event type.
	now := time.Now().UTC()
	for _, wh := range webhooks {
		if !wh.Active {
			continue
		}

		if !subscribesToEvent(wh.Events, eventType) {
			continue
		}

		// 3. Serialize the payload.
		payloadJSON, marshalErr := toJSONMap(payload)
		if marshalErr != nil {
			d.logger.Error("failed to serialize webhook payload",
				"webhook_id", wh.ID,
				"event_type", eventType,
				"error", marshalErr,
			)
			continue
		}

		// 4. Create a webhook_event record.
		event := &model.WebhookEvent{
			ID:        uuid.New(),
			WebhookID: wh.ID,
			EventType: eventType,
			Payload:   payloadJSON,
			Status:    EventStatusPending,
			Attempts:  0,
			CreatedAt: now,
		}

		if err := d.webhookEventRepo.Create(ctx, event); err != nil {
			d.logger.Error("failed to create webhook event",
				"webhook_id", wh.ID,
				"event_type", eventType,
				"error", err,
			)
			continue
		}

		// 5. Enqueue a webhook:deliver task.
		taskPayload, taskErr := json.Marshal(map[string]string{
			"webhook_event_id": event.ID.String(),
		})
		if taskErr != nil {
			d.logger.Error("failed to marshal webhook deliver task payload", "error", taskErr)
			continue
		}

		task := asynq.NewTask(taskWebhookDeliver, taskPayload, asynq.Queue("default"), asynq.MaxRetry(d.maxRetries))
		if _, enqErr := d.asynqClient.Enqueue(task); enqErr != nil {
			d.logger.Error("failed to enqueue webhook:deliver task",
				"webhook_event_id", event.ID,
				"error", enqErr,
			)
			continue
		}

		d.logger.Debug("webhook event enqueued",
			"webhook_id", wh.ID,
			"webhook_event_id", event.ID,
			"event_type", eventType,
		)
	}

	return nil
}

// Deliver performs the actual HTTP POST to the webhook endpoint for a given webhook event.
func (d *Dispatcher) Deliver(ctx context.Context, webhookEventID uuid.UUID) error {
	// 1. Get the webhook event.
	event, err := d.webhookEventRepo.GetByID(ctx, webhookEventID)
	if err != nil {
		return fmt.Errorf("fetching webhook event %s: %w", webhookEventID, err)
	}

	// 2. Get the webhook (for URL and signing secret).
	wh, err := d.webhookRepo.GetByID(ctx, event.WebhookID)
	if err != nil {
		return fmt.Errorf("fetching webhook %s: %w", event.WebhookID, err)
	}

	// 3. Build the JSON payload body.
	body, err := json.Marshal(map[string]interface{}{
		"type":       event.EventType,
		"created_at": event.CreatedAt.Format(time.RFC3339),
		"data":       event.Payload,
	})
	if err != nil {
		return fmt.Errorf("marshalling webhook payload: %w", err)
	}

	// 4. Sign the payload.
	timestamp := time.Now().UTC().Unix()
	signature := Sign(body, wh.SigningSecret, timestamp)

	// 5. Build the HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MailIt-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", wh.ID.String())
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", timestamp))

	// 6. Execute the request.
	event.Attempts++
	resp, httpErr := d.httpClient.Do(req)

	now := time.Now().UTC()

	if httpErr != nil {
		// Network error.
		event.Status = EventStatusFailed
		errMsg := httpErr.Error()
		event.ResponseBody = &errMsg
		if updateErr := d.webhookEventRepo.Update(ctx, event); updateErr != nil {
			d.logger.Error("failed to update webhook event after network error", "error", updateErr)
		}
		return fmt.Errorf("HTTP request to %s failed: %w", wh.URL, httpErr)
	}
	defer func() { _ = resp.Body.Close() }()

	// 7. Read the response.
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	respCode := resp.StatusCode
	respBodyStr := string(respBody)

	event.ResponseCode = &respCode
	event.ResponseBody = &respBodyStr

	// 8. Determine success or failure.
	if respCode >= 200 && respCode < 300 {
		event.Status = EventStatusDelivered
		d.logger.Info("webhook delivered successfully",
			"webhook_event_id", webhookEventID,
			"status_code", respCode,
		)
	} else {
		event.Status = EventStatusFailed

		// Schedule retry if we haven't exceeded max attempts.
		if event.Attempts < d.maxRetries {
			retryAt := calculateRetryTime(now, event.Attempts)
			event.NextRetryAt = &retryAt
			d.logger.Warn("webhook delivery failed, will retry",
				"webhook_event_id", webhookEventID,
				"status_code", respCode,
				"attempt", event.Attempts,
				"next_retry_at", retryAt.Format(time.RFC3339),
			)
		} else {
			d.logger.Error("webhook delivery failed permanently",
				"webhook_event_id", webhookEventID,
				"status_code", respCode,
				"attempts", event.Attempts,
			)
		}
	}

	// 9. Update the webhook event record.
	if updateErr := d.webhookEventRepo.Update(ctx, event); updateErr != nil {
		d.logger.Error("failed to update webhook event", "error", updateErr)
	}

	// Return an error for non-2xx responses so asynq can retry.
	if respCode < 200 || respCode >= 300 {
		return fmt.Errorf("webhook delivery to %s returned status %d", wh.URL, respCode)
	}

	return nil
}

// Sign creates an HMAC-SHA256 signature for a webhook payload.
// The signed content is "{timestamp}.{payload}" to prevent replay attacks.
func Sign(payload []byte, secret string, timestamp int64) string {
	signedContent := fmt.Sprintf("%d.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedContent))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies that a webhook signature is valid.
func VerifySignature(payload []byte, secret string, timestamp int64, signature string) bool {
	expected := Sign(payload, secret, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// subscribesToEvent checks if a webhook's event list includes the given event type.
// A wildcard "*" matches all events.
func subscribesToEvent(events []string, eventType string) bool {
	for _, e := range events {
		if e == "*" || e == eventType {
			return true
		}
	}
	return false
}

// toJSONMap converts an arbitrary value to a model.JSONMap via JSON marshalling.
func toJSONMap(v interface{}) (model.JSONMap, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m model.JSONMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// calculateRetryTime computes the next retry time using exponential backoff.
// Retry intervals: 30s, 2m, 10m, 30m, 2h (roughly).
func calculateRetryTime(now time.Time, attempt int) time.Time {
	backoffs := []time.Duration{
		30 * time.Second,
		2 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		2 * time.Hour,
	}

	idx := attempt - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(backoffs) {
		idx = len(backoffs) - 1
	}

	return now.Add(backoffs[idx])
}
