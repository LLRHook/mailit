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
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestAPIKeyService_Create_HappyPath(t *testing.T) {
	apiKeyRepo := new(tmock.MockAPIKeyRepository)
	svc := NewAPIKeyService(apiKeyRepo, "re_")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	apiKeyRepo.On("Create", ctx, mock.AnythingOfType("*model.APIKey")).Return(nil)

	req := &dto.CreateAPIKeyRequest{
		Name:       "My API Key",
		Permission: "full",
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "My API Key", resp.Name)
	assert.Equal(t, "full", resp.Permission)
	assert.NotEmpty(t, resp.KeyPrefix)

	apiKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_List_ReturnsKeysWithoutPlaintext(t *testing.T) {
	apiKeyRepo := new(tmock.MockAPIKeyRepository)
	svc := NewAPIKeyService(apiKeyRepo, "re_")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	key1 := *testutil.NewTestAPIKey()
	key2 := *testutil.NewTestAPIKey()
	key2.Name = "Second Key"
	apiKeyRepo.On("ListByTeamID", ctx, teamID).Return([]model.APIKey{key1, key2}, nil)

	resp, err := svc.List(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, resp, 2)
	// Token should be empty on list (only returned on create).
	assert.Empty(t, resp[0].Token)
	assert.Empty(t, resp[1].Token)
	assert.Equal(t, "Test Key", resp[0].Name)
	assert.Equal(t, "Second Key", resp[1].Name)

	apiKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_Delete_HappyPath(t *testing.T) {
	apiKeyRepo := new(tmock.MockAPIKeyRepository)
	svc := NewAPIKeyService(apiKeyRepo, "re_")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	key := *testutil.NewTestAPIKey()
	apiKeyRepo.On("ListByTeamID", ctx, teamID).Return([]model.APIKey{key}, nil)
	apiKeyRepo.On("Delete", ctx, key.ID).Return(nil)

	err := svc.Delete(ctx, teamID, key.ID)

	require.NoError(t, err)

	apiKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_Delete_NotFound(t *testing.T) {
	apiKeyRepo := new(tmock.MockAPIKeyRepository)
	svc := NewAPIKeyService(apiKeyRepo, "re_")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	key := *testutil.NewTestAPIKey()
	apiKeyRepo.On("ListByTeamID", ctx, teamID).Return([]model.APIKey{key}, nil)

	nonExistentID := uuid.New()
	err := svc.Delete(ctx, teamID, nonExistentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	apiKeyRepo.AssertExpectations(t)
}
