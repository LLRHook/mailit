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

// BroadcastSendHandler processes broadcast:send tasks by expanding a broadcast
// into individual email:send tasks for each contact in the target audience.
type BroadcastSendHandler struct {
	broadcastRepo postgres.BroadcastRepository
	contactRepo   postgres.ContactRepository
	audienceRepo  postgres.AudienceRepository
	emailRepo     postgres.EmailRepository
	asynqClient   *asynq.Client
	logger        *slog.Logger
}

// NewBroadcastSendHandler creates a new BroadcastSendHandler.
func NewBroadcastSendHandler(
	broadcastRepo postgres.BroadcastRepository,
	contactRepo postgres.ContactRepository,
	audienceRepo postgres.AudienceRepository,
	emailRepo postgres.EmailRepository,
	asynqClient *asynq.Client,
	logger *slog.Logger,
) *BroadcastSendHandler {
	return &BroadcastSendHandler{
		broadcastRepo: broadcastRepo,
		contactRepo:   contactRepo,
		audienceRepo:  audienceRepo,
		emailRepo:     emailRepo,
		asynqClient:   asynqClient,
		logger:        logger,
	}
}

// ProcessTask handles the broadcast:send task.
func (h *BroadcastSendHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p BroadcastSendPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling broadcast:send payload: %w", err)
	}

	log := h.logger.With("broadcast_id", p.BroadcastID, "team_id", p.TeamID)

	// 1. Get the broadcast from DB.
	broadcast, err := h.broadcastRepo.GetByID(ctx, p.BroadcastID)
	if err != nil {
		return fmt.Errorf("fetching broadcast %s: %w", p.BroadcastID, err)
	}

	// Only process queued broadcasts.
	if broadcast.Status != model.BroadcastStatusQueued {
		log.Info("skipping broadcast with non-queued status", "status", broadcast.Status)
		return nil
	}

	// 2. Verify the audience exists.
	if broadcast.AudienceID == nil {
		return fmt.Errorf("broadcast %s has no audience", p.BroadcastID)
	}

	_, err = h.audienceRepo.GetByID(ctx, *broadcast.AudienceID)
	if err != nil {
		return fmt.Errorf("fetching audience %s: %w", broadcast.AudienceID, err)
	}

	// 3. Update broadcast status to "sending".
	broadcast.Status = model.BroadcastStatusSending
	now := time.Now().UTC()
	broadcast.SentAt = &now
	broadcast.UpdatedAt = now
	if err := h.broadcastRepo.Update(ctx, broadcast); err != nil {
		return fmt.Errorf("updating broadcast to sending: %w", err)
	}

	// 4. Fetch all contacts from the audience in pages.
	const pageSize = 500
	var totalEnqueued int
	offset := 0

	for {
		contacts, total, err := h.contactRepo.List(ctx, *broadcast.AudienceID, pageSize, offset)
		if err != nil {
			return fmt.Errorf("listing contacts at offset %d: %w", offset, err)
		}

		for _, contact := range contacts {
			// Skip unsubscribed contacts.
			if contact.Unsubscribed {
				log.Debug("skipping unsubscribed contact", "contact_id", contact.ID, "email", contact.Email)
				continue
			}

			// 5. Create an email record for this contact.
			emailID := uuid.New()
			email := &model.Email{
				ID:          emailID,
				TeamID:      p.TeamID,
				DomainID:    nil,
				FromAddress: ptrToString(broadcast.FromAddress),
				ToAddresses: []string{contact.Email},
				Subject:     ptrToString(broadcast.Subject),
				HTMLBody:    broadcast.HTMLBody,
				TextBody:    broadcast.TextBody,
				Status:      model.EmailStatusQueued,
				Tags:        []string{"broadcast:" + p.BroadcastID.String()},
				Headers:     model.JSONMap{},
				Attachments: model.JSONArray{},
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := h.emailRepo.Create(ctx, email); err != nil {
				log.Error("failed to create email for contact", "contact_id", contact.ID, "error", err)
				continue
			}

			// 6. Enqueue the email:send task.
			task, taskErr := NewEmailSendTask(emailID, p.TeamID)
			if taskErr != nil {
				log.Error("failed to create email:send task", "email_id", emailID, "error", taskErr)
				continue
			}

			if _, enqErr := h.asynqClient.Enqueue(task); enqErr != nil {
				log.Error("failed to enqueue email:send task", "email_id", emailID, "error", enqErr)
				continue
			}

			totalEnqueued++
		}

		offset += pageSize
		if offset >= total {
			break
		}
	}

	// 7. Update broadcast with recipient count.
	broadcast.TotalRecipients = totalEnqueued
	broadcast.UpdatedAt = time.Now().UTC()

	// If no recipients were enqueued, mark as sent immediately.
	if totalEnqueued == 0 {
		broadcast.Status = model.BroadcastStatusSent
		log.Warn("broadcast had zero eligible recipients")
	}

	if err := h.broadcastRepo.Update(ctx, broadcast); err != nil {
		return fmt.Errorf("updating broadcast recipient count: %w", err)
	}

	log.Info("broadcast expanded into email tasks", "total_enqueued", totalEnqueued)
	return nil
}
