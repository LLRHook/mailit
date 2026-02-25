package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func newTestInboundEmail() model.InboundEmail {
	subject := "Inbound Test"
	return model.InboundEmail{
		ID:          uuid.New(),
		TeamID:      testutil.TestTeamID,
		FromAddress: "external@example.com",
		ToAddresses: []string{"inbox@example.com"},
		Subject:     &subject,
		Headers:     model.JSONMap{},
		Attachments: model.JSONArray{},
		Processed:   false,
		CreatedAt:   testutil.FixedTime,
	}
}

func TestInboundEmailService_List_Paginated(t *testing.T) {
	inboundRepo := new(tmock.MockInboundEmailRepository)
	svc := NewInboundEmailService(inboundRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email := newTestInboundEmail()
	inboundRepo.On("List", ctx, teamID, 20, 0).Return([]model.InboundEmail{email}, 1, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, email.ID, resp.Data[0].ID)

	inboundRepo.AssertExpectations(t)
}

func TestInboundEmailService_Get_HappyPath(t *testing.T) {
	inboundRepo := new(tmock.MockInboundEmailRepository)
	svc := NewInboundEmailService(inboundRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email := newTestInboundEmail()
	inboundRepo.On("GetByTeamAndID", ctx, teamID, email.ID).Return(&email, nil)

	resp, err := svc.Get(ctx, teamID, email.ID)

	require.NoError(t, err)
	assert.Equal(t, email.ID, resp.ID)
	assert.Equal(t, "external@example.com", resp.FromAddress)

	inboundRepo.AssertExpectations(t)
}

func TestInboundEmailService_Get_NotFound(t *testing.T) {
	inboundRepo := new(tmock.MockInboundEmailRepository)
	svc := NewInboundEmailService(inboundRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	inboundRepo.On("GetByTeamAndID", ctx, teamID, badID).Return(nil, assert.AnError)

	resp, err := svc.Get(ctx, teamID, badID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	inboundRepo.AssertExpectations(t)
}

func TestInboundEmailService_List_EmptyResult(t *testing.T) {
	inboundRepo := new(tmock.MockInboundEmailRepository)
	svc := NewInboundEmailService(inboundRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	inboundRepo.On("List", ctx, teamID, 20, 0).Return([]model.InboundEmail{}, 0, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Data)
	assert.False(t, resp.HasMore)

	inboundRepo.AssertExpectations(t)
}
