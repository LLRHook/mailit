package model

import (
	"time"

	"github.com/google/uuid"
)

// TeamInvitation represents a pending invitation to join a team.
type TeamInvitation struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TeamID     uuid.UUID  `json:"team_id" db:"team_id"`
	Email      string     `json:"email" db:"email"`
	Role       string     `json:"role" db:"role"`
	Token      string     `json:"-" db:"token"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}
