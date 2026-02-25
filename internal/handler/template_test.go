package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/testutil"
	mockpkg "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestTemplateHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	reqBody := dto.CreateTemplateRequest{Name: "Welcome Email"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.TemplateResponse{ID: uuid.New().String(), Name: "Welcome Email"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateTemplateRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/templates", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	// Empty name
	reqBody := dto.CreateTemplateRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/templates", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestTemplateHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	expected := &dto.PaginatedResponse[dto.TemplateResponse]{
		Data:    []dto.TemplateResponse{{ID: uuid.New().String(), Name: "Template 1"}},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/templates", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	templateID := uuid.New()
	expected := &dto.TemplateResponse{ID: templateID.String(), Name: "My Template"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, templateID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/templates/"+templateID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "templateId", templateID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/templates/{templateId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	templateID := uuid.New()
	name := "Updated Name"
	reqBody := dto.UpdateTemplateRequest{Name: &name}
	body, _ := json.Marshal(reqBody)

	expected := &dto.TemplateResponse{ID: templateID.String(), Name: "Updated Name"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, templateID, mock.AnythingOfType("*dto.UpdateTemplateRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/templates/"+templateID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "templateId", templateID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/templates/{templateId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	templateID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, templateID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/templates/"+templateID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "templateId", templateID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/templates/{templateId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Publish_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	templateID := uuid.New()
	expected := &dto.TemplateResponse{ID: templateID.String(), Name: "Published Template"}
	mockSvc.On("Publish", mock.Anything, testutil.TestTeamID, templateID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/templates/"+templateID.String()+"/publish", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "templateId", templateID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/templates/{templateId}/publish", h.Publish) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTemplateHandler_Publish_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockTemplateService)
	h := NewTemplateHandler(mockSvc)

	templateID := uuid.New()
	mockSvc.On("Publish", mock.Anything, testutil.TestTeamID, templateID).Return(nil, errors.New("no versions to publish"))

	req := httptest.NewRequest(http.MethodPost, "/templates/"+templateID.String()+"/publish", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "templateId", templateID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/templates/{templateId}/publish", h.Publish) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
