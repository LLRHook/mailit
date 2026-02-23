package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type suppressionRepository struct {
	pool *pgxpool.Pool
}

// NewSuppressionRepository creates a new SuppressionRepository backed by PostgreSQL.
func NewSuppressionRepository(pool *pgxpool.Pool) SuppressionRepository {
	return &suppressionRepository{pool: pool}
}

const suppressionColumns = `id, team_id, email, reason, details, created_at`

func (r *suppressionRepository) Create(ctx context.Context, entry *model.SuppressionEntry) error {
	query := fmt.Sprintf(`
		INSERT INTO suppression_list (%s)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, suppressionColumns, suppressionColumns)

	return r.pool.QueryRow(ctx, query,
		entry.ID, entry.TeamID, entry.Email, entry.Reason, entry.Details, entry.CreatedAt,
	).Scan(
		&entry.ID, &entry.TeamID, &entry.Email, &entry.Reason, &entry.Details, &entry.CreatedAt,
	)
}

func (r *suppressionRepository) GetByTeamAndEmail(ctx context.Context, teamID uuid.UUID, email string) (*model.SuppressionEntry, error) {
	query := fmt.Sprintf(`SELECT %s FROM suppression_list WHERE team_id = $1 AND email = $2`, suppressionColumns)

	entry := &model.SuppressionEntry{}
	err := r.pool.QueryRow(ctx, query, teamID, email).Scan(
		&entry.ID, &entry.TeamID, &entry.Email, &entry.Reason, &entry.Details, &entry.CreatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("suppression entry")
		}
		return nil, fmt.Errorf("get suppression entry by email: %w", err)
	}
	return entry, nil
}

func (r *suppressionRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.SuppressionEntry, int, error) {
	countQuery := `SELECT COUNT(*) FROM suppression_list WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count suppression entries: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM suppression_list WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, suppressionColumns)

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list suppression entries: %w", err)
	}
	defer rows.Close()

	entries, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.SuppressionEntry, error) {
		var e model.SuppressionEntry
		err := row.Scan(&e.ID, &e.TeamID, &e.Email, &e.Reason, &e.Details, &e.CreatedAt)
		return e, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect suppression entries: %w", err)
	}

	return entries, total, nil
}

func (r *suppressionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM suppression_list WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete suppression entry: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("suppression entry")
	}
	return nil
}
