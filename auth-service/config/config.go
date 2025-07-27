package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the auth service
type Config struct {
	// Server settings
	ServerPort string
	ServerHost string

	// JWT settings
	JWTSecret           string
	JWTExpirationTime   time.Duration
	JWTRefreshThreshold time.Duration

	// Database settings
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// Security settings
	BcryptCost        int
	MaxLoginAttempts  int
	LoginCooldownTime time.Duration

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		// Server settings
		ServerPort: getEnvString("AUTH_SERVER_PORT", "8081"),
		ServerHost: getEnvString("AUTH_SERVER_HOST", "0.0.0.0"),

		// JWT settings
		JWTSecret:           getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationTime:   getEnvDuration("JWT_EXPIRATION_TIME", "10m"),
		JWTRefreshThreshold: getEnvDuration("JWT_REFRESH_THRESHOLD", "2m"),

		// Database settings
		DatabaseHost:     getEnvString("DB_HOST", "localhost"),
		DatabasePort:     getEnvInt("DB_PORT", 5432),
		DatabaseUser:     getEnvString("DB_USER", "postgres"),
		DatabasePassword: getEnvString("DB_PASSWORD", "postgres123"),
		DatabaseName:     getEnvString("DB_NAME", "icecream_store"),
		DatabaseSSLMode:  getEnvString("DB_SSLMODE", "disable"),

		// Security settings
		BcryptCost:        getEnvInt("BCRYPT_COST", 12),
		MaxLoginAttempts:  getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
		LoginCooldownTime: getEnvDuration("LOGIN_COOLDOWN_TIME", "15m"),

		// Logging
		LogLevel: getEnvString("LOG_LEVEL", "info"),
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
