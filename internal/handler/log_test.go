package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/testutil"
	mockpkg "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestLogHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockLogService)
	h := NewLogHandler(mockSvc)

	expected := &dto.PaginatedResponse[model.Log]{
		Data: []model.Log{
			{
				ID:      uuid.New(),
				TeamID:  testutil.TestTeamID,
				Level:   "info",
				Message: "Email sent successfully",
			},
		},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, "", mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/logs", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestLogHandler_List_WithLevelFilter(t *testing.T) {
	mockSvc := new(mockpkg.MockLogService)
	h := NewLogHandler(mockSvc)

	expected := &dto.PaginatedResponse[model.Log]{
		Data: []model.Log{
			{
				ID:      uuid.New(),
				TeamID:  testutil.TestTeamID,
				Level:   "error",
				Message: "Failed to send email",
			},
		},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, "error", mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/logs?level=error", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/logs", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestLogHandler_List_Unauthorized(t *testing.T) {
	mockSvc := new(mockpkg.MockLogService)
	h := NewLogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/logs", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockLogService)
	h := NewLogHandler(mockSvc)

	mockSvc.On("List", mock.Anything, testutil.TestTeamID, "", mock.AnythingOfType("*dto.PaginationParams")).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/logs", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
