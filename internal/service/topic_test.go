package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func newTestTopic() *model.Topic {
	desc := "Monthly updates"
	return &model.Topic{
		ID:          uuid.New(),
		TeamID:      testutil.TestTeamID,
		Name:        "Monthly Newsletter",
		Description: &desc,
		CreatedAt:   testutil.FixedTime,
		UpdatedAt:   testutil.FixedTime,
	}
}

func TestTopicService_Create(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	topicRepo.On("Create", ctx, mock.AnythingOfType("*model.Topic")).Return(nil)

	req := &dto.CreateTopicRequest{
		Name:        "Product Updates",
		Description: testutil.StringPtr("Updates about our products"),
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Product Updates", resp.Name)

	topicRepo.AssertExpectations(t)
}

func TestTopicService_List(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	topic := *newTestTopic()
	topicRepo.On("ListByTeamID", ctx, teamID).Return([]model.Topic{topic}, nil)

	resp, err := svc.List(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Monthly Newsletter", resp.Data[0].Name)

	topicRepo.AssertExpectations(t)
}

func TestTopicService_Update_HappyPath(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	topic := newTestTopic()
	topicRepo.On("GetByID", ctx, topic.ID).Return(topic, nil)
	topicRepo.On("Update", ctx, mock.AnythingOfType("*model.Topic")).Return(nil)

	req := &dto.UpdateTopicRequest{
		Name: testutil.StringPtr("Updated Name"),
	}

	resp, err := svc.Update(ctx, teamID, topic.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.Name)

	topicRepo.AssertExpectations(t)
}

func TestTopicService_Update_WrongTeam(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	topic := newTestTopic()
	topicRepo.On("GetByID", ctx, topic.ID).Return(topic, nil)

	req := &dto.UpdateTopicRequest{
		Name: testutil.StringPtr("Updated"),
	}

	resp, err := svc.Update(ctx, wrongTeamID, topic.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	topicRepo.AssertExpectations(t)
}

func TestTopicService_Delete_HappyPath(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	topic := newTestTopic()
	topicRepo.On("GetByID", ctx, topic.ID).Return(topic, nil)
	topicRepo.On("Delete", ctx, topic.ID).Return(nil)

	err := svc.Delete(ctx, teamID, topic.ID)

	require.NoError(t, err)

	topicRepo.AssertExpectations(t)
}

func TestTopicService_Delete_WrongTeam(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	topic := newTestTopic()
	topicRepo.On("GetByID", ctx, topic.ID).Return(topic, nil)

	err := svc.Delete(ctx, wrongTeamID, topic.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	topicRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestTopicService_Delete_NotFound(t *testing.T) {
	topicRepo := new(tmock.MockTopicRepository)
	svc := NewTopicService(topicRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	topicRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	err := svc.Delete(ctx, teamID, badID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	topicRepo.AssertExpectations(t)
}
