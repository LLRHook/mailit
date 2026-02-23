package dto

type CreateAudienceRequest struct {
	Name string `json:"name" validate:"required"`
}

type AudienceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}
