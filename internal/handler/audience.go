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

type AudienceHandler struct {
	service service.AudienceService
}

func NewAudienceHandler(s service.AudienceService) *AudienceHandler {
	return &AudienceHandler{service: s}
}

// Create handles POST /audiences.
func (h *AudienceHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateAudienceRequest
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

// List handles GET /audiences.
func (h *AudienceHandler) List(w http.ResponseWriter, r *http.Request) {
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

// Get handles GET /audiences/{audienceId}.
func (h *AudienceHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.service.Get(r.Context(), auth.TeamID, audienceID)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /audiences/{audienceId}.
func (h *AudienceHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.Delete(r.Context(), auth.TeamID, audienceID); err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
