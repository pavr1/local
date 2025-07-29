package models

import (
	"time"
)

// SessionData represents a simple user session with essential information
type SessionData struct {
	// Core Session Info
	SessionID   string   `json:"session_id"`
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`

	// Token Management
	TokenHash string `json:"token_hash"` // SHA256 hash of JWT token for security

	// Timing
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`

	// Session State
	IsActive bool `json:"is_active"`
}

// SessionSummary provides a safe view of session data for user management
type SessionSummary struct {
	SessionID    string    `json:"session_id"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
	IsCurrent    bool      `json:"is_current"`
}

// SessionStats provides basic analytics about user sessions
type SessionStats struct {
	TotalSessions   int `json:"total_sessions"`
	ActiveSessions  int `json:"active_sessions"`
	ExpiredSessions int `json:"expired_sessions"`
}

// SessionValidationRequest represents a token validation request
type SessionValidationRequest struct {
	Token string `json:"token"`
}

// SessionValidationResponse represents the result of session validation
type SessionValidationResponse struct {
	IsValid       bool         `json:"is_valid"`
	SessionData   *SessionData `json:"session_data,omitempty"`
	ErrorCode     string       `json:"error_code,omitempty"`
	ErrorMessage  string       `json:"error_message,omitempty"`
	ShouldRefresh bool         `json:"should_refresh"`
	NewToken      string       `json:"new_token,omitempty"`
}

// SessionCreateRequest represents a session creation request
type SessionCreateRequest struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	RoleName    string    `json:"role_name"`
	Permissions []string  `json:"permissions"`
	RememberMe  bool      `json:"remember_me"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// SessionRevokeRequest represents a session revocation request
type SessionRevokeRequest struct {
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Token     string `json:"token,omitempty"`
	RevokeAll bool   `json:"revoke_all"` // Revoke all user sessions
}

// SessionConfig represents basic session management configuration
type SessionConfig struct {
	// Timing Configuration
	DefaultExpiration    time.Duration `json:"default_expiration"`
	RememberMeExpiration time.Duration `json:"remember_me_expiration"`
	RefreshThreshold     time.Duration `json:"refresh_threshold"`
	CleanupInterval      time.Duration `json:"cleanup_interval"`

	// Basic Security Configuration
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`

	// Storage Configuration
	StorageType string `json:"storage_type"` // "memory", "redis", "database"
}

// Default configuration with simple settings
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		DefaultExpiration:     30 * time.Minute,
		RememberMeExpiration:  7 * 24 * time.Hour, // 7 days
		RefreshThreshold:      5 * time.Minute,
		CleanupInterval:       10 * time.Minute,
		MaxConcurrentSessions: 5,
		StorageType:           "memory",
	}
}
