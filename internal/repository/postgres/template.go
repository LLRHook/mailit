package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type templateRepository struct {
	pool *pgxpool.Pool
}

// NewTemplateRepository creates a new TemplateRepository backed by PostgreSQL.
func NewTemplateRepository(pool *pgxpool.Pool) TemplateRepository {
	return &templateRepository{pool: pool}
}

const templateColumns = `id, team_id, name, description, created_at, updated_at`

func (r *templateRepository) Create(ctx context.Context, template *model.Template) error {
	query := fmt.Sprintf(`
		INSERT INTO templates (%s)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, templateColumns, templateColumns)

	return r.pool.QueryRow(ctx, query,
		template.ID, template.TeamID, template.Name, template.Description, template.CreatedAt, template.UpdatedAt,
	).Scan(
		&template.ID, &template.TeamID, &template.Name, &template.Description, &template.CreatedAt, &template.UpdatedAt,
	)
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	query := fmt.Sprintf(`SELECT %s FROM templates WHERE id = $1`, templateColumns)

	t := &model.Template{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("template")
		}
		return nil, fmt.Errorf("get template by id: %w", err)
	}
	return t, nil
}

func (r *templateRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Template, error) {
	query := fmt.Sprintf(`SELECT %s FROM templates WHERE team_id = $1 AND id = $2`, templateColumns)

	t := &model.Template{}
	err := r.pool.QueryRow(ctx, query, teamID, id).Scan(
		&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("template")
		}
		return nil, fmt.Errorf("get template by team and id: %w", err)
	}
	return t, nil
}

func (r *templateRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Template, int, error) {
	countQuery := `SELECT COUNT(*) FROM templates WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count templates: %w", err)
	}

	query := `
		SELECT t.id, t.team_id, t.name, t.description, t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM template_versions tv WHERE tv.template_id = t.id) AS version_count
		FROM templates t WHERE t.team_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	templates, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Template, error) {
		var t model.Template
		err := row.Scan(&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt, &t.VersionCount)
		return t, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect templates: %w", err)
	}

	return templates, total, nil
}

func (r *templateRepository) Update(ctx context.Context, template *model.Template) error {
	query := fmt.Sprintf(`
		UPDATE templates
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
		RETURNING %s`, templateColumns)

	err := r.pool.QueryRow(ctx, query,
		template.ID, template.Name, template.Description, template.UpdatedAt,
	).Scan(
		&template.ID, &template.TeamID, &template.Name, &template.Description, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("template")
		}
		return fmt.Errorf("update template: %w", err)
	}
	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM templates WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("template")
	}
	return nil
}

// --- TemplateVersionRepository ---

type templateVersionRepository struct {
	pool *pgxpool.Pool
}

// NewTemplateVersionRepository creates a new TemplateVersionRepository backed by PostgreSQL.
func NewTemplateVersionRepository(pool *pgxpool.Pool) TemplateVersionRepository {
	return &templateVersionRepository{pool: pool}
}

const templateVersionColumns = `id, template_id, version, subject, html_body, text_body, variables, published, created_at`

func scanTemplateVersionPtr(row pgx.Row) (*model.TemplateVersion, error) {
	v := &model.TemplateVersion{}
	err := row.Scan(
		&v.ID, &v.TemplateID, &v.Version, &v.Subject, &v.HTMLBody, &v.TextBody,
		&v.Variables, &v.Published, &v.CreatedAt,
	)
	return v, err
}

func (r *templateVersionRepository) Create(ctx context.Context, version *model.TemplateVersion) error {
	query := fmt.Sprintf(`
		INSERT INTO template_versions (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING %s`, templateVersionColumns, templateVersionColumns)

	row := r.pool.QueryRow(ctx, query,
		version.ID, version.TemplateID, version.Version, version.Subject,
		version.HTMLBody, version.TextBody, version.Variables, version.Published, version.CreatedAt,
	)
	scanned, err := scanTemplateVersionPtr(row)
	if err != nil {
		return fmt.Errorf("create template version: %w", err)
	}
	*version = *scanned
	return nil
}

func (r *templateVersionRepository) GetLatestByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM template_versions
		WHERE template_id = $1
		ORDER BY version DESC
		LIMIT 1`, templateVersionColumns)

	v, err := scanTemplateVersionPtr(r.pool.QueryRow(ctx, query, templateID))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("template version")
		}
		return nil, fmt.Errorf("get latest template version: %w", err)
	}
	return v, nil
}

func (r *templateVersionRepository) GetPublishedByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM template_versions
		WHERE template_id = $1 AND published = true
		ORDER BY version DESC
		LIMIT 1`, templateVersionColumns)

	v, err := scanTemplateVersionPtr(r.pool.QueryRow(ctx, query, templateID))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("template version")
		}
		return nil, fmt.Errorf("get published template version: %w", err)
	}
	return v, nil
}

func (r *templateVersionRepository) ListByTemplateID(ctx context.Context, templateID uuid.UUID) ([]model.TemplateVersion, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM template_versions
		WHERE template_id = $1
		ORDER BY version DESC`, templateVersionColumns)

	rows, err := r.pool.Query(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("list template versions: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.TemplateVersion, error) {
		var v model.TemplateVersion
		err := row.Scan(
			&v.ID, &v.TemplateID, &v.Version, &v.Subject, &v.HTMLBody, &v.TextBody,
			&v.Variables, &v.Published, &v.CreatedAt,
		)
		return v, err
	})
}

func (r *templateVersionRepository) Publish(ctx context.Context, versionID uuid.UUID) error {
	// Find the template_id for this version, then unpublish all and publish the target.
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var templateID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT template_id FROM template_versions WHERE id = $1`, versionID,
	).Scan(&templateID)
	if err != nil {
		if isNoRows(err) {
			return notFound("template version")
		}
		return fmt.Errorf("find template version: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE template_versions SET published = false WHERE template_id = $1`,
		templateID,
	)
	if err != nil {
		return fmt.Errorf("unpublish template versions: %w", err)
	}

	result, err := tx.Exec(ctx,
		`UPDATE template_versions SET published = true WHERE id = $1`,
		versionID,
	)
	if err != nil {
		return fmt.Errorf("publish template version: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("template version")
	}

	return tx.Commit(ctx)
}
