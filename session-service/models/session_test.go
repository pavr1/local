package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionData tests the SessionData struct and its JSON serialization
func TestSessionData(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	lastActivity := now.Add(-5 * time.Minute)

	sessionData := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write", "delete"},
		TokenHash:    "abc123hash",
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastActivity: lastActivity,
		IsActive:     true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(sessionData)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "session-123")
	assert.Contains(t, string(jsonData), "testuser")
	assert.Contains(t, string(jsonData), "admin")

	// Test JSON unmarshaling
	var unmarshaled SessionData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, sessionData.SessionID, unmarshaled.SessionID)
	assert.Equal(t, sessionData.UserID, unmarshaled.UserID)
	assert.Equal(t, sessionData.Username, unmarshaled.Username)
	assert.Equal(t, sessionData.RoleName, unmarshaled.RoleName)
	assert.Equal(t, sessionData.Permissions, unmarshaled.Permissions)
	assert.Equal(t, sessionData.TokenHash, unmarshaled.TokenHash)
	assert.Equal(t, sessionData.IsActive, unmarshaled.IsActive)

	// Time comparison with some tolerance due to JSON precision
	assert.WithinDuration(t, sessionData.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, sessionData.ExpiresAt, unmarshaled.ExpiresAt, time.Second)
	assert.WithinDuration(t, sessionData.LastActivity, unmarshaled.LastActivity, time.Second)
}

// TestSessionSummary tests the SessionSummary struct
func TestSessionSummary(t *testing.T) {
	now := time.Now()
	lastActivity := now.Add(-10 * time.Minute)

	summary := &SessionSummary{
		SessionID:    "session-789",
		CreatedAt:    now,
		LastActivity: lastActivity,
		IsActive:     true,
		IsCurrent:    false,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(summary)
	require.NoError(t, err)

	var unmarshaled SessionSummary
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, summary.SessionID, unmarshaled.SessionID)
	assert.Equal(t, summary.IsActive, unmarshaled.IsActive)
	assert.Equal(t, summary.IsCurrent, unmarshaled.IsCurrent)
	assert.WithinDuration(t, summary.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, summary.LastActivity, unmarshaled.LastActivity, time.Second)
}

// TestSessionStats tests the SessionStats struct
func TestSessionStats(t *testing.T) {
	stats := &SessionStats{
		TotalSessions:   100,
		ActiveSessions:  75,
		ExpiredSessions: 25,
	}

	// Test that totals make sense
	assert.Equal(t, stats.ActiveSessions+stats.ExpiredSessions, stats.TotalSessions)

	// Test JSON serialization
	jsonData, err := json.Marshal(stats)
	require.NoError(t, err)

	var unmarshaled SessionStats
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, stats.TotalSessions, unmarshaled.TotalSessions)
	assert.Equal(t, stats.ActiveSessions, unmarshaled.ActiveSessions)
	assert.Equal(t, stats.ExpiredSessions, unmarshaled.ExpiredSessions)
}

// TestSessionValidationRequest tests the SessionValidationRequest struct
func TestSessionValidationRequest(t *testing.T) {
	request := &SessionValidationRequest{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled SessionValidationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Token, unmarshaled.Token)
}

// TestSessionValidationResponse tests the SessionValidationResponse struct
func TestSessionValidationResponse(t *testing.T) {
	sessionData := &SessionData{
		SessionID: "test-session",
		UserID:    "test-user",
		Username:  "testuser",
		RoleName:  "user",
		IsActive:  true,
	}

	tests := []struct {
		name     string
		response *SessionValidationResponse
	}{
		{
			name: "valid response with session data",
			response: &SessionValidationResponse{
				IsValid:       true,
				SessionData:   sessionData,
				ShouldRefresh: false,
			},
		},
		{
			name: "invalid response with error",
			response: &SessionValidationResponse{
				IsValid:      false,
				ErrorCode:    "INVALID_TOKEN",
				ErrorMessage: "Token is expired",
			},
		},
		{
			name: "valid response requiring refresh",
			response: &SessionValidationResponse{
				IsValid:       true,
				SessionData:   sessionData,
				ShouldRefresh: true,
				NewToken:      "new-jwt-token-here",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			require.NoError(t, err)

			var unmarshaled SessionValidationResponse
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.response.IsValid, unmarshaled.IsValid)
			assert.Equal(t, tt.response.ErrorCode, unmarshaled.ErrorCode)
			assert.Equal(t, tt.response.ErrorMessage, unmarshaled.ErrorMessage)
			assert.Equal(t, tt.response.ShouldRefresh, unmarshaled.ShouldRefresh)
			assert.Equal(t, tt.response.NewToken, unmarshaled.NewToken)

			if tt.response.SessionData != nil {
				require.NotNil(t, unmarshaled.SessionData)
				assert.Equal(t, tt.response.SessionData.SessionID, unmarshaled.SessionData.SessionID)
			}
		})
	}
}

// TestSessionCreateRequest tests the SessionCreateRequest struct
func TestSessionCreateRequest(t *testing.T) {
	expiresAt := time.Now().Add(30 * time.Minute)

	request := &SessionCreateRequest{
		UserID:      "user-123",
		Username:    "testuser",
		RoleName:    "admin",
		Permissions: []string{"read", "write", "admin"},
		RememberMe:  true,
		ExpiresAt:   expiresAt,
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled SessionCreateRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.UserID, unmarshaled.UserID)
	assert.Equal(t, request.Username, unmarshaled.Username)
	assert.Equal(t, request.RoleName, unmarshaled.RoleName)
	assert.Equal(t, request.Permissions, unmarshaled.Permissions)
	assert.Equal(t, request.RememberMe, unmarshaled.RememberMe)
	assert.WithinDuration(t, request.ExpiresAt, unmarshaled.ExpiresAt, time.Second)
}

// TestSessionRevokeRequest tests the SessionRevokeRequest struct
func TestSessionRevokeRequest(t *testing.T) {
	tests := []struct {
		name    string
		request *SessionRevokeRequest
	}{
		{
			name: "revoke by session ID",
			request: &SessionRevokeRequest{
				SessionID: "session-123",
				RevokeAll: false,
			},
		},
		{
			name: "revoke by user ID",
			request: &SessionRevokeRequest{
				UserID:    "user-456",
				RevokeAll: false,
			},
		},
		{
			name: "revoke by token",
			request: &SessionRevokeRequest{
				Token:     "jwt-token-here",
				RevokeAll: false,
			},
		},
		{
			name: "revoke all user sessions",
			request: &SessionRevokeRequest{
				UserID:    "user-789",
				RevokeAll: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)

			var unmarshaled SessionRevokeRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.request.SessionID, unmarshaled.SessionID)
			assert.Equal(t, tt.request.UserID, unmarshaled.UserID)
			assert.Equal(t, tt.request.Token, unmarshaled.Token)
			assert.Equal(t, tt.request.RevokeAll, unmarshaled.RevokeAll)
		})
	}
}

// TestSessionConfig tests the SessionConfig struct
func TestSessionConfig(t *testing.T) {
	config := &SessionConfig{
		DefaultExpiration:     30 * time.Minute,
		RememberMeExpiration:  7 * 24 * time.Hour,
		RefreshThreshold:      5 * time.Minute,
		CleanupInterval:       10 * time.Minute,
		MaxConcurrentSessions: 5,
		StorageType:           "memory",
	}

	jsonData, err := json.Marshal(config)
	require.NoError(t, err)

	var unmarshaled SessionConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, config.DefaultExpiration, unmarshaled.DefaultExpiration)
	assert.Equal(t, config.RememberMeExpiration, unmarshaled.RememberMeExpiration)
	assert.Equal(t, config.RefreshThreshold, unmarshaled.RefreshThreshold)
	assert.Equal(t, config.CleanupInterval, unmarshaled.CleanupInterval)
	assert.Equal(t, config.MaxConcurrentSessions, unmarshaled.MaxConcurrentSessions)
	assert.Equal(t, config.StorageType, unmarshaled.StorageType)
}

// TestDefaultSessionConfig tests the default session configuration
func TestDefaultSessionConfig(t *testing.T) {
	config := DefaultSessionConfig()

	// Test default values
	assert.Equal(t, 30*time.Minute, config.DefaultExpiration)
	assert.Equal(t, 7*24*time.Hour, config.RememberMeExpiration)
	assert.Equal(t, 5*time.Minute, config.RefreshThreshold)
	assert.Equal(t, 10*time.Minute, config.CleanupInterval)
	assert.Equal(t, 5, config.MaxConcurrentSessions)
	assert.Equal(t, "memory", config.StorageType)

	// Test that refresh threshold is less than default expiration
	assert.True(t, config.RefreshThreshold < config.DefaultExpiration)

	// Test that remember me expiration is longer than default
	assert.True(t, config.RememberMeExpiration > config.DefaultExpiration)

	// Test that cleanup interval is reasonable
	assert.True(t, config.CleanupInterval > 0)
	assert.True(t, config.CleanupInterval < config.DefaultExpiration)

	// Test that max concurrent sessions is reasonable
	assert.True(t, config.MaxConcurrentSessions > 0)
}

// TestSessionConfigValidation tests session config validation logic
func TestSessionConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *SessionConfig
		valid  bool
	}{
		{
			name:   "valid default config",
			config: DefaultSessionConfig(),
			valid:  true,
		},
		{
			name: "valid custom config",
			config: &SessionConfig{
				DefaultExpiration:     60 * time.Minute,
				RememberMeExpiration:  30 * 24 * time.Hour,
				RefreshThreshold:      10 * time.Minute,
				CleanupInterval:       15 * time.Minute,
				MaxConcurrentSessions: 10,
				StorageType:           "redis",
			},
			valid: true,
		},
		{
			name: "zero default expiration",
			config: &SessionConfig{
				DefaultExpiration:     0,
				RememberMeExpiration:  24 * time.Hour,
				RefreshThreshold:      5 * time.Minute,
				CleanupInterval:       10 * time.Minute,
				MaxConcurrentSessions: 5,
				StorageType:           "memory",
			},
			valid: false,
		},
		{
			name: "negative max concurrent sessions",
			config: &SessionConfig{
				DefaultExpiration:     30 * time.Minute,
				RememberMeExpiration:  24 * time.Hour,
				RefreshThreshold:      5 * time.Minute,
				CleanupInterval:       10 * time.Minute,
				MaxConcurrentSessions: -1,
				StorageType:           "memory",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.config.DefaultExpiration > 0 &&
				tt.config.RememberMeExpiration > 0 &&
				tt.config.RefreshThreshold >= 0 &&
				tt.config.CleanupInterval > 0 &&
				tt.config.MaxConcurrentSessions >= 0 &&
				tt.config.StorageType != ""

			assert.Equal(t, tt.valid, isValid)
		})
	}
}

// TestSessionDataComparison tests comparison and equality of SessionData
func TestSessionDataComparison(t *testing.T) {
	now := time.Now()

	session1 := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write"},
		TokenHash:    "hash123",
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute),
		LastActivity: now,
		IsActive:     true,
	}

	session2 := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write"},
		TokenHash:    "hash123",
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute),
		LastActivity: now,
		IsActive:     true,
	}

	session3 := &SessionData{
		SessionID:    "session-789",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write"},
		TokenHash:    "hash123",
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute),
		LastActivity: now,
		IsActive:     true,
	}

	// Test that identical sessions have same fields
	assert.Equal(t, session1.SessionID, session2.SessionID)
	assert.Equal(t, session1.UserID, session2.UserID)
	assert.Equal(t, session1.Permissions, session2.Permissions)

	// Test that different sessions have different IDs
	assert.NotEqual(t, session1.SessionID, session3.SessionID)
}

// TestSessionDataWithNilPermissions tests SessionData with nil permissions
func TestSessionDataWithNilPermissions(t *testing.T) {
	sessionData := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "user",
		Permissions:  nil, // nil permissions
		TokenHash:    "hash123",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Test JSON marshaling with nil permissions
	jsonData, err := json.Marshal(sessionData)
	require.NoError(t, err)

	var unmarshaled SessionData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// nil slice should unmarshal as nil (not empty slice)
	assert.Nil(t, unmarshaled.Permissions)
}

// TestSessionDataWithEmptyPermissions tests SessionData with empty permissions
func TestSessionDataWithEmptyPermissions(t *testing.T) {
	sessionData := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "user",
		Permissions:  []string{}, // empty permissions slice
		TokenHash:    "hash123",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Test JSON marshaling with empty permissions
	jsonData, err := json.Marshal(sessionData)
	require.NoError(t, err)

	var unmarshaled SessionData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Empty slice should remain as empty slice
	assert.NotNil(t, unmarshaled.Permissions)
	assert.Len(t, unmarshaled.Permissions, 0)
}

// BenchmarkSessionDataMarshal benchmarks SessionData JSON marshaling
func BenchmarkSessionDataMarshal(b *testing.B) {
	sessionData := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write", "delete", "admin"},
		TokenHash:    "abc123hash",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(sessionData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSessionDataUnmarshal benchmarks SessionData JSON unmarshaling
func BenchmarkSessionDataUnmarshal(b *testing.B) {
	sessionData := &SessionData{
		SessionID:    "session-123",
		UserID:       "user-456",
		Username:     "testuser",
		RoleName:     "admin",
		Permissions:  []string{"read", "write", "delete", "admin"},
		TokenHash:    "abc123hash",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var unmarshaled SessionData
		err := json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			b.Fatal(err)
		}
	}
}
