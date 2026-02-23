package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// Bounce classification constants.
const (
	BounceTypeHard      = "hard"
	BounceTypeSoft      = "soft"
	BounceTypeComplaint = "complaint"
)

// BounceHandler processes bounce:process tasks by classifying bounces and
// updating suppression lists.
type BounceHandler struct {
	emailRepo       postgres.EmailRepository
	eventRepo       postgres.EmailEventRepository
	suppressionRepo postgres.SuppressionRepository
	logger          *slog.Logger
}

// NewBounceHandler creates a new BounceHandler.
func NewBounceHandler(
	emailRepo postgres.EmailRepository,
	eventRepo postgres.EmailEventRepository,
	suppressionRepo postgres.SuppressionRepository,
	logger *slog.Logger,
) *BounceHandler {
	return &BounceHandler{
		emailRepo:       emailRepo,
		eventRepo:       eventRepo,
		suppressionRepo: suppressionRepo,
		logger:          logger,
	}
}

// ProcessTask handles the bounce:process task.
func (h *BounceHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p BounceProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling bounce:process payload: %w", err)
	}

	log := h.logger.With("email_id", p.EmailID, "recipient", p.Recipient, "code", p.Code)

	// 1. Classify the bounce.
	bounceType := classifyBounce(p.Code, p.Message)
	log.Info("bounce classified", "bounce_type", bounceType)

	// 2. Get the email for team context.
	email, err := h.emailRepo.GetByID(ctx, p.EmailID)
	if err != nil {
		return fmt.Errorf("fetching email %s: %w", p.EmailID, err)
	}

	now := time.Now().UTC()

	switch bounceType {
	case BounceTypeHard:
		// Hard bounce: add to suppression list, update email to bounced.
		if err := h.addToSuppressionList(ctx, email.TeamID, p.Recipient, model.SuppressionBounce, p.Message); err != nil {
			log.Error("failed to add hard bounce to suppression list", "error", err)
		}

		email.Status = model.EmailStatusBounced
		email.LastError = &p.Message
		email.UpdatedAt = now
		if err := h.emailRepo.Update(ctx, email); err != nil {
			log.Error("failed to update email to bounced", "error", err)
		}

		h.createEvent(ctx, email.ID, model.EventBounced, &p.Recipient, model.JSONMap{
			"type":    BounceTypeHard,
			"code":    p.Code,
			"message": p.Message,
		})

	case BounceTypeSoft:
		// Soft bounce: create a bounce event but don't suppress.
		// The email handler retries via asynq's retry mechanism.
		h.createEvent(ctx, email.ID, model.EventBounced, &p.Recipient, model.JSONMap{
			"type":    BounceTypeSoft,
			"code":    p.Code,
			"message": p.Message,
		})
		log.Info("soft bounce recorded, email will be retried by send handler")

	case BounceTypeComplaint:
		// Complaint: add to suppression list with "complaint" reason.
		if err := h.addToSuppressionList(ctx, email.TeamID, p.Recipient, model.SuppressionComplaint, p.Message); err != nil {
			log.Error("failed to add complaint to suppression list", "error", err)
		}

		h.createEvent(ctx, email.ID, model.EventComplained, &p.Recipient, model.JSONMap{
			"code":    p.Code,
			"message": p.Message,
		})
	}

	return nil
}

// classifyBounce determines the type of bounce based on the SMTP response code and message.
func classifyBounce(code int, message string) string {
	// 5xx codes are permanent failures (hard bounces).
	// 4xx codes are temporary failures (soft bounces).
	// Certain 5xx codes indicate complaints or policy blocks.
	switch {
	case code >= 500 && code < 600:
		// Check for complaint indicators in the message.
		lowerMsg := lowerContains(message)
		if lowerMsg("spam") || lowerMsg("complaint") || lowerMsg("abuse") || lowerMsg("blocked") {
			return BounceTypeComplaint
		}
		return BounceTypeHard
	case code >= 400 && code < 500:
		return BounceTypeSoft
	default:
		// Unknown codes default to soft to avoid false suppressions.
		return BounceTypeSoft
	}
}

// lowerContains returns a closure that checks if a lowercased message contains the given substring.
func lowerContains(message string) func(string) bool {
	lower := toLower(message)
	return func(substr string) bool {
		return contains(lower, substr)
	}
}

// toLower is a simple ASCII lowercase conversion.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

// searchString performs a simple substring search.
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// addToSuppressionList adds a recipient to the team's suppression list if not already present.
func (h *BounceHandler) addToSuppressionList(ctx context.Context, teamID uuid.UUID, email, reason, details string) error {
	// Check if already suppressed.
	existing, _ := h.suppressionRepo.GetByTeamAndEmail(ctx, teamID, email)
	if existing != nil {
		h.logger.Debug("recipient already on suppression list", "email", email, "existing_reason", existing.Reason)
		return nil
	}

	entry := &model.SuppressionEntry{
		ID:        uuid.New(),
		TeamID:    teamID,
		Email:     email,
		Reason:    reason,
		Details:   &details,
		CreatedAt: time.Now().UTC(),
	}

	return h.suppressionRepo.Create(ctx, entry)
}

// createEvent is a helper to create an email event record.
func (h *BounceHandler) createEvent(ctx context.Context, emailID uuid.UUID, eventType string, recipient *string, payload model.JSONMap) {
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
