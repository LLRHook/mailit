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

// TopicService defines operations for managing subscription topics.
type TopicService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTopicRequest) (*dto.TopicResponse, error)
	List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.TopicResponse], error)
	Update(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID, req *dto.UpdateTopicRequest) (*dto.TopicResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID) error
}

type topicService struct {
	topicRepo postgres.TopicRepository
}

// NewTopicService creates a new TopicService.
func NewTopicService(topicRepo postgres.TopicRepository) TopicService {
	return &topicService{
		topicRepo: topicRepo,
	}
}

func (s *topicService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTopicRequest) (*dto.TopicResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	now := time.Now().UTC()

	topic := &model.Topic{
		ID:          uuid.New(),
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.topicRepo.Create(ctx, topic); err != nil {
		return nil, fmt.Errorf("creating topic: %w", err)
	}

	return topicToResponse(topic), nil
}

func (s *topicService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.TopicResponse], error) {
	topics, err := s.topicRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing topics: %w", err)
	}

	responses := make([]dto.TopicResponse, 0, len(topics))
	for _, t := range topics {
		responses = append(responses, *topicToResponse(&t))
	}

	return &dto.ListResponse[dto.TopicResponse]{Data: responses}, nil
}

func (s *topicService) Update(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID, req *dto.UpdateTopicRequest) (*dto.TopicResponse, error) {
	topic, err := s.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("topic not found: %w", err)
	}

	// Verify the topic belongs to the team.
	if topic.TeamID != teamID {
		return nil, fmt.Errorf("topic not found: %w", postgres.ErrNotFound)
	}

	if req.Name != nil {
		topic.Name = *req.Name
	}
	if req.Description != nil {
		topic.Description = req.Description
	}

	topic.UpdatedAt = time.Now().UTC()

	if err := s.topicRepo.Update(ctx, topic); err != nil {
		return nil, fmt.Errorf("updating topic: %w", err)
	}

	return topicToResponse(topic), nil
}

func (s *topicService) Delete(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID) error {
	topic, err := s.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		return fmt.Errorf("topic not found: %w", err)
	}

	// Verify the topic belongs to the team.
	if topic.TeamID != teamID {
		return fmt.Errorf("topic not found: %w", postgres.ErrNotFound)
	}

	if err := s.topicRepo.Delete(ctx, topicID); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}

	return nil
}

// topicToResponse converts a model.Topic to a dto.TopicResponse.
func topicToResponse(t *model.Topic) *dto.TopicResponse {
	return &dto.TopicResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
	}
}
