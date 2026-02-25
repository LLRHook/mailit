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

func TestInboundEmailHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	expected := &dto.PaginatedResponse[model.InboundEmail]{
		Data: []model.InboundEmail{
			{
				ID:          uuid.New(),
				TeamID:      testutil.TestTeamID,
				FromAddress: "sender@external.com",
				ToAddresses: []string{"inbox@example.com"},
			},
		},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInboundEmailHandler_List_Unauthorized(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails", nil)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestInboundEmailHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInboundEmailHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	emailID := uuid.New()
	expected := &model.InboundEmail{
		ID:          emailID,
		TeamID:      testutil.TestTeamID,
		FromAddress: "sender@external.com",
		ToAddresses: []string{"inbox@example.com"},
	}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, emailID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails/"+emailID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "emailId", emailID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails/{emailId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInboundEmailHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails/bad-id", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "emailId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails/{emailId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestInboundEmailHandler_Get_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockInboundEmailService)
	h := NewInboundEmailHandler(mockSvc)

	emailID := uuid.New()
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, emailID).Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/inbound/emails/"+emailID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "emailId", emailID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/inbound/emails/{emailId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
