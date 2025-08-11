package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatewayMiddleware_ValidateGateway(t *testing.T) {
	middleware := NewGatewayMiddleware()

	t.Run("valid gateway request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/sessions/validate", nil)
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")

		w := httptest.NewRecorder()
		handlerCalled := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		middleware.ValidateGateway(handler).ServeHTTP(w, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing gateway service header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/sessions/validate", nil)
		req.Header.Set("X-Gateway-Session-Managed", "true")

		w := httptest.NewRecorder()
		handlerCalled := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware.ValidateGateway(handler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "gateway_required")
	})

	t.Run("missing session managed header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/sessions/validate", nil)
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")

		w := httptest.NewRecorder()
		handlerCalled := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware.ValidateGateway(handler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "gateway_required")
	})

	t.Run("invalid gateway service", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/sessions/validate", nil)
		req.Header.Set("X-Gateway-Service", "invalid-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")

		w := httptest.NewRecorder()
		handlerCalled := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware.ValidateGateway(handler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_gateway")
	})

	t.Run("invalid session managed flag", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/sessions/validate", nil)
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "false")

		w := httptest.NewRecorder()
		handlerCalled := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware.ValidateGateway(handler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "session_not_managed")
	})
}

func TestNewGatewayMiddleware(t *testing.T) {
	middleware := NewGatewayMiddleware()
	assert.NotNil(t, middleware)
}
