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

func TestSegmentHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	reqBody := dto.CreateSegmentRequest{
		Name:       "Active Users",
		Conditions: map[string]interface{}{"field": "status", "value": "active"},
	}
	body, _ := json.Marshal(reqBody)

	expected := &dto.SegmentResponse{ID: uuid.New().String(), Name: "Active Users"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, audienceID, mock.AnythingOfType("*dto.CreateSegmentRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/audiences/"+audienceID.String()+"/segments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences/{audienceId}/segments", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSegmentHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	// Missing name and conditions
	reqBody := dto.CreateSegmentRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/audiences/"+audienceID.String()+"/segments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/audiences/{audienceId}/segments", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSegmentHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	expected := []dto.SegmentResponse{
		{ID: uuid.New().String(), Name: "Segment 1"},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID, audienceID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/audiences/"+audienceID.String()+"/segments", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "audienceId", audienceID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/audiences/{audienceId}/segments", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSegmentHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	segmentID := uuid.New()
	name := "Updated Segment"
	reqBody := dto.UpdateSegmentRequest{Name: &name}
	body, _ := json.Marshal(reqBody)

	expected := &dto.SegmentResponse{ID: segmentID.String(), Name: "Updated Segment"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, audienceID, segmentID, mock.AnythingOfType("*dto.UpdateSegmentRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/audiences/"+audienceID.String()+"/segments/"+segmentID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "segmentId": segmentID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/audiences/{audienceId}/segments/{segmentId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSegmentHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	segmentID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, audienceID, segmentID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/audiences/"+audienceID.String()+"/segments/"+segmentID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "segmentId": segmentID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}/segments/{segmentId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSegmentHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockSegmentService)
	h := NewSegmentHandler(mockSvc)

	audienceID := uuid.New()
	segmentID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, audienceID, segmentID).Return(errors.New("not found"))

	req := httptest.NewRequest(http.MethodDelete, "/audiences/"+audienceID.String()+"/segments/"+segmentID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParams(req, map[string]string{"audienceId": audienceID.String(), "segmentId": segmentID.String()})
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/audiences/{audienceId}/segments/{segmentId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
