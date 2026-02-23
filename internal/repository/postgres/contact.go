package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type contactRepository struct {
	pool *pgxpool.Pool
}

// NewContactRepository creates a new ContactRepository backed by PostgreSQL.
func NewContactRepository(pool *pgxpool.Pool) ContactRepository {
	return &contactRepository{pool: pool}
}

const contactColumns = `id, audience_id, email, first_name, last_name, unsubscribed, created_at, updated_at`

func scanContactPtr(row pgx.Row) (*model.Contact, error) {
	c := &model.Contact{}
	err := row.Scan(
		&c.ID, &c.AudienceID, &c.Email, &c.FirstName, &c.LastName,
		&c.Unsubscribed, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *contactRepository) Create(ctx context.Context, contact *model.Contact) error {
	query := fmt.Sprintf(`
		INSERT INTO contacts (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s`, contactColumns, contactColumns)

	row := r.pool.QueryRow(ctx, query,
		contact.ID, contact.AudienceID, contact.Email, contact.FirstName, contact.LastName,
		contact.Unsubscribed, contact.CreatedAt, contact.UpdatedAt,
	)
	scanned, err := scanContactPtr(row)
	if err != nil {
		return fmt.Errorf("create contact: %w", err)
	}
	*contact = *scanned
	return nil
}

func (r *contactRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Contact, error) {
	query := fmt.Sprintf(`SELECT %s FROM contacts WHERE id = $1`, contactColumns)

	c, err := scanContactPtr(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("contact")
		}
		return nil, fmt.Errorf("get contact by id: %w", err)
	}
	return c, nil
}

func (r *contactRepository) GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Contact, error) {
	query := fmt.Sprintf(`SELECT %s FROM contacts WHERE audience_id = $1 AND id = $2`, contactColumns)

	c, err := scanContactPtr(r.pool.QueryRow(ctx, query, audienceID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("contact")
		}
		return nil, fmt.Errorf("get contact by audience and id: %w", err)
	}
	return c, nil
}

func (r *contactRepository) GetByAudienceAndEmail(ctx context.Context, audienceID uuid.UUID, email string) (*model.Contact, error) {
	query := fmt.Sprintf(`SELECT %s FROM contacts WHERE audience_id = $1 AND email = $2`, contactColumns)

	c, err := scanContactPtr(r.pool.QueryRow(ctx, query, audienceID, email))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("contact")
		}
		return nil, fmt.Errorf("get contact by audience and email: %w", err)
	}
	return c, nil
}

func (r *contactRepository) List(ctx context.Context, audienceID uuid.UUID, limit, offset int) ([]model.Contact, int, error) {
	countQuery := `SELECT COUNT(*) FROM contacts WHERE audience_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, audienceID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count contacts: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM contacts WHERE audience_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, contactColumns)

	rows, err := r.pool.Query(ctx, query, audienceID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	contacts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Contact, error) {
		var c model.Contact
		err := row.Scan(
			&c.ID, &c.AudienceID, &c.Email, &c.FirstName, &c.LastName,
			&c.Unsubscribed, &c.CreatedAt, &c.UpdatedAt,
		)
		return c, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect contacts: %w", err)
	}

	return contacts, total, nil
}

func (r *contactRepository) Update(ctx context.Context, contact *model.Contact) error {
	query := fmt.Sprintf(`
		UPDATE contacts
		SET email = $2, first_name = $3, last_name = $4, unsubscribed = $5, updated_at = $6
		WHERE id = $1
		RETURNING %s`, contactColumns)

	row := r.pool.QueryRow(ctx, query,
		contact.ID, contact.Email, contact.FirstName, contact.LastName,
		contact.Unsubscribed, contact.UpdatedAt,
	)
	scanned, err := scanContactPtr(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("contact")
		}
		return fmt.Errorf("update contact: %w", err)
	}
	*contact = *scanned
	return nil
}

func (r *contactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contacts WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("contact")
	}
	return nil
}
