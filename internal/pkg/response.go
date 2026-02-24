package pkg

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code and data.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes a JSON error response matching the Resend-style error format.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]interface{}{
		"statusCode": status,
		"message":    message,
		"name":       http.StatusText(status),
	})
}

// DecodeJSON decodes a JSON request body into the given value.
// Unknown fields in the request body will cause an error.
func DecodeJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
