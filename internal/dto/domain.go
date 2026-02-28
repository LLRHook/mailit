package dto

type CreateDomainRequest struct {
	Name string `json:"name" validate:"required,fqdn"`
}

type DomainResponse struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Status    string              `json:"status"`
	Region    *string             `json:"region,omitempty"`
	DNSRecords []DNSRecordResponse `json:"dns_records"`
	CreatedAt string              `json:"created_at"`
}

type DNSRecordResponse struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
	Status   string `json:"status"`
	TTL      string `json:"ttl"`
}

type UpdateDomainRequest struct {
	OpenTracking  *bool   `json:"open_tracking,omitempty"`
	ClickTracking *bool   `json:"click_tracking,omitempty"`
	TLSPolicy     *string `json:"tls_policy,omitempty" validate:"omitempty,oneof=opportunistic enforce"`
}
