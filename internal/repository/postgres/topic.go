package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type topicRepository struct {
	pool *pgxpool.Pool
}

// NewTopicRepository creates a new TopicRepository backed by PostgreSQL.
func NewTopicRepository(pool *pgxpool.Pool) TopicRepository {
	return &topicRepository{pool: pool}
}

const topicColumns = `id, team_id, name, description, created_at, updated_at`

func (r *topicRepository) Create(ctx context.Context, topic *model.Topic) error {
	query := fmt.Sprintf(`
		INSERT INTO topics (%s)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, topicColumns, topicColumns)

	return r.pool.QueryRow(ctx, query,
		topic.ID, topic.TeamID, topic.Name, topic.Description, topic.CreatedAt, topic.UpdatedAt,
	).Scan(
		&topic.ID, &topic.TeamID, &topic.Name, &topic.Description, &topic.CreatedAt, &topic.UpdatedAt,
	)
}

func (r *topicRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Topic, error) {
	query := fmt.Sprintf(`SELECT %s FROM topics WHERE id = $1`, topicColumns)

	t := &model.Topic{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("topic")
		}
		return nil, fmt.Errorf("get topic by id: %w", err)
	}
	return t, nil
}

func (r *topicRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Topic, error) {
	query := fmt.Sprintf(`SELECT %s FROM topics WHERE team_id = $1 AND id = $2`, topicColumns)

	t := &model.Topic{}
	err := r.pool.QueryRow(ctx, query, teamID, id).Scan(
		&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("topic")
		}
		return nil, fmt.Errorf("get topic by team and id: %w", err)
	}
	return t, nil
}

func (r *topicRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Topic, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM topics WHERE team_id = $1
		ORDER BY created_at DESC`, topicColumns)

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Topic, error) {
		var t model.Topic
		err := row.Scan(&t.ID, &t.TeamID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
		return t, err
	})
}

func (r *topicRepository) Update(ctx context.Context, topic *model.Topic) error {
	query := fmt.Sprintf(`
		UPDATE topics
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
		RETURNING %s`, topicColumns)

	err := r.pool.QueryRow(ctx, query,
		topic.ID, topic.Name, topic.Description, topic.UpdatedAt,
	).Scan(
		&topic.ID, &topic.TeamID, &topic.Name, &topic.Description, &topic.CreatedAt, &topic.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("topic")
		}
		return fmt.Errorf("update topic: %w", err)
	}
	return nil
}

func (r *topicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM topics WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("topic")
	}
	return nil
}
