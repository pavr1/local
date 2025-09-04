package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	// RequestIDHeader is the HTTP header name for request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestIDContextKey is the context key for request ID
type RequestIDContextKey struct{}

type RequestIDMiddleware struct {
	logger *logrus.Logger
}

func NewRequestIDMiddleware(logger *logrus.Logger) *RequestIDMiddleware {
	return &RequestIDMiddleware{logger: logger}
}

// RequestIDMiddleware generates a unique request ID for each request
func (rm *RequestIDMiddleware) InjectRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID is already provided in header
		requestID := generateRequestID()

		// Add request ID to response headers
		r.Header.Set(RequestIDHeader, requestID)

		// Log the incoming request with request ID
		rm.logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
		}).Debug("Incoming request, request ID injected")

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// CheckRequestIDMiddleware validates request ID header and adds it to context
// healthExcludedPath: path to exclude from X-Request-ID validation (e.g., "/api/v1/sessions/p/health")
func (rm *RequestIDMiddleware) CheckRequestIDMiddleware(serviceName string, healthExcludedPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip X-Request-ID validation for health check endpoints
			if r.URL.Path == healthExcludedPath {
				next.ServeHTTP(w, r)
				return
			}

			// Validate request ID header exists (should be provided by gateway)
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				rm.logger.Error("Missing X-Request-ID header")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)

				errorResponse := map[string]interface{}{
					"error":   "missing_request_id",
					"message": "X-Request-ID header is required",
				}

				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Convert to hex string
	return hex.EncodeToString(bytes)
}
