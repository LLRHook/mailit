package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/dto"
	mockpkg "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		TeamName: "Test Team",
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.AuthResponse{Token: "jwt-token"}
	expected.User.ID = "user-id"
	expected.User.Email = "test@example.com"
	expected.User.Name = "Test User"

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/register", h.Register)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp dto.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "jwt-token", resp.Token)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	// Missing required fields
	reqBody := map[string]string{"email": "not-an-email"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/register", h.Register)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertNotCalled(t, "Register")
}

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		TeamName: "Test Team",
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).Return(nil, errors.New("user already exists"))

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/register", h.Register)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/register", h.Register)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.AuthResponse{Token: "jwt-token"}
	expected.User.ID = "user-id"
	expected.User.Email = "test@example.com"
	expected.User.Name = "Test User"

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/login", h.Login)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "jwt-token", resp.Token)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	// Password too short
	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "short",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/login", h.Login)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Login_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockAuthService)
	h := NewAuthHandler(mockSvc)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(nil, errors.New("invalid credentials"))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/auth/login", h.Login)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
