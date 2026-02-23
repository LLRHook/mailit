package dto

type CreateAPIKeyRequest struct {
	Name       string  `json:"name" validate:"required"`
	Permission string  `json:"permission" validate:"omitempty,oneof=full sending"`
	DomainID   *string `json:"domain_id,omitempty" validate:"omitempty,uuid"`
}

type APIKeyResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Token      string `json:"token,omitempty"` // only on create
	KeyPrefix  string `json:"key_prefix,omitempty"`
	Permission string `json:"permission"`
	CreatedAt  string `json:"created_at"`
}
