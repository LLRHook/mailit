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

type BroadcastHandler struct {
	service service.BroadcastService
}

func NewBroadcastHandler(s service.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{service: s}
}

// Create handles POST /broadcasts.
func (h *BroadcastHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateBroadcastRequest
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
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusCreated, resp)
}

// List handles GET /broadcasts.
func (h *BroadcastHandler) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	params := parsePagination(r)

	resp, err := h.service.List(r.Context(), auth.TeamID, &params)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Get handles GET /broadcasts/{broadcastId}.
func (h *BroadcastHandler) Get(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcastId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid broadcast id")
		return
	}

	resp, err := h.service.Get(r.Context(), auth.TeamID, broadcastID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /broadcasts/{broadcastId}.
func (h *BroadcastHandler) Update(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcastId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid broadcast id")
		return
	}

	var req dto.UpdateBroadcastRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, broadcastID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /broadcasts/{broadcastId}.
func (h *BroadcastHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcastId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid broadcast id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, broadcastID); err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// Send handles POST /broadcasts/{broadcastId}/send.
func (h *BroadcastHandler) Send(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcastId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid broadcast id")
		return
	}

	resp, err := h.service.Send(r.Context(), auth.TeamID, broadcastID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
