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

func TestEmailHandler_Send_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	html := "<p>Hello</p>"
	reqBody := dto.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		HTML:    &html,
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.SendEmailResponse{ID: uuid.New().String()}
	mockSvc.On("Send", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.SendEmailRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/emails", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.SendEmailResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	mockSvc.AssertExpectations(t)
}

func TestEmailHandler_Send_Unauthorized(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/emails", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestEmailHandler_Send_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	// Missing required fields
	reqBody := map[string]string{"from": "not-email"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/emails", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestEmailHandler_Send_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	html := "<p>Hello</p>"
	reqBody := dto.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		HTML:    &html,
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Send", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.SendEmailRequest")).Return(nil, errors.New("send failed"))

	req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/emails", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestEmailHandler_Send_WithIdempotencyKey(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	html := "<p>Hello</p>"
	reqBody := dto.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		HTML:    &html,
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.SendEmailResponse{ID: uuid.New().String()}
	mockSvc.On("Send", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.SendEmailRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "unique-key-123")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/emails", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestEmailHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	emailID := uuid.New()
	expected := &dto.EmailResponse{
		ID:      emailID.String(),
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Status:  "sent",
	}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, emailID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/emails/"+emailID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "emailId", emailID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/emails/{emailId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestEmailHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/emails/not-a-uuid", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "emailId", "not-a-uuid")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/emails/{emailId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEmailHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	expected := &dto.PaginatedResponse[dto.EmailResponse]{
		Data:       []dto.EmailResponse{{ID: uuid.New().String(), From: "test@example.com"}},
		Total:      1,
		Page:       1,
		PerPage:    20,
		TotalPages: 1,
		HasMore:    false,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/emails?page=1&per_page=20", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/emails", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestEmailHandler_List_Unauthorized(t *testing.T) {
	mockSvc := new(mockpkg.MockEmailService)
	h := NewEmailHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/emails", nil)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/emails", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
