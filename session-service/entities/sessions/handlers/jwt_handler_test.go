package handlers

import (
	"testing"
	"time"

	"session-service/entities/sessions/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTHandler(t *testing.T) {
	logger := logrus.New()
	secretKey := "test-secret-key"
	expirationTime := 1 * time.Hour

	handler := NewJWTHandler(secretKey, expirationTime, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, secretKey, handler.secretKey)
	assert.Equal(t, expirationTime, handler.expirationTime)
	assert.Equal(t, logger, handler.logger)
}

func TestGenerateSessionID(t *testing.T) {
	logger := logrus.New()
	handler := NewJWTHandler("test-secret", 1*time.Hour, logger)

	sessionID1, err := handler.GenerateSessionID()
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID1)
	assert.Len(t, sessionID1, 32) // 16 bytes = 32 hex chars

	sessionID2, err := handler.GenerateSessionID()
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID2)
	assert.NotEqual(t, sessionID1, sessionID2) // Should be unique
}

func TestGenerateToken(t *testing.T) {
	logger := logrus.New()
	handler := NewJWTHandler("test-secret-key", 1*time.Hour, logger)

	userProfile := &models.UserProfile{
		User: models.User{
			ID:       "user123",
			Username: "testuser",
		},
		Role: models.Role{
			RoleName: "admin",
		},
		Permissions: []models.Permission{
			{PermissionName: "read"},
			{PermissionName: "write"},
		},
	}

	sessionID := "session123"
	token, err := handler.GenerateToken(sessionID, userProfile)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	claims, err := handler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.RoleName)
	assert.Equal(t, []string{"read", "write"}, claims.Permissions)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestGenerateTokenHash(t *testing.T) {
	logger := logrus.New()
	handler := NewJWTHandler("test-secret", 1*time.Hour, logger)

	token := "test-token-string"
	hash1 := handler.GenerateTokenHash(token)
	hash2 := handler.GenerateTokenHash(token)

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2) // Same input should produce same hash
	assert.Len(t, hash1, 64)      // SHA256 hash = 64 hex chars

	// Different tokens should produce different hashes
	hash3 := handler.GenerateTokenHash("different-token")
	assert.NotEqual(t, hash1, hash3)
}

func TestValidateToken(t *testing.T) {
	logger := logrus.New()
	handler := NewJWTHandler("test-secret-key", 1*time.Hour, logger)

	userProfile := &models.UserProfile{
		User: models.User{
			ID:       "user123",
			Username: "testuser",
		},
		Role: models.Role{
			RoleName: "admin",
		},
		Permissions: []models.Permission{
			{PermissionName: "read"},
		},
	}

	token, err := handler.GenerateToken("session123", userProfile)
	require.NoError(t, err)

	// Test valid token
	claims, err := handler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.RoleName)
	assert.Equal(t, []string{"read"}, claims.Permissions)

	// Test invalid token
	_, err = handler.ValidateToken("invalid-token")
	assert.Error(t, err)

	// Test token with wrong secret
	wrongHandler := NewJWTHandler("wrong-secret", 1*time.Hour, logger)
	_, err = wrongHandler.ValidateToken(token)
	assert.Error(t, err)
}

func TestGetTokenExpiration(t *testing.T) {
	logger := logrus.New()
	handler := NewJWTHandler("test-secret-key", 1*time.Hour, logger)

	userProfile := &models.UserProfile{
		User: models.User{
			ID:       "user123",
			Username: "testuser",
		},
		Role: models.Role{
			RoleName: "admin",
		},
		Permissions: []models.Permission{},
	}

	token, err := handler.GenerateToken("session123", userProfile)
	require.NoError(t, err)

	expiration, err := handler.GetTokenExpiration(token)
	require.NoError(t, err)
	assert.True(t, expiration.After(time.Now()))
	assert.True(t, expiration.Before(time.Now().Add(2*time.Hour)))

	// Test invalid token
	_, err = handler.GetTokenExpiration("invalid-token")
	assert.Error(t, err)
}
