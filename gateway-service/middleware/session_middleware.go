package middleware

import (
	"encoding/json"
	sessionmanager "gateway-service/middleware/session-manager"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SessionMiddleware handles session validation for protected routes
type SessionMiddleware struct {
	sessionManager *sessionmanager.SessionManager
	logger         *logrus.Logger
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(sessionManager *sessionmanager.SessionManager, logger *logrus.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// ValidateSession middleware validates the session ID against the session service
func (sm *SessionMiddleware) ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm.logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Info("Session validation middleware triggered")

		// Extract session ID from Authorization header
		sessionId := extractSessionIdFromHeader(r)
		if sessionId == "" {
			sm.logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Warn("Session validation failed: missing session ID")
			sm.writeErrorResponse(w, http.StatusUnauthorized, "missing_session", "Session ID is required")
			return
		}

		sm.logger.WithFields(logrus.Fields{
			"session_id": sessionId,
			"method":     r.Method,
			"path":       r.URL.Path,
		}).Info("Validating session with session service")

		// Validate session ID with session service
		validation, err := sm.sessionManager.ValidateSession(sessionId)
		if err != nil {
			sm.logger.WithError(err).WithFields(logrus.Fields{
				"session_id": sessionId,
				"method":     r.Method,
				"path":       r.URL.Path,
			}).Error("Session validation error")
			sm.writeErrorResponse(w, http.StatusInternalServerError, "validation_error", "Failed to validate session")
			return
		}

		// Check if session is valid
		if !validation.Valid {
			sm.logger.WithFields(logrus.Fields{
				"session_id": sessionId,
				"method":     r.Method,
				"path":       r.URL.Path,
				"message":    validation.Message,
			}).Warn("Session validation failed: invalid session")
			sm.writeErrorResponse(w, http.StatusUnauthorized, "invalid_session", validation.Message)
			return
		}

		sm.logger.WithFields(logrus.Fields{
			"session_id": sessionId,
			"user_id":    validation.UserID,
			"username":   validation.Username,
			"method":     r.Method,
			"path":       r.URL.Path,
		}).Info("Session validation successful")

		// Add user context to request headers for backend services
		r.Header.Set("X-User-ID", validation.UserID)
		r.Header.Set("X-Username", validation.Username)
		r.Header.Set("X-User-Role", validation.RoleName)

		// Convert permissions to comma-separated string
		if len(validation.Permissions) > 0 {
			r.Header.Set("X-User-Permissions", strings.Join(validation.Permissions, ","))
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// LoginSession handles login and creates sessions
func (sm *SessionMiddleware) LoginSession(sessionServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sm.logger.WithError(err).Error("Failed to read request body")
			sm.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
			return
		}

		// Forward login request to session service with gateway headers
		req, err := http.NewRequest("POST", sessionServiceURL+"/api/v1/sessions/p/login", strings.NewReader(string(body)))
		if err != nil {
			sm.logger.WithError(err).Error("Failed to create login request")
			sm.writeErrorResponse(w, http.StatusInternalServerError, "request_error", "Failed to create login request")
			return
		}

		// Add gateway headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			sm.logger.WithError(err).WithFields(logrus.Fields{
				"session_service_url": sessionServiceURL,
				"method":              r.Method,
				"remote_addr":         r.RemoteAddr,
			}).Error("Login proxy error")
			sm.writeErrorResponse(w, http.StatusBadGateway, "service_unavailable", "Authentication service unavailable")
			return
		}
		defer resp.Body.Close()

		// Read response from session service
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			sm.logger.WithError(err).WithFields(logrus.Fields{
				"session_service_url": sessionServiceURL,
				"status_code":         resp.StatusCode,
			}).Error("Failed to read login response body")
			sm.writeErrorResponse(w, http.StatusInternalServerError, "response_error", "Failed to read login response")
			return
		}

		// Gateway acts as pure proxy - session service handles all session creation logic
		if resp.StatusCode == http.StatusOK {
			sm.logger.WithFields(logrus.Fields{
				"session_service_url": sessionServiceURL,
				"status_code":         resp.StatusCode,
				"response_length":     len(respBody),
			}).Info("Login successful - session service handled session creation")
		}

		// Copy headers from session service response
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set status code and write response
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}
}

// LogoutSession handles logout and revokes sessions
func (sm *SessionMiddleware) LogoutSession(sessionServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from request
		sessionId := extractSessionIdFromHeader(r)
		if sessionId != "" {
			sm.logger.WithFields(logrus.Fields{
				"session_id":  sessionId,
				"method":      r.Method,
				"remote_addr": r.RemoteAddr,
			}).Info("Attempting to revoke session")

			// Revoke session in session service
			if err := sm.sessionManager.LogoutSession(sessionId); err != nil {
				sm.logger.WithError(err).WithFields(logrus.Fields{
					"session_id": sessionId,
				}).Error("Failed to revoke session")
			} else {
				sm.logger.WithFields(logrus.Fields{
					"session_id": sessionId,
				}).Info("Session revoked successfully")
			}
		} else {
			sm.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"remote_addr": r.RemoteAddr,
			}).Warn("Logout request without session ID")
		}

		// Forward logout request to session service with gateway headers
		req, err := http.NewRequest("POST", sessionServiceURL+"/api/v1/sessions/logout", r.Body)
		if err != nil {
			sm.logger.WithError(err).Error("Failed to create logout request")
			sm.writeErrorResponse(w, http.StatusInternalServerError, "request_error", "Failed to create logout request")
			return
		}

		// Add gateway headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			sm.logger.WithError(err).WithFields(logrus.Fields{
				"session_service_url": sessionServiceURL,
				"method":              r.Method,
				"remote_addr":         r.RemoteAddr,
			}).Error("Logout proxy error")
			sm.writeErrorResponse(w, http.StatusBadGateway, "service_unavailable", "Authentication service unavailable")
			return
		}
		defer resp.Body.Close()

		// Copy response from session service
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			sm.logger.WithError(err).Error("Failed to read logout response body")
			sm.writeErrorResponse(w, http.StatusInternalServerError, "response_error", "Failed to read logout response")
			return
		}

		// Copy headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}
}

// Helper functions

func extractSessionIdFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer session ID
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return authHeader[len(bearerPrefix):]
	}

	return ""
}

func (sm *SessionMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     errorCode,
		"message":   message,
		"timestamp": time.Now(),
		"service":   "gateway",
	}

	json.NewEncoder(w).Encode(response)
}
