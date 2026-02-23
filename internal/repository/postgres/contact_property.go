package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type contactPropertyRepository struct {
	pool *pgxpool.Pool
}

// NewContactPropertyRepository creates a new ContactPropertyRepository backed by PostgreSQL.
func NewContactPropertyRepository(pool *pgxpool.Pool) ContactPropertyRepository {
	return &contactPropertyRepository{pool: pool}
}

const contactPropertyColumns = `id, team_id, name, label, type, created_at, updated_at`

func (r *contactPropertyRepository) Create(ctx context.Context, property *model.ContactProperty) error {
	query := fmt.Sprintf(`
		INSERT INTO contact_properties (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING %s`, contactPropertyColumns, contactPropertyColumns)

	return r.pool.QueryRow(ctx, query,
		property.ID, property.TeamID, property.Name, property.Label, property.Type, property.CreatedAt, property.UpdatedAt,
	).Scan(
		&property.ID, &property.TeamID, &property.Name, &property.Label, &property.Type, &property.CreatedAt, &property.UpdatedAt,
	)
}

func (r *contactPropertyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ContactProperty, error) {
	query := fmt.Sprintf(`SELECT %s FROM contact_properties WHERE id = $1`, contactPropertyColumns)

	p := &model.ContactProperty{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TeamID, &p.Name, &p.Label, &p.Type, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("contact property")
		}
		return nil, fmt.Errorf("get contact property by id: %w", err)
	}
	return p, nil
}

func (r *contactPropertyRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.ContactProperty, error) {
	query := fmt.Sprintf(`SELECT %s FROM contact_properties WHERE team_id = $1 AND id = $2`, contactPropertyColumns)

	p := &model.ContactProperty{}
	err := r.pool.QueryRow(ctx, query, teamID, id).Scan(
		&p.ID, &p.TeamID, &p.Name, &p.Label, &p.Type, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("contact property")
		}
		return nil, fmt.Errorf("get contact property by team and id: %w", err)
	}
	return p, nil
}

func (r *contactPropertyRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.ContactProperty, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM contact_properties WHERE team_id = $1
		ORDER BY created_at DESC`, contactPropertyColumns)

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list contact properties: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContactProperty, error) {
		var p model.ContactProperty
		err := row.Scan(&p.ID, &p.TeamID, &p.Name, &p.Label, &p.Type, &p.CreatedAt, &p.UpdatedAt)
		return p, err
	})
}

func (r *contactPropertyRepository) Update(ctx context.Context, property *model.ContactProperty) error {
	query := fmt.Sprintf(`
		UPDATE contact_properties
		SET name = $2, label = $3, type = $4, updated_at = $5
		WHERE id = $1
		RETURNING %s`, contactPropertyColumns)

	err := r.pool.QueryRow(ctx, query,
		property.ID, property.Name, property.Label, property.Type, property.UpdatedAt,
	).Scan(
		&property.ID, &property.TeamID, &property.Name, &property.Label, &property.Type, &property.CreatedAt, &property.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("contact property")
		}
		return fmt.Errorf("update contact property: %w", err)
	}
	return nil
}

func (r *contactPropertyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contact_properties WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete contact property: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("contact property")
	}
	return nil
}
