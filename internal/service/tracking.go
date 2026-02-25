package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/worker"
)

// TrackingService defines the interface for handling tracking events.
type TrackingService interface {
	HandleOpen(ctx context.Context, linkID uuid.UUID) error
	HandleClick(ctx context.Context, linkID uuid.UUID) (originalURL string, err error)
	HandleUnsubscribe(ctx context.Context, linkID uuid.UUID) error
}

type trackingService struct {
	trackingRepo    postgres.TrackingLinkRepository
	emailRepo       postgres.EmailRepository
	eventRepo       postgres.EmailEventRepository
	contactRepo     postgres.ContactRepository
	audienceRepo    postgres.AudienceRepository
	webhookDispatch worker.WebhookDispatchFunc
	metricsIncrement worker.MetricsIncrementFunc
}

func NewTrackingService(
	trackingRepo postgres.TrackingLinkRepository,
	emailRepo postgres.EmailRepository,
	eventRepo postgres.EmailEventRepository,
	contactRepo postgres.ContactRepository,
	audienceRepo postgres.AudienceRepository,
	webhookDispatch worker.WebhookDispatchFunc,
	metricsIncrement worker.MetricsIncrementFunc,
) TrackingService {
	return &trackingService{
		trackingRepo:     trackingRepo,
		emailRepo:        emailRepo,
		eventRepo:        eventRepo,
		contactRepo:      contactRepo,
		audienceRepo:     audienceRepo,
		webhookDispatch:  webhookDispatch,
		metricsIncrement: metricsIncrement,
	}
}

func (s *trackingService) HandleOpen(ctx context.Context, linkID uuid.UUID) error {
	link, err := s.trackingRepo.GetByID(ctx, linkID)
	if err != nil {
		return fmt.Errorf("tracking link not found: %w", err)
	}

	if link.Type != model.TrackingTypeOpen {
		return fmt.Errorf("invalid tracking link type: %s", link.Type)
	}

	// Create opened event.
	event := &model.EmailEvent{
		ID:        uuid.New(),
		EmailID:   link.EmailID,
		Type:      model.EventOpened,
		Payload:   model.JSONMap{"recipient": link.Recipient},
		Recipient: &link.Recipient,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return fmt.Errorf("creating opened event: %w", err)
	}

	// Increment metrics.
	if s.metricsIncrement != nil {
		s.metricsIncrement(ctx, link.TeamID, model.EventOpened)
	}

	// Dispatch webhook.
	if s.webhookDispatch != nil {
		s.webhookDispatch(ctx, link.TeamID, "email.opened", map[string]interface{}{
			"email_id":  link.EmailID.String(),
			"recipient": link.Recipient,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return nil
}

func (s *trackingService) HandleClick(ctx context.Context, linkID uuid.UUID) (string, error) {
	link, err := s.trackingRepo.GetByID(ctx, linkID)
	if err != nil {
		return "", fmt.Errorf("tracking link not found: %w", err)
	}

	if link.Type != model.TrackingTypeClick {
		return "", fmt.Errorf("invalid tracking link type: %s", link.Type)
	}

	if link.OriginalURL == nil {
		return "", fmt.Errorf("tracking link has no original URL")
	}

	// Create clicked event.
	event := &model.EmailEvent{
		ID:      uuid.New(),
		EmailID: link.EmailID,
		Type:    model.EventClicked,
		Payload: model.JSONMap{
			"recipient": link.Recipient,
			"url":       *link.OriginalURL,
		},
		Recipient: &link.Recipient,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return "", fmt.Errorf("creating clicked event: %w", err)
	}

	// Increment metrics.
	if s.metricsIncrement != nil {
		s.metricsIncrement(ctx, link.TeamID, model.EventClicked)
	}

	// Dispatch webhook.
	if s.webhookDispatch != nil {
		s.webhookDispatch(ctx, link.TeamID, "email.clicked", map[string]interface{}{
			"email_id":  link.EmailID.String(),
			"recipient": link.Recipient,
			"url":       *link.OriginalURL,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return *link.OriginalURL, nil
}

func (s *trackingService) HandleUnsubscribe(ctx context.Context, linkID uuid.UUID) error {
	link, err := s.trackingRepo.GetByID(ctx, linkID)
	if err != nil {
		return fmt.Errorf("tracking link not found: %w", err)
	}

	if link.Type != model.TrackingTypeUnsubscribe {
		return fmt.Errorf("invalid tracking link type: %s", link.Type)
	}

	// Find all contacts matching this email across all audiences for this team.
	audiences, err := s.audienceRepo.ListByTeamID(ctx, link.TeamID)
	if err != nil {
		return fmt.Errorf("listing audiences: %w", err)
	}

	for _, audience := range audiences {
		contact, err := s.contactRepo.GetByAudienceAndEmail(ctx, audience.ID, link.Recipient)
		if err != nil {
			continue // Not in this audience.
		}
		if !contact.Unsubscribed {
			contact.Unsubscribed = true
			contact.UpdatedAt = time.Now().UTC()
			if err := s.contactRepo.Update(ctx, contact); err != nil {
				return fmt.Errorf("unsubscribing contact %s: %w", contact.ID, err)
			}
		}
	}

	// Create unsubscribed event.
	event := &model.EmailEvent{
		ID:        uuid.New(),
		EmailID:   link.EmailID,
		Type:      model.EventUnsubscribed,
		Payload:   model.JSONMap{"recipient": link.Recipient},
		Recipient: &link.Recipient,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return fmt.Errorf("creating unsubscribed event: %w", err)
	}

	// Dispatch webhook.
	if s.webhookDispatch != nil {
		s.webhookDispatch(ctx, link.TeamID, "contact.unsubscribed", map[string]interface{}{
			"email_id":  link.EmailID.String(),
			"recipient": link.Recipient,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return nil
}
