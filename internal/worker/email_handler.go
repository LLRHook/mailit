package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// EmailSender defines the interface for actually delivering email messages.
// This is implemented by the engine.Sender type (not yet created).
type EmailSender interface {
	// SendEmail delivers an email via SMTP to each recipient's MX server.
	// It returns a result per recipient indicating success or failure.
	SendEmail(ctx context.Context, msg *OutboundMessage) ([]RecipientResult, error)
}

// OutboundMessage represents an email ready for SMTP delivery.
type OutboundMessage struct {
	MessageID    string
	From         string
	To           []string
	Cc           []string
	Bcc          []string
	ReplyTo      string
	Subject      string
	HTMLBody     string
	TextBody     string
	Headers      map[string]string
	DKIMDomain   string
	DKIMSelector string
	DKIMKey      []byte
}

// RecipientResult captures the delivery outcome for a single recipient.
type RecipientResult struct {
	Recipient string
	Success   bool
	Code      int
	Message   string
	Permanent bool // true if the error is permanent (5xx), false if temporary (4xx)
}

// WebhookDispatchFunc is the function signature for dispatching webhook events.
type WebhookDispatchFunc func(ctx context.Context, teamID uuid.UUID, eventType string, payload interface{})

// EmailSendHandler processes email:send tasks.
type EmailSendHandler struct {
	emailRepo       postgres.EmailRepository
	eventRepo       postgres.EmailEventRepository
	domainRepo      postgres.DomainRepository
	suppressionRepo postgres.SuppressionRepository
	sender          EmailSender
	webhookDispatch WebhookDispatchFunc
	logger          *slog.Logger
}

// NewEmailSendHandler creates a new EmailSendHandler.
func NewEmailSendHandler(
	emailRepo postgres.EmailRepository,
	eventRepo postgres.EmailEventRepository,
	domainRepo postgres.DomainRepository,
	suppressionRepo postgres.SuppressionRepository,
	sender EmailSender,
	webhookDispatch WebhookDispatchFunc,
	logger *slog.Logger,
) *EmailSendHandler {
	return &EmailSendHandler{
		emailRepo:       emailRepo,
		eventRepo:       eventRepo,
		domainRepo:      domainRepo,
		suppressionRepo: suppressionRepo,
		sender:          sender,
		webhookDispatch: webhookDispatch,
		logger:          logger,
	}
}

// ProcessTask handles the email:send task.
func (h *EmailSendHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p EmailSendPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling email:send payload: %w", err)
	}

	log := h.logger.With("email_id", p.EmailID, "team_id", p.TeamID)

	// 1. Get the email from DB.
	email, err := h.emailRepo.GetByID(ctx, p.EmailID)
	if err != nil {
		return fmt.Errorf("fetching email %s: %w", p.EmailID, err)
	}

	// If the email was cancelled or already sent, skip.
	if email.Status == model.EmailStatusCancelled || email.Status == model.EmailStatusSent || email.Status == model.EmailStatusDelivered {
		log.Info("skipping email with terminal status", "status", email.Status)
		return nil
	}

	// 2. Filter out suppressed recipients.
	filteredTo := h.filterSuppressed(ctx, email.TeamID, email.ToAddresses, log)
	filteredCc := h.filterSuppressed(ctx, email.TeamID, email.CcAddresses, log)
	filteredBcc := h.filterSuppressed(ctx, email.TeamID, email.BccAddresses, log)

	allRecipients := make([]string, 0, len(filteredTo)+len(filteredCc)+len(filteredBcc))
	allRecipients = append(allRecipients, filteredTo...)
	allRecipients = append(allRecipients, filteredCc...)
	allRecipients = append(allRecipients, filteredBcc...)

	if len(allRecipients) == 0 {
		log.Warn("all recipients are suppressed, marking email as failed")
		email.Status = model.EmailStatusFailed
		lastErr := "all recipients are on the suppression list"
		email.LastError = &lastErr
		email.UpdatedAt = time.Now().UTC()
		if updateErr := h.emailRepo.Update(ctx, email); updateErr != nil {
			log.Error("failed to update email status", "error", updateErr)
		}
		return nil
	}

	// 3. Get domain for DKIM signing.
	var dkimDomain, dkimSelector string
	var dkimKey []byte

	if email.DomainID != nil {
		domain, domErr := h.domainRepo.GetByID(ctx, *email.DomainID)
		if domErr != nil {
			log.Warn("failed to fetch domain for DKIM, sending without DKIM", "error", domErr)
		} else if domain.DKIMPrivateKey != nil {
			dkimDomain = domain.Name
			dkimSelector = domain.DKIMSelector
			dkimKey = []byte(*domain.DKIMPrivateKey)
		}
	} else {
		// Try to resolve the domain from the From address.
		fromDomain := extractDomain(email.FromAddress)
		if fromDomain != "" {
			domain, domErr := h.domainRepo.GetByTeamAndName(ctx, email.TeamID, fromDomain)
			if domErr == nil && domain.DKIMPrivateKey != nil {
				dkimDomain = domain.Name
				dkimSelector = domain.DKIMSelector
				dkimKey = []byte(*domain.DKIMPrivateKey)
			}
		}
	}

	// 4. Update email status to "sending".
	email.Status = model.EmailStatusSending
	email.UpdatedAt = time.Now().UTC()
	if err := h.emailRepo.Update(ctx, email); err != nil {
		return fmt.Errorf("updating email status to sending: %w", err)
	}

	// 5. Build outbound message.
	msg := &OutboundMessage{
		MessageID:    ptrToString(email.MessageID),
		From:         email.FromAddress,
		To:           filteredTo,
		Cc:           filteredCc,
		Bcc:          filteredBcc,
		ReplyTo:      ptrToString(email.ReplyTo),
		Subject:      email.Subject,
		HTMLBody:     ptrToString(email.HTMLBody),
		TextBody:     ptrToString(email.TextBody),
		Headers:      jsonMapToStringMap(email.Headers),
		DKIMDomain:   dkimDomain,
		DKIMSelector: dkimSelector,
		DKIMKey:      dkimKey,
	}

	// 6. Send via SMTP engine.
	results, err := h.sender.SendEmail(ctx, msg)
	if err != nil {
		// A transport-level error (e.g. DNS failure) is usually temporary.
		email.RetryCount++
		email.LastError = strPtr(err.Error())
		email.Status = model.EmailStatusQueued
		email.UpdatedAt = time.Now().UTC()
		if updateErr := h.emailRepo.Update(ctx, email); updateErr != nil {
			log.Error("failed to update email after send error", "error", updateErr)
		}
		return fmt.Errorf("sending email: %w", err)
	}

	// 7. Process per-recipient results.
	now := time.Now().UTC()
	var hasTemporaryFailure bool
	allSucceeded := true

	for _, r := range results {
		if r.Success {
			h.createEvent(ctx, email.ID, model.EventSent, &r.Recipient, model.JSONMap{
				"code":    r.Code,
				"message": r.Message,
			})
			if h.webhookDispatch != nil {
				h.webhookDispatch(ctx, email.TeamID, "email.sent", map[string]interface{}{
					"email_id":  email.ID.String(),
					"recipient": r.Recipient,
					"timestamp": now.Format(time.RFC3339),
				})
			}
		} else if r.Permanent {
			allSucceeded = false
			h.createEvent(ctx, email.ID, model.EventBounced, &r.Recipient, model.JSONMap{
				"code":    r.Code,
				"message": r.Message,
				"type":    "hard",
			})
			// Enqueue a bounce:process task for hard bounces.
			bounceTask, taskErr := NewBounceProcessTask(email.ID, r.Code, r.Message, r.Recipient)
			if taskErr != nil {
				log.Error("failed to create bounce task", "error", taskErr, "recipient", r.Recipient)
			} else {
				if _, enqErr := asynq.NewClient(asynq.RedisClientOpt{}).Enqueue(bounceTask); enqErr != nil {
					log.Error("failed to enqueue bounce task", "error", enqErr, "recipient", r.Recipient)
				}
			}
			if h.webhookDispatch != nil {
				h.webhookDispatch(ctx, email.TeamID, "email.bounced", map[string]interface{}{
					"email_id":  email.ID.String(),
					"recipient": r.Recipient,
					"code":      r.Code,
					"message":   r.Message,
					"timestamp": now.Format(time.RFC3339),
				})
			}
		} else {
			// Temporary failure: mark for retry.
			allSucceeded = false
			hasTemporaryFailure = true
			h.createEvent(ctx, email.ID, model.EventFailed, &r.Recipient, model.JSONMap{
				"code":      r.Code,
				"message":   r.Message,
				"type":      "temporary",
				"will_retry": true,
			})
		}
	}

	// 8. Update final email status.
	if allSucceeded {
		email.Status = model.EmailStatusSent
		email.SentAt = &now
	} else if hasTemporaryFailure {
		// Return an error so asynq retries the whole task.
		email.Status = model.EmailStatusQueued
		email.RetryCount++
		email.UpdatedAt = now
		if updateErr := h.emailRepo.Update(ctx, email); updateErr != nil {
			log.Error("failed to update email for retry", "error", updateErr)
		}
		return fmt.Errorf("temporary delivery failures for email %s, will retry", email.ID)
	} else {
		// All failures were permanent.
		email.Status = model.EmailStatusFailed
		email.LastError = strPtr("all recipients failed permanently")
	}

	email.UpdatedAt = now
	if err := h.emailRepo.Update(ctx, email); err != nil {
		log.Error("failed to update final email status", "error", err)
		return fmt.Errorf("updating final email status: %w", err)
	}

	return nil
}

// filterSuppressed removes suppressed addresses from the given list.
func (h *EmailSendHandler) filterSuppressed(ctx context.Context, teamID uuid.UUID, addresses []string, log *slog.Logger) []string {
	if len(addresses) == 0 {
		return addresses
	}
	filtered := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		entry, _ := h.suppressionRepo.GetByTeamAndEmail(ctx, teamID, addr)
		if entry != nil {
			log.Info("skipping suppressed recipient", "email", addr, "reason", entry.Reason)
			continue
		}
		filtered = append(filtered, addr)
	}
	return filtered
}

// createEvent is a helper to create an email event record.
func (h *EmailSendHandler) createEvent(ctx context.Context, emailID uuid.UUID, eventType string, recipient *string, payload model.JSONMap) {
	event := &model.EmailEvent{
		ID:        uuid.New(),
		EmailID:   emailID,
		Type:      eventType,
		Payload:   payload,
		Recipient: recipient,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.eventRepo.Create(ctx, event); err != nil {
		h.logger.Error("failed to create email event", "error", err, "email_id", emailID, "event_type", eventType)
	}
}

// extractDomain extracts the domain part from an email address.
func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// ptrToString safely dereferences a string pointer, returning empty string for nil.
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}

// jsonMapToStringMap converts a model.JSONMap to map[string]string.
func jsonMapToStringMap(m model.JSONMap) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}
