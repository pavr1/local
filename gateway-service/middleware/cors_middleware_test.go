package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestCORSMiddlewareStructure tests the CORSMiddleware struct creation
func TestCORSMiddlewareStructure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	middleware := NewCORSMiddleware(logger)

	assert.NotNil(t, middleware)
	assert.Equal(t, logger, middleware.logger)
}

// TestCORSMiddlewareHandleCORS tests the CORS middleware functionality
func TestCORSMiddlewareHandleCORS(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	middleware := NewCORSMiddleware(logger)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	corsHandler := middleware.HandleCORS(testHandler)

	t.Run("regular request with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check response status
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-User-ID, X-Username, X-User-Role, X-User-Permissions", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))

		// Check response body
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check response status
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-User-ID, X-Username, X-User-Role, X-User-Permissions", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))

		// Check that the next handler was not called (empty response body)
		assert.Empty(t, w.Body.String())
	})

	t.Run("POST request with custom headers", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check response status
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers are still set
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-User-ID, X-Username, X-User-Role, X-User-Permissions", w.Header().Get("Access-Control-Allow-Headers"))

		// Check response body
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("DELETE request without origin", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check response status
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers are still set
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))

		// Check response body
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("PUT request with complex origin", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/test", nil)
		req.Header.Set("Origin", "https://app.example.com:8080")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check response status
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))

		// Check response body
		assert.Equal(t, "test response", w.Body.String())
	})
}

// TestCORSMiddlewareHeaders tests that all required CORS headers are set correctly
func TestCORSMiddlewareHeaders(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	middleware := NewCORSMiddleware(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := middleware.HandleCORS(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	// Verify all CORS headers are present and correct
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-User-ID, X-Username, X-User-Role, X-User-Permissions",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "86400",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		assert.Equal(t, expectedValue, actualValue, "Header %s should be set correctly", header)
	}
}

// TestCORSMiddlewarePreflightResponse tests that preflight requests return immediately
func TestCORSMiddlewarePreflightResponse(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	middleware := NewCORSMiddleware(logger)

	// Create a handler that would normally write to response
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not be called"))
	})

	corsHandler := middleware.HandleCORS(testHandler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	// Verify preflight request was handled correctly
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String(), "Preflight request should not call next handler")
	assert.False(t, handlerCalled, "Next handler should not be called for OPTIONS request")
}

// TestCORSMiddlewareNextHandlerCalled tests that non-OPTIONS requests call the next handler
func TestCORSMiddlewareNextHandlerCalled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	middleware := NewCORSMiddleware(logger)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler called"))
	})

	corsHandler := middleware.HandleCORS(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	// Verify next handler was called
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "handler called", w.Body.String())
	assert.True(t, handlerCalled, "Next handler should be called for non-OPTIONS request")
}

