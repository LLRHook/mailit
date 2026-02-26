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

// AudienceService defines operations for managing audiences (contact lists).
type AudienceService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAudienceRequest) (*dto.AudienceResponse, error)
	List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.AudienceResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.AudienceResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) error
}

type audienceService struct {
	audienceRepo postgres.AudienceRepository
}

// NewAudienceService creates a new AudienceService.
func NewAudienceService(audienceRepo postgres.AudienceRepository) AudienceService {
	return &audienceService{
		audienceRepo: audienceRepo,
	}
}

func (s *audienceService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAudienceRequest) (*dto.AudienceResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	now := time.Now().UTC()

	audience := &model.Audience{
		ID:        uuid.New(),
		TeamID:    teamID,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.audienceRepo.Create(ctx, audience); err != nil {
		return nil, fmt.Errorf("creating audience: %w", err)
	}

	return audienceToResponse(audience), nil
}

func (s *audienceService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.AudienceResponse], error) {
	audiences, err := s.audienceRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing audiences: %w", err)
	}

	responses := make([]dto.AudienceResponse, 0, len(audiences))
	for _, a := range audiences {
		responses = append(responses, *audienceToResponse(&a))
	}

	return &dto.ListResponse[dto.AudienceResponse]{Data: responses}, nil
}

func (s *audienceService) Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.AudienceResponse, error) {
	audience, err := s.audienceRepo.GetByID(ctx, audienceID)
	if err != nil {
		return nil, fmt.Errorf("audience not found: %w", err)
	}

	// Verify the audience belongs to the team.
	if audience.TeamID != teamID {
		return nil, fmt.Errorf("audience not found: %w", postgres.ErrNotFound)
	}

	return audienceToResponse(audience), nil
}

func (s *audienceService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) error {
	audience, err := s.audienceRepo.GetByID(ctx, audienceID)
	if err != nil {
		return fmt.Errorf("audience not found: %w", err)
	}

	// Verify the audience belongs to the team.
	if audience.TeamID != teamID {
		return fmt.Errorf("audience not found: %w", postgres.ErrNotFound)
	}

	if err := s.audienceRepo.Delete(ctx, audienceID); err != nil {
		return fmt.Errorf("deleting audience: %w", err)
	}

	return nil
}

// audienceToResponse converts a model.Audience to a dto.AudienceResponse.
func audienceToResponse(a *model.Audience) *dto.AudienceResponse {
	return &dto.AudienceResponse{
		ID:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}
