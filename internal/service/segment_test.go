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

func newTestSegment(audienceID uuid.UUID) *model.Segment {
	return &model.Segment{
		ID:         uuid.New(),
		AudienceID: audienceID,
		Name:       "Active Users",
		Conditions: model.JSONArray{map[string]interface{}{"field": "email", "op": "contains", "value": "@example.com"}},
		CreatedAt:  testutil.FixedTime,
		UpdatedAt:  testutil.FixedTime,
	}
}

func TestSegmentService_Create_HappyPath(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("Create", ctx, mock.AnythingOfType("*model.Segment")).Return(nil)

	conditions := []interface{}{
		map[string]interface{}{"field": "email", "op": "contains", "value": "@test.com"},
	}
	req := &dto.CreateSegmentRequest{
		Name:       "Test Segment",
		Conditions: conditions,
	}

	resp, err := svc.Create(ctx, teamID, aud.ID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Test Segment", resp.Name)

	segmentRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestSegmentService_Create_AudienceOwnershipFails(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)

	conditions := []interface{}{"test"}
	req := &dto.CreateSegmentRequest{
		Name:       "Test Segment",
		Conditions: conditions,
	}

	resp, err := svc.Create(ctx, wrongTeamID, aud.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience not found")

	audienceRepo.AssertExpectations(t)
}

func TestSegmentService_List(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	seg := *newTestSegment(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("ListByAudienceID", ctx, aud.ID).Return([]model.Segment{seg}, nil)

	resp, err := svc.List(ctx, teamID, aud.ID)

	require.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Active Users", resp[0].Name)

	segmentRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestSegmentService_Update_HappyPath(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	seg := newTestSegment(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("GetByID", ctx, seg.ID).Return(seg, nil)
	segmentRepo.On("Update", ctx, mock.AnythingOfType("*model.Segment")).Return(nil)

	req := &dto.UpdateSegmentRequest{
		Name: testutil.StringPtr("Updated Segment"),
	}

	resp, err := svc.Update(ctx, teamID, aud.ID, seg.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated Segment", resp.Name)

	segmentRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestSegmentService_Delete_HappyPath(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	seg := newTestSegment(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("GetByID", ctx, seg.ID).Return(seg, nil)
	segmentRepo.On("Delete", ctx, seg.ID).Return(nil)

	err := svc.Delete(ctx, teamID, aud.ID, seg.ID)

	require.NoError(t, err)

	segmentRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestSegmentService_Delete_WrongAudience(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	// Segment belongs to a different audience.
	differentAudienceID := uuid.New()
	seg := newTestSegment(differentAudienceID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("GetByID", ctx, seg.ID).Return(seg, nil)

	err := svc.Delete(ctx, teamID, aud.ID, seg.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	segmentRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestSegmentService_Get_NotFound(t *testing.T) {
	segmentRepo := new(tmock.MockSegmentRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewSegmentService(segmentRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	badID := uuid.New()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	segmentRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	err := svc.Delete(ctx, teamID, aud.ID, badID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	segmentRepo.AssertExpectations(t)
}
