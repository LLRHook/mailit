package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type teamRepository struct {
	pool *pgxpool.Pool
}

// NewTeamRepository creates a new TeamRepository backed by PostgreSQL.
func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &teamRepository{pool: pool}
}

func (r *teamRepository) Create(ctx context.Context, team *model.Team) error {
	query := `
		INSERT INTO teams (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, slug, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		team.ID, team.Name, team.Slug, team.CreatedAt, team.UpdatedAt,
	).Scan(
		&team.ID, &team.Name, &team.Slug, &team.CreatedAt, &team.UpdatedAt,
	)
}

func (r *teamRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Team, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM teams WHERE id = $1`

	team := &model.Team{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Slug, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("team")
		}
		return nil, fmt.Errorf("get team by id: %w", err)
	}
	return team, nil
}

func (r *teamRepository) GetBySlug(ctx context.Context, slug string) (*model.Team, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM teams WHERE slug = $1`

	team := &model.Team{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&team.ID, &team.Name, &team.Slug, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("team")
		}
		return nil, fmt.Errorf("get team by slug: %w", err)
	}
	return team, nil
}

// --- TeamMemberRepository ---

type teamMemberRepository struct {
	pool *pgxpool.Pool
}

// NewTeamMemberRepository creates a new TeamMemberRepository backed by PostgreSQL.
func NewTeamMemberRepository(pool *pgxpool.Pool) TeamMemberRepository {
	return &teamMemberRepository{pool: pool}
}

func (r *teamMemberRepository) Create(ctx context.Context, member *model.TeamMember) error {
	query := `
		INSERT INTO team_members (id, team_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, team_id, user_id, role, created_at`

	return r.pool.QueryRow(ctx, query,
		member.ID, member.TeamID, member.UserID, member.Role, member.CreatedAt,
	).Scan(
		&member.ID, &member.TeamID, &member.UserID, &member.Role, &member.CreatedAt,
	)
}

func (r *teamMemberRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, created_at
		FROM team_members WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list team members by user: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.TeamMember, error) {
		var m model.TeamMember
		err := row.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Role, &m.CreatedAt)
		return m, err
	})
}

func (r *teamMemberRepository) GetByTeamAndUser(ctx context.Context, teamID, userID uuid.UUID) (*model.TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, created_at
		FROM team_members WHERE team_id = $1 AND user_id = $2`

	member := &model.TeamMember{}
	err := r.pool.QueryRow(ctx, query, teamID, userID).Scan(
		&member.ID, &member.TeamID, &member.UserID, &member.Role, &member.CreatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("team member")
		}
		return nil, fmt.Errorf("get team member: %w", err)
	}
	return member, nil
}
