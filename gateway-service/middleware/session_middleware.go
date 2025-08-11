package middleware

import (
	"encoding/json"
	sessionmanager "gateway-service/middleware/session-manager"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// SessionMiddleware handles session validation for protected routes
type SessionMiddleware struct {
	sessionManager *sessionmanager.SessionManager
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(sessionManager *sessionmanager.SessionManager) *SessionMiddleware {
	return &SessionMiddleware{
		sessionManager: sessionManager,
	}
}

// ValidateSession middleware validates the session ID against the session service
func (sm *SessionMiddleware) ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from Authorization header
		sessionId := extractSessionIdFromHeader(r)
		if sessionId == "" {
			sm.writeErrorResponse(w, http.StatusUnauthorized, "missing_session", "Session ID is required")
			return
		}

		// Validate session ID with session service
		validation, err := sm.sessionManager.ValidateSession(sessionId)
		if err != nil {
			log.Printf("Session validation error: %v", err)
			sm.writeErrorResponse(w, http.StatusInternalServerError, "validation_error", "Failed to validate session")
			return
		}

		// Check if session is valid
		if !validation.Valid {
			sm.writeErrorResponse(w, http.StatusUnauthorized, "invalid_session", validation.Message)
			return
		}

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
			sm.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
			return
		}

		// Forward login request to session service with gateway headers
		req, err := http.NewRequest("POST", sessionServiceURL+"/api/v1/sessions/p/login", strings.NewReader(string(body)))
		if err != nil {
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
			log.Printf("Login proxy error: %v", err)
			sm.writeErrorResponse(w, http.StatusBadGateway, "service_unavailable", "Authentication service unavailable")
			return
		}
		defer resp.Body.Close()

		// Read response from session service
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			sm.writeErrorResponse(w, http.StatusInternalServerError, "response_error", "Failed to read login response")
			return
		}

		// Gateway acts as pure proxy - session service handles all session creation logic
		if resp.StatusCode == http.StatusOK {
			log.Printf("Login successful - session service handled session creation")
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
			// Revoke session in session service
			if err := sm.sessionManager.LogoutSession(sessionId); err != nil {
				log.Printf("Failed to revoke session: %v", err)
			} else {
				log.Printf("Session revoked successfully")
			}
		}

		// Forward logout request to session service with gateway headers
		req, err := http.NewRequest("POST", sessionServiceURL+"/api/v1/sessions/logout", r.Body)
		if err != nil {
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
			log.Printf("Logout proxy error: %v", err)
			sm.writeErrorResponse(w, http.StatusBadGateway, "service_unavailable", "Authentication service unavailable")
			return
		}
		defer resp.Body.Close()

		// Copy response from session service
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
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

// Helper functions to safely extract values from interface{} maps
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return nil
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}
