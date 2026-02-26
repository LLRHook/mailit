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

// TemplateService defines operations for managing email templates and their versions.
type TemplateService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTemplateRequest) (*dto.TemplateResponse, error)
	List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.TemplateResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateDetailResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID, req *dto.UpdateTemplateRequest) (*dto.TemplateResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) error
	Publish(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateResponse, error)
}

type templateService struct {
	templateRepo        postgres.TemplateRepository
	templateVersionRepo postgres.TemplateVersionRepository
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(templateRepo postgres.TemplateRepository, templateVersionRepo postgres.TemplateVersionRepository) TemplateService {
	return &templateService{
		templateRepo:        templateRepo,
		templateVersionRepo: templateVersionRepo,
	}
}

func (s *templateService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTemplateRequest) (*dto.TemplateResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	now := time.Now().UTC()

	// Create the template record.
	template := &model.Template{
		ID:          uuid.New(),
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("creating template: %w", err)
	}

	// Create the initial template version (v1, unpublished).
	version := &model.TemplateVersion{
		ID:         uuid.New(),
		TemplateID: template.ID,
		Version:    1,
		Subject:    req.Subject,
		HTMLBody:   req.HTML,
		TextBody:   req.Text,
		Variables:  model.JSONArray{},
		Published:  false,
		CreatedAt:  now,
	}

	if err := s.templateVersionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("creating template version: %w", err)
	}

	return templateToResponse(template), nil
}

func (s *templateService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.TemplateResponse], error) {
	params.Normalize()

	templates, total, err := s.templateRepo.List(ctx, teamID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}

	data := make([]dto.TemplateResponse, 0, len(templates))
	for _, t := range templates {
		data = append(data, *templateToResponse(&t))
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[dto.TemplateResponse]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *templateService) Get(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateDetailResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Verify the template belongs to the team.
	if template.TeamID != teamID {
		return nil, fmt.Errorf("template not found: %w", postgres.ErrNotFound)
	}

	resp := &dto.TemplateDetailResponse{
		ID:          template.ID.String(),
		Name:        template.Name,
		Description: template.Description,
		CreatedAt:   template.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   template.UpdatedAt.Format(time.RFC3339),
	}

	// Fetch latest version to include content fields.
	latest, err := s.templateVersionRepo.GetLatestByTemplateID(ctx, template.ID)
	if err == nil && latest != nil {
		resp.Subject = latest.Subject
		resp.HTML = latest.HTMLBody
		resp.Text = latest.TextBody
		resp.Published = latest.Published
	}

	return resp, nil
}

func (s *templateService) Update(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID, req *dto.UpdateTemplateRequest) (*dto.TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Verify the template belongs to the team.
	if template.TeamID != teamID {
		return nil, fmt.Errorf("template not found: %w", postgres.ErrNotFound)
	}

	// Update template metadata.
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = req.Description
	}

	template.UpdatedAt = time.Now().UTC()

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("updating template: %w", err)
	}

	// If content fields are provided, create a new version.
	if req.Subject != nil || req.HTML != nil || req.Text != nil {
		latest, err := s.templateVersionRepo.GetLatestByTemplateID(ctx, template.ID)
		if err != nil {
			return nil, fmt.Errorf("getting latest template version: %w", err)
		}

		newVersion := &model.TemplateVersion{
			ID:         uuid.New(),
			TemplateID: template.ID,
			Version:    latest.Version + 1,
			Subject:    latest.Subject,
			HTMLBody:   latest.HTMLBody,
			TextBody:   latest.TextBody,
			Variables:  latest.Variables,
			Published:  false,
			CreatedAt:  time.Now().UTC(),
		}

		// Override with provided values.
		if req.Subject != nil {
			newVersion.Subject = req.Subject
		}
		if req.HTML != nil {
			newVersion.HTMLBody = req.HTML
		}
		if req.Text != nil {
			newVersion.TextBody = req.Text
		}

		if err := s.templateVersionRepo.Create(ctx, newVersion); err != nil {
			return nil, fmt.Errorf("creating new template version: %w", err)
		}
	}

	return templateToResponse(template), nil
}

func (s *templateService) Delete(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) error {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Verify the template belongs to the team.
	if template.TeamID != teamID {
		return fmt.Errorf("template not found: %w", postgres.ErrNotFound)
	}

	if err := s.templateRepo.Delete(ctx, templateID); err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}

	return nil
}

func (s *templateService) Publish(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Verify the template belongs to the team.
	if template.TeamID != teamID {
		return nil, fmt.Errorf("template not found: %w", postgres.ErrNotFound)
	}

	// Get the latest version and publish it.
	latest, err := s.templateVersionRepo.GetLatestByTemplateID(ctx, template.ID)
	if err != nil {
		return nil, fmt.Errorf("getting latest template version: %w", err)
	}

	if latest.Published {
		return nil, fmt.Errorf("latest version is already published")
	}

	if err := s.templateVersionRepo.Publish(ctx, latest.ID); err != nil {
		return nil, fmt.Errorf("publishing template version: %w", err)
	}

	template.UpdatedAt = time.Now().UTC()
	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("updating template: %w", err)
	}

	return templateToResponse(template), nil
}

// templateToResponse converts a model.Template to a dto.TemplateResponse.
func templateToResponse(t *model.Template) *dto.TemplateResponse {
	return &dto.TemplateResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
	}
}
