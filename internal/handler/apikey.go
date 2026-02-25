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

type APIKeyHandler struct {
	service service.APIKeyService
}

func NewAPIKeyHandler(s service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: s}
}

// Create handles POST /api-keys.
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateAPIKeyRequest
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

// List handles GET /api-keys.
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
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

// Delete handles DELETE /api-keys/{apiKeyId}.
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	apiKeyID, err := uuid.Parse(chi.URLParam(r, "apiKeyId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid api key id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, apiKeyID); err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
