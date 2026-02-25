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

func TestContactService_Create_HappyPath(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("GetByAudienceAndEmail", ctx, aud.ID, "john@example.com").Return(nil, postgres.ErrNotFound)
	contactRepo.On("Create", ctx, mock.AnythingOfType("*model.Contact")).Return(nil)

	req := &dto.CreateContactRequest{
		Email:     "john@example.com",
		FirstName: testutil.StringPtr("John"),
		LastName:  testutil.StringPtr("Doe"),
	}

	resp, err := svc.Create(ctx, teamID, aud.ID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "john@example.com", resp.Email)
	assert.Equal(t, testutil.StringPtr("John"), resp.FirstName)

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestContactService_Create_DuplicateEmail(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	existingContact := testutil.NewTestContact(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("GetByAudienceAndEmail", ctx, aud.ID, "john@example.com").Return(existingContact, nil)

	req := &dto.CreateContactRequest{
		Email: "john@example.com",
	}

	resp, err := svc.Create(ctx, teamID, aud.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestContactService_Create_AudienceOwnershipCheck(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	aud := testutil.NewTestAudience() // belongs to testutil.TestTeamID
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)

	req := &dto.CreateContactRequest{
		Email: "john@example.com",
	}

	resp, err := svc.Create(ctx, wrongTeamID, aud.ID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience not found")

	audienceRepo.AssertExpectations(t)
}

func TestContactService_List(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	contact := *testutil.NewTestContact(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("List", ctx, aud.ID, 20, 0).Return([]model.Contact{contact}, 1, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, aud.ID, params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestContactService_Get_HappyPath(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	contact := testutil.NewTestContact(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("GetByID", ctx, contact.ID).Return(contact, nil)

	resp, err := svc.Get(ctx, teamID, aud.ID, contact.ID)

	require.NoError(t, err)
	assert.Equal(t, contact.ID.String(), resp.ID)

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestContactService_Update(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	contact := testutil.NewTestContact(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("GetByID", ctx, contact.ID).Return(contact, nil)
	contactRepo.On("Update", ctx, mock.AnythingOfType("*model.Contact")).Return(nil)

	req := &dto.UpdateContactRequest{
		FirstName:    testutil.StringPtr("Jane"),
		Unsubscribed: testutil.BoolPtr(true),
	}

	resp, err := svc.Update(ctx, teamID, aud.ID, contact.ID, req)

	require.NoError(t, err)
	assert.Equal(t, testutil.StringPtr("Jane"), resp.FirstName)
	assert.True(t, resp.Unsubscribed)

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}

func TestContactService_Delete(t *testing.T) {
	contactRepo := new(tmock.MockContactRepository)
	audienceRepo := new(tmock.MockAudienceRepository)
	svc := NewContactService(contactRepo, audienceRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	aud := testutil.NewTestAudience()
	contact := testutil.NewTestContact(aud.ID)
	audienceRepo.On("GetByID", ctx, aud.ID).Return(aud, nil)
	contactRepo.On("GetByID", ctx, contact.ID).Return(contact, nil)
	contactRepo.On("Delete", ctx, contact.ID).Return(nil)

	err := svc.Delete(ctx, teamID, aud.ID, contact.ID)

	require.NoError(t, err)

	contactRepo.AssertExpectations(t)
	audienceRepo.AssertExpectations(t)
}
