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

func TestAPIKeyHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	reqBody := dto.CreateAPIKeyRequest{Name: "My API Key", Permission: "full"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.APIKeyResponse{
		ID:         uuid.New().String(),
		Name:       "My API Key",
		Token:      "re_1234abcd_secret",
		Permission: "full",
	}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateAPIKeyRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/api-keys", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp dto.APIKeyResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	mockSvc.AssertExpectations(t)
}

func TestAPIKeyHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	// Missing required name
	reqBody := dto.CreateAPIKeyRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/api-keys", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAPIKeyHandler_Create_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	reqBody := dto.CreateAPIKeyRequest{Name: "My API Key"}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateAPIKeyRequest")).Return(nil, errors.New("limit reached"))

	req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/api-keys", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAPIKeyHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	expected := &dto.ListResponse[dto.APIKeyResponse]{
		Data: []dto.APIKeyResponse{
			{ID: uuid.New().String(), Name: "Key 1", Permission: "full"},
		},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/api-keys", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAPIKeyHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	apiKeyID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, apiKeyID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+apiKeyID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "apiKeyId", apiKeyID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/api-keys/{apiKeyId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAPIKeyHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockAPIKeyService)
	h := NewAPIKeyHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/bad-id", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "apiKeyId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/api-keys/{apiKeyId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
