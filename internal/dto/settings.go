package dto

// UsageResponse holds usage count data for the settings page.
type UsageResponse struct {
	EmailsSentToday int `json:"emails_sent_today"`
	EmailsSentMonth int `json:"emails_sent_month"`
	Domains         int `json:"domains"`
	APIKeys         int `json:"api_keys"`
	Webhooks        int `json:"webhooks"`
	Contacts        int `json:"contacts"`
}

// TeamMemberResponse represents a team member with user info.
type TeamMemberResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// TeamResponse holds team info and member list.
type TeamResponse struct {
	ID      string               `json:"id"`
	Name    string               `json:"name"`
	Slug    string               `json:"slug"`
	Members []TeamMemberResponse `json:"members"`
}

// UpdateTeamRequest is the request body for PATCH /settings/team.
type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// SMTPConfigResponse holds SMTP configuration for display.
type SMTPConfigResponse struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"`
}

// InviteMemberRequest is the request body for POST /settings/team/invite.
type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member"`
}

// AcceptInviteRequest is the request body for POST /auth/accept-invite.
type AcceptInviteRequest struct {
	Token    string `json:"token" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}
