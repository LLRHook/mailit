package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type logRepository struct {
	pool *pgxpool.Pool
}

// NewLogRepository creates a new LogRepository backed by PostgreSQL.
func NewLogRepository(pool *pgxpool.Pool) LogRepository {
	return &logRepository{pool: pool}
}

const logColumns = `id, team_id, level, message, metadata, request_id, method, path, status_code, duration_ms, ip_address, created_at`

func (r *logRepository) Create(ctx context.Context, log *model.Log) error {
	query := fmt.Sprintf(`
		INSERT INTO logs (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING %s`, logColumns, logColumns)

	return r.pool.QueryRow(ctx, query,
		log.ID, log.TeamID, log.Level, log.Message, log.Metadata, log.RequestID,
		log.Method, log.Path, log.StatusCode, log.DurationMs, log.IPAddress, log.CreatedAt,
	).Scan(
		&log.ID, &log.TeamID, &log.Level, &log.Message, &log.Metadata, &log.RequestID,
		&log.Method, &log.Path, &log.StatusCode, &log.DurationMs, &log.IPAddress, &log.CreatedAt,
	)
}

func (r *logRepository) List(ctx context.Context, teamID uuid.UUID, level string, limit, offset int) ([]model.Log, int, error) {
	var countQuery string
	var dataQuery string
	var countArgs []interface{}
	var dataArgs []interface{}

	if level != "" {
		countQuery = `SELECT COUNT(*) FROM logs WHERE team_id = $1 AND level = $2`
		countArgs = []interface{}{teamID, level}

		dataQuery = fmt.Sprintf(`
			SELECT %s FROM logs WHERE team_id = $1 AND level = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`, logColumns)
		dataArgs = []interface{}{teamID, level, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM logs WHERE team_id = $1`
		countArgs = []interface{}{teamID}

		dataQuery = fmt.Sprintf(`
			SELECT %s FROM logs WHERE team_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`, logColumns)
		dataArgs = []interface{}{teamID, limit, offset}
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count logs: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list logs: %w", err)
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Log, error) {
		var l model.Log
		err := row.Scan(
			&l.ID, &l.TeamID, &l.Level, &l.Message, &l.Metadata, &l.RequestID,
			&l.Method, &l.Path, &l.StatusCode, &l.DurationMs, &l.IPAddress, &l.CreatedAt,
		)
		return l, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect logs: %w", err)
	}

	return logs, total, nil
}
