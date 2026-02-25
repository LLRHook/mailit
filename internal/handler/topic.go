package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type TopicHandler struct {
	service service.TopicService
}

func NewTopicHandler(s service.TopicService) *TopicHandler {
	return &TopicHandler{service: s}
}

// Create handles POST /topics.
func (h *TopicHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateTopicRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp, err := h.service.Create(r.Context(), auth.TeamID, &req)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusCreated, resp)
}

// List handles GET /topics.
func (h *TopicHandler) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.List(r.Context(), auth.TeamID)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /topics/{topicId}.
func (h *TopicHandler) Update(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	topicID, err := uuid.Parse(chi.URLParam(r, "topicId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	var req dto.UpdateTopicRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, topicID, &req)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /topics/{topicId}.
func (h *TopicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	topicID, err := uuid.Parse(chi.URLParam(r, "topicId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, topicID); err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
