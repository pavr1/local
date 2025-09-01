package handlers

import (
	"database/sql"
	"fmt"
	"session-service/entities/sessions/models"
	sessionSQL "session-service/entities/sessions/sql"
	sharedConfig "shared/config"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// DBHandler handles database operations for sessions
type DBHandler struct {
	db         *sql.DB
	queries    sessionSQL.Queries
	jwtHandler *JWTHandler
	logger     *logrus.Logger
}

// NewDBHandler creates a new database handler with internal database connection
func NewDBHandler(cfg *sharedConfig.Config, jwtHandler *JWTHandler, logger *logrus.Logger) (*DBHandler, error) {
	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Load SQL queries
	queries, err := sessionSQL.LoadQueries()
	if err != nil {
		db.Close() // Close connection if queries fail to load
		return nil, fmt.Errorf("failed to load SQL queries: %w", err)
	}

	return &DBHandler{
		db:         db,
		queries:    *queries,
		jwtHandler: jwtHandler,
		logger:     logger,
	}, nil
}

// Close closes the database connection
func (h *DBHandler) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// connectToDatabase connects to the PostgreSQL database
func connectToDatabase(cfg *sharedConfig.Config, logger *logrus.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.GetString("DB_HOST", "localhost"),
		cfg.GetString("DB_PORT", "5432"),
		cfg.GetString("DB_USER", "postgres"),
		cfg.GetString("DB_PASSWORD", "postgres123"),
		cfg.GetString("DB_NAME", "icecream_store"),
		cfg.GetString("DB_SSL_MODE", "disable"))

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")
	return db, nil
}

// CreateSession creates a new session for a user
func (h *DBHandler) CreateSession(req *models.SessionCreateRequest) (*models.SessionCreateResponse, error) {
	// Authenticate user
	userProfile, err := h.authenticateUser(req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Generate session ID using JWT handler
	sessionID, err := h.jwtHandler.GenerateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate JWT token using JWT handler
	tokenString, err := h.jwtHandler.GenerateToken(sessionID, userProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Generate token hash using JWT handler
	// Store the full token for development/debugging purposes

	// Store session in database
	err = h.storeSession(sessionID, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Update last login
	err = h.updateLastLogin(userProfile.User.ID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to update last login")
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"username":   userProfile.User.Username,
		"user_id":    userProfile.User.ID,
	}).Info("Session created successfully")

	// Convert permissions to string slice for JSON response
	var permissionNames []string
	for _, perm := range userProfile.Permissions {
		permissionNames = append(permissionNames, perm.PermissionName)
	}

	return &models.SessionCreateResponse{
		SessionID:   sessionID,
		Message:     "Session created successfully",
		User:        &userProfile.User,
		Role:        &userProfile.Role,
		Permissions: permissionNames,
	}, nil
}

// authenticateUser validates user credentials and returns user profile
func (h *DBHandler) authenticateUser(username, password string) (*models.UserProfile, error) {
	// Get user and role information
	query, err := h.queries.Get("get_user_by_username")
	if err != nil {
		return nil, fmt.Errorf("failed to get user query: %w", err)
	}

	var user models.User
	var role models.Role
	var passwordHash string

	err = h.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.RoleName, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// Get user permissions
	permissions, err := h.getUserPermissions(role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return &models.UserProfile{
		User:        user,
		Role:        role,
		Permissions: permissions,
	}, nil
}

// getUserPermissions retrieves permissions for a role
func (h *DBHandler) getUserPermissions(roleID string) ([]models.Permission, error) {
	query, err := h.queries.Get("get_user_permissions")
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions query: %w", err)
	}

	rows, err := h.db.Query(query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(
			&perm.ID, &perm.PermissionName, &perm.Description,
			&perm.RoleID, &perm.CreatedAt, &perm.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// storeSession stores the session in the database
func (h *DBHandler) storeSession(sessionID, token string) error {
	query, err := h.queries.Get("create_session")
	if err != nil {
		return fmt.Errorf("failed to get create session query: %w", err)
	}

	_, err = h.db.Exec(query, sessionID, token)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// updateLastLogin updates the user's last login timestamp
func (h *DBHandler) updateLastLogin(userID string) error {
	query, err := h.queries.Get("update_last_login")
	if err != nil {
		return fmt.Errorf("failed to get update last login query: %w", err)
	}

	_, err = h.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// ValidateSession validates a session and handles expiration/renewal
func (h *DBHandler) ValidateSession(sessionID string) (*models.SessionValidationResponse, error) {
	// Get session from database
	session, err := h.getSessionByID(sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.SessionValidationResponse{
				Valid:   false,
				Message: "Session not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Validate token and extract claims
	claims, err := h.jwtHandler.ValidateToken(session.Token)
	if err != nil {
		// Token is invalid, delete the session
		if deleteErr := h.deleteSession(sessionID); deleteErr != nil {
			h.logger.Errorf("Failed to delete invalid session %s: %v", sessionID, deleteErr)
		}
		return &models.SessionValidationResponse{
			Valid:   false,
			Message: "Invalid token",
		}, nil
	}

	// Check if token is expired by checking the expiration time
	if time.Now().After(claims.ExpiresAt.Time) {
		// Token is expired, delete the session
		if deleteErr := h.deleteSession(sessionID); deleteErr != nil {
			h.logger.Errorf("Failed to delete expired session %s: %v", sessionID, deleteErr)
		}
		return &models.SessionValidationResponse{
			Valid:   false,
			Message: "Session expired",
		}, nil
	}

	// Additional validation: Check if token is about to expire (within 5 minutes)
	// If so, we'll renew it proactively
	shouldRenew := time.Until(claims.ExpiresAt.Time) < 5*time.Minute

	if shouldRenew {
		// Token is valid but expiring soon, renew it
		// Create a new user profile for token generation
		userProfile := &models.UserProfile{
			User: models.User{
				ID:       claims.UserID,
				Username: claims.Username,
			},
			Role: models.Role{
				RoleName: claims.RoleName,
			},
			Permissions: make([]models.Permission, len(claims.Permissions)),
		}

		// Convert permissions back to Permission structs
		for i, permName := range claims.Permissions {
			userProfile.Permissions[i] = models.Permission{
				PermissionName: permName,
			}
		}

		newToken, err := h.jwtHandler.GenerateToken(sessionID, userProfile)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new token: %w", err)
		}

		// Update session with new token
		if err := h.updateSessionToken(sessionID, newToken); err != nil {
			return nil, fmt.Errorf("failed to update session token: %w", err)
		}

		return &models.SessionValidationResponse{
			Valid:       true,
			SessionID:   sessionID,
			Message:     "Session validated and renewed",
			UserID:      claims.UserID,
			Username:    claims.Username,
			RoleName:    claims.RoleName,
			Permissions: claims.Permissions,
		}, nil
	}

	// Token is valid and not expiring soon, just return validation
	return &models.SessionValidationResponse{
		Valid:       true,
		SessionID:   sessionID,
		Message:     "Session validated",
		UserID:      claims.UserID,
		Username:    claims.Username,
		RoleName:    claims.RoleName,
		Permissions: claims.Permissions,
	}, nil
}

// getSessionByID retrieves a session by ID
func (h *DBHandler) getSessionByID(sessionID string) (*models.Session, error) {
	query, err := h.queries.Get(sessionSQL.GetSessionByIDQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}

	var session models.Session
	err = h.db.QueryRow(query, sessionID).Scan(&session.SessionID, &session.Token)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// deleteSession deletes a session by ID
func (h *DBHandler) deleteSession(sessionID string) error {
	query, err := h.queries.Get(sessionSQL.DeleteSessionQuery)
	if err != nil {
		return fmt.Errorf("failed to get query: %w", err)
	}

	_, err = h.db.Exec(query, sessionID)
	return err
}

// updateSessionToken updates the token hash for a session
func (h *DBHandler) updateSessionToken(sessionID, token string) error {
	query, err := h.queries.Get(sessionSQL.UpdateSessionTokenQuery)
	if err != nil {
		return fmt.Errorf("failed to get query: %w", err)
	}

	_, err = h.db.Exec(query, sessionID, token)
	return err
}

// DeleteSession deletes a session (logout functionality)
func (h *DBHandler) DeleteSession(sessionID string) (*models.SessionLogoutResponse, error) {
	// First check if session exists
	session, err := h.getSessionByID(sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.SessionLogoutResponse{
				Success:   false,
				SessionID: sessionID,
				Message:   "Session not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Validate token to get user information for logging
	claims, err := h.jwtHandler.ValidateToken(session.Token)
	if err != nil {
		// Token is invalid, but we'll still try to delete the session
		h.logger.Warnf("Invalid token for session %s during logout, proceeding with deletion", sessionID)
	} else {
		h.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    claims.UserID,
			"username":   claims.Username,
		}).Info("User logging out")
	}

	// Delete the session
	if err := h.deleteSession(sessionID); err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &models.SessionLogoutResponse{
		Success:   true,
		SessionID: sessionID,
		Message:   "Session successfully logged out",
	}, nil
}
