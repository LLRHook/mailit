package pkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	t.Run("correct Content-Type header", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSON(w, http.StatusOK, map[string]string{"key": "value"})

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("correct status code 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSON(w, http.StatusOK, map[string]string{"key": "value"})

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("correct status code 201", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSON(w, http.StatusCreated, map[string]string{"id": "123"})

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("correct body encoding", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]interface{}{
			"name":  "Alice",
			"count": 42,
		}
		JSON(w, http.StatusOK, data)

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, float64(42), result["count"])
	})

	t.Run("struct body encoding", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "abc", Name: "Test"}
		JSON(w, http.StatusOK, data)

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "abc", result["id"])
		assert.Equal(t, "Test", result["name"])
	})

	t.Run("nil body encodes to null", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSON(w, http.StatusOK, nil)

		assert.Equal(t, "null\n", w.Body.String())
	})

	t.Run("slice body encoding", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := []string{"a", "b", "c"}
		JSON(w, http.StatusOK, data)

		var result []string
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})
}

func TestError(t *testing.T) {
	t.Run("correct error format for 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusBadRequest, "invalid input")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(400), result["statusCode"])
		assert.Equal(t, "invalid input", result["message"])
		assert.Equal(t, "Bad Request", result["name"])
	})

	t.Run("correct error format for 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusNotFound, "resource not found")

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(404), result["statusCode"])
		assert.Equal(t, "resource not found", result["message"])
		assert.Equal(t, "Not Found", result["name"])
	})

	t.Run("correct error format for 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusInternalServerError, "something went wrong")

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(500), result["statusCode"])
		assert.Equal(t, "something went wrong", result["message"])
		assert.Equal(t, "Internal Server Error", result["name"])
	})

	t.Run("correct error format for 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusUnauthorized, "missing api key")

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(401), result["statusCode"])
		assert.Equal(t, "missing api key", result["message"])
		assert.Equal(t, "Unauthorized", result["name"])
	})

	t.Run("correct error format for 429", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusTooManyRequests, "rate limit exceeded")

		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(429), result["statusCode"])
		assert.Equal(t, "rate limit exceeded", result["message"])
		assert.Equal(t, "Too Many Requests", result["name"])
	})
}

func TestDecodeJSON(t *testing.T) {
	t.Run("valid JSON decodes correctly", func(t *testing.T) {
		body := `{"name":"Alice","age":30}`
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		var result struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		err := DecodeJSON(r, &result)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		body := `{invalid json`
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		var result map[string]interface{}
		err := DecodeJSON(r, &result)
		assert.Error(t, err)
	})

	t.Run("unknown fields are rejected", func(t *testing.T) {
		body := `{"name":"Alice","unknown_field":"value"}`
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		var result struct {
			Name string `json:"name"`
		}
		err := DecodeJSON(r, &result)
		assert.Error(t, err, "should reject unknown fields due to DisallowUnknownFields")
	})

	t.Run("empty body returns error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

		var result map[string]interface{}
		err := DecodeJSON(r, &result)
		assert.Error(t, err)
	})

	t.Run("null JSON body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("null"))

		var result *struct{ Name string }
		err := DecodeJSON(r, &result)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
