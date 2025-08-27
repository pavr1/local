package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	// RequestIDHeader is the HTTP header name for request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID is already provided in header
			requestID := r.Header.Get(RequestIDHeader)
			
			// Generate new request ID if not provided
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Add request ID to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "request_id", requestID)
			r = r.WithContext(ctx)

			// Log the incoming request with request ID
			logger.WithFields(logrus.Fields{
				"request_id":  requestID,
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
			}).Info("Incoming request")

			// Call next handler
			next.ServeHTTP(w, r)

			// Log the completed request
			logger.WithField("request_id", requestID).Info("Request completed")
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
