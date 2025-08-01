package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"session-service/models"

	"github.com/sirupsen/logrus"
)

// SessionManager handles basic session management
type SessionManager struct {
	// Core dependencies
	jwtManager *JWTManager
	logger     *logrus.Logger
	config     *models.SessionConfig

	// Storage
	storage SessionStorage

	// Synchronization
	mutex      sync.RWMutex
	cleanupMux sync.Mutex

	// Basic metrics
	metrics *SessionMetrics
}

// SessionStorage defines the interface for session storage backends
type SessionStorage interface {
	Store(sessionID string, session *models.SessionData) error
	Get(sessionID string) (*models.SessionData, error)
	GetByTokenHash(tokenHash string) (*models.SessionData, error)
	GetUserSessions(userID string) ([]*models.SessionData, error)
	Update(sessionID string, session *models.SessionData) error
	Delete(sessionID string) error
	DeleteUserSessions(userID string) error
	GetAllSessions() ([]*models.SessionData, error)
	Cleanup() error
}

// ExtendedSessionStorage defines additional methods available for some storage backends
type ExtendedSessionStorage interface {
	SessionStorage
	CleanupUserExpiredSessions(userID string) error
}

// SessionMetrics tracks basic session-related metrics
type SessionMetrics struct {
	TotalSessions  int64
	ActiveSessions int64
	LastCleanup    time.Time
	mutex          sync.RWMutex
}

// NewSessionManager creates a new session manager with database storage
func NewSessionManager(jwtManager *JWTManager, config *models.SessionConfig, storage SessionStorage, logger *logrus.Logger) *SessionManager {
	if config == nil {
		config = models.DefaultSessionConfig()
	}

	sm := &SessionManager{
		jwtManager: jwtManager,
		logger:     logger,
		config:     config,
		storage:    storage,
		metrics:    &SessionMetrics{},
	}

	// Start background cleanup process
	go sm.startCleanupProcess()

	logger.WithFields(logrus.Fields{
		"max_sessions":     config.MaxConcurrentSessions,
		"cleanup_interval": config.CleanupInterval,
		"storage_type":     "database",
	}).Info("Session manager initialized with database storage")

	return sm
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(req *models.SessionCreateRequest) (*models.SessionData, string, error) {
	// Clean up expired sessions for this user first (in background to avoid blocking)
	go sm.cleanupUserExpiredSessions(req.UserID)

	// Check concurrent session limits
	if err := sm.checkConcurrentSessions(req.UserID); err != nil {
		return nil, "", err
	}

	// Generate session ID and token
	sessionID := sm.generateSessionID()
	token, _, err := sm.jwtManager.GenerateToken(&models.UserProfile{
		User: models.User{
			ID:       req.UserID,
			Username: req.Username,
			RoleID:   req.RoleName, // Using RoleName for RoleID temporarily
		},
		Role: models.Role{
			RoleName: req.RoleName,
		},
		Permissions: make([]models.Permission, 0), // Convert from strings if needed
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session data
	now := time.Now()
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		if req.RememberMe {
			expiresAt = now.Add(sm.config.RememberMeExpiration)
		} else {
			expiresAt = now.Add(sm.config.DefaultExpiration)
		}
	}

	session := &models.SessionData{
		SessionID:    sessionID,
		UserID:       req.UserID,
		Username:     req.Username,
		RoleName:     req.RoleName,
		Permissions:  req.Permissions,
		TokenHash:    sm.hashToken(token),
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastActivity: now,
		IsActive:     true,
	}

	// Store session
	if err := sm.storage.Store(sessionID, session); err != nil {
		return nil, "", fmt.Errorf("failed to store session: %w", err)
	}

	// Update metrics
	sm.updateMetrics(func(m *SessionMetrics) {
		m.TotalSessions++
		m.ActiveSessions++
	})

	// Log session creation
	sm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    req.UserID,
		"username":   req.Username,
		"expires_at": expiresAt,
	}).Info("Session created successfully")

	return session, token, nil
}

// ValidateSession validates a token or session ID against stored sessions
func (sm *SessionManager) ValidateSession(req *models.SessionValidationRequest) (*models.SessionValidationResponse, error) {
	var session *models.SessionData
	var err error

	// Validate input parameters
	if req.Token == "" && req.SessionID == "" {
		return &models.SessionValidationResponse{
			IsValid:      false,
			ErrorCode:    "missing_parameter",
			ErrorMessage: "Either token or session_id must be provided",
		}, nil
	}

	// Get session from storage - prefer session ID for better performance
	if req.SessionID != "" {
		session, err = sm.storage.Get(req.SessionID)
	} else {
		tokenHash := sm.hashToken(req.Token)
		session, err = sm.storage.GetByTokenHash(tokenHash)
	}

	if err != nil {
		return &models.SessionValidationResponse{
			IsValid:      false,
			ErrorCode:    "session_not_found",
			ErrorMessage: "Session not found",
		}, nil
	}

	// Check session validity
	if !session.IsActive {
		return &models.SessionValidationResponse{
			IsValid:      false,
			ErrorCode:    "session_inactive",
			ErrorMessage: "Session is not active",
		}, nil
	}

	// Check expiration
	now := time.Now()
	if now.After(session.ExpiresAt) {
		sm.expireSession(session.SessionID)
		return &models.SessionValidationResponse{
			IsValid:      false,
			ErrorCode:    "session_expired",
			ErrorMessage: "Session has expired",
		}, nil
	}

	// Update session activity
	session.LastActivity = now
	sm.storage.Update(session.SessionID, session)

	// Check if token needs refresh
	response := &models.SessionValidationResponse{
		IsValid:     true,
		SessionData: session,
	}

	refreshTime := session.ExpiresAt.Add(-sm.config.RefreshThreshold)
	if now.After(refreshTime) {
		newToken, newExp, err := sm.refreshSessionToken(session)
		if err != nil {
			sm.logger.WithError(err).Warn("Failed to refresh token")
		} else {
			response.ShouldRefresh = true
			response.NewToken = newToken
			session.ExpiresAt = newExp
			sm.storage.Update(session.SessionID, session)
		}
	}

	return response, nil
}

// RevokeSession revokes a session or all sessions for a user
func (sm *SessionManager) RevokeSession(req *models.SessionRevokeRequest) error {
	if req.RevokeAll && req.UserID != "" {
		// Revoke all user sessions
		return sm.storage.DeleteUserSessions(req.UserID)
	}

	// Revoke single session
	var sessionID string

	if req.SessionID != "" {
		sessionID = req.SessionID
	} else if req.Token != "" {
		tokenHash := sm.hashToken(req.Token)
		session, err := sm.storage.GetByTokenHash(tokenHash)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}
		sessionID = session.SessionID
	} else {
		return fmt.Errorf("either session_id or token must be provided")
	}

	return sm.storage.Delete(sessionID)
}

// GetUserSessions returns a summary of user sessions for management
func (sm *SessionManager) GetUserSessions(userID string, currentSessionID string) ([]*models.SessionSummary, error) {
	sessions, err := sm.storage.GetUserSessions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	summaries := make([]*models.SessionSummary, 0, len(sessions))
	for _, session := range sessions {
		if !session.IsActive {
			continue
		}

		summary := &models.SessionSummary{
			SessionID:    session.SessionID,
			CreatedAt:    session.CreatedAt,
			LastActivity: session.LastActivity,
			IsActive:     session.IsActive,
			IsCurrent:    session.SessionID == currentSessionID,
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetSessionStats returns basic analytics about sessions
func (sm *SessionManager) GetSessionStats() *models.SessionStats {
	sm.metrics.mutex.RLock()
	defer sm.metrics.mutex.RUnlock()

	return &models.SessionStats{
		TotalSessions:   int(sm.metrics.TotalSessions),
		ActiveSessions:  int(sm.metrics.ActiveSessions),
		ExpiredSessions: 0, // Can be calculated if needed
	}
}

// Helper methods

func (sm *SessionManager) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (sm *SessionManager) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (sm *SessionManager) checkConcurrentSessions(userID string) error {
	// Use database-optimized count (only database storage is supported)
	dbStorage, ok := sm.storage.(*DatabaseSessionStorage)
	if !ok {
		return fmt.Errorf("unsupported storage type - only database storage is supported")
	}

	activeCount, err := dbStorage.CountUserActiveSessions(userID)
	if err != nil {
		return err
	}

	if activeCount >= sm.config.MaxConcurrentSessions {
		// Get user sessions to find oldest and remove it
		sessions, err := sm.storage.GetUserSessions(userID)
		if err != nil {
			return err
		}

		// Find oldest active session
		var oldestSession *models.SessionData
		now := time.Now()
		for _, session := range sessions {
			if session.IsActive && now.Before(session.ExpiresAt) {
				if oldestSession == nil || session.CreatedAt.Before(oldestSession.CreatedAt) {
					oldestSession = session
				}
			}
		}

		if oldestSession != nil {
			sm.storage.Delete(oldestSession.SessionID)
		}
	}

	return nil
}

func (sm *SessionManager) expireSession(sessionID string) {
	session, err := sm.storage.Get(sessionID)
	if err != nil {
		return
	}

	session.IsActive = false
	sm.storage.Update(sessionID, session)

	sm.updateMetrics(func(m *SessionMetrics) {
		m.ActiveSessions--
	})
}

func (sm *SessionManager) refreshSessionToken(session *models.SessionData) (string, time.Time, error) {
	// Create user profile for token generation
	profile := &models.UserProfile{
		User: models.User{
			ID:       session.UserID,
			Username: session.Username,
			RoleID:   session.RoleName,
		},
		Role: models.Role{
			RoleName: session.RoleName,
		},
	}

	newToken, newExp, err := sm.jwtManager.GenerateToken(profile)
	if err != nil {
		return "", time.Time{}, err
	}

	// Update session with new token
	session.TokenHash = sm.hashToken(newToken)

	return newToken, newExp, nil
}

func (sm *SessionManager) updateMetrics(fn func(*SessionMetrics)) {
	sm.metrics.mutex.Lock()
	defer sm.metrics.mutex.Unlock()
	fn(sm.metrics)
}

// Background processes

func (sm *SessionManager) startCleanupProcess() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.performCleanup()
	}
}

func (sm *SessionManager) performCleanup() {
	sm.cleanupMux.Lock()
	defer sm.cleanupMux.Unlock()

	// Clean up expired sessions
	sm.storage.Cleanup()

	sm.updateMetrics(func(m *SessionMetrics) {
		m.LastCleanup = time.Now()
	})

	sm.logger.Debug("Session cleanup completed")
}

// cleanupUserExpiredSessions performs user-specific cleanup in background
func (sm *SessionManager) cleanupUserExpiredSessions(userID string) {
	sm.logger.WithField("user_id", userID).Debug("Starting user-specific expired session cleanup")

	// Check if storage supports user-specific cleanup
	if extStorage, ok := sm.storage.(ExtendedSessionStorage); ok {
		if err := extStorage.CleanupUserExpiredSessions(userID); err != nil {
			sm.logger.WithError(err).WithField("user_id", userID).Warn("Failed to cleanup user expired sessions")
		} else {
			sm.logger.WithField("user_id", userID).Debug("User expired session cleanup completed successfully")
		}
	} else {
		// Fallback: perform general cleanup (should not happen with database storage)
		sm.logger.WithField("user_id", userID).Warn("Database storage doesn't implement ExtendedSessionStorage, using general cleanup")
		if err := sm.storage.Cleanup(); err != nil {
			sm.logger.WithError(err).Warn("Failed to perform general cleanup during user login")
		}
	}
}

// CleanupUserExpiredSessions provides a public method to trigger user-specific cleanup (for testing/admin purposes)
func (sm *SessionManager) CleanupUserExpiredSessions(userID string) error {
	if extStorage, ok := sm.storage.(ExtendedSessionStorage); ok {
		return extStorage.CleanupUserExpiredSessions(userID)
	}
	// Fallback to general cleanup (should not happen with database storage)
	sm.logger.Warn("Database storage doesn't implement ExtendedSessionStorage, using general cleanup")
	return sm.storage.Cleanup()
}

// Session manager now only supports database storage - memory storage has been removed
