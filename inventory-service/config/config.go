package config

import (
	"os"
	"strconv"
)

// Config holds the configuration for the inventory service
type Config struct {
	// Server settings
	ServerPort string
	ServerHost string

	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		// Server settings
		ServerPort: getEnvString("INVENTORY_SERVER_PORT", "8082"),
		ServerHost: getEnvString("INVENTORY_SERVER_HOST", "0.0.0.0"),

		// Database settings
		DBHost:     getEnvString("DB_HOST", "localhost"),
		DBPort:     getEnvString("DB_PORT", "5432"),
		DBUser:     getEnvString("DB_USER", "postgres"),
		DBPassword: getEnvString("DB_PASSWORD", "postgres123"),
		DBName:     getEnvString("DB_NAME", "icecream_store"),
		DBSSLMode:  getEnvString("DB_SSLMODE", "disable"),

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
