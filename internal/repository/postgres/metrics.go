package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type metricsRepository struct {
	pool *pgxpool.Pool
}

// NewMetricsRepository creates a new MetricsRepository backed by PostgreSQL.
func NewMetricsRepository(pool *pgxpool.Pool) MetricsRepository {
	return &metricsRepository{pool: pool}
}

const metricsColumns = `id, team_id, period_start, period_type, sent, delivered, bounced, failed, opened, clicked, complained`

func (r *metricsRepository) Upsert(ctx context.Context, m *model.EmailMetrics) error {
	query := `
		INSERT INTO email_metrics (team_id, period_start, period_type, sent, delivered, bounced, failed, opened, clicked, complained)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (team_id, period_start, period_type) DO UPDATE SET
			sent = email_metrics.sent + EXCLUDED.sent,
			delivered = email_metrics.delivered + EXCLUDED.delivered,
			bounced = email_metrics.bounced + EXCLUDED.bounced,
			failed = email_metrics.failed + EXCLUDED.failed,
			opened = email_metrics.opened + EXCLUDED.opened,
			clicked = email_metrics.clicked + EXCLUDED.clicked,
			complained = email_metrics.complained + EXCLUDED.complained
		RETURNING id`

	return r.pool.QueryRow(ctx, query,
		m.TeamID, m.PeriodStart, m.PeriodType,
		m.Sent, m.Delivered, m.Bounced, m.Failed, m.Opened, m.Clicked, m.Complained,
	).Scan(&m.ID)
}

func (r *metricsRepository) ListByTeam(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) ([]model.EmailMetrics, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM email_metrics
		WHERE team_id = $1 AND period_type = $2 AND period_start >= $3 AND period_start < $4
		ORDER BY period_start ASC`, metricsColumns)

	rows, err := r.pool.Query(ctx, query, teamID, periodType, from, to)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	defer rows.Close()

	metrics, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.EmailMetrics, error) {
		var m model.EmailMetrics
		err := row.Scan(
			&m.ID, &m.TeamID, &m.PeriodStart, &m.PeriodType,
			&m.Sent, &m.Delivered, &m.Bounced, &m.Failed, &m.Opened, &m.Clicked, &m.Complained,
		)
		return m, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect metrics: %w", err)
	}

	return metrics, nil
}

func (r *metricsRepository) AggregateTotals(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) (*model.EmailMetrics, error) {
	query := `
		SELECT
			COALESCE(SUM(sent), 0),
			COALESCE(SUM(delivered), 0),
			COALESCE(SUM(bounced), 0),
			COALESCE(SUM(failed), 0),
			COALESCE(SUM(opened), 0),
			COALESCE(SUM(clicked), 0),
			COALESCE(SUM(complained), 0)
		FROM email_metrics
		WHERE team_id = $1 AND period_type = $2 AND period_start >= $3 AND period_start < $4`

	var m model.EmailMetrics
	m.TeamID = teamID
	m.PeriodType = periodType

	err := r.pool.QueryRow(ctx, query, teamID, periodType, from, to).Scan(
		&m.Sent, &m.Delivered, &m.Bounced, &m.Failed, &m.Opened, &m.Clicked, &m.Complained,
	)
	if err != nil {
		return nil, fmt.Errorf("aggregate metrics: %w", err)
	}

	return &m, nil
}
