package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"session-service/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestJWTManager creates a test JWT manager for use in tests
func setupTestJWTManager() *JWTManager {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	return NewJWTManager("test-secret-key", 30*time.Minute, logger)
}

// createTestUserProfile creates a test user profile for JWT generation
func createTestUserProfile() *models.UserProfile {
	now := time.Now()

	return &models.UserProfile{
		User: models.User{
			ID:        "user-123",
			Username:  "testuser",
			FullName:  "Test User",
			RoleID:    "role-456",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Role: models.Role{
			ID:          "role-456",
			RoleName:    "admin",
			Description: "Administrator role",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Permissions: []models.Permission{
			{
				ID:             "perm-1",
				PermissionName: "read",
				Description:    "Read permission",
				RoleID:         "role-456",
			},
			{
				ID:             "perm-2",
				PermissionName: "write",
				Description:    "Write permission",
				RoleID:         "role-456",
			},
		},
	}
}

// TestNewJWTManager tests the JWT manager constructor
func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		expiration time.Duration
		logger     *logrus.Logger
	}{
		{
			name:       "valid parameters",
			secret:     "test-secret",
			expiration: 30 * time.Minute,
			logger:     logrus.New(),
		},
		{
			name:       "empty secret",
			secret:     "",
			expiration: 30 * time.Minute,
			logger:     logrus.New(),
		},
		{
			name:       "zero expiration",
			secret:     "test-secret",
			expiration: 0,
			logger:     logrus.New(),
		},
		{
			name:       "nil logger",
			secret:     "test-secret",
			expiration: 30 * time.Minute,
			logger:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtManager := NewJWTManager(tt.secret, tt.expiration, tt.logger)

			assert.NotNil(t, jwtManager)
			assert.Equal(t, []byte(tt.secret), jwtManager.secret)
			assert.Equal(t, tt.expiration, jwtManager.expiration)
			assert.Equal(t, tt.logger, jwtManager.logger)
		})
	}
}

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	token, expiresAt, err := jwtManager.GenerateToken(profile)

	// Test successful generation
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))

	// Token should have 3 parts (header.payload.signature)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
	assert.Contains(t, token, ".")

	// Test expiration time is approximately correct
	expectedExpiration := time.Now().Add(30 * time.Minute)
	assert.WithinDuration(t, expectedExpiration, expiresAt, time.Minute)
}

// TestGenerateTokenWithNilProfile tests token generation with nil profile
func TestGenerateTokenWithNilProfile(t *testing.T) {
	jwtManager := setupTestJWTManager()

	// This should panic or return an error
	assert.Panics(t, func() {
		jwtManager.GenerateToken(nil)
	})
}

// TestGenerateTokenWithEmptyPermissions tests token generation with empty permissions
func TestGenerateTokenWithEmptyPermissions(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()
	profile.Permissions = []models.Permission{} // Empty permissions

	token, expiresAt, err := jwtManager.GenerateToken(profile)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))

	// Validate the token and check claims
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims.Permissions)
	assert.Len(t, claims.Permissions, 0)
}

// TestGenerateTokenWithNilPermissions tests token generation with nil permissions
func TestGenerateTokenWithNilPermissions(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()
	profile.Permissions = nil // Nil permissions

	token, expiresAt, err := jwtManager.GenerateToken(profile)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))

	// Validate the token and check claims
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims.Permissions)
	assert.Len(t, claims.Permissions, 0) // Should be empty slice, not nil
}

// TestValidateToken tests JWT token validation
func TestValidateToken(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	// Generate a valid token
	token, _, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Test valid token
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)

	// Verify claims content
	assert.Equal(t, profile.User.ID, claims.UserID)
	assert.Equal(t, profile.User.Username, claims.Username)
	assert.Equal(t, profile.User.FullName, claims.FullName)
	assert.Equal(t, profile.Role.RoleName, claims.RoleName)
	assert.Len(t, claims.Permissions, 2)
	assert.Contains(t, claims.Permissions, "read")
	assert.Contains(t, claims.Permissions, "write")

	// Verify registered claims
	assert.Equal(t, profile.User.ID, claims.Subject)
	assert.Equal(t, "icecream-session-service", claims.Issuer)
	assert.Contains(t, claims.Audience, "icecream-store")
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

// TestValidateTokenErrors tests various error scenarios for token validation
func TestValidateTokenErrors(t *testing.T) {
	jwtManager := setupTestJWTManager()

	tests := []struct {
		name     string
		token    string
		errorMsg string
	}{
		{
			name:     "empty token",
			token:    "",
			errorMsg: "invalid token",
		},
		{
			name:     "malformed token",
			token:    "invalid.token.format",
			errorMsg: "invalid token",
		},
		{
			name:     "token with wrong signature",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			errorMsg: "invalid token",
		},
		{
			name:     "token with wrong parts",
			token:    "header.payload",
			errorMsg: "invalid token",
		},
		{
			name:     "non-jwt string",
			token:    "this-is-not-a-jwt-token",
			errorMsg: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)

			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// TestValidateExpiredToken tests validation of expired tokens
func TestValidateExpiredToken(t *testing.T) {
	// Create JWT manager with very short expiration
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	jwtManager := NewJWTManager("test-secret", 1*time.Millisecond, logger)

	profile := createTestUserProfile()
	token, _, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	claims, err := jwtManager.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

// TestValidateTokenWithDifferentSecret tests validation with wrong secret
func TestValidateTokenWithDifferentSecret(t *testing.T) {
	// Generate token with one manager
	jwtManager1 := NewJWTManager("secret1", 30*time.Minute, logrus.New())
	profile := createTestUserProfile()
	token, _, err := jwtManager1.GenerateToken(profile)
	require.NoError(t, err)

	// Try to validate with different secret
	jwtManager2 := NewJWTManager("secret2", 30*time.Minute, logrus.New())
	claims, err := jwtManager2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

// TestRefreshToken tests JWT token refresh functionality
func TestRefreshToken(t *testing.T) {
	// Create JWT manager with longer expiration for refresh testing
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	jwtManager := NewJWTManager("test-secret", 10*time.Minute, logger)

	profile := createTestUserProfile()
	originalToken, originalExpiry, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Wait a bit to ensure new token has different issued time
	time.Sleep(1 * time.Second)

	// Refresh token (should work because we're within refresh threshold)
	refreshThreshold := 15 * time.Minute // Longer than token expiration
	newToken, newExpiry, err := jwtManager.RefreshToken(originalToken, refreshThreshold)

	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.True(t, newExpiry.After(originalExpiry))
	// Note: tokens might be the same if generated at the same second, but that's OK

	// Validate new token
	claims, err := jwtManager.ValidateToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, profile.User.ID, claims.UserID)
	assert.Equal(t, profile.User.Username, claims.Username)
}

// TestRefreshTokenNotNeeded tests refresh when token doesn't need refreshing yet
func TestRefreshTokenNotNeeded(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	token, _, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Try to refresh with short threshold (token doesn't need refresh yet)
	refreshThreshold := 1 * time.Minute
	newToken, newExpiry, err := jwtManager.RefreshToken(token, refreshThreshold)

	assert.Error(t, err)
	assert.Empty(t, newToken)
	assert.True(t, newExpiry.IsZero())
	assert.Contains(t, err.Error(), "token refresh not needed yet")
}

// TestRefreshInvalidToken tests refresh with invalid token
func TestRefreshInvalidToken(t *testing.T) {
	jwtManager := setupTestJWTManager()

	invalidToken := "invalid.jwt.token"
	refreshThreshold := 5 * time.Minute

	newToken, newExpiry, err := jwtManager.RefreshToken(invalidToken, refreshThreshold)

	assert.Error(t, err)
	assert.Empty(t, newToken)
	assert.True(t, newExpiry.IsZero())
	assert.Contains(t, err.Error(), "cannot refresh invalid token")
}

// TestRefreshExpiredToken tests refresh with expired token
func TestRefreshExpiredToken(t *testing.T) {
	// Create JWT manager with very short expiration
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	jwtManager := NewJWTManager("test-secret", 1*time.Millisecond, logger)

	profile := createTestUserProfile()
	token, _, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to refresh expired token
	refreshThreshold := 5 * time.Minute
	newToken, newExpiry, err := jwtManager.RefreshToken(token, refreshThreshold)

	assert.Error(t, err)
	assert.Empty(t, newToken)
	assert.True(t, newExpiry.IsZero())
	assert.Contains(t, err.Error(), "cannot refresh invalid token")
}

// TestGetTokenInfo tests token information extraction
func TestGetTokenInfo(t *testing.T) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	token, expiresAt, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Test valid token info
	info := jwtManager.GetTokenInfo(token)
	require.NotNil(t, info)

	assert.True(t, info.Valid)
	assert.Equal(t, profile.User.ID, info.UserID)
	assert.Equal(t, profile.User.Username, info.Username)
	assert.Equal(t, profile.Role.RoleName, info.RoleName)
	assert.Len(t, info.Permissions, 2)
	assert.Contains(t, info.Permissions, "read")
	assert.Contains(t, info.Permissions, "write")
	assert.WithinDuration(t, expiresAt, info.ExpiresAt, time.Second)
	assert.Empty(t, info.Error)
}

// TestGetTokenInfoInvalid tests token info for invalid tokens
func TestGetTokenInfoInvalid(t *testing.T) {
	jwtManager := setupTestJWTManager()

	tests := []struct {
		name     string
		token    string
		errorMsg string
	}{
		{
			name:     "empty token",
			token:    "",
			errorMsg: "token contains an invalid number of segments",
		},
		{
			name:     "malformed token",
			token:    "invalid.token.format",
			errorMsg: "illegal base64 data",
		},
		{
			name:     "token with wrong signature",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			errorMsg: "signature is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := jwtManager.GetTokenInfo(tt.token)
			require.NotNil(t, info)

			assert.False(t, info.Valid)
			assert.NotEmpty(t, info.Error)
			assert.Empty(t, info.UserID)
			assert.Empty(t, info.Username)
		})
	}
}

// TestGetTokenInfoExpired tests token info for expired tokens
func TestGetTokenInfoExpired(t *testing.T) {
	// Create JWT manager with very short expiration
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	jwtManager := NewJWTManager("test-secret", 1*time.Millisecond, logger)

	profile := createTestUserProfile()
	token, _, err := jwtManager.GenerateToken(profile)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Get info for expired token
	info := jwtManager.GetTokenInfo(token)
	require.NotNil(t, info)

	assert.False(t, info.Valid)
	assert.Contains(t, info.Error, "token is expired")
	// For expired tokens, the implementation may not extract user info
	// This is acceptable behavior for security reasons
}

// TestJWTManagerWithDifferentExpirations tests JWT manager with various expiration times
func TestJWTManagerWithDifferentExpirations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	profile := createTestUserProfile()

	expirations := []time.Duration{
		1 * time.Second,
		5 * time.Minute,
		1 * time.Hour,
		24 * time.Hour,
	}

	for _, expiration := range expirations {
		t.Run(fmt.Sprintf("expiration_%v", expiration), func(t *testing.T) {
			jwtManager := NewJWTManager("test-secret", expiration, logger)

			token, expiresAt, err := jwtManager.GenerateToken(profile)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Check expiration time is approximately correct
			expectedExpiration := time.Now().Add(expiration)
			assert.WithinDuration(t, expectedExpiration, expiresAt, time.Minute)

			// Validate token
			claims, err := jwtManager.ValidateToken(token)
			require.NoError(t, err)
			assert.NotNil(t, claims)
		})
	}
}

// TestJWTManagerWithDifferentSecrets tests JWT manager with various secret lengths
func TestJWTManagerWithDifferentSecrets(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	profile := createTestUserProfile()

	secrets := []string{
		"short",
		"medium-length-secret-key",
		"very-long-secret-key-with-many-characters-to-test-security",
		"special!@#$%^&*()characters",
		"unicode-秘密-ключ-سر",
	}

	for _, secret := range secrets {
		t.Run(fmt.Sprintf("secret_%d_chars", len(secret)), func(t *testing.T) {
			jwtManager := NewJWTManager(secret, 30*time.Minute, logger)

			token, _, err := jwtManager.GenerateToken(profile)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate token
			claims, err := jwtManager.ValidateToken(token)
			require.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, profile.User.ID, claims.UserID)
		})
	}
}

// TestTokenCrossCompatibility tests that tokens generated with one manager can be validated by another with same secret
func TestTokenCrossCompatibility(t *testing.T) {
	secret := "shared-secret-key"
	expiration := 30 * time.Minute
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create two managers with same secret
	jwtManager1 := NewJWTManager(secret, expiration, logger)
	jwtManager2 := NewJWTManager(secret, expiration, logger)

	profile := createTestUserProfile()

	// Generate token with first manager
	token, _, err := jwtManager1.GenerateToken(profile)
	require.NoError(t, err)

	// Validate with second manager
	claims, err := jwtManager2.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, profile.User.ID, claims.UserID)

	// Refresh with second manager
	refreshThreshold := 35 * time.Minute
	newToken, _, err := jwtManager2.RefreshToken(token, refreshThreshold)
	require.NoError(t, err)

	// Validate refreshed token with first manager
	claims, err = jwtManager1.ValidateToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, profile.User.ID, claims.UserID)
}

// BenchmarkGenerateToken benchmarks JWT token generation
func BenchmarkGenerateToken(b *testing.B) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := jwtManager.GenerateToken(profile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateToken benchmarks JWT token validation
func BenchmarkValidateToken(b *testing.B) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	token, _, err := jwtManager.GenerateToken(profile)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtManager.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRefreshToken benchmarks JWT token refresh
func BenchmarkRefreshToken(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	jwtManager := NewJWTManager("test-secret", 10*time.Minute, logger)

	profile := createTestUserProfile()
	token, _, err := jwtManager.GenerateToken(profile)
	if err != nil {
		b.Fatal(err)
	}

	refreshThreshold := 15 * time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := jwtManager.RefreshToken(token, refreshThreshold)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTokenInfo benchmarks token info extraction
func BenchmarkGetTokenInfo(b *testing.B) {
	jwtManager := setupTestJWTManager()
	profile := createTestUserProfile()

	token, _, err := jwtManager.GenerateToken(profile)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info := jwtManager.GetTokenInfo(token)
		if !info.Valid {
			b.Fatal("Token should be valid")
		}
	}
}
