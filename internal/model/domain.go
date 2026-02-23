package model

import (
	"time"

	"github.com/google/uuid"
)

type Domain struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TeamID         uuid.UUID `json:"team_id" db:"team_id"`
	Name           string    `json:"name" db:"name"`
	Status         string    `json:"status" db:"status"`
	Region         *string   `json:"region,omitempty" db:"region"`
	DKIMPrivateKey *string   `json:"-" db:"dkim_private_key"`
	DKIMSelector   string    `json:"dkim_selector" db:"dkim_selector"`
	OpenTracking   bool      `json:"open_tracking" db:"open_tracking"`
	ClickTracking  bool      `json:"click_tracking" db:"click_tracking"`
	TLSPolicy      string    `json:"tls_policy" db:"tls_policy"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// DomainDNSRecord is a DNS record associated with a domain.
type DomainDNSRecord struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	DomainID      uuid.UUID  `json:"domain_id" db:"domain_id"`
	RecordType    string     `json:"record_type" db:"record_type"`    // SPF, DKIM, MX, DMARC, RETURN_PATH
	DNSType       string     `json:"dns_type" db:"dns_type"`          // TXT, MX, CNAME, A, AAAA
	Name          string     `json:"name" db:"name"`
	Value         string     `json:"value" db:"value"`
	Priority      *int       `json:"priority,omitempty" db:"priority"`
	Status        string     `json:"status" db:"status"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty" db:"last_checked_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

const (
	DomainStatusPending  = "pending"
	DomainStatusVerified = "verified"
	DomainStatusFailed   = "failed"
)
