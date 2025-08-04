package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"session-service/models"
	"session-service/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SessionAPI handles REST API endpoints for session management
type SessionAPI struct {
	sessionHandler *SessionHandler
	logger         *logrus.Logger
	jwtManager     *utils.JWTManager
	db             *sql.DB
}

// NewSessionAPI creates a new session API handler
func NewSessionAPI(sessionManager *utils.SessionManager, jwtManager *utils.JWTManager, db *sql.DB, logger *logrus.Logger) *SessionAPI {
	return &SessionAPI{
		sessionHandler: NewSessionHandler(sessionManager, jwtManager, logger),
		logger:         logger,
		jwtManager:     jwtManager,
		db:             db,
	}
}

// CreateSession creates a new session (called by gateway during login)
func (api *SessionAPI) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req models.SessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Username == "" || req.RoleName == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_fields", "UserID, Username, and RoleName are required")
		return
	}

	session, token, err := api.sessionHandler.sessionManager.CreateSession(&req)
	if err != nil {
		api.logger.WithError(err).Error("Failed to create session")
		api.writeErrorResponse(w, http.StatusInternalServerError, "session_creation_failed", "Failed to create session")
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Session created successfully",
		"session_id": session.SessionID,
		"token":      token,
		"expires_at": session.ExpiresAt,
		"user": map[string]interface{}{
			"id":       session.UserID,
			"username": session.Username,
			"role":     session.RoleName,
		},
	}

	api.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"username":   session.Username,
	}).Info("Session created via API")

	api.writeJSONResponse(w, http.StatusCreated, response)
}

// ValidateSession validates a session token
func (api *SessionAPI) ValidateSession(w http.ResponseWriter, r *http.Request) {
	var req models.SessionValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	response, err := api.sessionHandler.sessionManager.ValidateSession(&req)
	if err != nil {
		api.logger.WithError(err).Error("Session validation failed")
		api.writeErrorResponse(w, http.StatusInternalServerError, "validation_error", "Session validation failed")
		return
	}

	// Add additional context for valid sessions
	if response.IsValid && response.SessionData != nil {
		responseData := map[string]interface{}{
			"is_valid": true,
			"session": map[string]interface{}{
				"session_id":    response.SessionData.SessionID,
				"user_id":       response.SessionData.UserID,
				"username":      response.SessionData.Username,
				"role_name":     response.SessionData.RoleName,
				"permissions":   response.SessionData.Permissions,
				"created_at":    response.SessionData.CreatedAt,
				"expires_at":    response.SessionData.ExpiresAt,
				"last_activity": response.SessionData.LastActivity,
			},
			"should_refresh": response.ShouldRefresh,
		}

		if response.NewToken != "" {
			responseData["new_token"] = response.NewToken
		}

		api.writeJSONResponse(w, http.StatusOK, responseData)
	} else {
		api.writeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"is_valid":      false,
			"error_code":    response.ErrorCode,
			"error_message": response.ErrorMessage,
		})
	}
}

// RefreshSession refreshes a session token
func (api *SessionAPI) RefreshSession(w http.ResponseWriter, r *http.Request) {
	api.sessionHandler.RefreshSession(w, r)
}

// GetUserSessions returns all sessions for a user
func (api *SessionAPI) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_user_id", "User ID is required")
		return
	}

	// Get current session ID from token if provided
	currentSessionID := api.getCurrentSessionIDFromToken(r)

	sessions, err := api.sessionHandler.sessionManager.GetUserSessions(userID, currentSessionID)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get user sessions")
		api.writeErrorResponse(w, http.StatusInternalServerError, "fetch_error", "Failed to retrieve sessions")
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"user_id":  userID,
		"sessions": sessions,
		"count":    len(sessions),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

// RevokeSession revokes a specific session
func (api *SessionAPI) RevokeSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionID"]

	if sessionID == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_session_id", "Session ID is required")
		return
	}

	err := api.sessionHandler.sessionManager.RevokeSession(&models.SessionRevokeRequest{
		SessionID: sessionID,
	})

	if err != nil {
		api.logger.WithError(err).Error("Failed to revoke session")
		api.writeErrorResponse(w, http.StatusInternalServerError, "revoke_error", "Failed to revoke session")
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Session revoked successfully",
		"session_id": sessionID,
	}

	api.logger.WithField("session_id", sessionID).Info("Session revoked via API")
	api.writeJSONResponse(w, http.StatusOK, response)
}

// RevokeAllUserSessions revokes all sessions for a user
func (api *SessionAPI) RevokeAllUserSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_user_id", "User ID is required")
		return
	}

	// Check if we should exclude current session
	excludeCurrent := r.URL.Query().Get("exclude_current") == "true"
	currentSessionID := ""
	if excludeCurrent {
		currentSessionID = api.getCurrentSessionIDFromToken(r)
	}

	// Get sessions first to count them
	sessions, err := api.sessionHandler.sessionManager.GetUserSessions(userID, currentSessionID)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get user sessions for revocation")
		api.writeErrorResponse(w, http.StatusInternalServerError, "fetch_error", "Failed to retrieve sessions")
		return
	}

	// Revoke each session except current if excluded
	revokedCount := 0
	for _, session := range sessions {
		if excludeCurrent && session.SessionID == currentSessionID {
			continue
		}

		err := api.sessionHandler.sessionManager.RevokeSession(&models.SessionRevokeRequest{
			SessionID: session.SessionID,
		})
		if err != nil {
			api.logger.WithError(err).Warn("Failed to revoke session during bulk revocation")
		} else {
			revokedCount++
		}
	}

	response := map[string]interface{}{
		"success":        true,
		"message":        "User sessions revoked successfully",
		"user_id":        userID,
		"revoked_count":  revokedCount,
		"total_sessions": len(sessions),
	}

	api.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"revoked_count": revokedCount,
	}).Info("All user sessions revoked via API")

	api.writeJSONResponse(w, http.StatusOK, response)
}

// RevokeSessionByToken revokes a session by token (for logout)
func (api *SessionAPI) RevokeSessionByToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	if req.Token == "" {
		// Try to get token from Authorization header
		req.Token = api.extractTokenFromHeader(r)
	}

	if req.Token == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Token is required")
		return
	}

	err := api.sessionHandler.sessionManager.RevokeSession(&models.SessionRevokeRequest{
		Token: req.Token,
	})

	if err != nil {
		api.logger.WithError(err).Warn("Failed to revoke session by token")
		// Don't fail logout if session revocation fails
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Session revoked successfully",
	}

	api.logger.Info("Session revoked by token via API")
	api.writeJSONResponse(w, http.StatusOK, response)
}

// GetSessionStats returns session statistics
func (api *SessionAPI) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	stats := api.sessionHandler.sessionManager.GetSessionStats()

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

// HealthCheck returns the health status of the session service
func (api *SessionAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost:8086/health")
	if err != nil {
		api.logger.WithError(err).Error("Failed to connect to data-service during health check")
		response := map[string]interface{}{
			"success": false,
			"service": "session-service",
			"status":  "unhealthy",
			"message": "Data service connection failed",
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		api.logger.WithField("status_code", resp.StatusCode).Error("Data service health check failed")
		response := map[string]interface{}{
			"success": false,
			"service": "session-service",
			"status":  "unhealthy",
			"message": "Data service is unhealthy",
			"error":   fmt.Sprintf("Data service returned status %d", resp.StatusCode),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"service": "session-service",
		"status":  "healthy",
		"message": "Session service is operational",
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (api *SessionAPI) getCurrentSessionIDFromToken(r *http.Request) string {
	token := api.extractTokenFromHeader(r)
	if token == "" {
		return ""
	}

	// Validate session to get session ID
	req := &models.SessionValidationRequest{
		Token: token,
	}

	response, err := api.sessionHandler.sessionManager.ValidateSession(req)
	if err != nil || !response.IsValid || response.SessionData == nil {
		return ""
	}

	return response.SessionData.SessionID
}

func (api *SessionAPI) extractTokenFromHeader(r *http.Request) string {
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

func (api *SessionAPI) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (api *SessionAPI) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	response := models.ErrorResponse{
		Error:   errorCode,
		Message: message,
		Code:    errorCode,
	}

	api.writeJSONResponse(w, statusCode, response)
}

// authenticateUser validates user credentials against the database
func (api *SessionAPI) authenticateUser(username, password string) (*models.UserProfile, error) {
	// Query to get user with role information
	query := `
		SELECT u.id, u.username, u.password_hash, u.full_name, u.role_id, u.is_active,
		       r.id as role_id, r.role_name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 AND u.is_active = true
	`

	var user models.User
	var role models.Role
	var passwordHash string

	err := api.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &role.ID, &role.RoleName,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Verify password using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, nil // Invalid password
	}

	// Get user permissions
	permQuery := `
		SELECT permission_name, description
		FROM permissions
		WHERE role_id = $1
	`

	rows, err := api.db.Query(permQuery, role.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.PermissionName, &perm.Description); err != nil {
			return nil, err
		}
		perm.RoleID = role.ID
		permissions = append(permissions, perm)
	}

	return &models.UserProfile{
		User:        user,
		Role:        role,
		Permissions: permissions,
	}, nil
}

// Login handles user authentication (database-backed implementation)
func (api *SessionAPI) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		api.writeErrorResponse(w, http.StatusBadRequest, "missing_credentials", "Username and password are required")
		return
	}

	// Authenticate user against database
	profile, err := api.authenticateUser(req.Username, req.Password)
	if err != nil {
		api.logger.WithError(err).Warn("Authentication failed for user: " + req.Username)
		api.writeErrorResponse(w, http.StatusUnauthorized, "authentication_failed", "Invalid username or password")
		return
	}

	if profile != nil {

		// Create session properly using SessionManager
		session, token, err := api.sessionHandler.CreateSessionFromLogin(profile, r, false)
		if err != nil {
			api.logger.WithError(err).Error("Failed to create session")
			api.writeErrorResponse(w, http.StatusInternalServerError, "session_creation_failed", "Failed to create session")
			return
		}

		// Return response in expected format with session ID
		response := models.LoginResponse{
			User:      profile.User,
			Role:      profile.Role,
			Token:     token,
			SessionID: session.SessionID,
		}

		api.writeJSONResponse(w, http.StatusOK, response)
		return
	}

	// Invalid credentials
	api.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password")
}
