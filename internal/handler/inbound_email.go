package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type InboundEmailHandler struct {
	service service.InboundEmailService
}

func NewInboundEmailHandler(s service.InboundEmailService) *InboundEmailHandler {
	return &InboundEmailHandler{service: s}
}

// List handles GET /inbound/emails.
func (h *InboundEmailHandler) List(w http.ResponseWriter, r *http.Request) {
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

// Get handles GET /inbound/emails/{emailId}.
func (h *InboundEmailHandler) Get(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	emailID, err := uuid.Parse(chi.URLParam(r, "emailId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid email id")
		return
	}

	resp, err := h.service.Get(r.Context(), auth.TeamID, emailID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
