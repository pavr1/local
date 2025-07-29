package config

import (
	"os"
	"strconv"
	"time"

	"session-service/models"
)

// Config holds the configuration for the session service
type Config struct {
	// Server settings
	ServerPort string
	ServerHost string

	// JWT settings
	JWTSecret           string
	JWTExpirationTime   time.Duration
	JWTRefreshThreshold time.Duration

	// Session Management settings
	SessionDefaultExpiration    time.Duration
	SessionRememberMeExpiration time.Duration
	SessionCleanupInterval      time.Duration
	SessionMaxConcurrent        int
	SessionStorageType          string

	// Basic security settings
	BcryptCost        int
	MaxLoginAttempts  int
	LoginCooldownTime time.Duration

	// Database settings
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		// Server settings
		ServerPort: getEnvString("SESSION_SERVER_PORT", "8081"),
		ServerHost: getEnvString("SESSION_SERVER_HOST", "0.0.0.0"),

		// JWT settings
		JWTSecret:           getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationTime:   getEnvDuration("JWT_EXPIRATION_TIME", "30m"),
		JWTRefreshThreshold: getEnvDuration("JWT_REFRESH_THRESHOLD", "5m"),

		// Session Management settings
		SessionDefaultExpiration:    getEnvDuration("SESSION_DEFAULT_EXPIRATION", "30m"),
		SessionRememberMeExpiration: getEnvDuration("SESSION_REMEMBER_ME_EXPIRATION", "168h"), // 7 days
		SessionCleanupInterval:      getEnvDuration("SESSION_CLEANUP_INTERVAL", "10m"),
		SessionMaxConcurrent:        getEnvInt("SESSION_MAX_CONCURRENT", 5),
		SessionStorageType:          getEnvString("SESSION_STORAGE_TYPE", "memory"),

		// Basic security settings
		BcryptCost:        getEnvInt("BCRYPT_COST", 12),
		MaxLoginAttempts:  getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
		LoginCooldownTime: getEnvDuration("LOGIN_COOLDOWN_TIME", "15m"),

		// Database settings
		DatabaseHost:     getEnvString("DB_HOST", "localhost"),
		DatabasePort:     getEnvInt("DB_PORT", 5432),
		DatabaseUser:     getEnvString("DB_USER", "postgres"),
		DatabasePassword: getEnvString("DB_PASSWORD", "postgres123"),
		DatabaseName:     getEnvString("DB_NAME", "icecream_store"),
		DatabaseSSLMode:  getEnvString("DB_SSLMODE", "disable"),

		// Logging
		LogLevel: getEnvString("LOG_LEVEL", "info"),
	}
}

// ToSessionConfig converts the main config to session-specific config
func (c *Config) ToSessionConfig() *models.SessionConfig {
	return &models.SessionConfig{
		DefaultExpiration:     c.SessionDefaultExpiration,
		RememberMeExpiration:  c.SessionRememberMeExpiration,
		RefreshThreshold:      c.JWTRefreshThreshold,
		CleanupInterval:       c.SessionCleanupInterval,
		MaxConcurrentSessions: c.SessionMaxConcurrent,
		StorageType:           c.SessionStorageType,
	}
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnvString(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// Fallback to default if parsing fails
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 10 * time.Minute // Ultimate fallback
}
