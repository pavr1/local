package models

import (
	"time"
)

// Session represents a user session
type Session struct {
	SessionID string `json:"session_id"`
	TokenHash string `json:"token_hash"`
}

// SessionCreateRequest represents a session creation request
type SessionCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SessionCreateResponse represents a session creation response
type SessionCreateResponse struct {
	SessionID string `json:"session_id"`
}

// User represents a user from the database
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"password_hash"`
	FullName     string     `json:"full_name"`
	RoleID       string     `json:"role_id"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Role represents a user role
type Role struct {
	ID        string    `json:"id"`
	RoleName  string    `json:"role_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Permission represents a user permission
type Permission struct {
	ID             string    `json:"id"`
	PermissionName string    `json:"permission_name"`
	Description    string    `json:"description"`
	RoleID         string    `json:"role_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserProfile represents a complete user profile with role and permissions
type UserProfile struct {
	User        User         `json:"user"`
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
