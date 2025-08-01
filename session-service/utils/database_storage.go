package utils

import (
	"database/sql"
	"fmt"

	"session-service/models"
	sessionSQL "session-service/sql"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// DatabaseSessionStorage implements SessionStorage interface using PostgreSQL
type DatabaseSessionStorage struct {
	db      *sql.DB
	queries sessionSQL.Queries
	logger  *logrus.Logger
}

// NewDatabaseSessionStorage creates a new database-based session storage
func NewDatabaseSessionStorage(db *sql.DB, logger *logrus.Logger) (*DatabaseSessionStorage, error) {
	queries, err := sessionSQL.LoadQueries()
	if err != nil {
		return nil, fmt.Errorf("failed to load SQL queries: %w", err)
	}

	storage := &DatabaseSessionStorage{
		db:      db,
		queries: queries,
		logger:  logger,
	}

	logger.Info("Database session storage initialized successfully")
	return storage, nil
}

// Store saves a session in the database
func (s *DatabaseSessionStorage) Store(sessionID string, session *models.SessionData) error {
	query, err := s.queries.Get("insert_session")
	if err != nil {
		return fmt.Errorf("failed to get insert query: %w", err)
	}

	_, err = s.db.Exec(query,
		session.SessionID,
		session.UserID,
		session.Username,
		session.RoleName,
		pq.Array(session.Permissions), // Convert to PostgreSQL array
		session.TokenHash,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
		session.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    session.UserID,
	}).Debug("Session stored in database")

	return nil
}

// Get retrieves a session by session ID
func (s *DatabaseSessionStorage) Get(sessionID string) (*models.SessionData, error) {
	query, err := s.queries.Get("get_session_by_id")
	if err != nil {
		return nil, fmt.Errorf("failed to get session query: %w", err)
	}

	session := &models.SessionData{}
	var permissions pq.StringArray

	err = s.db.QueryRow(query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.RoleName,
		&permissions,
		&session.TokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.Permissions = []string(permissions)
	return session, nil
}

// GetByTokenHash retrieves a session by token hash
func (s *DatabaseSessionStorage) GetByTokenHash(tokenHash string) (*models.SessionData, error) {
	query, err := s.queries.Get("get_session_by_token_hash")
	if err != nil {
		return nil, fmt.Errorf("failed to get session query: %w", err)
	}

	session := &models.SessionData{}
	var permissions pq.StringArray

	err = s.db.QueryRow(query, tokenHash).Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.RoleName,
		&permissions,
		&session.TokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	session.Permissions = []string(permissions)

	// Log debug info about retrieved session for troubleshooting
	s.logger.WithFields(logrus.Fields{
		"session_id":    session.SessionID,
		"user_id":       session.UserID,
		"username":      session.Username,
		"created_at":    session.CreatedAt,
		"expires_at":    session.ExpiresAt,
		"last_activity": session.LastActivity,
	}).Debug("Retrieved session by token hash (most recent)")

	return session, nil
}

// GetUserSessions returns all sessions for a user
func (s *DatabaseSessionStorage) GetUserSessions(userID string) ([]*models.SessionData, error) {
	query, err := s.queries.Get("get_user_sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions query: %w", err)
	}

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.SessionData

	for rows.Next() {
		session := &models.SessionData{}
		var permissions pq.StringArray

		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.Username,
			&session.RoleName,
			&permissions,
			&session.TokenHash,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.LastActivity,
			&session.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		session.Permissions = []string(permissions)
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// Update updates a session in the database
func (s *DatabaseSessionStorage) Update(sessionID string, session *models.SessionData) error {
	query, err := s.queries.Get("update_session_activity")
	if err != nil {
		return fmt.Errorf("failed to get update query: %w", err)
	}

	_, err = s.db.Exec(query,
		sessionID,
		session.LastActivity,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
	}).Debug("Session updated in database")

	return nil
}

// Delete deactivates a session (soft delete)
func (s *DatabaseSessionStorage) Delete(sessionID string) error {
	query, err := s.queries.Get("deactivate_session")
	if err != nil {
		return fmt.Errorf("failed to get deactivate query: %w", err)
	}

	_, err = s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
	}).Info("Session deactivated")

	return nil
}

// DeleteUserSessions deactivates all sessions for a user
func (s *DatabaseSessionStorage) DeleteUserSessions(userID string) error {
	query, err := s.queries.Get("deactivate_user_sessions")
	if err != nil {
		return fmt.Errorf("failed to get deactivate user sessions query: %w", err)
	}

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get rows affected count")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"rows_affected": rowsAffected,
	}).Info("User sessions deactivated")

	return nil
}

// GetAllSessions returns all sessions (for admin purposes)
func (s *DatabaseSessionStorage) GetAllSessions() ([]*models.SessionData, error) {
	// For now, just return user sessions without specific user
	// This could be optimized with a separate query if needed
	return nil, fmt.Errorf("get all sessions not implemented for database storage")
}

// Cleanup removes expired sessions
func (s *DatabaseSessionStorage) Cleanup() error {
	query, err := s.queries.Get("cleanup_expired_sessions")
	if err != nil {
		return fmt.Errorf("failed to get cleanup query: %w", err)
	}

	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get cleanup rows affected count")
	} else {
		s.logger.WithFields(logrus.Fields{
			"expired_sessions": rowsAffected,
		}).Info("Expired sessions cleaned up")
	}

	return nil
}

// CountUserActiveSessions counts active sessions for a user (for concurrent session limits)
func (s *DatabaseSessionStorage) CountUserActiveSessions(userID string) (int, error) {
	query, err := s.queries.Get("count_user_active_sessions")
	if err != nil {
		return 0, fmt.Errorf("failed to get count query: %w", err)
	}

	var count int
	err = s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user active sessions: %w", err)
	}

	return count, nil
}

// CleanupUserExpiredSessions removes expired sessions for a specific user
func (s *DatabaseSessionStorage) CleanupUserExpiredSessions(userID string) error {
	query, err := s.queries.Get("cleanup_user_expired_sessions")
	if err != nil {
		return fmt.Errorf("failed to get user cleanup query: %w", err)
	}

	result, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to cleanup user expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user cleanup rows affected count")
	} else if rowsAffected > 0 {
		s.logger.WithFields(logrus.Fields{
			"user_id":          userID,
			"expired_sessions": rowsAffected,
		}).Info("User expired sessions cleaned up")
	}

	return nil
}
