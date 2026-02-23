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

type ContactHandler struct {
	service service.ContactService
}

func NewContactHandler(s service.ContactService) *ContactHandler {
	return &ContactHandler{service: s}
}

// Create handles POST /audiences/{audienceId}/contacts.
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateContactRequest
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

// List handles GET /audiences/{audienceId}/contacts.
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
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

	params := parsePagination(r)

	resp, err := h.service.List(r.Context(), auth.TeamID, audienceID, &params)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Get handles GET /audiences/{audienceId}/contacts/{contactId}.
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}

	resp, err := h.service.Get(r.Context(), auth.TeamID, audienceID, contactID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /audiences/{audienceId}/contacts/{contactId}.
func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}

	var req dto.UpdateContactRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, audienceID, contactID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /audiences/{audienceId}/contacts/{contactId}.
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, audienceID, contactID); err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
