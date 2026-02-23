package model

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TeamMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TeamID    uuid.UUID `json:"team_id" db:"team_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)
