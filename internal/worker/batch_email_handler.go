package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

// BatchEmailSendHandler processes email:send_batch tasks by expanding
// them into individual email:send tasks.
type BatchEmailSendHandler struct {
	asynqClient *asynq.Client
	logger      *slog.Logger
}

// NewBatchEmailSendHandler creates a new BatchEmailSendHandler.
func NewBatchEmailSendHandler(asynqClient *asynq.Client, logger *slog.Logger) *BatchEmailSendHandler {
	return &BatchEmailSendHandler{
		asynqClient: asynqClient,
		logger:      logger,
	}
}

// ProcessTask handles the email:send_batch task.
func (h *BatchEmailSendHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p EmailBatchSendPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling email:send_batch payload: %w", err)
	}

	log := h.logger.With("team_id", p.TeamID, "batch_size", len(p.EmailIDs))
	log.Info("processing batch email send")

	var enqueueErrors int
	for _, emailID := range p.EmailIDs {
		task, err := NewEmailSendTask(emailID, p.TeamID)
		if err != nil {
			log.Error("failed to create email:send task", "email_id", emailID, "error", err)
			enqueueErrors++
			continue
		}
		if _, err := h.asynqClient.EnqueueContext(ctx, task); err != nil {
			log.Error("failed to enqueue email:send task", "email_id", emailID, "error", err)
			enqueueErrors++
			continue
		}
	}

	if enqueueErrors > 0 {
		log.Warn("batch completed with errors", "failed", enqueueErrors, "total", len(p.EmailIDs))
	} else {
		log.Info("batch email send completed", "total", len(p.EmailIDs))
	}

	return nil
}
