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

func TestBroadcastHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	reqBody := dto.CreateBroadcastRequest{Name: "Newsletter #1"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.BroadcastResponse{ID: uuid.New().String(), Name: "Newsletter #1", Status: "draft"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateBroadcastRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/broadcasts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/broadcasts", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	reqBody := dto.CreateBroadcastRequest{} // Missing name
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/broadcasts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/broadcasts", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestBroadcastHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	expected := &dto.PaginatedResponse[dto.BroadcastResponse]{
		Data:    []dto.BroadcastResponse{{ID: uuid.New().String(), Name: "Broadcast 1"}},
		Total:   1,
		Page:    1,
		PerPage: 20,
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.PaginationParams")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/broadcasts", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/broadcasts", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Get_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	broadcastID := uuid.New()
	expected := &dto.BroadcastResponse{ID: broadcastID.String(), Name: "My Broadcast"}
	mockSvc.On("Get", mock.Anything, testutil.TestTeamID, broadcastID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/broadcasts/"+broadcastID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", broadcastID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/broadcasts/{broadcastId}", h.Get) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	broadcastID := uuid.New()
	name := "Updated Broadcast"
	reqBody := dto.UpdateBroadcastRequest{Name: &name}
	body, _ := json.Marshal(reqBody)

	expected := &dto.BroadcastResponse{ID: broadcastID.String(), Name: "Updated Broadcast"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, broadcastID, mock.AnythingOfType("*dto.UpdateBroadcastRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/broadcasts/"+broadcastID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", broadcastID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/broadcasts/{broadcastId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	broadcastID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, broadcastID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/broadcasts/"+broadcastID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", broadcastID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/broadcasts/{broadcastId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Send_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	broadcastID := uuid.New()
	expected := &dto.BroadcastResponse{ID: broadcastID.String(), Name: "My Broadcast", Status: "queued"}
	mockSvc.On("Send", mock.Anything, testutil.TestTeamID, broadcastID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/broadcasts/"+broadcastID.String()+"/send", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", broadcastID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/broadcasts/{broadcastId}/send", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Send_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	broadcastID := uuid.New()
	mockSvc.On("Send", mock.Anything, testutil.TestTeamID, broadcastID).Return(nil, errors.New("no audience set"))

	req := httptest.NewRequest(http.MethodPost, "/broadcasts/"+broadcastID.String()+"/send", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", broadcastID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/broadcasts/{broadcastId}/send", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBroadcastHandler_Send_InvalidID(t *testing.T) {
	mockSvc := new(mockpkg.MockBroadcastService)
	h := NewBroadcastHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/broadcasts/bad-id/send", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "broadcastId", "bad-id")
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/broadcasts/{broadcastId}/send", h.Send) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
