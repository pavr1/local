package middlewares

import (
	"net/http"
	sharedLogger "shared/logger"

	"github.com/sirupsen/logrus"
)

// AddGatewayHeadersMiddleware adds gateway-specific headers to requests
func InjectGatewayHeadersMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := sharedLogger.GetRequestLogger(r, serviceName)

			// Add gateway headers
			// Only set Content-Type if not already set (preserve multipart/form-data for file uploads)
			if r.Header.Get("Content-Type") == "" {
				r.Header.Set("Content-Type", "application/json")
			}
			r.Header.Set("X-Gateway-Service", "ice-cream-gateway")
			r.Header.Set("X-Gateway-Session-Managed", "true")

			logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Debug("Added gateway headers to request")

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateGateway checks if the request comes from the gateway or internal services
func CheckGatewayMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)
			// Check for gateway headers
			gatewayService := r.Header.Get("X-Gateway-Service")
			gatewaySessionManaged := r.Header.Get("X-Gateway-Session-Managed")

			// Allow requests from gateway or internal services
			if gatewayService != "" && gatewaySessionManaged == "true" {
				// Valid gateway or internal service request, continue
				next.ServeHTTP(w, r)
				return
			}

			// Block direct access
			logger.WithFields(logrus.Fields{
				"remote_addr":               r.RemoteAddr,
				"method":                    r.Method,
				"path":                      r.URL.Path,
				"X-Gateway-Service":         gatewayService,
				"X-Gateway-Session-Managed": gatewaySessionManaged,
			}).Warn("Direct access attempt blocked - requests must go through gateway")

			http.Error(w, "Direct access not allowed. All requests must go through the gateway.", http.StatusForbidden)
		})
	}
}
