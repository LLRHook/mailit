package handler

import (
	"net/http"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type SettingsHandler struct {
	service service.SettingsService
}

func NewSettingsHandler(s service.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: s}
}

// GetUsage handles GET /settings/usage.
func (h *SettingsHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	usage, err := h.service.GetUsage(r.Context(), auth.TeamID)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, usage)
}

// GetTeam handles GET /settings/team.
func (h *SettingsHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	team, err := h.service.GetTeam(r.Context(), auth.TeamID)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, team)
}

// UpdateTeam handles PATCH /settings/team.
func (h *SettingsHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateTeamRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.service.UpdateTeam(r.Context(), auth.TeamID, &req); err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, map[string]string{"message": "team updated"})
}

// GetSMTP handles GET /settings/smtp.
func (h *SettingsHandler) GetSMTP(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	smtp := h.service.GetSMTPConfig()
	pkg.JSON(w, http.StatusOK, smtp)
}

// InviteMember handles POST /settings/team/invite.
func (h *SettingsHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.InviteMemberRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	invitation, err := h.service.InviteMember(r.Context(), auth.TeamID, &req)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusCreated, invitation)
}

// AcceptInvite handles POST /auth/accept-invite.
func (h *SettingsHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req dto.AcceptInviteRequest
	if err := pkg.DecodeJSON(r, &req); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := pkg.Validate(&req); err != nil {
		pkg.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp, err := h.service.AcceptInvite(r.Context(), &req)
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
