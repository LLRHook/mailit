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

type SegmentHandler struct {
	service service.SegmentService
}

func NewSegmentHandler(s service.SegmentService) *SegmentHandler {
	return &SegmentHandler{service: s}
}

// Create handles POST /audiences/{audienceId}/segments.
func (h *SegmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	audienceID, err := uuid.Parse(chi.URLParam(r, "audienceId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid audience id")
		return
	}

	var req dto.CreateSegmentRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp, err := h.service.Create(r.Context(), auth.TeamID, audienceID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusCreated, resp)
}

// List handles GET /audiences/{audienceId}/segments.
func (h *SegmentHandler) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	audienceID, err := uuid.Parse(chi.URLParam(r, "audienceId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid audience id")
		return
	}

	resp, err := h.service.List(r.Context(), auth.TeamID, audienceID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /audiences/{audienceId}/segments/{segmentId}.
func (h *SegmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	audienceID, err := uuid.Parse(chi.URLParam(r, "audienceId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid audience id")
		return
	}

	segmentID, err := uuid.Parse(chi.URLParam(r, "segmentId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid segment id")
		return
	}

	var req dto.UpdateSegmentRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, audienceID, segmentID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /audiences/{audienceId}/segments/{segmentId}.
func (h *SegmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	audienceID, err := uuid.Parse(chi.URLParam(r, "audienceId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid audience id")
		return
	}

	segmentID, err := uuid.Parse(chi.URLParam(r, "segmentId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid segment id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, audienceID, segmentID); err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
