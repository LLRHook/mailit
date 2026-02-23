package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository backed by PostgreSQL.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, password_hash, name, email_verified, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.EmailVerified, user.CreatedAt, user.UpdatedAt,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, email_verified, created_at, updated_at
		FROM users WHERE id = $1`

	user := &model.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("user")
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, email_verified, created_at, updated_at
		FROM users WHERE email = $1`

	user := &model.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("user")
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, name = $4, email_verified = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, email, password_hash, name, email_verified, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.EmailVerified, user.UpdatedAt,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("user")
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
