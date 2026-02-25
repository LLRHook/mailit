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

func TestDomainHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	reqBody := dto.CreateDomainRequest{Name: "example.com"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.DomainResponse{ID: uuid.New().String(), Name: "example.com", Status: "pending"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateDomainRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/domains", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/domains", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	// Empty name
	reqBody := dto.CreateDomainRequest{Name: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/domains", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/domains", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestDomainHandler_Create_Unauthorized(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/domains", bytes.NewReader([]byte(`{"name":"example.com"}`)))
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/domains", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDomainHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	expected := &dto.PaginatedResponse[dto.DomainResponse]{
		Data:    []dto.DomainResponse{{ID: uuid.New().String(), Name: "example.com"}},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/domains", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	expected := &dto.DomainResponse{ID: domainID.String(), Name: "example.com", Status: "verified"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, domainID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/domains/"+domainID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/domains/{domainId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/domains/bad-id", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/domains/{domainId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDomainHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	openTracking := true
	reqBody := dto.UpdateDomainRequest{OpenTracking: &openTracking}
	body, _ := json.Marshal(reqBody)

	expected := &dto.DomainResponse{ID: domainID.String(), Name: "example.com"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, domainID, mock.AnythingOfType("*dto.UpdateDomainRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/domains/"+domainID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/domains/{domainId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, domainID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/domains/"+domainID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/domains/{domainId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, domainID).Return(errors.New("not found"))

	req := httptest.NewRequest(http.MethodDelete, "/domains/"+domainID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/domains/{domainId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Verify_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	expected := &dto.DomainResponse{ID: domainID.String(), Name: "example.com", Status: "verified"}
	mockSvc.On("Verify", mock.Anything, testutil.TestTeamID, domainID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/domains/"+domainID.String()+"/verify", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/domains/{domainId}/verify", h.Verify) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestDomainHandler_Verify_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockDomainService)
	h := NewDomainHandler(mockSvc)

	domainID := uuid.New()
	mockSvc.On("Verify", mock.Anything, testutil.TestTeamID, domainID).Return(nil, errors.New("DNS check failed"))

	req := httptest.NewRequest(http.MethodPost, "/domains/"+domainID.String()+"/verify", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "domainId", domainID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/domains/{domainId}/verify", h.Verify) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
