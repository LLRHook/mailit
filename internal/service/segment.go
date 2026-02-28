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

// SegmentService defines operations for managing audience segments.
type SegmentService interface {
	Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateSegmentRequest) (*dto.SegmentResponse, error)
	List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.ListResponse[dto.SegmentResponse], error)
	Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID, req *dto.UpdateSegmentRequest) (*dto.SegmentResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID) error
}

type segmentService struct {
	segmentRepo  postgres.SegmentRepository
	audienceRepo postgres.AudienceRepository
}

// NewSegmentService creates a new SegmentService.
func NewSegmentService(segmentRepo postgres.SegmentRepository, audienceRepo postgres.AudienceRepository) SegmentService {
	return &segmentService{
		segmentRepo:  segmentRepo,
		audienceRepo: audienceRepo,
	}
}

// verifyAudienceOwnership checks that the audience exists and belongs to the team.
func (s *segmentService) verifyAudienceOwnership(ctx context.Context, teamID, audienceID uuid.UUID) error {
	audience, err := s.audienceRepo.GetByID(ctx, audienceID)
	if err != nil {
		return fmt.Errorf("audience not found: %w", err)
	}
	if audience.TeamID != teamID {
		return fmt.Errorf("audience not found: %w", postgres.ErrNotFound)
	}
	return nil
}

func (s *segmentService) Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateSegmentRequest) (*dto.SegmentResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	// Convert conditions to JSONArray.
	conditions, err := toJSONArray(req.Conditions)
	if err != nil {
		return nil, fmt.Errorf("invalid conditions format: %w", err)
	}

	now := time.Now().UTC()

	segment := &model.Segment{
		ID:         uuid.New(),
		AudienceID: audienceID,
		Name:       req.Name,
		Conditions: conditions,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.segmentRepo.Create(ctx, segment); err != nil {
		return nil, fmt.Errorf("creating segment: %w", err)
	}

	return segmentToResponse(segment), nil
}

func (s *segmentService) List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.ListResponse[dto.SegmentResponse], error) {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	segments, err := s.segmentRepo.ListByAudienceID(ctx, audienceID)
	if err != nil {
		return nil, fmt.Errorf("listing segments: %w", err)
	}

	responses := make([]dto.SegmentResponse, 0, len(segments))
	for _, seg := range segments {
		responses = append(responses, *segmentToResponse(&seg))
	}

	return &dto.ListResponse[dto.SegmentResponse]{Data: responses}, nil
}

func (s *segmentService) Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID, req *dto.UpdateSegmentRequest) (*dto.SegmentResponse, error) {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return nil, err
	}

	segment, err := s.segmentRepo.GetByID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("segment not found: %w", err)
	}

	// Verify the segment belongs to the audience.
	if segment.AudienceID != audienceID {
		return nil, fmt.Errorf("segment not found: %w", postgres.ErrNotFound)
	}

	if req.Name != nil {
		segment.Name = *req.Name
	}
	if req.Conditions != nil {
		conditions, err := toJSONArray(req.Conditions)
		if err != nil {
			return nil, fmt.Errorf("invalid conditions format: %w", err)
		}
		segment.Conditions = conditions
	}

	segment.UpdatedAt = time.Now().UTC()

	if err := s.segmentRepo.Update(ctx, segment); err != nil {
		return nil, fmt.Errorf("updating segment: %w", err)
	}

	return segmentToResponse(segment), nil
}

func (s *segmentService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID) error {
	if err := s.verifyAudienceOwnership(ctx, teamID, audienceID); err != nil {
		return err
	}

	segment, err := s.segmentRepo.GetByID(ctx, segmentID)
	if err != nil {
		return fmt.Errorf("segment not found: %w", err)
	}

	// Verify the segment belongs to the audience.
	if segment.AudienceID != audienceID {
		return fmt.Errorf("segment not found: %w", postgres.ErrNotFound)
	}

	if err := s.segmentRepo.Delete(ctx, segmentID); err != nil {
		return fmt.Errorf("deleting segment: %w", err)
	}

	return nil
}

// segmentToResponse converts a model.Segment to a dto.SegmentResponse.
func segmentToResponse(seg *model.Segment) *dto.SegmentResponse {
	var conditions interface{} = seg.Conditions
	return &dto.SegmentResponse{
		ID:         seg.ID.String(),
		Name:       seg.Name,
		Conditions: conditions,
		CreatedAt:  seg.CreatedAt.Format(time.RFC3339),
	}
}

// toJSONArray converts an arbitrary value to a model.JSONArray.
func toJSONArray(v interface{}) (model.JSONArray, error) {
	if v == nil {
		return model.JSONArray{}, nil
	}

	// If it's already a JSONArray, return as-is.
	if ja, ok := v.(model.JSONArray); ok {
		return ja, nil
	}

	// If it's a slice of interface{}, convert directly.
	if slice, ok := v.([]interface{}); ok {
		return model.JSONArray(slice), nil
	}

	// Otherwise, wrap the value as a single-element array.
	return model.JSONArray{v}, nil
}
