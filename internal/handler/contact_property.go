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

type ContactPropertyHandler struct {
	service service.ContactPropertyService
}

func NewContactPropertyHandler(s service.ContactPropertyService) *ContactPropertyHandler {
	return &ContactPropertyHandler{service: s}
}

// Create handles POST /contact-properties.
func (h *ContactPropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateContactPropertyRequest
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

// List handles GET /contact-properties.
func (h *ContactPropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.List(r.Context(), auth.TeamID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /contact-properties/{propertyId}.
func (h *ContactPropertyHandler) Update(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid property id")
		return
	}

	var req dto.UpdateContactPropertyRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, propertyID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /contact-properties/{propertyId}.
func (h *ContactPropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	propertyID, err := uuid.Parse(chi.URLParam(r, "propertyId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid property id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, propertyID); err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
