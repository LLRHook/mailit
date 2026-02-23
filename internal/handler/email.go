package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type EmailHandler struct {
	service service.EmailService
}

func NewEmailHandler(s service.EmailService) *EmailHandler {
	return &EmailHandler{service: s}
}

// Send handles POST /emails.
func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.SendEmailRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if key := r.Header.Get("Idempotency-Key"); key != "" {
		req.IdempotencyKey = &key
	}

	resp, err := h.service.Send(r.Context(), auth.TeamID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// BatchSend handles POST /emails/batch.
func (h *EmailHandler) BatchSend(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.BatchSendEmailRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp, err := h.service.BatchSend(r.Context(), auth.TeamID, &req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// List handles GET /emails.
func (h *EmailHandler) List(w http.ResponseWriter, r *http.Request) {
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

// Get handles GET /emails/{emailId}.
func (h *EmailHandler) Get(w http.ResponseWriter, r *http.Request) {
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

// Update handles PATCH /emails/{emailId}.
func (h *EmailHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req map[string]interface{}
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), auth.TeamID, emailID, req)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// Cancel handles POST /emails/{emailId}/cancel.
func (h *EmailHandler) Cancel(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.service.Cancel(r.Context(), auth.TeamID, emailID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}

// parsePagination extracts page and per_page from query params with defaults.
func parsePagination(r *http.Request) dto.PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	params := dto.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
	params.Normalize()
	return params
}
