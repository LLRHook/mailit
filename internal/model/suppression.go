package model

import (
	"time"

	"github.com/google/uuid"
)

type SuppressionEntry struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TeamID    uuid.UUID `json:"team_id" db:"team_id"`
	Email     string    `json:"email" db:"email"`
	Reason    string    `json:"reason" db:"reason"`
	Details   *string   `json:"details,omitempty" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

const (
	SuppressionBounce      = "bounce"
	SuppressionComplaint   = "complaint"
	SuppressionUnsubscribe = "unsubscribe"
	SuppressionManual      = "manual"
)
