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

func TestContactHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	reqBody := dto.CreateContactRequest{Email: "john@example.com"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.ContactResponse{ID: uuid.New().String(), Email: "john@example.com"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, audienceID, mock.AnythingOfType("*dto.CreateContactRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/audiences/"+audienceID.String()+"/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences/{audienceId}/contacts", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	// Invalid email
	reqBody := dto.CreateContactRequest{Email: "not-email"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/audiences/"+audienceID.String()+"/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences/{audienceId}/contacts", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestContactHandler_Create_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	reqBody := dto.CreateContactRequest{Email: "john@example.com"}
	body, _ := json.Marshal(reqBody)

	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, audienceID, mock.AnythingOfType("*dto.CreateContactRequest")).Return(nil, errors.New("duplicate email"))

	req := httptest.NewRequest(http.MethodPost, "/audiences/"+audienceID.String()+"/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences/{audienceId}/contacts", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	expected := &dto.PaginatedResponse[dto.ContactResponse]{
		Data:    []dto.ContactResponse{{ID: uuid.New().String(), Email: "john@example.com"}},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, audienceID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/audiences/"+audienceID.String()+"/contacts", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/audiences/{audienceId}/contacts", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	contactID := uuid.New()
	expected := &dto.ContactResponse{ID: contactID.String(), Email: "john@example.com"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, audienceID, contactID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/audiences/"+audienceID.String()+"/contacts/"+contactID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "contactId": contactID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/audiences/{audienceId}/contacts/{contactId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	contactID := uuid.New()
	first := "Jane"
	reqBody := dto.UpdateContactRequest{FirstName: &first}
	body, _ := json.Marshal(reqBody)

	expected := &dto.ContactResponse{ID: contactID.String(), Email: "john@example.com", FirstName: &first}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, audienceID, contactID, mock.AnythingOfType("*dto.UpdateContactRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/audiences/"+audienceID.String()+"/contacts/"+contactID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "contactId": contactID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/audiences/{audienceId}/contacts/{contactId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	audienceID := uuid.New()
	contactID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, audienceID, contactID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/audiences/"+audienceID.String()+"/contacts/"+contactID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "contactId": contactID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}/contacts/{contactId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestContactHandler_Delete_InvalidAudienceID(t *testing.T) {
	mockSvc := new(mockpkg.MockContactService)
	h := NewContactHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/audiences/bad/contacts/"+uuid.New().String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": "bad-id", "contactId": uuid.New().String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}/contacts/{contactId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
