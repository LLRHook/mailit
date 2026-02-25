package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// InboundHandler processes inbound:process tasks by marking inbound emails
// as processed and dispatching webhook events.
type InboundHandler struct {
	inboundRepo     postgres.InboundEmailRepository
	webhookDispatch WebhookDispatchFunc
	logger          *slog.Logger
}

// NewInboundHandler creates a new InboundHandler.
func NewInboundHandler(
	inboundRepo postgres.InboundEmailRepository,
	webhookDispatch WebhookDispatchFunc,
	logger *slog.Logger,
) *InboundHandler {
	return &InboundHandler{
		inboundRepo:     inboundRepo,
		webhookDispatch: webhookDispatch,
		logger:          logger,
	}
}

// ProcessTask handles the inbound:process task.
func (h *InboundHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p InboundProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling inbound:process payload: %w", err)
	}

	log := h.logger.With("inbound_email_id", p.InboundEmailID)

	// 1. Get the inbound email from DB.
	inbound, err := h.inboundRepo.GetByID(ctx, p.InboundEmailID)
	if err != nil {
		return fmt.Errorf("fetching inbound email %s: %w", p.InboundEmailID, err)
	}

	// Skip if already processed.
	if inbound.Processed {
		log.Info("inbound email already processed, skipping")
		return nil
	}

	// 2. Mark as processed.
	inbound.Processed = true
	if err := h.inboundRepo.Update(ctx, inbound); err != nil {
		return fmt.Errorf("marking inbound email as processed: %w", err)
	}

	// 3. Dispatch "email.inbound" webhook event with full payload.
	if h.webhookDispatch != nil {
		webhookPayload := buildInboundWebhookPayload(inbound)
		h.webhookDispatch(ctx, inbound.TeamID, "email.inbound", webhookPayload)
		log.Info("dispatched email.inbound webhook")
	}

	log.Info("inbound email processed successfully")
	return nil
}

// buildInboundWebhookPayload constructs the webhook payload for an inbound email event.
func buildInboundWebhookPayload(inbound *model.InboundEmail) map[string]interface{} {
	payload := map[string]interface{}{
		"inbound_email_id": inbound.ID.String(),
		"team_id":          inbound.TeamID.String(),
		"from":             inbound.FromAddress,
		"to":               inbound.ToAddresses,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}
	if inbound.Subject != nil {
		payload["subject"] = *inbound.Subject
	}
	if inbound.HTMLBody != nil {
		payload["html_body"] = *inbound.HTMLBody
	}
	if inbound.TextBody != nil {
		payload["text_body"] = *inbound.TextBody
	}
	if len(inbound.CcAddresses) > 0 {
		payload["cc"] = inbound.CcAddresses
	}
	if len(inbound.Attachments) > 0 {
		payload["attachments"] = inbound.Attachments
	}
	return payload
}
