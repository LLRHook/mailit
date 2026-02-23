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

// ContactPropertyService defines operations for managing custom contact properties.
type ContactPropertyService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateContactPropertyRequest) (*dto.ContactPropertyResponse, error)
	List(ctx context.Context, teamID uuid.UUID) ([]dto.ContactPropertyResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID, req *dto.UpdateContactPropertyRequest) (*dto.ContactPropertyResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID) error
}

type contactPropertyService struct {
	propertyRepo postgres.ContactPropertyRepository
}

// NewContactPropertyService creates a new ContactPropertyService.
func NewContactPropertyService(propertyRepo postgres.ContactPropertyRepository) ContactPropertyService {
	return &contactPropertyService{
		propertyRepo: propertyRepo,
	}
}

func (s *contactPropertyService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateContactPropertyRequest) (*dto.ContactPropertyResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	now := time.Now().UTC()

	property := &model.ContactProperty{
		ID:        uuid.New(),
		TeamID:    teamID,
		Name:      req.Name,
		Label:     req.Label,
		Type:      req.Type,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.propertyRepo.Create(ctx, property); err != nil {
		return nil, fmt.Errorf("creating contact property: %w", err)
	}

	return contactPropertyToResponse(property), nil
}

func (s *contactPropertyService) List(ctx context.Context, teamID uuid.UUID) ([]dto.ContactPropertyResponse, error) {
	properties, err := s.propertyRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing contact properties: %w", err)
	}

	responses := make([]dto.ContactPropertyResponse, 0, len(properties))
	for _, p := range properties {
		responses = append(responses, *contactPropertyToResponse(&p))
	}

	return responses, nil
}

func (s *contactPropertyService) Update(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID, req *dto.UpdateContactPropertyRequest) (*dto.ContactPropertyResponse, error) {
	property, err := s.propertyRepo.GetByID(ctx, propertyID)
	if err != nil {
		return nil, fmt.Errorf("contact property not found: %w", err)
	}

	// Verify the property belongs to the team.
	if property.TeamID != teamID {
		return nil, fmt.Errorf("contact property not found: %w", postgres.ErrNotFound)
	}

	if req.Label != nil {
		property.Label = *req.Label
	}

	property.UpdatedAt = time.Now().UTC()

	if err := s.propertyRepo.Update(ctx, property); err != nil {
		return nil, fmt.Errorf("updating contact property: %w", err)
	}

	return contactPropertyToResponse(property), nil
}

func (s *contactPropertyService) Delete(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID) error {
	property, err := s.propertyRepo.GetByID(ctx, propertyID)
	if err != nil {
		return fmt.Errorf("contact property not found: %w", err)
	}

	// Verify the property belongs to the team.
	if property.TeamID != teamID {
		return fmt.Errorf("contact property not found: %w", postgres.ErrNotFound)
	}

	if err := s.propertyRepo.Delete(ctx, propertyID); err != nil {
		return fmt.Errorf("deleting contact property: %w", err)
	}

	return nil
}

// contactPropertyToResponse converts a model.ContactProperty to a dto.ContactPropertyResponse.
func contactPropertyToResponse(p *model.ContactProperty) *dto.ContactPropertyResponse {
	return &dto.ContactPropertyResponse{
		ID:        p.ID.String(),
		Name:      p.Name,
		Label:     p.Label,
		Type:      p.Type,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}
