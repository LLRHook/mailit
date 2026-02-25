package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type teamInvitationRepository struct {
	pool *pgxpool.Pool
}

// NewTeamInvitationRepository creates a new TeamInvitationRepository backed by PostgreSQL.
func NewTeamInvitationRepository(pool *pgxpool.Pool) TeamInvitationRepository {
	return &teamInvitationRepository{pool: pool}
}

const invitationColumns = `id, team_id, email, role, token, expires_at, accepted_at, created_at`

func (r *teamInvitationRepository) Create(ctx context.Context, invitation *model.TeamInvitation) error {
	query := fmt.Sprintf(`
		INSERT INTO team_invitations (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s`, invitationColumns, invitationColumns)

	return r.pool.QueryRow(ctx, query,
		invitation.ID, invitation.TeamID, invitation.Email, invitation.Role,
		invitation.Token, invitation.ExpiresAt, invitation.AcceptedAt, invitation.CreatedAt,
	).Scan(
		&invitation.ID, &invitation.TeamID, &invitation.Email, &invitation.Role,
		&invitation.Token, &invitation.ExpiresAt, &invitation.AcceptedAt, &invitation.CreatedAt,
	)
}

func (r *teamInvitationRepository) GetByToken(ctx context.Context, token string) (*model.TeamInvitation, error) {
	query := fmt.Sprintf(`SELECT %s FROM team_invitations WHERE token = $1`, invitationColumns)

	var inv model.TeamInvitation
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.TeamID, &inv.Email, &inv.Role,
		&inv.Token, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("invitation")
		}
		return nil, fmt.Errorf("getting invitation by token: %w", err)
	}
	return &inv, nil
}

func (r *teamInvitationRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.TeamInvitation, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM team_invitations
		WHERE team_id = $1 AND accepted_at IS NULL
		ORDER BY created_at DESC`, invitationColumns)

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing invitations: %w", err)
	}
	defer rows.Close()

	invitations, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.TeamInvitation, error) {
		var inv model.TeamInvitation
		err := row.Scan(
			&inv.ID, &inv.TeamID, &inv.Email, &inv.Role,
			&inv.Token, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt,
		)
		return inv, err
	})
	if err != nil {
		return nil, fmt.Errorf("collecting invitations: %w", err)
	}

	return invitations, nil
}

func (r *teamInvitationRepository) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE team_invitations SET accepted_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("marking invitation as accepted: %w", err)
	}
	return nil
}

func (r *teamInvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM team_invitations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting invitation: %w", err)
	}
	return nil
}
