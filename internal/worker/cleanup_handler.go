package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// Default retention periods for cleanup.
const (
	WebhookEventRetention = 30 * 24 * time.Hour // 30 days
	LogRetention          = 90 * 24 * time.Hour  // 90 days
)

// CleanupHandler processes cleanup:expired tasks by removing old data
// based on retention policies.
type CleanupHandler struct {
	webhookEventRepo postgres.WebhookEventRepository
	logRepo          postgres.LogRepository
	logger           *slog.Logger
}

// NewCleanupHandler creates a new CleanupHandler.
func NewCleanupHandler(
	webhookEventRepo postgres.WebhookEventRepository,
	logRepo postgres.LogRepository,
	logger *slog.Logger,
) *CleanupHandler {
	return &CleanupHandler{
		webhookEventRepo: webhookEventRepo,
		logRepo:          logRepo,
		logger:           logger,
	}
}

// ProcessTask handles the cleanup:expired task.
func (h *CleanupHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	log := h.logger.With("task", TaskCleanupExpired)
	log.Info("starting cleanup of expired data")

	var errs []error

	// 1. Clean up old webhook events.
	webhookCutoff := time.Now().UTC().Add(-WebhookEventRetention)
	deletedWebhookEvents, err := h.webhookEventRepo.DeleteOlderThan(ctx, webhookCutoff)
	if err != nil {
		log.Error("failed to clean up webhook events", "error", err)
		errs = append(errs, fmt.Errorf("webhook events cleanup: %w", err))
	} else {
		log.Info("cleaned up old webhook events", "deleted", deletedWebhookEvents, "cutoff", webhookCutoff.Format(time.RFC3339))
	}

	// 2. Additional cleanup can be added here as retention policies grow:
	//    - Old email events
	//    - Old logs
	//    - Expired Redis keys
	//    - Orphaned attachments

	if len(errs) > 0 {
		return fmt.Errorf("cleanup completed with %d errors: %v", len(errs), errs)
	}

	log.Info("cleanup completed successfully")
	return nil
}
