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

type DomainHandler struct {
	service service.DomainService
}

func NewDomainHandler(s service.DomainService) *DomainHandler {
	return &DomainHandler{service: s}
}

// Create handles POST /domains.
func (h *DomainHandler) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateDomainRequest
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

// List handles GET /domains.
func (h *DomainHandler) List(w http.ResponseWriter, r *http.Request) {
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

// Get handles GET /domains/{domainId}.
func (h *DomainHandler) Get(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	domainID, err := uuid.Parse(chi.URLParam(r, "domainId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid domain id")
		return
	}

	resp, err := h.service.Get(r.Context(), auth.TeamID, domainID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Update handles PATCH /domains/{domainId}.
func (h *DomainHandler) Update(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	domainID, err := uuid.Parse(chi.URLParam(r, "domainId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid domain id")
		return
	}

	var req dto.UpdateDomainRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, domainID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /domains/{domainId}.
func (h *DomainHandler) Delete(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	domainID, err := uuid.Parse(chi.URLParam(r, "domainId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid domain id")
		return
	}

	if err := h.service.Delete(r.Context(), auth.TeamID, domainID); err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// Verify handles POST /domains/{domainId}/verify.
func (h *DomainHandler) Verify(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	domainID, err := uuid.Parse(chi.URLParam(r, "domainId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid domain id")
		return
	}

	resp, err := h.service.Verify(r.Context(), auth.TeamID, domainID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
