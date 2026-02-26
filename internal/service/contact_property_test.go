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

func newTestContactProperty() *model.ContactProperty {
	return &model.ContactProperty{
		ID:        uuid.New(),
		TeamID:    testutil.TestTeamID,
		Name:      "company",
		Label:     "Company Name",
		Type:      "string",
		CreatedAt: testutil.FixedTime,
		UpdatedAt: testutil.FixedTime,
	}
}

func TestContactPropertyService_Create(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	propRepo.On("Create", ctx, mock.AnythingOfType("*model.ContactProperty")).Return(nil)

	req := &dto.CreateContactPropertyRequest{
		Name:  "company",
		Label: "Company Name",
		Type:  "string",
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "company", resp.Name)
	assert.Equal(t, "Company Name", resp.Label)
	assert.Equal(t, "string", resp.Type)

	propRepo.AssertExpectations(t)
}

func TestContactPropertyService_List(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	prop := *newTestContactProperty()
	propRepo.On("ListByTeamID", ctx, teamID).Return([]model.ContactProperty{prop}, nil)

	resp, err := svc.List(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "company", resp.Data[0].Name)

	propRepo.AssertExpectations(t)
}

func TestContactPropertyService_Update_HappyPath(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	prop := newTestContactProperty()
	propRepo.On("GetByID", ctx, prop.ID).Return(prop, nil)
	propRepo.On("Update", ctx, mock.AnythingOfType("*model.ContactProperty")).Return(nil)

	req := &dto.UpdateContactPropertyRequest{
		Label: testutil.StringPtr("Updated Label"),
	}

	resp, err := svc.Update(ctx, teamID, prop.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated Label", resp.Label)

	propRepo.AssertExpectations(t)
}

func TestContactPropertyService_Update_WrongTeam(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	prop := newTestContactProperty()
	propRepo.On("GetByID", ctx, prop.ID).Return(prop, nil)

	req := &dto.UpdateContactPropertyRequest{
		Label: testutil.StringPtr("Updated"),
	}

	resp, err := svc.Update(ctx, wrongTeamID, prop.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	propRepo.AssertExpectations(t)
}

func TestContactPropertyService_Delete_HappyPath(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	prop := newTestContactProperty()
	propRepo.On("GetByID", ctx, prop.ID).Return(prop, nil)
	propRepo.On("Delete", ctx, prop.ID).Return(nil)

	err := svc.Delete(ctx, teamID, prop.ID)

	require.NoError(t, err)

	propRepo.AssertExpectations(t)
}

func TestContactPropertyService_Delete_WrongTeam(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	prop := newTestContactProperty()
	propRepo.On("GetByID", ctx, prop.ID).Return(prop, nil)

	err := svc.Delete(ctx, wrongTeamID, prop.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	propRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestContactPropertyService_Delete_NotFound(t *testing.T) {
	propRepo := new(tmock.MockContactPropertyRepository)
	svc := NewContactPropertyService(propRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	propRepo.On("GetByID", ctx, badID).Return(nil, postgres.ErrNotFound)

	err := svc.Delete(ctx, teamID, badID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	propRepo.AssertExpectations(t)
}
