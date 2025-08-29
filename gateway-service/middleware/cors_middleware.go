package middleware

import (
	"net/http"
	sharedLogger "shared/logger"

	"github.com/sirupsen/logrus"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
type CORSMiddleware struct {
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// HandleCORS middleware sets CORS headers and handles preflight requests
func (cm *CORSMiddleware) HandleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)
		logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"origin": r.Header.Get("Origin"),
		}).Debug("CORS middleware triggered")

		// Set CORS headers - only the gateway sets these
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Username, X-User-Role, X-User-Permissions")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"origin": r.Header.Get("Origin"),
			}).Debug("Handling CORS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
