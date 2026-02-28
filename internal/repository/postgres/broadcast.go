package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type broadcastRepository struct {
	pool *pgxpool.Pool
}

// NewBroadcastRepository creates a new BroadcastRepository backed by PostgreSQL.
func NewBroadcastRepository(pool *pgxpool.Pool) BroadcastRepository {
	return &broadcastRepository{pool: pool}
}

const broadcastColumns = `id, team_id, name, audience_id, segment_id, template_id, from_address,
	subject, html_body, text_body, status, scheduled_at, sent_at,
	total_recipients, sent_count, created_at, updated_at`

func scanBroadcastPtr(row pgx.Row) (*model.Broadcast, error) {
	b := &model.Broadcast{}
	err := row.Scan(
		&b.ID, &b.TeamID, &b.Name, &b.AudienceID, &b.SegmentID, &b.TemplateID,
		&b.FromAddress, &b.Subject, &b.HTMLBody, &b.TextBody, &b.Status,
		&b.ScheduledAt, &b.SentAt, &b.TotalRecipients, &b.SentCount,
		&b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}

func (r *broadcastRepository) Create(ctx context.Context, broadcast *model.Broadcast) error {
	query := fmt.Sprintf(`
		INSERT INTO broadcasts (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING %s`, broadcastColumns, broadcastColumns)

	row := r.pool.QueryRow(ctx, query,
		broadcast.ID, broadcast.TeamID, broadcast.Name, broadcast.AudienceID, broadcast.SegmentID,
		broadcast.TemplateID, broadcast.FromAddress, broadcast.Subject, broadcast.HTMLBody,
		broadcast.TextBody, broadcast.Status, broadcast.ScheduledAt, broadcast.SentAt,
		broadcast.TotalRecipients, broadcast.SentCount, broadcast.CreatedAt, broadcast.UpdatedAt,
	)
	scanned, err := scanBroadcastPtr(row)
	if err != nil {
		return fmt.Errorf("create broadcast: %w", err)
	}
	*broadcast = *scanned
	return nil
}

func (r *broadcastRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Broadcast, error) {
	query := fmt.Sprintf(`SELECT %s FROM broadcasts WHERE id = $1`, broadcastColumns)

	b, err := scanBroadcastPtr(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("broadcast")
		}
		return nil, fmt.Errorf("get broadcast by id: %w", err)
	}
	return b, nil
}

func (r *broadcastRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Broadcast, error) {
	query := fmt.Sprintf(`SELECT %s FROM broadcasts WHERE team_id = $1 AND id = $2`, broadcastColumns)

	b, err := scanBroadcastPtr(r.pool.QueryRow(ctx, query, teamID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("broadcast")
		}
		return nil, fmt.Errorf("get broadcast by team and id: %w", err)
	}
	return b, nil
}

func (r *broadcastRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Broadcast, int, error) {
	countQuery := `SELECT COUNT(*) FROM broadcasts WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count broadcasts: %w", err)
	}

	query := `
		SELECT b.id, b.team_id, b.name, b.audience_id, b.segment_id, b.template_id,
			b.from_address, b.subject, b.html_body, b.text_body, b.status,
			b.scheduled_at, b.sent_at, b.total_recipients, b.sent_count,
			b.created_at, b.updated_at, a.name AS audience_name
		FROM broadcasts b
		LEFT JOIN audiences a ON a.id = b.audience_id
		WHERE b.team_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list broadcasts: %w", err)
	}
	defer rows.Close()

	broadcasts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Broadcast, error) {
		var b model.Broadcast
		err := row.Scan(
			&b.ID, &b.TeamID, &b.Name, &b.AudienceID, &b.SegmentID, &b.TemplateID,
			&b.FromAddress, &b.Subject, &b.HTMLBody, &b.TextBody, &b.Status,
			&b.ScheduledAt, &b.SentAt, &b.TotalRecipients, &b.SentCount,
			&b.CreatedAt, &b.UpdatedAt, &b.AudienceName,
		)
		return b, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect broadcasts: %w", err)
	}

	return broadcasts, total, nil
}

func (r *broadcastRepository) Update(ctx context.Context, broadcast *model.Broadcast) error {
	query := fmt.Sprintf(`
		UPDATE broadcasts
		SET name = $2, audience_id = $3, segment_id = $4, template_id = $5, from_address = $6,
		    subject = $7, html_body = $8, text_body = $9, status = $10, scheduled_at = $11,
		    sent_at = $12, total_recipients = $13, sent_count = $14, updated_at = $15
		WHERE id = $1
		RETURNING %s`, broadcastColumns)

	row := r.pool.QueryRow(ctx, query,
		broadcast.ID, broadcast.Name, broadcast.AudienceID, broadcast.SegmentID, broadcast.TemplateID,
		broadcast.FromAddress, broadcast.Subject, broadcast.HTMLBody, broadcast.TextBody,
		broadcast.Status, broadcast.ScheduledAt, broadcast.SentAt,
		broadcast.TotalRecipients, broadcast.SentCount, broadcast.UpdatedAt,
	)
	scanned, err := scanBroadcastPtr(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("broadcast")
		}
		return fmt.Errorf("update broadcast: %w", err)
	}
	*broadcast = *scanned
	return nil
}

func (r *broadcastRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM broadcasts WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete broadcast: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("broadcast")
	}
	return nil
}
