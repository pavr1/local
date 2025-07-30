package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the inventory service
type Config struct {
	ServerPort string
	ServerHost string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	LogLevel   string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		ServerPort: getEnvString("INVENTORY_SERVER_PORT", "8082"),
		ServerHost: getEnvString("INVENTORY_SERVER_HOST", "0.0.0.0"),
		DBHost:     getEnvString("DB_HOST", "localhost"),
		DBPort:     getEnvString("DB_PORT", "5432"),
		DBUser:     getEnvString("DB_USER", "postgres"),
		DBPassword: getEnvString("DB_PASSWORD", "postgres123"),
		DBName:     getEnvString("DB_NAME", "icecream_store"),
		DBSSLMode:  getEnvString("DB_SSLMODE", "disable"),
		LogLevel:   getEnvString("LOG_LEVEL", "info"),
	}
}

// getEnvString returns the environment variable value or default if not set
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the environment variable value as int or default if not set
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration returns the environment variable value as duration or default if not set
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
