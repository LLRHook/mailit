package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// MetricsIncrementFunc is called to increment a metrics counter in real-time.
type MetricsIncrementFunc func(ctx context.Context, teamID uuid.UUID, eventType string)

// MetricsAggregateHandler processes the periodic metrics:aggregate task.
// It reconciles email_metrics by querying email_events joined with emails
// for the last period.
type MetricsAggregateHandler struct {
	pool        *pgxpool.Pool
	metricsRepo postgres.MetricsRepository
	logger      *slog.Logger
}

// NewMetricsAggregateHandler creates a new MetricsAggregateHandler.
func NewMetricsAggregateHandler(
	pool *pgxpool.Pool,
	metricsRepo postgres.MetricsRepository,
	logger *slog.Logger,
) *MetricsAggregateHandler {
	return &MetricsAggregateHandler{
		pool:        pool,
		metricsRepo: metricsRepo,
		logger:      logger,
	}
}

// ProcessTask runs the aggregation for the previous hour.
func (h *MetricsAggregateHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	now := time.Now().UTC()
	hourStart := now.Truncate(time.Hour).Add(-time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	h.logger.Info("running metrics aggregation", "period_start", hourStart, "period_end", hourEnd)

	// Query email_events joined with emails to get per-team, per-event-type counts for this hour.
	query := `
		SELECT e.team_id, ee.type, COUNT(*)
		FROM email_events ee
		JOIN emails e ON e.id = ee.email_id
		WHERE ee.created_at >= $1 AND ee.created_at < $2
		GROUP BY e.team_id, ee.type`

	rows, err := h.pool.Query(ctx, query, hourStart, hourEnd)
	if err != nil {
		return fmt.Errorf("querying events for aggregation: %w", err)
	}
	defer rows.Close()

	type bucket struct {
		teamID    uuid.UUID
		eventType string
		count     int
	}

	var buckets []bucket
	for rows.Next() {
		var b bucket
		if err := rows.Scan(&b.teamID, &b.eventType, &b.count); err != nil {
			return fmt.Errorf("scanning aggregation row: %w", err)
		}
		buckets = append(buckets, b)
	}
	if rows.Err() != nil {
		return fmt.Errorf("iterating aggregation rows: %w", rows.Err())
	}

	// Group by team and build metrics.
	type teamMetrics struct {
		sent, delivered, bounced, failed, opened, clicked, complained int
	}
	teams := make(map[uuid.UUID]*teamMetrics)

	for _, b := range buckets {
		tm, ok := teams[b.teamID]
		if !ok {
			tm = &teamMetrics{}
			teams[b.teamID] = tm
		}
		switch b.eventType {
		case model.EventSent:
			tm.sent = b.count
		case model.EventDelivered:
			tm.delivered = b.count
		case model.EventBounced:
			tm.bounced = b.count
		case model.EventFailed:
			tm.failed = b.count
		case model.EventOpened:
			tm.opened = b.count
		case model.EventClicked:
			tm.clicked = b.count
		case model.EventComplained:
			tm.complained = b.count
		}
	}

	// Upsert hourly metrics per team.
	// We use a "replace" approach for the aggregation: delete the row first, then insert.
	// Actually, the ON CONFLICT ... DO UPDATE SET ... = EXCLUDED.x approach replaces with the aggregate value.
	// But since real-time increments also use upsert with addition, we need a different approach here.
	// For the periodic reconciliation, we directly set the values using a raw query.
	for teamID, tm := range teams {
		_, err := h.pool.Exec(ctx, `
			INSERT INTO email_metrics (team_id, period_start, period_type, sent, delivered, bounced, failed, opened, clicked, complained)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (team_id, period_start, period_type) DO UPDATE SET
				sent = GREATEST(email_metrics.sent, $4),
				delivered = GREATEST(email_metrics.delivered, $5),
				bounced = GREATEST(email_metrics.bounced, $6),
				failed = GREATEST(email_metrics.failed, $7),
				opened = GREATEST(email_metrics.opened, $8),
				clicked = GREATEST(email_metrics.clicked, $9),
				complained = GREATEST(email_metrics.complained, $10)`,
			teamID, hourStart, model.PeriodTypeHourly,
			tm.sent, tm.delivered, tm.bounced, tm.failed, tm.opened, tm.clicked, tm.complained,
		)
		if err != nil {
			h.logger.Error("failed to upsert hourly metrics", "team_id", teamID, "error", err)
			continue
		}

		// Also aggregate into daily bucket.
		dayStart := hourStart.Truncate(24 * time.Hour)
		_, err = h.pool.Exec(ctx, `
			INSERT INTO email_metrics (team_id, period_start, period_type, sent, delivered, bounced, failed, opened, clicked, complained)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (team_id, period_start, period_type) DO UPDATE SET
				sent = GREATEST(email_metrics.sent, $4),
				delivered = GREATEST(email_metrics.delivered, $5),
				bounced = GREATEST(email_metrics.bounced, $6),
				failed = GREATEST(email_metrics.failed, $7),
				opened = GREATEST(email_metrics.opened, $8),
				clicked = GREATEST(email_metrics.clicked, $9),
				complained = GREATEST(email_metrics.complained, $10)`,
			teamID, dayStart, model.PeriodTypeDaily,
			tm.sent, tm.delivered, tm.bounced, tm.failed, tm.opened, tm.clicked, tm.complained,
		)
		if err != nil {
			h.logger.Error("failed to upsert daily metrics", "team_id", teamID, "error", err)
		}
	}

	h.logger.Info("metrics aggregation complete", "teams_processed", len(teams))
	return nil
}
