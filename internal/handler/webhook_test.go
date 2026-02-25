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

func TestWebhookHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	reqBody := dto.CreateWebhookRequest{
		URL:    "https://example.com/webhook",
		Events: []string{"email.sent", "email.bounced"},
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.WebhookResponse{
		ID:     uuid.New().String(),
		URL:    "https://example.com/webhook",
		Events: []string{"email.sent", "email.bounced"},
		Active: true,
	}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateWebhookRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/webhooks", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestWebhookHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	// Missing URL and events
	reqBody := dto.CreateWebhookRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/webhooks", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestWebhookHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	expected := []dto.WebhookResponse{
		{ID: uuid.New().String(), URL: "https://example.com/webhook", Events: []string{"email.sent"}, Active: true},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/webhooks", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestWebhookHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	webhookID := uuid.New()
	expected := &dto.WebhookResponse{ID: webhookID.String(), URL: "https://example.com/webhook"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, webhookID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "webhookId", webhookID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/webhooks/{webhookId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestWebhookHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	webhookID := uuid.New()
	active := false
	reqBody := dto.UpdateWebhookRequest{Active: &active}
	body, _ := json.Marshal(reqBody)

	expected := &dto.WebhookResponse{ID: webhookID.String(), Active: false}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, webhookID, mock.AnythingOfType("*dto.UpdateWebhookRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/webhooks/"+webhookID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "webhookId", webhookID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/webhooks/{webhookId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestWebhookHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	webhookID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, webhookID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+webhookID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "webhookId", webhookID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/webhooks/{webhookId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestWebhookHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockWebhookService)
	h := NewWebhookHandler(mockSvc)

	webhookID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, webhookID).Return(errors.New("not found"))

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+webhookID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "webhookId", webhookID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/webhooks/{webhookId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
