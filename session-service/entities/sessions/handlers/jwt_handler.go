package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"session-service/entities/sessions/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the custom claims for our JWT tokens
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTHandler handles JWT token operations
type JWTHandler struct {
	secretKey      string
	expirationTime time.Duration
	logger         *logrus.Logger
}

// NewJWTHandler creates a new JWT handler
func NewJWTHandler(secretKey string, expirationTime time.Duration, logger *logrus.Logger) *JWTHandler {
	return &JWTHandler{
		secretKey:      secretKey,
		expirationTime: expirationTime,
		logger:         logger,
	}
}

// GenerateSessionID generates a unique session ID
func (h *JWTHandler) GenerateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateToken creates a JWT token for a user profile and returns the token string
func (h *JWTHandler) GenerateToken(sessionID string, userProfile *models.UserProfile) (string, error) {
	// Convert permissions to string slice
	permissions := make([]string, len(userProfile.Permissions))
	for i, perm := range userProfile.Permissions {
		permissions[i] = perm.PermissionName
	}

	// Create claims
	now := time.Now()
	expiresAt := now.Add(h.expirationTime)

	claims := JWTClaims{
		UserID:      userProfile.User.ID,
		Username:    userProfile.User.Username,
		RoleName:    userProfile.Role.RoleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "session-service",
			Subject:   userProfile.User.ID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userProfile.User.ID,
		"username":   userProfile.User.Username,
		"expires_at": expiresAt,
	}).Debug("JWT token generated successfully")

	return tokenString, nil
}

// GenerateTokenHash generates a SHA256 hash of the JWT token
func (h *JWTHandler) GenerateTokenHash(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

// ValidateToken validates and parses a JWT token
func (h *JWTHandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetTokenExpiration returns the expiration time of a token
func (h *JWTHandler) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := h.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}
