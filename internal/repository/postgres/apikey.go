package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type apiKeyRepository struct {
	pool *pgxpool.Pool
}

// NewAPIKeyRepository creates a new APIKeyRepository backed by PostgreSQL.
func NewAPIKeyRepository(pool *pgxpool.Pool) APIKeyRepository {
	return &apiKeyRepository{pool: pool}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	query := `
		INSERT INTO api_keys (id, team_id, name, key_hash, key_prefix, permission, domain_id, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, team_id, name, key_hash, key_prefix, permission, domain_id, last_used_at, created_at`

	return r.pool.QueryRow(ctx, query,
		key.ID, key.TeamID, key.Name, key.KeyHash, key.KeyPrefix, key.Permission, key.DomainID, key.LastUsedAt, key.CreatedAt,
	).Scan(
		&key.ID, &key.TeamID, &key.Name, &key.KeyHash, &key.KeyPrefix, &key.Permission, &key.DomainID, &key.LastUsedAt, &key.CreatedAt,
	)
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	query := `
		SELECT id, team_id, name, key_hash, key_prefix, permission, domain_id, last_used_at, created_at
		FROM api_keys WHERE key_hash = $1`

	key := &model.APIKey{}
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&key.ID, &key.TeamID, &key.Name, &key.KeyHash, &key.KeyPrefix, &key.Permission, &key.DomainID, &key.LastUsedAt, &key.CreatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("api key")
		}
		return nil, fmt.Errorf("get api key by hash: %w", err)
	}
	return key, nil
}

func (r *apiKeyRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.APIKey, error) {
	query := `
		SELECT id, team_id, name, key_hash, key_prefix, permission, domain_id, last_used_at, created_at
		FROM api_keys WHERE team_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.APIKey, error) {
		var k model.APIKey
		err := row.Scan(
			&k.ID, &k.TeamID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.Permission, &k.DomainID, &k.LastUsedAt, &k.CreatedAt,
		)
		return k, err
	})
}

func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("api key")
	}
	return nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string, usedAt time.Time) error {
	query := `UPDATE api_keys SET last_used_at = $2 WHERE key_hash = $1`

	result, err := r.pool.Exec(ctx, query, keyHash, usedAt)
	if err != nil {
		return fmt.Errorf("update api key last used: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("api key")
	}
	return nil
}
