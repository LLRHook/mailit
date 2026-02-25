package model

import (
	"time"

	"github.com/google/uuid"
)

// EmailMetrics holds aggregated email event counts for a time period.
type EmailMetrics struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TeamID      uuid.UUID `json:"team_id" db:"team_id"`
	PeriodStart time.Time `json:"period_start" db:"period_start"`
	PeriodType  string    `json:"period_type" db:"period_type"`
	Sent        int       `json:"sent" db:"sent"`
	Delivered   int       `json:"delivered" db:"delivered"`
	Bounced     int       `json:"bounced" db:"bounced"`
	Failed      int       `json:"failed" db:"failed"`
	Opened      int       `json:"opened" db:"opened"`
	Clicked     int       `json:"clicked" db:"clicked"`
	Complained  int       `json:"complained" db:"complained"`
}

const (
	PeriodTypeHourly = "hourly"
	PeriodTypeDaily  = "daily"
)
