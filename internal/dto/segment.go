package dto

type CreateSegmentRequest struct {
	Name       string      `json:"name" validate:"required"`
	Conditions interface{} `json:"conditions" validate:"required"`
}

type UpdateSegmentRequest struct {
	Name       *string     `json:"name,omitempty"`
	Conditions interface{} `json:"conditions,omitempty"`
}

type SegmentResponse struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"conditions"`
	CreatedAt  string      `json:"created_at"`
}
