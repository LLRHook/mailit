package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mailit-dev/mailit/internal/model"
)

type trackingLinkRepository struct {
	pool *pgxpool.Pool
}

func NewTrackingLinkRepository(pool *pgxpool.Pool) TrackingLinkRepository {
	return &trackingLinkRepository{pool: pool}
}

func (r *trackingLinkRepository) Create(ctx context.Context, link *model.TrackingLink) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_tracking_links (id, email_id, team_id, type, original_url, recipient, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		link.ID, link.EmailID, link.TeamID, link.Type, link.OriginalURL, link.Recipient, link.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting tracking link: %w", err)
	}
	return nil
}

func (r *trackingLinkRepository) CreateBatch(ctx context.Context, links []*model.TrackingLink) error {
	if len(links) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, link := range links {
		batch.Queue(
			`INSERT INTO email_tracking_links (id, email_id, team_id, type, original_url, recipient, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			link.ID, link.EmailID, link.TeamID, link.Type, link.OriginalURL, link.Recipient, link.CreatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	for range links {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch inserting tracking links: %w", err)
		}
	}
	return nil
}

func (r *trackingLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.TrackingLink, error) {
	var link model.TrackingLink
	err := r.pool.QueryRow(ctx,
		`SELECT id, email_id, team_id, type, original_url, recipient, created_at
		 FROM email_tracking_links WHERE id = $1`, id,
	).Scan(&link.ID, &link.EmailID, &link.TeamID, &link.Type, &link.OriginalURL, &link.Recipient, &link.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("tracking link")
		}
		return nil, fmt.Errorf("querying tracking link: %w", err)
	}
	return &link, nil
}
