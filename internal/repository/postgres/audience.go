package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type audienceRepository struct {
	pool *pgxpool.Pool
}

// NewAudienceRepository creates a new AudienceRepository backed by PostgreSQL.
func NewAudienceRepository(pool *pgxpool.Pool) AudienceRepository {
	return &audienceRepository{pool: pool}
}

func (r *audienceRepository) Create(ctx context.Context, audience *model.Audience) error {
	query := `
		INSERT INTO audiences (id, team_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, team_id, name, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		audience.ID, audience.TeamID, audience.Name, audience.CreatedAt, audience.UpdatedAt,
	).Scan(
		&audience.ID, &audience.TeamID, &audience.Name, &audience.CreatedAt, &audience.UpdatedAt,
	)
}

func (r *audienceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Audience, error) {
	query := `
		SELECT id, team_id, name, created_at, updated_at
		FROM audiences WHERE id = $1`

	a := &model.Audience{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.TeamID, &a.Name, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("audience")
		}
		return nil, fmt.Errorf("get audience by id: %w", err)
	}
	return a, nil
}

func (r *audienceRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Audience, error) {
	query := `
		SELECT id, team_id, name, created_at, updated_at
		FROM audiences WHERE team_id = $1 AND id = $2`

	a := &model.Audience{}
	err := r.pool.QueryRow(ctx, query, teamID, id).Scan(
		&a.ID, &a.TeamID, &a.Name, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("audience")
		}
		return nil, fmt.Errorf("get audience by team and id: %w", err)
	}
	return a, nil
}

func (r *audienceRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Audience, error) {
	query := `
		SELECT id, team_id, name, created_at, updated_at
		FROM audiences WHERE team_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list audiences: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Audience, error) {
		var a model.Audience
		err := row.Scan(&a.ID, &a.TeamID, &a.Name, &a.CreatedAt, &a.UpdatedAt)
		return a, err
	})
}

func (r *audienceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM audiences WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete audience: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("audience")
	}
	return nil
}
