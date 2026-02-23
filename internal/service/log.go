package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// LogService defines operations for querying application logs.
type LogService interface {
	List(ctx context.Context, teamID uuid.UUID, level string, params *dto.PaginationParams) (*dto.PaginatedResponse[model.Log], error)
}

type logService struct {
	logRepo postgres.LogRepository
}

// NewLogService creates a new LogService.
func NewLogService(logRepo postgres.LogRepository) LogService {
	return &logService{
		logRepo: logRepo,
	}
}

func (s *logService) List(ctx context.Context, teamID uuid.UUID, level string, params *dto.PaginationParams) (*dto.PaginatedResponse[model.Log], error) {
	params.Normalize()

	logs, total, err := s.logRepo.List(ctx, teamID, level, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing logs: %w", err)
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[model.Log]{
		Data:       logs,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}
