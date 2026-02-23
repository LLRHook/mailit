package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// WebhookService defines operations for managing webhook endpoints.
type WebhookService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateWebhookRequest) (*dto.WebhookResponse, error)
	List(ctx context.Context, teamID uuid.UUID) ([]dto.WebhookResponse, error)
	Get(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) (*dto.WebhookResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID, req *dto.UpdateWebhookRequest) (*dto.WebhookResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) error
}

type webhookService struct {
	webhookRepo postgres.WebhookRepository
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(webhookRepo postgres.WebhookRepository) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
	}
}

func (s *webhookService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateWebhookRequest) (*dto.WebhookResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Generate a signing secret for this webhook.
	signingSecret, err := pkg.GenerateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("generating signing secret: %w", err)
	}

	now := time.Now().UTC()

	webhook := &model.Webhook{
		ID:            uuid.New(),
		TeamID:        teamID,
		URL:           req.URL,
		Events:        req.Events,
		SigningSecret: signingSecret,
		Active:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, fmt.Errorf("creating webhook: %w", err)
	}

	return webhookToResponse(webhook), nil
}

func (s *webhookService) List(ctx context.Context, teamID uuid.UUID) ([]dto.WebhookResponse, error) {
	webhooks, err := s.webhookRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks: %w", err)
	}

	responses := make([]dto.WebhookResponse, 0, len(webhooks))
	for _, w := range webhooks {
		responses = append(responses, *webhookToResponse(&w))
	}

	return responses, nil
}

func (s *webhookService) Get(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) (*dto.WebhookResponse, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	// Verify the webhook belongs to the team.
	if webhook.TeamID != teamID {
		return nil, fmt.Errorf("webhook not found: %w", postgres.ErrNotFound)
	}

	return webhookToResponse(webhook), nil
}

func (s *webhookService) Update(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID, req *dto.UpdateWebhookRequest) (*dto.WebhookResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	// Verify the webhook belongs to the team.
	if webhook.TeamID != teamID {
		return nil, fmt.Errorf("webhook not found: %w", postgres.ErrNotFound)
	}

	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		webhook.Events = req.Events
	}
	if req.Active != nil {
		webhook.Active = *req.Active
	}

	webhook.UpdatedAt = time.Now().UTC()

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, fmt.Errorf("updating webhook: %w", err)
	}

	return webhookToResponse(webhook), nil
}

func (s *webhookService) Delete(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) error {
	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("webhook not found: %w", err)
	}

	// Verify the webhook belongs to the team.
	if webhook.TeamID != teamID {
		return fmt.Errorf("webhook not found: %w", postgres.ErrNotFound)
	}

	if err := s.webhookRepo.Delete(ctx, webhookID); err != nil {
		return fmt.Errorf("deleting webhook: %w", err)
	}

	return nil
}

// webhookToResponse converts a model.Webhook to a dto.WebhookResponse.
func webhookToResponse(w *model.Webhook) *dto.WebhookResponse {
	return &dto.WebhookResponse{
		ID:        w.ID.String(),
		URL:       w.URL,
		Events:    w.Events,
		Active:    w.Active,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
	}
}
