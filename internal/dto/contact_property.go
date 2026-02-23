package dto

type CreateContactPropertyRequest struct {
	Name  string `json:"name" validate:"required"`
	Label string `json:"label" validate:"required"`
	Type  string `json:"type" validate:"required,oneof=string number boolean date"`
}

type UpdateContactPropertyRequest struct {
	Label *string `json:"label,omitempty"`
}

type ContactPropertyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}
