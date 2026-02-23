package pkg

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate runs struct validation using go-playground/validator tags.
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidationErrors extracts a map of field names to failed validation tags
// from a validator.ValidationErrors error.
func ValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Tag()
		}
	}
	return errors
}
