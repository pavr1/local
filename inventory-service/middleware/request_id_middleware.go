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

// RequestIDMiddleware validates request ID header and adds it to context
func RequestIDMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID is already provided in header
			requestID := r.Header.Get(RequestIDHeader)
			
			// Generate new request ID if not provided
			if requestID == "" {
				requestID = generateRequestID()
				logger.WithField("request_id", requestID).Debug("Generated new request ID")
			}

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

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
