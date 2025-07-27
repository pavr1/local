package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"auth-service/config"
	"auth-service/models"
	"auth-service/utils"

	"github.com/sirupsen/logrus"
)

// AuthHandler interface defines the authentication operations
type AuthHandler interface {
	// Authentication operations
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	ValidateToken(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)

	// Utility operations
	HealthCheck(w http.ResponseWriter, r *http.Request)
	GetTokenInfo(w http.ResponseWriter, r *http.Request)

	// Middleware
	AuthMiddleware(next http.Handler) http.Handler
	RequirePermission(permission string) func(http.Handler) http.Handler
}

// authHandler implements the AuthHandler interface
type authHandler struct {
	db              *sql.DB
	config          *config.Config
	logger          *logrus.Logger
	jwtManager      *utils.JWTManager
	passwordManager *utils.PasswordManager
}

// New creates a new auth handler instance
func New(db *sql.DB, cfg *config.Config, logger *logrus.Logger) AuthHandler {
	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpirationTime, logger)
	passwordManager := utils.NewPasswordManager(cfg.BcryptCost, logger)

	return &authHandler{
		db:              db,
		config:          cfg,
		logger:          logger,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

// Login authenticates a user and returns a JWT token
func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		h.logger.WithError(err).Warn("Invalid login request format")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Validate required fields
	if loginReq.Username == "" || loginReq.Password == "" {
		h.logger.Warn("Login attempt with missing credentials")
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_credentials", "Username and password are required")
		return
	}

	// Get user profile from database
	profile, err := h.getUserProfile(loginReq.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithField("username", loginReq.Username).Warn("Login attempt with non-existent user")
			h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password")
			return
		}
		h.logger.WithError(err).Error("Database error during login")
		h.writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Internal server error")
		return
	}

	// Check if user is active
	if !profile.User.IsActive {
		h.logger.WithField("username", loginReq.Username).Warn("Login attempt with inactive user")
		h.writeErrorResponse(w, http.StatusUnauthorized, "user_inactive", "User account is inactive")
		return
	}

	// Validate password
	if err := h.passwordManager.ValidatePassword(loginReq.Password, profile.User.PasswordHash); err != nil {
		h.logger.WithFields(logrus.Fields{
			"username": loginReq.Username,
			"user_id":  profile.User.ID,
		}).Warn("Login attempt with incorrect password")
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password")
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtManager.GenerateToken(profile)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate JWT token")
		h.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_error", "Failed to generate authentication token")
		return
	}

	// Update last login time
	if err := h.updateLastLogin(profile.User.ID); err != nil {
		h.logger.WithError(err).Warn("Failed to update last login time")
		// Don't fail the login for this
	}

	// Prepare response
	response := models.LoginResponse{
		User:        profile.User,
		Role:        profile.Role,
		Permissions: profile.Permissions,
		Token:       token,
		ExpiresAt:   expiresAt,
		RefreshAt:   expiresAt.Add(-h.config.JWTRefreshThreshold),
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    profile.User.ID,
		"username":   profile.User.Username,
		"role":       profile.Role.RoleName,
		"expires_at": expiresAt,
	}).Info("User logged in successfully")

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Logout invalidates a user's token (for now, just logs the logout)
func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Authorization token is required")
		return
	}

	// Validate token to get user info for logging
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		h.logger.WithError(err).Debug("Logout attempt with invalid token")
		// Still return success for logout even with invalid token
	} else {
		h.logger.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"username": claims.Username,
		}).Info("User logged out successfully")
	}

	// TODO: In production, implement token blacklisting
	response := models.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RefreshToken generates a new token if the current one is within refresh threshold
func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshReq models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshReq); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	newToken, expiresAt, err := h.jwtManager.RefreshToken(refreshReq.Token, h.config.JWTRefreshThreshold)
	if err != nil {
		h.logger.WithError(err).Debug("Token refresh failed")
		h.writeErrorResponse(w, http.StatusBadRequest, "refresh_failed", err.Error())
		return
	}

	response := models.LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		RefreshAt: expiresAt.Add(-h.config.JWTRefreshThreshold),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// ValidateToken validates a JWT token and returns user information
func (h *authHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Authorization token is required")
		return
	}

	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_token", err.Error())
		return
	}

	response := models.AuthStatus{
		IsAuthenticated: true,
		User: &models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			FullName: claims.FullName,
			RoleID:   claims.RoleID,
			IsActive: true, // Token validation implies active user
		},
		Role: &models.Role{
			ID:       claims.RoleID,
			RoleName: claims.RoleName,
		},
		ExpiresAt: claims.ExpiresAt.Time,
		RefreshAt: claims.ExpiresAt.Time.Add(-h.config.JWTRefreshThreshold),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetProfile returns the current user's profile
func (h *authHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*models.JWTClaims)

	profile, err := h.getUserProfileByID(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user profile")
		h.writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to get user profile")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, profile)
}

// HealthCheck returns the health status of the auth service
func (h *authHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Test database connection
	if err := h.db.Ping(); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "database_unavailable", "Database connection failed")
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Auth service is healthy",
		Data: map[string]interface{}{
			"service": "auth-service",
			"status":  "healthy",
			"time":    time.Now(),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTokenInfo returns detailed token information (for debugging/admin)
func (h *authHandler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Authorization token is required")
		return
	}

	tokenInfo := h.jwtManager.GetTokenInfo(token)
	h.writeJSONResponse(w, http.StatusOK, tokenInfo)
}

// Helper methods

// extractTokenFromHeader extracts the JWT token from the Authorization header
func (h *authHandler) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// writeJSONResponse writes a JSON response
func (h *authHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes an error response
func (h *authHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	response := models.ErrorResponse{
		Error:   code,
		Message: message,
		Code:    code,
	}
	h.writeJSONResponse(w, statusCode, response)
}

// Database helper methods

// getUserProfile retrieves a user's complete profile by username
func (h *authHandler) getUserProfile(username string) (*models.UserProfile, error) {
	h.logger.WithField("username", username).Debug("Retrieving user profile by username")

	// Query to get user with role information
	userQuery := `
		SELECT u.id, u.username, u.password_hash, u.full_name, u.role_id, u.is_active, 
		       u.last_login, u.created_at, u.updated_at,
		       r.id, r.role_name, r.description, r.created_at, r.updated_at
		FROM users u
		INNER JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 AND u.is_active = true`

	var user models.User
	var role models.Role
	var lastLogin sql.NullTime

	err := h.db.QueryRow(userQuery, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive,
		&lastLogin, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.RoleName, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithField("username", username).Debug("User not found")
			return nil, sql.ErrNoRows
		}
		h.logger.WithError(err).Error("Failed to query user profile")
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Handle nullable last_login
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	// Query to get user permissions
	permissions, err := h.getUserPermissions(user.RoleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user permissions")
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	profile := &models.UserProfile{
		User:        user,
		Role:        role,
		Permissions: permissions,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        role.RoleName,
		"permissions": len(permissions),
	}).Debug("User profile retrieved successfully")

	return profile, nil
}

// getUserProfileByID retrieves a user's complete profile by ID
func (h *authHandler) getUserProfileByID(userID string) (*models.UserProfile, error) {
	h.logger.WithField("user_id", userID).Debug("Retrieving user profile by ID")

	// Query to get user with role information
	userQuery := `
		SELECT u.id, u.username, u.password_hash, u.full_name, u.role_id, u.is_active, 
		       u.last_login, u.created_at, u.updated_at,
		       r.id, r.role_name, r.description, r.created_at, r.updated_at
		FROM users u
		INNER JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active = true`

	var user models.User
	var role models.Role
	var lastLogin sql.NullTime

	err := h.db.QueryRow(userQuery, userID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive,
		&lastLogin, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.RoleName, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithField("user_id", userID).Debug("User not found")
			return nil, sql.ErrNoRows
		}
		h.logger.WithError(err).Error("Failed to query user profile")
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Handle nullable last_login
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	// Query to get user permissions
	permissions, err := h.getUserPermissions(user.RoleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user permissions")
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	profile := &models.UserProfile{
		User:        user,
		Role:        role,
		Permissions: permissions,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        role.RoleName,
		"permissions": len(permissions),
	}).Debug("User profile retrieved successfully")

	return profile, nil
}

// getUserPermissions retrieves all permissions for a given role
func (h *authHandler) getUserPermissions(roleID string) ([]models.Permission, error) {
	h.logger.WithField("role_id", roleID).Debug("Retrieving permissions for role")

	permissionsQuery := `
		SELECT id, permission_name, description, role_id, created_at, updated_at
		FROM permissions
		WHERE role_id = $1
		ORDER BY permission_name`

	rows, err := h.db.Query(permissionsQuery, roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to query permissions")
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		err := rows.Scan(
			&perm.ID, &perm.PermissionName, &perm.Description, &perm.RoleID,
			&perm.CreatedAt, &perm.UpdatedAt,
		)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan permission row")
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	if err := rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error iterating permission rows")
		return nil, fmt.Errorf("error reading permissions: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"role_id":           roleID,
		"permissions_count": len(permissions),
	}).Debug("Permissions retrieved successfully")

	return permissions, nil
}

// updateLastLogin updates the user's last login timestamp
func (h *authHandler) updateLastLogin(userID string) error {
	h.logger.WithField("user_id", userID).Debug("Updating last login timestamp")

	updateQuery := `
		UPDATE users 
		SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := h.db.Exec(updateQuery, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update last login")
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Warn("Could not get rows affected for last login update")
		// Don't fail for this
	} else if rowsAffected == 0 {
		h.logger.WithField("user_id", userID).Warn("No rows affected when updating last login")
		return fmt.Errorf("user not found for last login update")
	}

	h.logger.WithField("user_id", userID).Debug("Last login timestamp updated successfully")
	return nil
}

// AuthMiddleware validates JWT tokens and adds user context to requests
func (h *authHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := h.extractTokenFromHeader(r)
		if token == "" {
			h.writeErrorResponse(w, http.StatusUnauthorized, "missing_token", "Authorization token is required")
			return
		}

		// Validate token
		claims, err := h.jwtManager.ValidateToken(token)
		if err != nil {
			h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_token", err.Error())
			return
		}

		// Add user claims to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.RoleName)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission creates middleware that checks for a specific permission
func (h *authHandler) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context (should be set by AuthMiddleware)
			claims, ok := r.Context().Value("user").(*models.JWTClaims)
			if !ok {
				h.writeErrorResponse(w, http.StatusUnauthorized, "missing_auth_context", "Authentication context is missing")
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, userPerm := range claims.Permissions {
				if userPerm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				h.logger.WithFields(logrus.Fields{
					"user_id":             claims.UserID,
					"username":            claims.Username,
					"required_permission": permission,
					"user_permissions":    claims.Permissions,
				}).Warn("Access denied: insufficient permissions")

				h.writeErrorResponse(w, http.StatusForbidden, "insufficient_permissions",
					fmt.Sprintf("Required permission '%s' not found", permission))
				return
			}

			// User has permission, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
