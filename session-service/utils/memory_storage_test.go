package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"session-service/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestSession creates a test session for use in tests
func createTestSession(sessionID, userID, username string) *models.SessionData {
	now := time.Now()
	return &models.SessionData{
		SessionID:    sessionID,
		UserID:       userID,
		Username:     username,
		RoleName:     "user",
		Permissions:  []string{"read"},
		TokenHash:    fmt.Sprintf("hash-%s", sessionID),
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute),
		LastActivity: now,
		IsActive:     true,
	}
}

// TestNewMemorySessionStorage tests the constructor
func TestNewMemorySessionStorage(t *testing.T) {
	storage := NewMemorySessionStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.sessions)
	assert.NotNil(t, storage.tokenIndex)
	assert.NotNil(t, storage.userIndex)
	assert.Len(t, storage.sessions, 0)
	assert.Len(t, storage.tokenIndex, 0)
	assert.Len(t, storage.userIndex, 0)
}

// TestStore tests session storage functionality
func TestStore(t *testing.T) {
	storage := NewMemorySessionStorage()
	session := createTestSession("session-123", "user-456", "testuser")

	err := storage.Store("session-123", session)
	require.NoError(t, err)

	// Verify session is stored
	assert.Len(t, storage.sessions, 1)
	assert.Contains(t, storage.sessions, "session-123")

	// Verify token index is updated
	assert.Len(t, storage.tokenIndex, 1)
	assert.Contains(t, storage.tokenIndex, session.TokenHash)
	assert.Equal(t, "session-123", storage.tokenIndex[session.TokenHash])

	// Verify user index is updated
	assert.Len(t, storage.userIndex, 1)
	assert.Contains(t, storage.userIndex, "user-456")
	assert.Contains(t, storage.userIndex["user-456"], "session-123")
}

// TestStoreMultipleSessions tests storing multiple sessions
func TestStoreMultipleSessions(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Store sessions for different users
	session1 := createTestSession("session-1", "user-1", "user1")
	session2 := createTestSession("session-2", "user-2", "user2")
	session3 := createTestSession("session-3", "user-1", "user1") // Same user, different session

	err := storage.Store("session-1", session1)
	require.NoError(t, err)
	err = storage.Store("session-2", session2)
	require.NoError(t, err)
	err = storage.Store("session-3", session3)
	require.NoError(t, err)

	// Verify all sessions are stored
	assert.Len(t, storage.sessions, 3)
	assert.Len(t, storage.tokenIndex, 3)

	// Verify user index has correct sessions
	assert.Len(t, storage.userIndex, 2)           // Two unique users
	assert.Len(t, storage.userIndex["user-1"], 2) // User-1 has 2 sessions
	assert.Len(t, storage.userIndex["user-2"], 1) // User-2 has 1 session
}

// TestGet tests session retrieval by session ID
func TestGet(t *testing.T) {
	storage := NewMemorySessionStorage()
	originalSession := createTestSession("session-123", "user-456", "testuser")

	// Store session
	err := storage.Store("session-123", originalSession)
	require.NoError(t, err)

	// Retrieve session
	retrievedSession, err := storage.Get("session-123")
	require.NoError(t, err)
	assert.NotNil(t, retrievedSession)

	// Verify session data
	assert.Equal(t, originalSession.SessionID, retrievedSession.SessionID)
	assert.Equal(t, originalSession.UserID, retrievedSession.UserID)
	assert.Equal(t, originalSession.Username, retrievedSession.Username)
	assert.Equal(t, originalSession.TokenHash, retrievedSession.TokenHash)

	// Verify it's a copy (not the same pointer)
	assert.NotSame(t, originalSession, retrievedSession)
}

// TestGetNonexistentSession tests retrieving a session that doesn't exist
func TestGetNonexistentSession(t *testing.T) {
	storage := NewMemorySessionStorage()

	session, err := storage.Get("nonexistent-session")
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found")
}

// TestGetByTokenHash tests session retrieval by token hash
func TestGetByTokenHash(t *testing.T) {
	storage := NewMemorySessionStorage()
	originalSession := createTestSession("session-123", "user-456", "testuser")

	// Store session
	err := storage.Store("session-123", originalSession)
	require.NoError(t, err)

	// Retrieve by token hash
	retrievedSession, err := storage.GetByTokenHash(originalSession.TokenHash)
	require.NoError(t, err)
	assert.NotNil(t, retrievedSession)
	assert.Equal(t, originalSession.SessionID, retrievedSession.SessionID)
}

// TestGetByTokenHashNonexistent tests retrieving by nonexistent token hash
func TestGetByTokenHashNonexistent(t *testing.T) {
	storage := NewMemorySessionStorage()

	session, err := storage.GetByTokenHash("nonexistent-token-hash")
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found for token")
}

// TestGetUserSessions tests retrieving all sessions for a user
func TestGetUserSessions(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Create multiple sessions for the same user
	session1 := createTestSession("session-1", "user-123", "testuser")
	session2 := createTestSession("session-2", "user-123", "testuser")
	session3 := createTestSession("session-3", "user-456", "otheruser")

	// Store sessions
	err := storage.Store("session-1", session1)
	require.NoError(t, err)
	err = storage.Store("session-2", session2)
	require.NoError(t, err)
	err = storage.Store("session-3", session3)
	require.NoError(t, err)

	// Get sessions for user-123
	userSessions, err := storage.GetUserSessions("user-123")
	require.NoError(t, err)
	assert.Len(t, userSessions, 2)

	// Verify correct sessions are returned
	sessionIDs := make([]string, len(userSessions))
	for i, session := range userSessions {
		sessionIDs[i] = session.SessionID
	}
	assert.Contains(t, sessionIDs, "session-1")
	assert.Contains(t, sessionIDs, "session-2")
	assert.NotContains(t, sessionIDs, "session-3")
}

// TestGetUserSessionsNonexistent tests getting sessions for a user that doesn't exist
func TestGetUserSessionsNonexistent(t *testing.T) {
	storage := NewMemorySessionStorage()

	sessions, err := storage.GetUserSessions("nonexistent-user")
	require.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 0)
}

// TestUpdate tests session updates
func TestUpdate(t *testing.T) {
	storage := NewMemorySessionStorage()
	originalSession := createTestSession("session-123", "user-456", "testuser")

	// Store original session
	err := storage.Store("session-123", originalSession)
	require.NoError(t, err)

	// Update session
	updatedSession := *originalSession
	updatedSession.LastActivity = time.Now().Add(10 * time.Minute)
	updatedSession.RoleName = "admin"

	err = storage.Update("session-123", &updatedSession)
	require.NoError(t, err)

	// Retrieve and verify update
	retrievedSession, err := storage.Get("session-123")
	require.NoError(t, err)
	assert.Equal(t, "admin", retrievedSession.RoleName)
	assert.True(t, retrievedSession.LastActivity.After(originalSession.LastActivity))
}

// TestUpdateNonexistentSession tests updating a session that doesn't exist
func TestUpdateNonexistentSession(t *testing.T) {
	storage := NewMemorySessionStorage()
	session := createTestSession("session-123", "user-456", "testuser")

	err := storage.Update("nonexistent-session", session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestDelete tests session deletion
func TestDelete(t *testing.T) {
	storage := NewMemorySessionStorage()
	session := createTestSession("session-123", "user-456", "testuser")

	// Store session
	err := storage.Store("session-123", session)
	require.NoError(t, err)

	// Verify session exists
	_, err = storage.Get("session-123")
	require.NoError(t, err)

	// Delete session
	err = storage.Delete("session-123")
	require.NoError(t, err)

	// Verify session is deleted
	_, err = storage.Get("session-123")
	assert.Error(t, err)

	// Verify indexes are cleaned up
	assert.NotContains(t, storage.sessions, "session-123")
	assert.NotContains(t, storage.tokenIndex, session.TokenHash)

	// Check user index cleanup
	userSessions, err := storage.GetUserSessions("user-456")
	require.NoError(t, err)
	assert.Len(t, userSessions, 0)
}

// TestDeleteNonexistentSession tests deleting a session that doesn't exist
func TestDeleteNonexistentSession(t *testing.T) {
	storage := NewMemorySessionStorage()

	err := storage.Delete("nonexistent-session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestDeleteUserSessions tests deleting all sessions for a user
func TestDeleteUserSessions(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Create sessions for multiple users
	session1 := createTestSession("session-1", "user-123", "testuser")
	session2 := createTestSession("session-2", "user-123", "testuser")
	session3 := createTestSession("session-3", "user-456", "otheruser")

	// Store sessions
	err := storage.Store("session-1", session1)
	require.NoError(t, err)
	err = storage.Store("session-2", session2)
	require.NoError(t, err)
	err = storage.Store("session-3", session3)
	require.NoError(t, err)

	// Delete all sessions for user-123
	err = storage.DeleteUserSessions("user-123")
	require.NoError(t, err)

	// Verify user-123 sessions are deleted
	userSessions, err := storage.GetUserSessions("user-123")
	require.NoError(t, err)
	assert.Len(t, userSessions, 0)

	// Verify user-456 session still exists
	userSessions, err = storage.GetUserSessions("user-456")
	require.NoError(t, err)
	assert.Len(t, userSessions, 1)

	// Verify total sessions
	allSessions, err := storage.GetAllSessions()
	require.NoError(t, err)
	assert.Len(t, allSessions, 1)
}

// TestDeleteUserSessionsNonexistent tests deleting sessions for a user that doesn't exist
func TestDeleteUserSessionsNonexistent(t *testing.T) {
	storage := NewMemorySessionStorage()

	err := storage.DeleteUserSessions("nonexistent-user")
	require.NoError(t, err) // Should not error, just no-op
}

// TestGetAllSessions tests retrieving all sessions
func TestGetAllSessions(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Store multiple sessions
	session1 := createTestSession("session-1", "user-1", "user1")
	session2 := createTestSession("session-2", "user-2", "user2")
	session3 := createTestSession("session-3", "user-1", "user1")

	err := storage.Store("session-1", session1)
	require.NoError(t, err)
	err = storage.Store("session-2", session2)
	require.NoError(t, err)
	err = storage.Store("session-3", session3)
	require.NoError(t, err)

	// Get all sessions
	allSessions, err := storage.GetAllSessions()
	require.NoError(t, err)
	assert.Len(t, allSessions, 3)

	// Verify all sessions are present
	sessionIDs := make(map[string]bool)
	for _, session := range allSessions {
		sessionIDs[session.SessionID] = true
	}
	assert.True(t, sessionIDs["session-1"])
	assert.True(t, sessionIDs["session-2"])
	assert.True(t, sessionIDs["session-3"])
}

// TestCleanup tests expired session cleanup
func TestCleanup(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Create sessions with different expiration times
	now := time.Now()
	expiredSession := createTestSession("expired-session", "user-1", "user1")
	expiredSession.ExpiresAt = now.Add(-1 * time.Hour) // Expired

	validSession := createTestSession("valid-session", "user-2", "user2")
	validSession.ExpiresAt = now.Add(1 * time.Hour) // Still valid

	// Store sessions
	err := storage.Store("expired-session", expiredSession)
	require.NoError(t, err)
	err = storage.Store("valid-session", validSession)
	require.NoError(t, err)

	// Verify both sessions exist before cleanup
	allSessions, err := storage.GetAllSessions()
	require.NoError(t, err)
	assert.Len(t, allSessions, 2)

	// Run cleanup
	err = storage.Cleanup()
	require.NoError(t, err)

	// Verify only valid session remains
	allSessions, err = storage.GetAllSessions()
	require.NoError(t, err)
	assert.Len(t, allSessions, 1)
	assert.Equal(t, "valid-session", allSessions[0].SessionID)
}

// TestConcurrentAccess tests concurrent access to memory storage
func TestConcurrentAccess(t *testing.T) {
	storage := NewMemorySessionStorage()
	const numGoroutines = 10
	const numOperationsPerGoroutine = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// Concurrent store operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				sessionID := fmt.Sprintf("session-%d-%d", goroutineID, j)
				userID := fmt.Sprintf("user-%d", goroutineID)
				session := createTestSession(sessionID, userID, fmt.Sprintf("user%d", goroutineID))

				if err := storage.Store(sessionID, session); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Fatal(err)
	}

	// Verify all sessions were stored
	allSessions, err := storage.GetAllSessions()
	require.NoError(t, err)
	assert.Len(t, allSessions, numGoroutines*numOperationsPerGoroutine)
}

// TestConcurrentReadWrite tests concurrent read and write operations
func TestConcurrentReadWrite(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Pre-populate with some sessions
	for i := 0; i < 10; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		session := createTestSession(sessionID, fmt.Sprintf("user-%d", i), fmt.Sprintf("user%d", i))
		err := storage.Store(sessionID, session)
		require.NoError(t, err)
	}

	const numReaders = 5
	const numWriters = 3
	const operationsPerGoroutine = 20

	var wg sync.WaitGroup
	errors := make(chan error, (numReaders+numWriters)*operationsPerGoroutine)

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				sessionID := fmt.Sprintf("session-%d", j%10)
				_, err := storage.Get(sessionID)
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				sessionID := fmt.Sprintf("new-session-%d-%d", writerID, j)
				session := createTestSession(sessionID, fmt.Sprintf("new-user-%d-%d", writerID, j), "newuser")
				if err := storage.Store(sessionID, session); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Fatal(err)
	}
}

// TestSessionCopy tests that sessions are properly copied
func TestSessionCopy(t *testing.T) {
	storage := NewMemorySessionStorage()
	originalSession := createTestSession("session-123", "user-456", "testuser")

	// Store session
	err := storage.Store("session-123", originalSession)
	require.NoError(t, err)

	// Get session
	retrievedSession, err := storage.Get("session-123")
	require.NoError(t, err)

	// Modify retrieved session
	retrievedSession.RoleName = "modified"
	retrievedSession.IsActive = false

	// Get session again and verify original is unchanged
	secondRetrieval, err := storage.Get("session-123")
	require.NoError(t, err)

	assert.Equal(t, originalSession.RoleName, secondRetrieval.RoleName)
	assert.Equal(t, originalSession.IsActive, secondRetrieval.IsActive)
	assert.NotEqual(t, retrievedSession.RoleName, secondRetrieval.RoleName)
}

// TestIndexConsistency tests that all indexes remain consistent
func TestIndexConsistency(t *testing.T) {
	storage := NewMemorySessionStorage()

	// Store sessions
	session1 := createTestSession("session-1", "user-1", "user1")
	session2 := createTestSession("session-2", "user-1", "user1")
	session3 := createTestSession("session-3", "user-2", "user2")

	err := storage.Store("session-1", session1)
	require.NoError(t, err)
	err = storage.Store("session-2", session2)
	require.NoError(t, err)
	err = storage.Store("session-3", session3)
	require.NoError(t, err)

	// Verify index consistency
	assert.Len(t, storage.sessions, 3)
	assert.Len(t, storage.tokenIndex, 3)
	assert.Len(t, storage.userIndex, 2)

	// Delete one session and verify cleanup
	err = storage.Delete("session-1")
	require.NoError(t, err)

	// Check that indexes are properly cleaned up
	assert.Len(t, storage.sessions, 2)
	assert.Len(t, storage.tokenIndex, 2)
	assert.NotContains(t, storage.tokenIndex, session1.TokenHash)

	// User should still exist with one session
	userSessions, err := storage.GetUserSessions("user-1")
	require.NoError(t, err)
	assert.Len(t, userSessions, 1)
	assert.Equal(t, "session-2", userSessions[0].SessionID)
}

// BenchmarkStore benchmarks session storage
func BenchmarkStore(b *testing.B) {
	storage := NewMemorySessionStorage()
	session := createTestSession("session-123", "user-456", "testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		err := storage.Store(sessionID, session)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGet benchmarks session retrieval
func BenchmarkGet(b *testing.B) {
	storage := NewMemorySessionStorage()

	// Pre-populate storage
	for i := 0; i < 1000; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		session := createTestSession(sessionID, fmt.Sprintf("user-%d", i), "testuser")
		storage.Store(sessionID, session)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("session-%d", i%1000)
		_, err := storage.Get(sessionID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetByTokenHash benchmarks token hash lookups
func BenchmarkGetByTokenHash(b *testing.B) {
	storage := NewMemorySessionStorage()
	tokenHashes := make([]string, 1000)

	// Pre-populate storage
	for i := 0; i < 1000; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		session := createTestSession(sessionID, fmt.Sprintf("user-%d", i), "testuser")
		tokenHashes[i] = session.TokenHash
		storage.Store(sessionID, session)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenHash := tokenHashes[i%1000]
		_, err := storage.GetByTokenHash(tokenHash)
		if err != nil {
			b.Fatal(err)
		}
	}
}
