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

type webhookRepository struct {
	pool *pgxpool.Pool
}

// NewWebhookRepository creates a new WebhookRepository backed by PostgreSQL.
func NewWebhookRepository(pool *pgxpool.Pool) WebhookRepository {
	return &webhookRepository{pool: pool}
}

const webhookColumns = `id, team_id, url, events, signing_secret, active, created_at, updated_at`

func scanWebhookPtr(row pgx.Row) (*model.Webhook, error) {
	w := &model.Webhook{}
	err := row.Scan(
		&w.ID, &w.TeamID, &w.URL, &w.Events, &w.SigningSecret, &w.Active, &w.CreatedAt, &w.UpdatedAt,
	)
	return w, err
}

func (r *webhookRepository) Create(ctx context.Context, webhook *model.Webhook) error {
	query := fmt.Sprintf(`
		INSERT INTO webhooks (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s`, webhookColumns, webhookColumns)

	row := r.pool.QueryRow(ctx, query,
		webhook.ID, webhook.TeamID, webhook.URL, webhook.Events,
		webhook.SigningSecret, webhook.Active, webhook.CreatedAt, webhook.UpdatedAt,
	)
	scanned, err := scanWebhookPtr(row)
	if err != nil {
		return fmt.Errorf("create webhook: %w", err)
	}
	*webhook = *scanned
	return nil
}

func (r *webhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Webhook, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhooks WHERE id = $1`, webhookColumns)

	w, err := scanWebhookPtr(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("webhook")
		}
		return nil, fmt.Errorf("get webhook by id: %w", err)
	}
	return w, nil
}

func (r *webhookRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Webhook, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhooks WHERE team_id = $1 AND id = $2`, webhookColumns)

	w, err := scanWebhookPtr(r.pool.QueryRow(ctx, query, teamID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("webhook")
		}
		return nil, fmt.Errorf("get webhook by team and id: %w", err)
	}
	return w, nil
}

func (r *webhookRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Webhook, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM webhooks WHERE team_id = $1
		ORDER BY created_at DESC`, webhookColumns)

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Webhook, error) {
		var w model.Webhook
		err := row.Scan(&w.ID, &w.TeamID, &w.URL, &w.Events, &w.SigningSecret, &w.Active, &w.CreatedAt, &w.UpdatedAt)
		return w, err
	})
}

func (r *webhookRepository) Update(ctx context.Context, webhook *model.Webhook) error {
	query := fmt.Sprintf(`
		UPDATE webhooks
		SET url = $2, events = $3, signing_secret = $4, active = $5, updated_at = $6
		WHERE id = $1
		RETURNING %s`, webhookColumns)

	row := r.pool.QueryRow(ctx, query,
		webhook.ID, webhook.URL, webhook.Events, webhook.SigningSecret, webhook.Active, webhook.UpdatedAt,
	)
	scanned, err := scanWebhookPtr(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("webhook")
		}
		return fmt.Errorf("update webhook: %w", err)
	}
	*webhook = *scanned
	return nil
}

func (r *webhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("webhook")
	}
	return nil
}

// --- WebhookEventRepository ---

type webhookEventRepository struct {
	pool *pgxpool.Pool
}

// NewWebhookEventRepository creates a new WebhookEventRepository backed by PostgreSQL.
func NewWebhookEventRepository(pool *pgxpool.Pool) WebhookEventRepository {
	return &webhookEventRepository{pool: pool}
}

const webhookEventColumns = `id, webhook_id, event_type, payload, status, response_code, response_body, attempts, next_retry_at, created_at`

func scanWebhookEventPtr(row pgx.Row) (*model.WebhookEvent, error) {
	ev := &model.WebhookEvent{}
	err := row.Scan(
		&ev.ID, &ev.WebhookID, &ev.EventType, &ev.Payload, &ev.Status,
		&ev.ResponseCode, &ev.ResponseBody, &ev.Attempts, &ev.NextRetryAt, &ev.CreatedAt,
	)
	return ev, err
}

func (r *webhookEventRepository) Create(ctx context.Context, event *model.WebhookEvent) error {
	query := fmt.Sprintf(`
		INSERT INTO webhook_events (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING %s`, webhookEventColumns, webhookEventColumns)

	row := r.pool.QueryRow(ctx, query,
		event.ID, event.WebhookID, event.EventType, event.Payload, event.Status,
		event.ResponseCode, event.ResponseBody, event.Attempts, event.NextRetryAt, event.CreatedAt,
	)
	scanned, err := scanWebhookEventPtr(row)
	if err != nil {
		return fmt.Errorf("create webhook event: %w", err)
	}
	*event = *scanned
	return nil
}

func (r *webhookEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookEvent, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhook_events WHERE id = $1`, webhookEventColumns)

	ev, err := scanWebhookEventPtr(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("webhook event")
		}
		return nil, fmt.Errorf("get webhook event by id: %w", err)
	}
	return ev, nil
}

func (r *webhookEventRepository) Update(ctx context.Context, event *model.WebhookEvent) error {
	query := fmt.Sprintf(`
		UPDATE webhook_events
		SET status = $2, response_code = $3, response_body = $4, attempts = $5, next_retry_at = $6
		WHERE id = $1
		RETURNING %s`, webhookEventColumns)

	row := r.pool.QueryRow(ctx, query,
		event.ID, event.Status, event.ResponseCode, event.ResponseBody, event.Attempts, event.NextRetryAt,
	)
	scanned, err := scanWebhookEventPtr(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("webhook event")
		}
		return fmt.Errorf("update webhook event: %w", err)
	}
	*event = *scanned
	return nil
}

func (r *webhookEventRepository) ListByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]model.WebhookEvent, int, error) {
	countQuery := `SELECT COUNT(*) FROM webhook_events WHERE webhook_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, webhookID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count webhook events: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM webhook_events WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, webhookEventColumns)

	rows, err := r.pool.Query(ctx, query, webhookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list webhook events: %w", err)
	}
	defer rows.Close()

	events, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.WebhookEvent, error) {
		var ev model.WebhookEvent
		err := row.Scan(
			&ev.ID, &ev.WebhookID, &ev.EventType, &ev.Payload, &ev.Status,
			&ev.ResponseCode, &ev.ResponseBody, &ev.Attempts, &ev.NextRetryAt, &ev.CreatedAt,
		)
		return ev, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect webhook events: %w", err)
	}

	return events, total, nil
}

func (r *webhookEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM webhook_events WHERE created_at < $1`

	result, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("delete old webhook events: %w", err)
	}
	return result.RowsAffected(), nil
}
