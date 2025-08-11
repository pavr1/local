package handlers

import (
	"database/sql"
	"fmt"
	"session-service/config"
	"session-service/entities/sessions/models"
	sessionSQL "session-service/entities/sessions/sql"

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
func NewDBHandler(cfg *config.Config, jwtHandler *JWTHandler, logger *logrus.Logger) (*DBHandler, error) {
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
func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabaseSSLMode)

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
	tokenHash := h.jwtHandler.GenerateTokenHash(tokenString)

	// Store session in database
	err = h.storeSession(sessionID, tokenHash)
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

	return &models.SessionCreateResponse{
		SessionID: sessionID,
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
func (h *DBHandler) storeSession(sessionID, tokenHash string) error {
	query, err := h.queries.Get("create_session")
	if err != nil {
		return fmt.Errorf("failed to get create session query: %w", err)
	}

	_, err = h.db.Exec(query, sessionID, tokenHash)
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
