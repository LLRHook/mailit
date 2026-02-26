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

func TestTopicHandler_Create_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	reqBody := dto.CreateTopicRequest{Name: "Product Updates"}
	body, _ := json.Marshal(reqBody)

	expected := &dto.TopicResponse{ID: uuid.New().String(), Name: "Product Updates"}
	mockSvc.On("Create", mock.Anything, testutil.TestTeamID, mock.AnythingOfType("*dto.CreateTopicRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/topics", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTopicHandler_Create_ValidationError(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	reqBody := dto.CreateTopicRequest{} // Missing name
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Post("/topics", h.Create) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestTopicHandler_List_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	expected := &dto.ListResponse[dto.TopicResponse]{
		Data: []dto.TopicResponse{
			{ID: uuid.New().String(), Name: "Topic 1"},
		},
	}
	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/topics", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTopicHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	mockSvc.On("List", mock.Anything, testutil.TestTeamID).Return((*dto.ListResponse[dto.TopicResponse])(nil), errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Get("/topics", h.List) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTopicHandler_Update_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	topicID := uuid.New()
	name := "Updated Topic"
	reqBody := dto.UpdateTopicRequest{Name: &name}
	body, _ := json.Marshal(reqBody)

	expected := &dto.TopicResponse{ID: topicID.String(), Name: "Updated Topic"}
	mockSvc.On("Update", mock.Anything, testutil.TestTeamID, topicID, mock.AnythingOfType("*dto.UpdateTopicRequest")).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPatch, "/topics/"+topicID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "topicId", topicID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Patch("/topics/{topicId}", h.Update) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTopicHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	topicID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, topicID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/topics/"+topicID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "topicId", topicID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/topics/{topicId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestTopicHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(mockpkg.MockTopicService)
	h := NewTopicHandler(mockSvc)

	topicID := uuid.New()
	mockSvc.On("Delete", mock.Anything, testutil.TestTeamID, topicID).Return(errors.New("not found"))

	req := httptest.NewRequest(http.MethodDelete, "/topics/"+topicID.String(), nil)
	req = testutil.AuthenticatedRequest(req, testutil.TestTeamID, testutil.TestUserID)
	req = testutil.WithURLParam(req, "topicId", topicID.String())
	rec := httptest.NewRecorder()

	r := testutil.SetupRouter(func(r chi.Router) { r.Delete("/topics/{topicId}", h.Delete) })
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
