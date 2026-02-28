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

func TestTemplateService_Create_HappyPath(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	templateRepo.On("Create", ctx, mock.AnythingOfType("*model.Template")).Return(nil)
	versionRepo.On("Create", ctx, mock.AnythingOfType("*model.TemplateVersion")).Return(nil)

	req := &dto.CreateTemplateRequest{
		Name:    "Welcome Email",
		Subject: testutil.StringPtr("Welcome!"),
		HTML:    testutil.StringPtr("<p>Welcome</p>"),
	}

	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.Equal(t, "Welcome Email", resp.Name)
	assert.NotEmpty(t, resp.ID)

	templateRepo.AssertExpectations(t)
	versionRepo.AssertExpectations(t)
}

func TestTemplateService_List_Paginated(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := *testutil.NewTestTemplate()
	templateRepo.On("List", ctx, teamID, 20, 0).Return([]model.Template{tmpl}, 1, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)

	templateRepo.AssertExpectations(t)
}

func TestTemplateService_Get_HappyPath(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := testutil.NewTestTemplate()
	version := testutil.NewTestTemplateVersion(tmpl.ID)
	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)
	versionRepo.On("GetLatestByTemplateID", ctx, tmpl.ID).Return(version, nil)

	resp, err := svc.Get(ctx, teamID, tmpl.ID)

	require.NoError(t, err)
	assert.Equal(t, tmpl.ID.String(), resp.ID)
	assert.Equal(t, version.Subject, resp.Subject)
	assert.Equal(t, version.HTMLBody, resp.HTML)

	templateRepo.AssertExpectations(t)
	versionRepo.AssertExpectations(t)
}

func TestTemplateService_Get_WrongTeam(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	tmpl := testutil.NewTestTemplate()
	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)

	resp, err := svc.Get(ctx, wrongTeamID, tmpl.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	templateRepo.AssertExpectations(t)
	// versionRepo should not have been called since team check failed first.
}

func TestTemplateService_Update_WithContentCreatesNewVersion(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := testutil.NewTestTemplate()
	version := testutil.NewTestTemplateVersion(tmpl.ID)

	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)
	templateRepo.On("Update", ctx, mock.AnythingOfType("*model.Template")).Return(nil)
	versionRepo.On("GetLatestByTemplateID", ctx, tmpl.ID).Return(version, nil)
	versionRepo.On("Create", ctx, mock.AnythingOfType("*model.TemplateVersion")).Return(nil)

	req := &dto.UpdateTemplateRequest{
		HTML: testutil.StringPtr("<p>Updated</p>"),
	}

	resp, err := svc.Update(ctx, teamID, tmpl.ID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	templateRepo.AssertExpectations(t)
	versionRepo.AssertExpectations(t)
}

func TestTemplateService_Delete_HappyPath(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := testutil.NewTestTemplate()
	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)
	templateRepo.On("Delete", ctx, tmpl.ID).Return(nil)

	err := svc.Delete(ctx, teamID, tmpl.ID)

	require.NoError(t, err)

	templateRepo.AssertExpectations(t)
}

func TestTemplateService_Publish_HappyPath(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := testutil.NewTestTemplate()
	version := testutil.NewTestTemplateVersion(tmpl.ID)
	version.Published = false

	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)
	versionRepo.On("GetLatestByTemplateID", ctx, tmpl.ID).Return(version, nil)
	versionRepo.On("Publish", ctx, version.ID).Return(nil)
	templateRepo.On("Update", ctx, mock.AnythingOfType("*model.Template")).Return(nil)

	resp, err := svc.Publish(ctx, teamID, tmpl.ID)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	templateRepo.AssertExpectations(t)
	versionRepo.AssertExpectations(t)
}

func TestTemplateService_Publish_AlreadyPublished(t *testing.T) {
	templateRepo := new(tmock.MockTemplateRepository)
	versionRepo := new(tmock.MockTemplateVersionRepository)
	svc := NewTemplateService(templateRepo, versionRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	tmpl := testutil.NewTestTemplate()
	version := testutil.NewTestTemplateVersion(tmpl.ID)
	version.Published = true

	templateRepo.On("GetByID", ctx, tmpl.ID).Return(tmpl, nil)
	versionRepo.On("GetLatestByTemplateID", ctx, tmpl.ID).Return(version, nil)

	resp, err := svc.Publish(ctx, teamID, tmpl.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already published")

	templateRepo.AssertExpectations(t)
	versionRepo.AssertExpectations(t)
}
