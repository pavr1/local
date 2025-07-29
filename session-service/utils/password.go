package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager handles password hashing and validation
type PasswordManager struct {
	cost   int
	logger *logrus.Logger
}

// NewPasswordManager creates a new password manager instance
func NewPasswordManager(cost int, logger *logrus.Logger) *PasswordManager {
	return &PasswordManager{
		cost:   cost,
		logger: logger,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (p *PasswordManager) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		p.logger.WithError(err).Error("Failed to hash password")
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	p.logger.Debug("Password hashed successfully")
	return string(hashedBytes), nil
}

// ValidatePassword validates a plain text password against a hashed password
func (p *PasswordManager) ValidatePassword(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			p.logger.Debug("Password validation failed: incorrect password")
			return fmt.Errorf("incorrect password")
		}
		p.logger.WithError(err).Error("Password validation error")
		return fmt.Errorf("password validation error: %w", err)
	}

	p.logger.Debug("Password validation successful")
	return nil
}
