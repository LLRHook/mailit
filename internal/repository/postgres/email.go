package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type emailRepository struct {
	pool *pgxpool.Pool
}

// NewEmailRepository creates a new EmailRepository backed by PostgreSQL.
func NewEmailRepository(pool *pgxpool.Pool) EmailRepository {
	return &emailRepository{pool: pool}
}

const emailColumns = `id, team_id, domain_id, from_address, to_addresses,
	cc_addresses, bcc_addresses, reply_to, subject,
	html_body, text_body, status, scheduled_at, sent_at,
	delivered_at, tags, headers, attachments,
	idempotency_key, message_id, last_error, retry_count,
	created_at, updated_at`

func scanEmail(row pgx.Row) (*model.Email, error) {
	e := &model.Email{}
	err := row.Scan(
		&e.ID, &e.TeamID, &e.DomainID, &e.FromAddress, &e.ToAddresses,
		&e.CcAddresses, &e.BccAddresses, &e.ReplyTo, &e.Subject,
		&e.HTMLBody, &e.TextBody, &e.Status, &e.ScheduledAt, &e.SentAt,
		&e.DeliveredAt, &e.Tags, &e.Headers, &e.Attachments,
		&e.IdempotencyKey, &e.MessageID, &e.LastError, &e.RetryCount,
		&e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

func scanEmailValue(row pgx.CollectableRow) (model.Email, error) {
	var e model.Email
	err := row.Scan(
		&e.ID, &e.TeamID, &e.DomainID, &e.FromAddress, &e.ToAddresses,
		&e.CcAddresses, &e.BccAddresses, &e.ReplyTo, &e.Subject,
		&e.HTMLBody, &e.TextBody, &e.Status, &e.ScheduledAt, &e.SentAt,
		&e.DeliveredAt, &e.Tags, &e.Headers, &e.Attachments,
		&e.IdempotencyKey, &e.MessageID, &e.LastError, &e.RetryCount,
		&e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

func (r *emailRepository) Create(ctx context.Context, email *model.Email) error {
	query := fmt.Sprintf(`
		INSERT INTO emails (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		RETURNING %s`, emailColumns, emailColumns)

	row := r.pool.QueryRow(ctx, query,
		email.ID, email.TeamID, email.DomainID, email.FromAddress, email.ToAddresses,
		email.CcAddresses, email.BccAddresses, email.ReplyTo, email.Subject,
		email.HTMLBody, email.TextBody, email.Status, email.ScheduledAt, email.SentAt,
		email.DeliveredAt, email.Tags, email.Headers, email.Attachments,
		email.IdempotencyKey, email.MessageID, email.LastError, email.RetryCount,
		email.CreatedAt, email.UpdatedAt,
	)
	scanned, err := scanEmail(row)
	if err != nil {
		return fmt.Errorf("create email: %w", err)
	}
	*email = *scanned
	return nil
}

func (r *emailRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Email, error) {
	query := fmt.Sprintf(`SELECT %s FROM emails WHERE id = $1`, emailColumns)

	email, err := scanEmail(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("email")
		}
		return nil, fmt.Errorf("get email by id: %w", err)
	}
	return email, nil
}

func (r *emailRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Email, error) {
	query := fmt.Sprintf(`SELECT %s FROM emails WHERE team_id = $1 AND id = $2`, emailColumns)

	email, err := scanEmail(r.pool.QueryRow(ctx, query, teamID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("email")
		}
		return nil, fmt.Errorf("get email by team and id: %w", err)
	}
	return email, nil
}

func (r *emailRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Email, int, error) {
	countQuery := `SELECT COUNT(*) FROM emails WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count emails: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM emails WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, emailColumns)

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list emails: %w", err)
	}
	defer rows.Close()

	emails, err := pgx.CollectRows(rows, scanEmailValue)
	if err != nil {
		return nil, 0, fmt.Errorf("collect emails: %w", err)
	}

	return emails, total, nil
}

func (r *emailRepository) Update(ctx context.Context, email *model.Email) error {
	query := fmt.Sprintf(`
		UPDATE emails
		SET domain_id = $2, from_address = $3, to_addresses = $4, cc_addresses = $5, bcc_addresses = $6,
		    reply_to = $7, subject = $8, html_body = $9, text_body = $10, status = $11,
		    scheduled_at = $12, sent_at = $13, delivered_at = $14, tags = $15, headers = $16,
		    attachments = $17, idempotency_key = $18, message_id = $19, last_error = $20,
		    retry_count = $21, updated_at = $22
		WHERE id = $1
		RETURNING %s`, emailColumns)

	row := r.pool.QueryRow(ctx, query,
		email.ID, email.DomainID, email.FromAddress, email.ToAddresses, email.CcAddresses, email.BccAddresses,
		email.ReplyTo, email.Subject, email.HTMLBody, email.TextBody, email.Status,
		email.ScheduledAt, email.SentAt, email.DeliveredAt, email.Tags, email.Headers,
		email.Attachments, email.IdempotencyKey, email.MessageID, email.LastError,
		email.RetryCount, email.UpdatedAt,
	)
	scanned, err := scanEmail(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("email")
		}
		return fmt.Errorf("update email: %w", err)
	}
	*email = *scanned
	return nil
}

// --- EmailEventRepository ---

type emailEventRepository struct {
	pool *pgxpool.Pool
}

// NewEmailEventRepository creates a new EmailEventRepository backed by PostgreSQL.
func NewEmailEventRepository(pool *pgxpool.Pool) EmailEventRepository {
	return &emailEventRepository{pool: pool}
}

func (r *emailEventRepository) Create(ctx context.Context, event *model.EmailEvent) error {
	query := `
		INSERT INTO email_events (id, email_id, type, payload, recipient, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email_id, type, payload, recipient, created_at`

	return r.pool.QueryRow(ctx, query,
		event.ID, event.EmailID, event.Type, event.Payload, event.Recipient, event.CreatedAt,
	).Scan(
		&event.ID, &event.EmailID, &event.Type, &event.Payload, &event.Recipient, &event.CreatedAt,
	)
}

func (r *emailEventRepository) ListByEmailID(ctx context.Context, emailID uuid.UUID) ([]model.EmailEvent, error) {
	query := `
		SELECT id, email_id, type, payload, recipient, created_at
		FROM email_events WHERE email_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, emailID)
	if err != nil {
		return nil, fmt.Errorf("list email events: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.EmailEvent, error) {
		var ev model.EmailEvent
		err := row.Scan(&ev.ID, &ev.EmailID, &ev.Type, &ev.Payload, &ev.Recipient, &ev.CreatedAt)
		return ev, err
	})
}
