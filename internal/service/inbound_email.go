package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// InboundEmailService defines operations for viewing received inbound emails.
type InboundEmailService interface {
	List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[model.InboundEmail], error)
	Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*model.InboundEmail, error)
}

type inboundEmailService struct {
	inboundEmailRepo postgres.InboundEmailRepository
}

// NewInboundEmailService creates a new InboundEmailService.
func NewInboundEmailService(inboundEmailRepo postgres.InboundEmailRepository) InboundEmailService {
	return &inboundEmailService{
		inboundEmailRepo: inboundEmailRepo,
	}
}

func (s *inboundEmailService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[model.InboundEmail], error) {
	params.Normalize()

	emails, total, err := s.inboundEmailRepo.List(ctx, teamID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing inbound emails: %w", err)
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[model.InboundEmail]{
		Data:       emails,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *inboundEmailService) Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*model.InboundEmail, error) {
	email, err := s.inboundEmailRepo.GetByTeamAndID(ctx, teamID, emailID)
	if err != nil {
		return nil, fmt.Errorf("inbound email not found: %w", err)
	}

	return email, nil
}
