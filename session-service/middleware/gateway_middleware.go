package middleware

import (
	"encoding/json"
	"net/http"
	"time"
)

// GatewayMiddleware validates that requests come through the gateway
type GatewayMiddleware struct{}

// NewGatewayMiddleware creates a new gateway middleware
func NewGatewayMiddleware() *GatewayMiddleware {
	return &GatewayMiddleware{}
}

// ValidateGateway checks if the request comes from the gateway
func (gm *GatewayMiddleware) ValidateGateway(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for gateway headers
		gatewayService := r.Header.Get("X-Gateway-Service")
		gatewaySessionManaged := r.Header.Get("X-Gateway-Session-Managed")

		if gatewayService == "" || gatewaySessionManaged == "" {
			gm.writeErrorResponse(w, http.StatusForbidden, "gateway_required", "Direct access not allowed. All requests must go through the gateway.")
			return
		}

		// Validate gateway service identifier
		if gatewayService != "ice-cream-gateway" {
			gm.writeErrorResponse(w, http.StatusForbidden, "invalid_gateway", "Invalid gateway service identifier.")
			return
		}

		// Validate session managed flag
		if gatewaySessionManaged != "true" {
			gm.writeErrorResponse(w, http.StatusForbidden, "session_not_managed", "Session management must be handled by gateway.")
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes an error response
func (gm *GatewayMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     errorCode,
		"message":   message,
		"timestamp": time.Now(),
		"service":   "session-service",
	}

	json.NewEncoder(w).Encode(response)
}
