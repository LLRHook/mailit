package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/worker"
)


// BroadcastService defines operations for managing and sending broadcasts.
type BroadcastService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateBroadcastRequest) (*dto.BroadcastResponse, error)
	List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.BroadcastResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID, req *dto.UpdateBroadcastRequest) (*dto.BroadcastResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) error
	Send(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error)
}

type broadcastService struct {
	broadcastRepo postgres.BroadcastRepository
	asynqClient   *asynq.Client
}

// NewBroadcastService creates a new BroadcastService.
func NewBroadcastService(broadcastRepo postgres.BroadcastRepository, asynqClient *asynq.Client) BroadcastService {
	return &broadcastService{
		broadcastRepo: broadcastRepo,
		asynqClient:   asynqClient,
	}
}

func (s *broadcastService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateBroadcastRequest) (*dto.BroadcastResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	now := time.Now().UTC()

	broadcast := &model.Broadcast{
		ID:          uuid.New(),
		TeamID:      teamID,
		Name:        req.Name,
		Status:      model.BroadcastStatusDraft,
		FromAddress: req.From,
		Subject:     req.Subject,
		HTMLBody:    req.HTML,
		TextBody:    req.Text,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Parse optional UUID references.
	if req.AudienceID != nil && *req.AudienceID != "" {
		id, err := uuid.Parse(*req.AudienceID)
		if err != nil {
			return nil, fmt.Errorf("invalid audience_id: %w", err)
		}
		broadcast.AudienceID = &id
	}
	if req.SegmentID != nil && *req.SegmentID != "" {
		id, err := uuid.Parse(*req.SegmentID)
		if err != nil {
			return nil, fmt.Errorf("invalid segment_id: %w", err)
		}
		broadcast.SegmentID = &id
	}
	if req.TemplateID != nil && *req.TemplateID != "" {
		id, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template_id: %w", err)
		}
		broadcast.TemplateID = &id
	}

	if err := s.broadcastRepo.Create(ctx, broadcast); err != nil {
		return nil, fmt.Errorf("creating broadcast: %w", err)
	}

	return broadcastToResponse(broadcast), nil
}

func (s *broadcastService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.BroadcastResponse], error) {
	params.Normalize()

	broadcasts, total, err := s.broadcastRepo.List(ctx, teamID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing broadcasts: %w", err)
	}

	data := make([]dto.BroadcastResponse, 0, len(broadcasts))
	for _, b := range broadcasts {
		data = append(data, *broadcastToResponse(&b))
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[dto.BroadcastResponse]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *broadcastService) Get(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error) {
	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("broadcast not found: %w", err)
	}

	// Verify the broadcast belongs to the team.
	if broadcast.TeamID != teamID {
		return nil, fmt.Errorf("broadcast not found: %w", postgres.ErrNotFound)
	}

	return broadcastToResponse(broadcast), nil
}

func (s *broadcastService) Update(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID, req *dto.UpdateBroadcastRequest) (*dto.BroadcastResponse, error) {
	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("broadcast not found: %w", err)
	}

	// Verify the broadcast belongs to the team.
	if broadcast.TeamID != teamID {
		return nil, fmt.Errorf("broadcast not found: %w", postgres.ErrNotFound)
	}

	// Only draft broadcasts can be updated.
	if broadcast.Status != model.BroadcastStatusDraft {
		return nil, fmt.Errorf("only draft broadcasts can be updated")
	}

	if req.Name != nil {
		broadcast.Name = *req.Name
	}
	if req.From != nil {
		broadcast.FromAddress = req.From
	}
	if req.Subject != nil {
		broadcast.Subject = req.Subject
	}
	if req.HTML != nil {
		broadcast.HTMLBody = req.HTML
	}
	if req.Text != nil {
		broadcast.TextBody = req.Text
	}

	if req.AudienceID != nil && *req.AudienceID != "" {
		id, err := uuid.Parse(*req.AudienceID)
		if err != nil {
			return nil, fmt.Errorf("invalid audience_id: %w", err)
		}
		broadcast.AudienceID = &id
	}
	if req.SegmentID != nil && *req.SegmentID != "" {
		id, err := uuid.Parse(*req.SegmentID)
		if err != nil {
			return nil, fmt.Errorf("invalid segment_id: %w", err)
		}
		broadcast.SegmentID = &id
	}

	broadcast.UpdatedAt = time.Now().UTC()

	if err := s.broadcastRepo.Update(ctx, broadcast); err != nil {
		return nil, fmt.Errorf("updating broadcast: %w", err)
	}

	return broadcastToResponse(broadcast), nil
}

func (s *broadcastService) Delete(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) error {
	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("broadcast not found: %w", err)
	}

	// Verify the broadcast belongs to the team.
	if broadcast.TeamID != teamID {
		return fmt.Errorf("broadcast not found: %w", postgres.ErrNotFound)
	}

	// Prevent deleting broadcasts that are actively sending.
	if broadcast.Status == model.BroadcastStatusSending {
		return fmt.Errorf("cannot delete a broadcast that is currently sending")
	}

	if err := s.broadcastRepo.Delete(ctx, broadcastID); err != nil {
		return fmt.Errorf("deleting broadcast: %w", err)
	}

	return nil
}

func (s *broadcastService) Send(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error) {
	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("broadcast not found: %w", err)
	}

	// Verify the broadcast belongs to the team.
	if broadcast.TeamID != teamID {
		return nil, fmt.Errorf("broadcast not found: %w", postgres.ErrNotFound)
	}

	// Validate the broadcast is ready to send.
	if broadcast.Status != model.BroadcastStatusDraft {
		return nil, fmt.Errorf("only draft broadcasts can be sent")
	}
	if broadcast.AudienceID == nil {
		return nil, fmt.Errorf("broadcast must have an audience before sending")
	}
	if broadcast.FromAddress == nil || *broadcast.FromAddress == "" {
		return nil, fmt.Errorf("broadcast must have a from address before sending")
	}
	// Must have content: either a template or inline HTML/text.
	hasTemplate := broadcast.TemplateID != nil
	hasInlineContent := (broadcast.HTMLBody != nil && *broadcast.HTMLBody != "") || (broadcast.TextBody != nil && *broadcast.TextBody != "")
	if !hasTemplate && !hasInlineContent {
		return nil, fmt.Errorf("broadcast must have content (template or inline HTML/text) before sending")
	}
	if broadcast.Subject == nil || *broadcast.Subject == "" {
		return nil, fmt.Errorf("broadcast must have a subject before sending")
	}

	// Update status to queued.
	broadcast.Status = model.BroadcastStatusQueued
	broadcast.UpdatedAt = time.Now().UTC()

	if err := s.broadcastRepo.Update(ctx, broadcast); err != nil {
		return nil, fmt.Errorf("updating broadcast status: %w", err)
	}

	// Enqueue the broadcast:send task.
	payload, err := json.Marshal(map[string]string{
		"broadcast_id": broadcast.ID.String(),
		"team_id":      teamID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling task payload: %w", err)
	}

	task := asynq.NewTask(worker.TaskBroadcastSend, payload)
	if _, err := s.asynqClient.Enqueue(task, asynq.Queue(worker.QueueCritical), asynq.MaxRetry(3)); err != nil {
		return nil, fmt.Errorf("enqueueing broadcast send task: %w", err)
	}

	return broadcastToResponse(broadcast), nil
}

// broadcastToResponse converts a model.Broadcast to a dto.BroadcastResponse.
func broadcastToResponse(b *model.Broadcast) *dto.BroadcastResponse {
	resp := &dto.BroadcastResponse{
		ID:         b.ID.String(),
		Name:       b.Name,
		Status:     b.Status,
		Recipients: b.TotalRecipients,
		Sent:       b.SentCount,
		CreatedAt:  b.CreatedAt.Format(time.RFC3339),
	}

	if b.AudienceID != nil {
		aid := b.AudienceID.String()
		resp.AudienceID = &aid
	}

	return resp
}
