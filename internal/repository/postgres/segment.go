package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type segmentRepository struct {
	pool *pgxpool.Pool
}

// NewSegmentRepository creates a new SegmentRepository backed by PostgreSQL.
func NewSegmentRepository(pool *pgxpool.Pool) SegmentRepository {
	return &segmentRepository{pool: pool}
}

const segmentColumns = `id, audience_id, name, conditions, created_at, updated_at`

func (r *segmentRepository) Create(ctx context.Context, segment *model.Segment) error {
	query := fmt.Sprintf(`
		INSERT INTO segments (%s)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, segmentColumns, segmentColumns)

	return r.pool.QueryRow(ctx, query,
		segment.ID, segment.AudienceID, segment.Name, segment.Conditions, segment.CreatedAt, segment.UpdatedAt,
	).Scan(
		&segment.ID, &segment.AudienceID, &segment.Name, &segment.Conditions, &segment.CreatedAt, &segment.UpdatedAt,
	)
}

func (r *segmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Segment, error) {
	query := fmt.Sprintf(`SELECT %s FROM segments WHERE id = $1`, segmentColumns)

	s := &model.Segment{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.AudienceID, &s.Name, &s.Conditions, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("segment")
		}
		return nil, fmt.Errorf("get segment by id: %w", err)
	}
	return s, nil
}

func (r *segmentRepository) GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Segment, error) {
	query := fmt.Sprintf(`SELECT %s FROM segments WHERE audience_id = $1 AND id = $2`, segmentColumns)

	s := &model.Segment{}
	err := r.pool.QueryRow(ctx, query, audienceID, id).Scan(
		&s.ID, &s.AudienceID, &s.Name, &s.Conditions, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("segment")
		}
		return nil, fmt.Errorf("get segment by audience and id: %w", err)
	}
	return s, nil
}

func (r *segmentRepository) ListByAudienceID(ctx context.Context, audienceID uuid.UUID) ([]model.Segment, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM segments WHERE audience_id = $1
		ORDER BY created_at DESC`, segmentColumns)

	rows, err := r.pool.Query(ctx, query, audienceID)
	if err != nil {
		return nil, fmt.Errorf("list segments: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Segment, error) {
		var s model.Segment
		err := row.Scan(&s.ID, &s.AudienceID, &s.Name, &s.Conditions, &s.CreatedAt, &s.UpdatedAt)
		return s, err
	})
}

func (r *segmentRepository) Update(ctx context.Context, segment *model.Segment) error {
	query := fmt.Sprintf(`
		UPDATE segments
		SET name = $2, conditions = $3, updated_at = $4
		WHERE id = $1
		RETURNING %s`, segmentColumns)

	err := r.pool.QueryRow(ctx, query,
		segment.ID, segment.Name, segment.Conditions, segment.UpdatedAt,
	).Scan(
		&segment.ID, &segment.AudienceID, &segment.Name, &segment.Conditions, &segment.CreatedAt, &segment.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("segment")
		}
		return fmt.Errorf("update segment: %w", err)
	}
	return nil
}

func (r *segmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM segments WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete segment: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("segment")
	}
	return nil
}
