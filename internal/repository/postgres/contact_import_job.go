package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mailit-dev/mailit/internal/model"
)

type contactImportJobRepository struct {
	pool *pgxpool.Pool
}

func NewContactImportJobRepository(pool *pgxpool.Pool) ContactImportJobRepository {
	return &contactImportJobRepository{pool: pool}
}

func (r *contactImportJobRepository) Create(ctx context.Context, job *model.ContactImportJob) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO contact_import_jobs (id, team_id, audience_id, status, total_rows, csv_data, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		job.ID, job.TeamID, job.AudienceID, job.Status, job.TotalRows, job.CSVData, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting import job: %w", err)
	}
	return nil
}

func (r *contactImportJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ContactImportJob, error) {
	var job model.ContactImportJob
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, audience_id, status, total_rows, processed_rows,
		        created_rows, updated_rows, skipped_rows, failed_rows,
		        error, csv_data, created_at, updated_at
		 FROM contact_import_jobs WHERE id = $1`, id,
	).Scan(
		&job.ID, &job.TeamID, &job.AudienceID, &job.Status, &job.TotalRows,
		&job.ProcessedRows, &job.CreatedRows, &job.UpdatedRows, &job.SkippedRows,
		&job.FailedRows, &job.Error, &job.CSVData, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("import job")
		}
		return nil, fmt.Errorf("querying import job: %w", err)
	}
	return &job, nil
}

func (r *contactImportJobRepository) Update(ctx context.Context, job *model.ContactImportJob) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE contact_import_jobs
		 SET status = $2, processed_rows = $3, created_rows = $4,
		     updated_rows = $5, skipped_rows = $6, failed_rows = $7,
		     error = $8, updated_at = $9
		 WHERE id = $1`,
		job.ID, job.Status, job.ProcessedRows, job.CreatedRows,
		job.UpdatedRows, job.SkippedRows, job.FailedRows,
		job.Error, job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating import job: %w", err)
	}
	return nil
}
