package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/dto"
)

type settingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository creates a new SettingsRepository backed by PostgreSQL.
func NewSettingsRepository(pool *pgxpool.Pool) SettingsRepository {
	return &settingsRepository{pool: pool}
}

func (r *settingsRepository) GetUsageCounts(ctx context.Context, teamID uuid.UUID) (*dto.UsageResponse, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := `
		SELECT
			(SELECT COUNT(*) FROM emails WHERE team_id = $1 AND created_at >= $2) AS emails_today,
			(SELECT COUNT(*) FROM emails WHERE team_id = $1 AND created_at >= $3) AS emails_month,
			(SELECT COUNT(*) FROM domains WHERE team_id = $1) AS domains,
			(SELECT COUNT(*) FROM api_keys WHERE team_id = $1) AS api_keys,
			(SELECT COUNT(*) FROM webhooks WHERE team_id = $1) AS webhooks,
			(SELECT COUNT(*) FROM contacts c JOIN audiences a ON c.audience_id = a.id WHERE a.team_id = $1) AS contacts`

	var usage dto.UsageResponse
	err := r.pool.QueryRow(ctx, query, teamID, startOfDay, startOfMonth).Scan(
		&usage.EmailsSentToday,
		&usage.EmailsSentMonth,
		&usage.Domains,
		&usage.APIKeys,
		&usage.Webhooks,
		&usage.Contacts,
	)
	if err != nil {
		return nil, fmt.Errorf("getting usage counts: %w", err)
	}

	return &usage, nil
}

func (r *settingsRepository) GetTeamWithMembers(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error) {
	// Get team info.
	var team dto.TeamResponse
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug FROM teams WHERE id = $1`, teamID,
	).Scan(&team.ID, &team.Name, &team.Slug)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("team")
		}
		return nil, fmt.Errorf("getting team: %w", err)
	}

	// Get members with user info.
	rows, err := r.pool.Query(ctx, `
		SELECT tm.id, u.name, u.email, tm.role
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY tm.created_at ASC`, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	defer rows.Close()

	members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (dto.TeamMemberResponse, error) {
		var m dto.TeamMemberResponse
		err := row.Scan(&m.ID, &m.Name, &m.Email, &m.Role)
		return m, err
	})
	if err != nil {
		return nil, fmt.Errorf("collecting team members: %w", err)
	}

	team.Members = members
	return &team, nil
}

func (r *settingsRepository) UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE teams SET name = $1, updated_at = $2 WHERE id = $3`,
		name, time.Now().UTC(), teamID)
	if err != nil {
		return fmt.Errorf("updating team name: %w", err)
	}
	return nil
}
