package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// SessionMiddleware handles session validation for protected routes
type SessionMiddleware struct {
	sessionManager *SessionManager
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(sessionManager *SessionManager) *SessionMiddleware {
	return &SessionMiddleware{
		sessionManager: sessionManager,
	}
}

// ValidateSession middleware validates the JWT token against the session service
func (sm *SessionMiddleware) ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := extractTokenFromHeader(r)
		if token == "" {
			sm.writeErrorResponse(w, http.StatusUnauthorized, "missing_token", "Authorization token is required")
			return
		}

		// Validate token with session service
		validation, err := sm.sessionManager.ValidateSession(token)
		if err != nil {
			log.Printf("Session validation error: %v", err)
			sm.writeErrorResponse(w, http.StatusInternalServerError, "validation_error", "Failed to validate session")
			return
		}

		// Check if session is valid
		if !validation.IsValid {
			sm.writeErrorResponse(w, http.StatusUnauthorized, validation.ErrorCode, validation.ErrorMessage)
			return
		}

		// Add user context to request headers for backend services
		if validation.Session != nil {
			r.Header.Set("X-User-ID", validation.Session.UserID)
			r.Header.Set("X-Username", validation.Session.Username)
			r.Header.Set("X-User-Role", validation.Session.RoleName)

			// Convert permissions to comma-separated string
			if len(validation.Session.Permissions) > 0 {
				r.Header.Set("X-User-Permissions", strings.Join(validation.Session.Permissions, ","))
			}
		}

		// Handle token refresh if needed
		if validation.ShouldRefresh && validation.NewToken != "" {
			w.Header().Set("X-New-Token", validation.NewToken)
			log.Printf("Token refreshed for user %s", validation.Session.Username)
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// SessionAwareLoginHandler handles login and creates sessions
func (sm *SessionMiddleware) SessionAwareLoginHandler(sessionServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sm.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
			return
		}

		// Forward login request to session service with gateway headers
		req, err := http.NewRequest("POST", sessionServiceURL+"/api/v1/sessions/login", strings.NewReader(string(body)))
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

		// If login was successful (200 OK), create a session
		if resp.StatusCode == http.StatusOK {
			var loginResp map[string]interface{}
			if err := json.Unmarshal(respBody, &loginResp); err == nil {
				// Extract user information from login response
				if userInfo, ok := loginResp["user"].(map[string]interface{}); ok {
					// Create session request
					sessionReq := &SessionCreateRequest{
						UserID:      getString(userInfo, "id"),
						Username:    getString(userInfo, "username"),
						RoleName:    getString(userInfo, "role"),
						Permissions: getStringSlice(userInfo, "permissions"),
						RememberMe:  getBool(loginResp, "remember_me"),
					}

					// Create session in session service
					sessionResp, err := sm.sessionManager.CreateSession(sessionReq)
					if err != nil {
						log.Printf("Failed to create session: %v", err)
						// Return original login response even if session creation fails
					} else {
						// Update response with session token instead of original token
						loginResp["token"] = sessionResp.Token
						loginResp["session_id"] = sessionResp.SessionID
						loginResp["expires_at"] = sessionResp.ExpiresAt

						// Re-marshal the updated response
						if updatedBody, err := json.Marshal(loginResp); err == nil {
							respBody = updatedBody
						}

						log.Printf("Session created for user %s (session: %s)", sessionReq.Username, sessionResp.SessionID)
					}
				}
			}
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

// SessionAwareLogoutHandler handles logout and revokes sessions
func (sm *SessionMiddleware) SessionAwareLogoutHandler(sessionServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request
		token := extractTokenFromHeader(r)
		if token != "" {
			// Revoke session in session service
			if err := sm.sessionManager.LogoutSession(token); err != nil {
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

func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
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
