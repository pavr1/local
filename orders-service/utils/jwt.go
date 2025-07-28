package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secret []byte
}

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	FullName    string    `json:"full_name"`
	RoleID      uuid.UUID `json:"role_id"`
	RoleName    string    `json:"role_name"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
	}
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Debug logging
	fmt.Printf("🔍 JWT Validation Debug:\n")
	fmt.Printf("Secret length: %d\n", len(j.secret))
	fmt.Printf("Token length: %d\n", len(tokenString))
	fmt.Printf("Secret first 20 chars: %s...\n", string(j.secret[:20]))
	fmt.Printf("Token first 50 chars: %s...\n", tokenString[:50])

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Printf("❌ Unexpected signing method: %v\n", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		fmt.Printf("✅ Signing method is correct: %v\n", token.Header["alg"])
		return j.secret, nil
	})

	if err != nil {
		fmt.Printf("❌ Parse error: %v\n", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		fmt.Printf("❌ Token is not valid\n")
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		fmt.Printf("❌ Invalid token claims type\n")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		fmt.Printf("❌ Token has expired: %v < %v\n", claims.ExpiresAt.Time, time.Now())
		return nil, fmt.Errorf("token has expired")
	}

	fmt.Printf("✅ JWT validation successful for user: %s\n", claims.Username)
	return claims, nil
}

// HasPermission checks if the user has a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (c *Claims) IsAdmin() bool {
	return c.RoleName == "super_admin" || c.RoleName == "admin"
}

// HasOrdersPermission checks if user has any orders-related permission
func (c *Claims) HasOrdersPermission(action string) bool {
	permission := "orders-" + action
	return c.HasPermission(permission) || c.IsAdmin()
}

// ExtractUserID extracts user ID from claims
func (c *Claims) ExtractUserID() uuid.UUID {
	return c.UserID
}
