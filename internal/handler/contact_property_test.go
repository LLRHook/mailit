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

func TestContactPropertyHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	reqBody := dto.CreateContactPropertyRequest{
		Name:  "company",
		Label: "Company",
		Type:  "string",
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.ContactPropertyResponse{
		ID:    uuid.New().String(),
		Name:  "company",
		Label: "Company",
		Type:  "string",
	}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateContactPropertyRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/contact-properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/contact-properties", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactPropertyHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	// Missing required fields
	reqBody := dto.CreateContactPropertyRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/contact-properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/contact-properties", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestContactPropertyHandler_Create_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	reqBody := dto.CreateContactPropertyRequest{
		Name:  "company",
		Label: "Company",
		Type:  "string",
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateContactPropertyRequest")).Return(nil, errors.New("duplicate"))

	req := httptest.NewRequest(http.MethodPost, "/contact-properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/contact-properties", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactPropertyHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	expected := &dto.ListResponse[dto.ContactPropertyResponse]{
		Data: []dto.ContactPropertyResponse{
			{ID: uuid.New().String(), Name: "company", Label: "Company", Type: "string"},
		},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/contact-properties", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/contact-properties", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactPropertyHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	propertyID := uuid.New()
	label := "Updated Label"
	reqBody := dto.UpdateContactPropertyRequest{Label: &label}
	body, _ := json.Marshal(reqBody)

	expected := &dto.ContactPropertyResponse{ID: propertyID.String(), Name: "company", Label: "Updated Label", Type: "string"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, propertyID, mock.AnythingOfType("*dto.UpdateContactPropertyRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/contact-properties/"+propertyID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "propertyId", propertyID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/contact-properties/{propertyId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactPropertyHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	propertyID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, propertyID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/contact-properties/"+propertyID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "propertyId", propertyID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/contact-properties/{propertyId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactPropertyHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockContactPropertyService)
	h := NewContactPropertyHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/contact-properties/bad-id", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "propertyId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/contact-properties/{propertyId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
