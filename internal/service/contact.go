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

// ContactService defines operations for managing contacts within audiences.
type ContactService interface {
	Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateContactRequest) (*dto.ContactResponse, error)
	List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.ContactResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) (*dto.ContactResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID, req *dto.UpdateContactRequest) (*dto.ContactResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) error
}

type contactService struct {
	contactRepo  postgres.ContactRepository
	audienceRepo postgres.AudienceRepository
}

// NewContactService creates a new ContactService.
func NewContactService(contactRepo postgres.ContactRepository, audienceRepo postgres.AudienceRepository) ContactService {
	return &contactService{
		contactRepo:  contactRepo,
		audienceRepo: audienceRepo,
	}
}

// verifyAudienceOwnership checks that the audience exists and belongs to the team.
func (s *contactService) verifyAudienceOwnership(ctx context.Context, teamID, audienceID uuid.UUID) error {
	audience, err := s.audienceRepo.GetByID(ctx, audienceID)
	if err != nil {
		return fmt.Errorf("audience not found: %w", err)
	}
	if audience.TeamID != teamID {
		return fmt.Errorf("audience not found: %w", postgres.ErrNotFound)
	}
	return nil
}

func (s *contactService) Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateContactRequest) (*dto.ContactResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	// Check for duplicate email within the audience.
	existing, _ := s.contactRepo.GetByAudienceAndEmail(ctx, audienceID, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("a contact with email %s already exists in this audience", req.Email)
	}

	now := time.Now().UTC()

	contact := &model.Contact{
		ID:         uuid.New(),
		AudienceID: audienceID,
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if req.Unsubscribed != nil {
		contact.Unsubscribed = *req.Unsubscribed
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("creating contact: %w", err)
	}

	return contactToResponse(contact), nil
}

func (s *contactService) List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.ContactResponse], error) {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	params.Normalize()

	contacts, total, err := s.contactRepo.List(ctx, audienceID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing contacts: %w", err)
	}

	data := make([]dto.ContactResponse, 0, len(contacts))
	for _, c := range contacts {
		data = append(data, *contactToResponse(&c))
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[dto.ContactResponse]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *contactService) Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) (*dto.ContactResponse, error) {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	// Verify the contact belongs to the audience.
	if contact.AudienceID != audienceID {
		return nil, fmt.Errorf("contact not found: %w", postgres.ErrNotFound)
	}

	return contactToResponse(contact), nil
}

func (s *contactService) Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID, req *dto.UpdateContactRequest) (*dto.ContactResponse, error) {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	// Verify the contact belongs to the audience.
	if contact.AudienceID != audienceID {
		return nil, fmt.Errorf("contact not found: %w", postgres.ErrNotFound)
	}

	if req.FirstName != nil {
		contact.FirstName = req.FirstName
	}
	if req.LastName != nil {
		contact.LastName = req.LastName
	}
	if req.Unsubscribed != nil {
		contact.Unsubscribed = *req.Unsubscribed
	}

	contact.UpdatedAt = time.Now().UTC()

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, fmt.Errorf("updating contact: %w", err)
	}

	return contactToResponse(contact), nil
}

func (s *contactService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) error {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return err
	}

	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("contact not found: %w", err)
	}

	// Verify the contact belongs to the audience.
	if contact.AudienceID != audienceID {
		return fmt.Errorf("contact not found: %w", postgres.ErrNotFound)
	}

	if err := s.contactRepo.Delete(ctx, contactID); err != nil {
		return fmt.Errorf("deleting contact: %w", err)
	}

	return nil
}

// contactToResponse converts a model.Contact to a dto.ContactResponse.
func contactToResponse(c *model.Contact) *dto.ContactResponse {
	return &dto.ContactResponse{
		ID:           c.ID.String(),
		Email:        c.Email,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		Unsubscribed: c.Unsubscribed,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
	}
}
