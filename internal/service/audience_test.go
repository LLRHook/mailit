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

func TestAudienceService_Create_HappyPath(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	audienceRepo.On("Create", ctx, mock.AnythingOfType("*model.Audience")).Return(nil)

	req := &dto.CreateAudienceRequest{Name: "Newsletter Subscribers"}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Newsletter Subscribers", resp.Name)

	audienceRepo.AssertExpectations(t)
}

func TestAudienceService_List(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud1 := *testutil.NewTestAudience()
	aud2 := *testutil.NewTestAudience()
	aud2.Name = "VIP Users"
	audienceRepo.On("ListByTeamID", ctx, teamID).Return([]model.Audience{aud1, aud2}, nil)

	resp, err := svc.List(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)

	audienceRepo.AssertExpectations(t)
}

func TestAudienceService_Get_HappyPath(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)

	resp, err := svc.Get(ctx, teamID, aud.ID)

	require.NoError(t, err)
	assert.Equal(t, aud.ID.String(), resp.ID)

	audienceRepo.AssertExpectations(t)
}

func TestAudienceService_Get_WrongTeam(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)

	resp, err := svc.Get(ctx, wrongTeamID, aud.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	audienceRepo.AssertExpectations(t)
}

func TestAudienceService_Delete_HappyPath(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	audienceRepo.On("Delete", ctx, aud.ID).Return(nil)

	err := svc.Delete(ctx, teamID, aud.ID)

	require.NoError(t, err)

	audienceRepo.AssertExpectations(t)
}

func TestAudienceService_Delete_WrongTeam(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)

	err := svc.Delete(ctx, wrongTeamID, aud.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	audienceRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestAudienceService_Get_NotFound(t *testing.T) {
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewAudienceService(audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	audienceRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	resp, err := svc.Get(ctx, teamID, badID)

	assert.Nil(t, resp)
	assert.Error(t, err)

	audienceRepo.AssertExpectations(t)
}
