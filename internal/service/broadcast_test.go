package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func newBroadcastTestDeps(t *testing.T) (*tmock.MockBroadcastRepository, *asynq.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	return new(tmock.MockBroadcastRepository), asynqClient
}

func TestBroadcastService_Create_HappyPath(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	broadcastRepo.On("Create", ctx, mock.AnythingOfType("*model.Broadcast")).Return(nil)

	req := &dto.CreateBroadcastRequest{
		Name:    "Weekly Newsletter",
		From:    testutil.StringPtr("news@example.com"),
		Subject: testutil.StringPtr("This week's update"),
		HTML:    testutil.StringPtr("<p>Newsletter content</p>"),
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Weekly Newsletter", resp.Name)
	assert.Equal(t, model.BroadcastStatusDraft, resp.Status)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_List_Paginated(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := *testutil.NewTestBroadcast()
	broadcastRepo.On("List", ctx, teamID, 20, 0).Return([]model.Broadcast{bc}, 1, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Get_HappyPath(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Get(ctx, teamID, bc.ID)

	require.NoError(t, err)
	assert.Equal(t, bc.ID.String(), resp.ID)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Get_WrongTeam(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	bc := testutil.NewTestBroadcast()
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Get(ctx, wrongTeamID, bc.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Update_OnlyDraft(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	// Draft broadcast can be updated.
	bc := testutil.NewTestBroadcast()
	bc.Status = model.BroadcastStatusDraft
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)
	broadcastRepo.On("Update", ctx, mock.AnythingOfType("*model.Broadcast")).Return(nil)

	req := &dto.UpdateBroadcastRequest{
		Name: testutil.StringPtr("Updated Name"),
	}

	resp, err := svc.Update(ctx, teamID, bc.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.Name)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Update_NonDraftFails(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.Status = model.BroadcastStatusSending
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	req := &dto.UpdateBroadcastRequest{
		Name: testutil.StringPtr("Updated Name"),
	}

	resp, err := svc.Update(ctx, teamID, bc.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only draft")

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Delete_CannotDeleteSending(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.Status = model.BroadcastStatusSending
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	err := svc.Delete(ctx, teamID, bc.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currently sending")

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Delete_DraftOK(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.Status = model.BroadcastStatusDraft
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)
	broadcastRepo.On("Delete", ctx, bc.ID).Return(nil)

	err := svc.Delete(ctx, teamID, bc.ID)

	require.NoError(t, err)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Send_HappyPath(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast() // has audience, from, html, subject
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)
	broadcastRepo.On("Update", ctx, mock.AnythingOfType("*model.Broadcast")).Return(nil)

	resp, err := svc.Send(ctx, teamID, bc.ID)

	require.NoError(t, err)
	assert.Equal(t, model.BroadcastStatusQueued, resp.Status)

	broadcastRepo.AssertExpectations(t)
}

func TestBroadcastService_Send_NotDraftFails(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.Status = model.BroadcastStatusSent
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Send(ctx, teamID, bc.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only draft")
}

func TestBroadcastService_Send_NoAudienceFails(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.AudienceID = nil
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Send(ctx, teamID, bc.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience")
}

func TestBroadcastService_Send_NoFromFails(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.FromAddress = nil
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Send(ctx, teamID, bc.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from address")
}

func TestBroadcastService_Send_NoSubjectFails(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	bc := testutil.NewTestBroadcast()
	bc.Subject = nil
	broadcastRepo.On("GetByID", ctx, bc.ID).Return(bc, nil)

	resp, err := svc.Send(ctx, teamID, bc.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subject")
}

func TestBroadcastService_Get_NotFound(t *testing.T) {
	broadcastRepo, asynqClient := newBroadcastTestDeps(t)
	svc := NewBroadcastService(broadcastRepo, asynqClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	broadcastRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	resp, err := svc.Get(ctx, teamID, badID)

	assert.Nil(t, resp)
	assert.Error(t, err)

	broadcastRepo.AssertExpectations(t)
}
