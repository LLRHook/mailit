package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/webhook"
)

// WebhookDeliverHandler processes webhook:deliver tasks by delegating to the
// webhook dispatcher for the actual HTTP delivery.
type WebhookDeliverHandler struct {
	dispatcher *webhook.Dispatcher
	logger     *slog.Logger
}

// NewWebhookDeliverHandler creates a new WebhookDeliverHandler.
func NewWebhookDeliverHandler(
	dispatcher *webhook.Dispatcher,
	logger *slog.Logger,
) *WebhookDeliverHandler {
	return &WebhookDeliverHandler{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// ProcessTask handles the webhook:deliver task.
func (h *WebhookDeliverHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p WebhookDeliverPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling webhook:deliver payload: %w", err)
	}

	log := h.logger.With("webhook_event_id", p.WebhookEventID)
	log.Info("delivering webhook event")

	if err := h.dispatcher.Deliver(ctx, p.WebhookEventID); err != nil {
		log.Error("webhook delivery failed", "error", err)
		return fmt.Errorf("delivering webhook event %s: %w", p.WebhookEventID, err)
	}

	log.Info("webhook event delivered successfully")
	return nil
}
