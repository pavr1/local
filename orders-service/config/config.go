package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server configuration
	ServerHost string
	ServerPort string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT configuration (for token validation)
	JWTSecret string

	// Logging
	LogLevel string

	// Business configuration
	DefaultTaxRate     float64
	DefaultServiceRate float64
	OrderTimeout       int // minutes
}

func LoadConfig() *Config {
	return &Config{
		// Server
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8083"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres123"),
		DBName:     getEnv("DB_NAME", "icecream_store"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Business
		DefaultTaxRate:     getEnvFloat("DEFAULT_TAX_RATE", 13.0),     // 13% IVA
		DefaultServiceRate: getEnvFloat("DEFAULT_SERVICE_RATE", 10.0), // 10% servicio
		OrderTimeout:       getEnvInt("ORDER_TIMEOUT", 30),            // 30 minutes
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
