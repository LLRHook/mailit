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

func TestAudienceHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	reqBody := dto.CreateAudienceRequest{Name: "Newsletter Subscribers"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.AudienceResponse{ID: uuid.New().String(), Name: "Newsletter Subscribers"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateAudienceRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/audiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAudienceHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	reqBody := dto.CreateAudienceRequest{} // Missing name
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/audiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAudienceHandler_Create_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	reqBody := dto.CreateAudienceRequest{Name: "My Audience"}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateAudienceRequest")).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/audiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAudienceHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	expected := []dto.AudienceResponse{
		{ID: uuid.New().String(), Name: "Audience 1"},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/audiences", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/audiences", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAudienceHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	audienceID := uuid.New()
	expected := &dto.AudienceResponse{ID: audienceID.String(), Name: "My Audience"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, audienceID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/audiences/"+audienceID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/audiences/{audienceId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAudienceHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	audienceID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, audienceID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/audiences/"+audienceID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAudienceHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockAudienceService)
	h := NewAudienceHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/audiences/bad-id", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
