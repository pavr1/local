package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// GatewayMiddleware ensures requests come through the gateway
type GatewayMiddleware struct {
	logger *logrus.Logger
}

// NewGatewayMiddleware creates a new gateway validation middleware
func NewGatewayMiddleware(logger *logrus.Logger) *GatewayMiddleware {
	return &GatewayMiddleware{
		logger: logger,
	}
}

// ValidateGateway ensures the request comes through the gateway
func (gm *GatewayMiddleware) ValidateGateway(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health checks to bypass gateway validation for monitoring
		if r.URL.Path == "/api/v1/sessions/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for gateway headers
		gatewayService := r.Header.Get("X-Gateway-Service")
		gatewaySessionManaged := r.Header.Get("X-Gateway-Session-Managed")

		if gatewayService != "ice-cream-gateway" || gatewaySessionManaged != "true" {
			gm.logger.WithFields(logrus.Fields{
				"remote_addr":     r.RemoteAddr,
				"method":          r.Method,
				"path":            r.URL.Path,
				"gateway_service": gatewayService,
				"gateway_managed": gatewaySessionManaged,
			}).Warn("Direct access attempt blocked - requests must go through gateway")

			gm.writeErrorResponse(w, http.StatusForbidden, "gateway_required", "Direct access not allowed. All requests must go through the gateway.")
			return
		}

		// Valid gateway request, continue
		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes a JSON error response
func (gm *GatewayMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     errorCode,
		"message":   message,
		"service":   "session-service",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}
