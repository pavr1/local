package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"session-service/models"
	"session-service/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SessionHandler handles session management endpoints with comprehensive security
type SessionHandler struct {
	sessionManager *utils.SessionManager
	jwtManager     *utils.JWTManager
	logger         *logrus.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager *utils.SessionManager, jwtManager *utils.JWTManager, logger *logrus.Logger) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
		jwtManager:     jwtManager,
		logger:         logger,
	}
}

// ValidateSessionToken validates a token against the session store
func (h *SessionHandler) ValidateSessionToken(w http.ResponseWriter, r *http.Request) {
	var req models.SessionValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Validate session
	response, err := h.sessionManager.ValidateSession(&req)
	if err != nil {
		h.logger.WithError(err).Error("Session validation failed")
		h.writeErrorResponse(w, http.StatusInternalServerError, "validation_error", "Session validation failed")
		return
	}

	// Return validation response
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetUserSessions returns all active sessions for a user
func (h *SessionHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL or token
	vars := mux.Vars(r)
	userID := vars["userID"]

	// Get current session ID from token
	currentSessionID := h.getCurrentSessionID(r)

	sessions, err := h.sessionManager.GetUserSessions(userID, currentSessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user sessions")
		h.writeErrorResponse(w, http.StatusInternalServerError, "fetch_error", "Failed to retrieve sessions")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeSession revokes a specific session
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	var req models.SessionRevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Simplified revoke request (no reason tracking)

	err := h.sessionManager.RevokeSession(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to revoke session")
		h.writeErrorResponse(w, http.StatusInternalServerError, "revoke_error", "Failed to revoke session")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// RevokeAllUserSessions revokes all sessions for a user except current one
func (h *SessionHandler) RevokeAllUserSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	currentSessionID := h.getCurrentSessionID(r)

	// Get all user sessions
	sessions, err := h.sessionManager.GetUserSessions(userID, currentSessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user sessions for revocation")
		h.writeErrorResponse(w, http.StatusInternalServerError, "fetch_error", "Failed to retrieve sessions")
		return
	}

	// Revoke all sessions except current one
	revokedCount := 0
	for _, session := range sessions {
		if session.SessionID != currentSessionID {
			err := h.sessionManager.RevokeSession(&models.SessionRevokeRequest{
				SessionID: session.SessionID,
			})
			if err != nil {
				h.logger.WithError(err).Warn("Failed to revoke session during bulk revocation")
			} else {
				revokedCount++
			}
		}
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "Sessions revoked successfully",
		"revoked_count":  revokedCount,
		"total_sessions": len(sessions),
	})
}

// GetSessionStats returns session analytics
func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	stats := h.sessionManager.GetSessionStats()
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// RefreshSession refreshes a session token
func (h *SessionHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Authorization token is required")
		return
	}

	// Validate and potentially refresh the session
	req := &models.SessionValidationRequest{
		Token: token,
	}

	response, err := h.sessionManager.ValidateSession(req)
	if err != nil {
		h.logger.WithError(err).Error("Session refresh validation failed")
		h.writeErrorResponse(w, http.StatusUnauthorized, "refresh_failed", "Session refresh failed")
		return
	}

	if !response.IsValid {
		h.writeErrorResponse(w, http.StatusUnauthorized, response.ErrorCode, response.ErrorMessage)
		return
	}

	// Return refresh response
	refreshResponse := map[string]interface{}{
		"success": true,
		"message": "Session refreshed successfully",
	}

	if response.ShouldRefresh && response.NewToken != "" {
		refreshResponse["token"] = response.NewToken
		refreshResponse["refreshed"] = true
	} else {
		refreshResponse["refreshed"] = false
	}

	h.writeJSONResponse(w, http.StatusOK, refreshResponse)
}

// CreateSessionFromLogin creates a new session after successful login
func (h *SessionHandler) CreateSessionFromLogin(userProfile *models.UserProfile, r *http.Request, rememberMe bool) (*models.SessionData, string, error) {
	// Convert permissions to string slice
	permissions := make([]string, len(userProfile.Permissions))
	for i, perm := range userProfile.Permissions {
		permissions[i] = perm.PermissionName
	}

	// Create session request
	req := &models.SessionCreateRequest{
		UserID:      userProfile.User.ID,
		Username:    userProfile.User.Username,
		RoleName:    userProfile.Role.RoleName,
		Permissions: permissions,
		RememberMe:  rememberMe,
	}

	// Create session
	session, token, err := h.sessionManager.CreateSession(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return session, token, nil
}

// LogoutFromSession handles logout with session revocation
func (h *SessionHandler) LogoutFromSession(w http.ResponseWriter, r *http.Request) {
	token := h.extractTokenFromHeader(r)
	if token == "" {
		// No token provided, just return success
		h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}

	// Revoke the session
	err := h.sessionManager.RevokeSession(&models.SessionRevokeRequest{
		Token: token,
	})

	if err != nil {
		h.logger.WithError(err).Warn("Failed to revoke session during logout")
		// Don't fail the logout if session revocation fails
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// Middleware for session validation
func (h *SessionHandler) SessionValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractTokenFromHeader(r)
		if token == "" {
			h.writeErrorResponse(w, http.StatusUnauthorized, "missing_token", "Authorization token is required")
			return
		}

		// Validate session
		req := &models.SessionValidationRequest{
			Token: token,
		}

		response, err := h.sessionManager.ValidateSession(req)
		if err != nil || !response.IsValid {
			h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_session", "Session is not valid")
			return
		}

		// Add session data to request context
		ctx := r.Context()
		ctx = addSessionToContext(ctx, response.SessionData)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper methods

func (h *SessionHandler) extractTokenFromHeader(r *http.Request) string {
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

func (h *SessionHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (h *SessionHandler) getCurrentSessionID(r *http.Request) string {
	// Extract session ID from context if available
	session := getSessionFromContext(r.Context())
	if session != nil {
		return session.SessionID
	}

	// Try to get it from token if context not available
	token := h.extractTokenFromHeader(r)
	if token == "" {
		return ""
	}

	// Validate session to get session ID
	req := &models.SessionValidationRequest{
		Token: token,
	}

	response, err := h.sessionManager.ValidateSession(req)
	if err != nil || !response.IsValid || response.SessionData == nil {
		return ""
	}

	return response.SessionData.SessionID
}

func (h *SessionHandler) generateDeviceID(r *http.Request) string {
	// Simple device ID generation based on User-Agent and other headers
	userAgent := r.UserAgent()
	acceptLang := r.Header.Get("Accept-Language")
	acceptEnc := r.Header.Get("Accept-Encoding")

	deviceString := fmt.Sprintf("%s|%s|%s", userAgent, acceptLang, acceptEnc)

	// Create a simple hash
	hash := fmt.Sprintf("%x", deviceString)
	if len(hash) > 32 {
		hash = hash[:32]
	}

	return hash
}

func (h *SessionHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *SessionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	response := models.ErrorResponse{
		Error:   errorCode,
		Message: message,
		Code:    errorCode,
	}

	h.writeJSONResponse(w, statusCode, response)
}
