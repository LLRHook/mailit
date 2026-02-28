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

func TestWebhookService_Create_HappyPath(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	webhookRepo.On("Create", ctx, mock.AnythingOfType("*model.Webhook")).Return(nil)

	req := &dto.CreateWebhookRequest{
		URL:    "https://example.com/webhook",
		Events: []string{"email.sent", "email.bounced"},
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "https://example.com/webhook", resp.URL)
	assert.Equal(t, []string{"email.sent", "email.bounced"}, resp.Events)
	assert.True(t, resp.Active)

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_List_ReturnsWebhooks(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	wh := *testutil.NewTestWebhook()
	webhookRepo.On("ListByTeamID", ctx, teamID).Return([]model.Webhook{wh}, nil)

	resp, err := svc.List(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, wh.URL, resp.Data[0].URL)

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_Get_HappyPath(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	wh := testutil.NewTestWebhook()
	webhookRepo.On("GetByID", ctx, wh.ID).Return(wh, nil)

	resp, err := svc.Get(ctx, teamID, wh.ID)

	require.NoError(t, err)
	assert.Equal(t, wh.ID.String(), resp.ID)
	assert.Equal(t, wh.URL, resp.URL)

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_Get_WrongTeam(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	wh := testutil.NewTestWebhook()
	webhookRepo.On("GetByID", ctx, wh.ID).Return(wh, nil)

	resp, err := svc.Get(ctx, wrongTeamID, wh.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_Update_Fields(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	wh := testutil.NewTestWebhook()
	webhookRepo.On("GetByID", ctx, wh.ID).Return(wh, nil)
	webhookRepo.On("Update", ctx, mock.AnythingOfType("*model.Webhook")).Return(nil)

	newURL := "https://updated.com/webhook"
	req := &dto.UpdateWebhookRequest{
		URL:    &newURL,
		Active: testutil.BoolPtr(false),
	}

	resp, err := svc.Update(ctx, teamID, wh.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "https://updated.com/webhook", resp.URL)
	assert.False(t, resp.Active)

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_Delete_HappyPath(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	wh := testutil.NewTestWebhook()
	webhookRepo.On("GetByID", ctx, wh.ID).Return(wh, nil)
	webhookRepo.On("Delete", ctx, wh.ID).Return(nil)

	err := svc.Delete(ctx, teamID, wh.ID)

	require.NoError(t, err)

	webhookRepo.AssertExpectations(t)
}

func TestWebhookService_Delete_WrongTeam(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	wh := testutil.NewTestWebhook()
	webhookRepo.On("GetByID", ctx, wh.ID).Return(wh, nil)

	err := svc.Delete(ctx, wrongTeamID, wh.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Delete should not have been called because the webhook belongs to a different team.
	webhookRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestWebhookService_Get_NotFound(t *testing.T) {
	webhookRepo := new(tmock.MockWebhookRepository)
	svc := NewWebhookService(webhookRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	webhookRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	resp, err := svc.Get(ctx, teamID, badID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	webhookRepo.AssertExpectations(t)
}
