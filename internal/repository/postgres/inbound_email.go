package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type inboundEmailRepository struct {
	pool *pgxpool.Pool
}

// NewInboundEmailRepository creates a new InboundEmailRepository backed by PostgreSQL.
func NewInboundEmailRepository(pool *pgxpool.Pool) InboundEmailRepository {
	return &inboundEmailRepository{pool: pool}
}

const inboundEmailColumns = `id, team_id, domain_id, from_address, to_addresses, cc_addresses,
	subject, html_body, text_body, raw_message, headers, attachments,
	spam_score, processed, created_at`

func scanInboundEmailPtr(row pgx.Row) (*model.InboundEmail, error) {
	e := &model.InboundEmail{}
	err := row.Scan(
		&e.ID, &e.TeamID, &e.DomainID, &e.FromAddress, &e.ToAddresses, &e.CcAddresses,
		&e.Subject, &e.HTMLBody, &e.TextBody, &e.RawMessage, &e.Headers, &e.Attachments,
		&e.SpamScore, &e.Processed, &e.CreatedAt,
	)
	return e, err
}

func (r *inboundEmailRepository) Create(ctx context.Context, email *model.InboundEmail) error {
	query := fmt.Sprintf(`
		INSERT INTO inbound_emails (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING %s`, inboundEmailColumns, inboundEmailColumns)

	row := r.pool.QueryRow(ctx, query,
		email.ID, email.TeamID, email.DomainID, email.FromAddress, email.ToAddresses, email.CcAddresses,
		email.Subject, email.HTMLBody, email.TextBody, email.RawMessage, email.Headers, email.Attachments,
		email.SpamScore, email.Processed, email.CreatedAt,
	)
	scanned, err := scanInboundEmailPtr(row)
	if err != nil {
		return fmt.Errorf("create inbound email: %w", err)
	}
	*email = *scanned
	return nil
}

func (r *inboundEmailRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InboundEmail, error) {
	query := fmt.Sprintf(`SELECT %s FROM inbound_emails WHERE id = $1`, inboundEmailColumns)

	email, err := scanInboundEmailPtr(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("inbound email")
		}
		return nil, fmt.Errorf("get inbound email by id: %w", err)
	}
	return email, nil
}

func (r *inboundEmailRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.InboundEmail, error) {
	query := fmt.Sprintf(`SELECT %s FROM inbound_emails WHERE team_id = $1 AND id = $2`, inboundEmailColumns)

	email, err := scanInboundEmailPtr(r.pool.QueryRow(ctx, query, teamID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("inbound email")
		}
		return nil, fmt.Errorf("get inbound email by team and id: %w", err)
	}
	return email, nil
}

func (r *inboundEmailRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.InboundEmail, int, error) {
	countQuery := `SELECT COUNT(*) FROM inbound_emails WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count inbound emails: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM inbound_emails WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, inboundEmailColumns)

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list inbound emails: %w", err)
	}
	defer rows.Close()

	emails, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.InboundEmail, error) {
		var e model.InboundEmail
		err := row.Scan(
			&e.ID, &e.TeamID, &e.DomainID, &e.FromAddress, &e.ToAddresses, &e.CcAddresses,
			&e.Subject, &e.HTMLBody, &e.TextBody, &e.RawMessage, &e.Headers, &e.Attachments,
			&e.SpamScore, &e.Processed, &e.CreatedAt,
		)
		return e, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect inbound emails: %w", err)
	}

	return emails, total, nil
}

func (r *inboundEmailRepository) Update(ctx context.Context, email *model.InboundEmail) error {
	query := fmt.Sprintf(`
		UPDATE inbound_emails
		SET domain_id = $2, from_address = $3, to_addresses = $4, cc_addresses = $5,
		    subject = $6, html_body = $7, text_body = $8, raw_message = $9, headers = $10,
		    attachments = $11, spam_score = $12, processed = $13
		WHERE id = $1
		RETURNING %s`, inboundEmailColumns)

	row := r.pool.QueryRow(ctx, query,
		email.ID, email.DomainID, email.FromAddress, email.ToAddresses, email.CcAddresses,
		email.Subject, email.HTMLBody, email.TextBody, email.RawMessage, email.Headers,
		email.Attachments, email.SpamScore, email.Processed,
	)
	scanned, err := scanInboundEmailPtr(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("inbound email")
		}
		return fmt.Errorf("update inbound email: %w", err)
	}
	*email = *scanned
	return nil
}
