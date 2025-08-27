package middleware

import (
	"context"
	"encoding/json"
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
			// Validate request ID header exists
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				logger.Error("Missing X-Request-ID header")

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

			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}
