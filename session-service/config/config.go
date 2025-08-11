package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server settings
	ServerPort string
	ServerHost string

	// Database settings
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// JWT settings
	JWTSecret         string
	JWTExpirationTime time.Duration

	// Logging
	LogLevel string
}

func LoadConfig() *Config {
	return &Config{
		// Server settings
		ServerPort: getEnvString("SESSION_SERVER_PORT", "8081"),
		ServerHost: getEnvString("SESSION_SERVER_HOST", "0.0.0.0"),

		// Database settings
		DatabaseHost:     getEnvString("DB_HOST", "localhost"),
		DatabasePort:     getEnvInt("DB_PORT", 5432),
		DatabaseUser:     getEnvString("DB_USER", "postgres"),
		DatabasePassword: getEnvString("DB_PASSWORD", "postgres123"),
		DatabaseName:     getEnvString("DB_NAME", "icecream_store"),
		DatabaseSSLMode:  getEnvString("DB_SSLMODE", "disable"),

		// JWT settings
		JWTSecret:         getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationTime: getEnvDuration("JWT_EXPIRATION_TIME", "30m"),

		// Logging
		LogLevel: getEnvString("LOG_LEVEL", "info"),
	}
}

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

func getEnvDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 30 * time.Minute // fallback
}
