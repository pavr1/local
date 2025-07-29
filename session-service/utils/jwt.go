package utils

import (
	"fmt"
	"time"

	"session-service/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secret     []byte
	expiration time.Duration
	logger     *logrus.Logger
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager(secret string, expiration time.Duration, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expiration: expiration,
		logger:     logger,
	}
}

// GenerateToken generates a JWT token for a user with their profile
func (j *JWTManager) GenerateToken(profile *models.UserProfile) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.expiration)

	// Convert permissions to string slice
	permissions := make([]string, len(profile.Permissions))
	for i, perm := range profile.Permissions {
		permissions[i] = perm.PermissionName
	}

	// Create claims
	claims := &models.JWTClaims{
		UserID:      profile.User.ID,
		Username:    profile.User.Username,
		FullName:    profile.User.FullName,
		RoleID:      profile.User.RoleID,
		RoleName:    profile.Role.RoleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   profile.User.ID,
			Issuer:    "icecream-session-service",
			Audience:  []string{"icecream-store"},
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		j.logger.WithError(err).Error("Failed to sign JWT token")
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":    profile.User.ID,
		"username":   profile.User.Username,
		"role":       profile.Role.RoleName,
		"expires_at": expiresAt,
	}).Info("JWT token generated successfully")

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		j.logger.WithError(err).Warn("JWT token validation failed")
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		j.logger.Warn("JWT token claims are invalid")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		j.logger.WithFields(logrus.Fields{
			"user_id":    claims.UserID,
			"expires_at": claims.ExpiresAt.Time,
		}).Warn("JWT token has expired")
		return nil, fmt.Errorf("token has expired")
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.RoleName,
	}).Debug("JWT token validated successfully")

	return claims, nil
}

// RefreshToken generates a new token if the current one is valid and within refresh threshold
func (j *JWTManager) RefreshToken(tokenString string, refreshThreshold time.Duration) (string, time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Check if token is within refresh threshold
	if claims.ExpiresAt != nil {
		timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
		if timeUntilExpiry > refreshThreshold {
			j.logger.WithFields(logrus.Fields{
				"user_id":           claims.UserID,
				"time_until_expiry": timeUntilExpiry,
				"refresh_threshold": refreshThreshold,
			}).Info("Token refresh not needed yet")
			return "", time.Time{}, fmt.Errorf("token refresh not needed yet")
		}
	}

	// Create new token with updated expiration
	now := time.Now()
	expiresAt := now.Add(j.expiration)

	newClaims := &models.JWTClaims{
		UserID:      claims.UserID,
		Username:    claims.Username,
		FullName:    claims.FullName,
		RoleID:      claims.RoleID,
		RoleName:    claims.RoleName,
		Permissions: claims.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   claims.UserID,
			Issuer:    "icecream-session-service",
			Audience:  []string{"icecream-store"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := token.SignedString(j.secret)
	if err != nil {
		j.logger.WithError(err).Error("Failed to sign refreshed JWT token")
		return "", time.Time{}, fmt.Errorf("failed to refresh token: %w", err)
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"expires_at": expiresAt,
	}).Info("JWT token refreshed successfully")

	return newTokenString, expiresAt, nil
}

// GetTokenInfo extracts token information for debugging/admin purposes
func (j *JWTManager) GetTokenInfo(tokenString string) *models.TokenInfo {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	info := &models.TokenInfo{}

	if err != nil {
		info.Valid = false
		info.Error = err.Error()
		return info
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		info.Valid = false
		info.Error = "invalid token claims"
		return info
	}

	info.Valid = token.Valid
	info.UserID = claims.UserID
	info.Username = claims.Username
	info.RoleName = claims.RoleName
	info.Permissions = claims.Permissions

	if claims.IssuedAt != nil {
		info.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		info.ExpiresAt = claims.ExpiresAt.Time
		if claims.ExpiresAt.Time.Before(time.Now()) {
			info.Valid = false
			info.Error = "token has expired"
		}
	}

	return info
}
