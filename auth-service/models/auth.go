package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a user in the system
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never serialize password
	FullName     string     `json:"full_name" db:"full_name"`
	RoleID       string     `json:"role_id" db:"role_id"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Role represents a user role
type Role struct {
	ID          string    `json:"id" db:"id"`
	RoleName    string    `json:"role_name" db:"role_name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID             string    `json:"id" db:"id"`
	PermissionName string    `json:"permission_name" db:"permission_name"`
	Description    string    `json:"description" db:"description"`
	RoleID         string    `json:"role_id" db:"role_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile represents the complete user profile with role and permissions
type UserProfile struct {
	User        User         `json:"user"`
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User        User         `json:"user"`
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
	Token       string       `json:"token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	RefreshAt   time.Time    `json:"refresh_at"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	Token string `json:"token" validate:"required"`
}

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	RoleID      string   `json:"role_id"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents multiple validation errors
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Errors []ValidationError `json:"errors"`
}

// AuthStatus represents the current authentication status
type AuthStatus struct {
	IsAuthenticated bool      `json:"is_authenticated"`
	User            *User     `json:"user,omitempty"`
	Role            *Role     `json:"role,omitempty"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	RefreshAt       time.Time `json:"refresh_at,omitempty"`
}

// TokenInfo represents token information for debugging/admin purposes
type TokenInfo struct {
	Valid       bool      `json:"valid"`
	UserID      string    `json:"user_id,omitempty"`
	Username    string    `json:"username,omitempty"`
	RoleName    string    `json:"role_name,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	IssuedAt    time.Time `json:"issued_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}
