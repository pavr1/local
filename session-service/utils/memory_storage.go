package utils

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"session-service/models"
)

// MemorySessionStorage implements SessionStorage interface using in-memory storage
type MemorySessionStorage struct {
	sessions   map[string]*models.SessionData // sessionID -> session
	tokenIndex map[string]string              // tokenHash -> sessionID
	userIndex  map[string][]string            // userID -> []sessionID
	mutex      sync.RWMutex
}

// NewMemorySessionStorage creates a new memory-based session storage
func NewMemorySessionStorage() *MemorySessionStorage {
	return &MemorySessionStorage{
		sessions:   make(map[string]*models.SessionData),
		tokenIndex: make(map[string]string),
		userIndex:  make(map[string][]string),
	}
}

// Store saves a session in memory
func (s *MemorySessionStorage) Store(sessionID string, session *models.SessionData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store session
	s.sessions[sessionID] = session

	// Update token index
	s.tokenIndex[session.TokenHash] = sessionID

	// Update user index
	userSessions := s.userIndex[session.UserID]
	userSessions = append(userSessions, sessionID)
	s.userIndex[session.UserID] = userSessions

	return nil
}

// Get retrieves a session by session ID
func (s *MemorySessionStorage) Get(sessionID string) (*models.SessionData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Return a copy to prevent external modifications
	return s.copySession(session), nil
}

// GetByTokenHash retrieves a session by token hash
func (s *MemorySessionStorage) GetByTokenHash(tokenHash string) (*models.SessionData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessionID, exists := s.tokenIndex[tokenHash]
	if !exists {
		return nil, fmt.Errorf("session not found for token")
	}

	session, exists := s.sessions[sessionID]
	if !exists {
		// Clean up orphaned token index entry
		delete(s.tokenIndex, tokenHash)
		return nil, fmt.Errorf("session not found")
	}

	return s.copySession(session), nil
}

// GetUserSessions retrieves all sessions for a user
func (s *MemorySessionStorage) GetUserSessions(userID string) ([]*models.SessionData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessionIDs, exists := s.userIndex[userID]
	if !exists {
		return []*models.SessionData{}, nil
	}

	sessions := make([]*models.SessionData, 0, len(sessionIDs))
	validSessionIDs := make([]string, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[sessionID]; exists {
			sessions = append(sessions, s.copySession(session))
			validSessionIDs = append(validSessionIDs, sessionID)
		}
	}

	// Update user index to remove orphaned session IDs
	s.userIndex[userID] = validSessionIDs

	// Sort sessions by creation time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	return sessions, nil
}

// Update modifies an existing session
func (s *MemorySessionStorage) Update(sessionID string, session *models.SessionData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	existingSession, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Update token index if token hash changed
	if existingSession.TokenHash != session.TokenHash {
		// Remove old token index entry
		delete(s.tokenIndex, existingSession.TokenHash)
		// Add new token index entry
		s.tokenIndex[session.TokenHash] = sessionID
	}

	// Store updated session
	s.sessions[sessionID] = session

	return nil
}

// Delete removes a session
func (s *MemorySessionStorage) Delete(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Remove from sessions map
	delete(s.sessions, sessionID)

	// Remove from token index
	delete(s.tokenIndex, session.TokenHash)

	// Remove from user index
	userSessions := s.userIndex[session.UserID]
	for i, sid := range userSessions {
		if sid == sessionID {
			// Remove element at index i
			s.userIndex[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
			break
		}
	}

	// Clean up empty user index entries
	if len(s.userIndex[session.UserID]) == 0 {
		delete(s.userIndex, session.UserID)
	}

	return nil
}

// DeleteUserSessions removes all sessions for a user
func (s *MemorySessionStorage) DeleteUserSessions(userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessionIDs, exists := s.userIndex[userID]
	if !exists {
		return nil // No sessions to delete
	}

	// Remove all user sessions
	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[sessionID]; exists {
			// Remove from sessions map
			delete(s.sessions, sessionID)
			// Remove from token index
			delete(s.tokenIndex, session.TokenHash)
		}
	}

	// Remove user index entry
	delete(s.userIndex, userID)

	return nil
}

// GetAllSessions retrieves all sessions (for admin/metrics purposes)
func (s *MemorySessionStorage) GetAllSessions() ([]*models.SessionData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessions := make([]*models.SessionData, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, s.copySession(session))
	}

	return sessions, nil
}

// Cleanup removes expired and inactive sessions
func (s *MemorySessionStorage) Cleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	expiredSessions := make([]string, 0)

	// Find expired sessions
	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) || !session.IsActive {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	// Remove expired sessions
	for _, sessionID := range expiredSessions {
		session := s.sessions[sessionID]

		// Remove from sessions map
		delete(s.sessions, sessionID)

		// Remove from token index
		delete(s.tokenIndex, session.TokenHash)

		// Remove from user index
		userSessions := s.userIndex[session.UserID]
		for i, sid := range userSessions {
			if sid == sessionID {
				s.userIndex[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
				break
			}
		}

		// Clean up empty user index entries
		if len(s.userIndex[session.UserID]) == 0 {
			delete(s.userIndex, session.UserID)
		}
	}

	return nil
}

// GetStats returns storage statistics
func (s *MemorySessionStorage) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"total_sessions":      len(s.sessions),
		"unique_users":        len(s.userIndex),
		"token_index_entries": len(s.tokenIndex),
		"storage_type":        "memory",
	}
}

// Helper methods

// copySession creates a deep copy of a session to prevent external modifications
func (s *MemorySessionStorage) copySession(original *models.SessionData) *models.SessionData {
	if original == nil {
		return nil
	}

	// Copy basic fields
	sessionCopy := &models.SessionData{
		SessionID:    original.SessionID,
		UserID:       original.UserID,
		Username:     original.Username,
		RoleName:     original.RoleName,
		TokenHash:    original.TokenHash,
		CreatedAt:    original.CreatedAt,
		ExpiresAt:    original.ExpiresAt,
		LastActivity: original.LastActivity,
		IsActive:     original.IsActive,
	}

	// Copy permissions slice
	if original.Permissions != nil {
		sessionCopy.Permissions = make([]string, len(original.Permissions))
		copy(sessionCopy.Permissions, original.Permissions)
	}

	return sessionCopy
}

// Validate storage integrity (for debugging/testing)
func (s *MemorySessionStorage) ValidateIntegrity() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	issues := make([]string, 0)

	// Check token index integrity
	for tokenHash, sessionID := range s.tokenIndex {
		session, exists := s.sessions[sessionID]
		if !exists {
			issues = append(issues, fmt.Sprintf("Token index references non-existent session: %s", sessionID))
			continue
		}
		if session.TokenHash != tokenHash {
			issues = append(issues, fmt.Sprintf("Token index mismatch for session %s", sessionID))
		}
	}

	// Check user index integrity
	for userID, sessionIDs := range s.userIndex {
		for _, sessionID := range sessionIDs {
			session, exists := s.sessions[sessionID]
			if !exists {
				issues = append(issues, fmt.Sprintf("User index references non-existent session: %s for user %s", sessionID, userID))
				continue
			}
			if session.UserID != userID {
				issues = append(issues, fmt.Sprintf("User index mismatch for session %s", sessionID))
			}
		}
	}

	// Check that all sessions are properly indexed
	for sessionID, session := range s.sessions {
		// Check token index
		if indexedSessionID, exists := s.tokenIndex[session.TokenHash]; !exists || indexedSessionID != sessionID {
			issues = append(issues, fmt.Sprintf("Session %s not properly indexed by token", sessionID))
		}

		// Check user index
		userSessions, exists := s.userIndex[session.UserID]
		if !exists {
			issues = append(issues, fmt.Sprintf("Session %s not in user index", sessionID))
			continue
		}

		found := false
		for _, sid := range userSessions {
			if sid == sessionID {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, fmt.Sprintf("Session %s not found in user index for user %s", sessionID, session.UserID))
		}
	}

	return issues
}
