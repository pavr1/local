package config

import (
	"os"
	"strconv"

	"github.com/pablovillalobos/local/data-service/pkg/settings"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// Server configuration
	ServerHost string
	ServerPort string

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

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "icecream-super-secret-jwt-key-change-in-production-2024"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Business
		DefaultTaxRate:     getEnvFloat("DEFAULT_TAX_RATE", 13.0),     // 13% IVA
		DefaultServiceRate: getEnvFloat("DEFAULT_SERVICE_RATE", 10.0), // 10% servicio
		OrderTimeout:       getEnvInt("ORDER_TIMEOUT", 30),            // 30 minutes
	}
}

// LoadConfigFromDatabase loads configuration from the database using the ServiceConfig
func LoadConfigFromDatabase(logger *logrus.Logger) (*Config, error) {
	// Create service config instance
	serviceConfig, err := settings.NewServiceConfig("Orders", logger)
	if err != nil {
		return nil, err
	}

	// Create config with database values, falling back to environment variables
	config := &Config{
		// Server
		ServerHost: serviceConfig.GetOrDefault("SERVER_HOST", getEnv("SERVER_HOST", "0.0.0.0")),
		ServerPort: serviceConfig.GetOrDefault("SERVER_PORT", getEnv("SERVER_PORT", "8083")),

		// JWT
		JWTSecret: serviceConfig.GetOrDefault("JWT_SECRET", getEnv("JWT_SECRET", "icecream-super-secret-jwt-key-change-in-production-2024")),

		// Logging
		LogLevel: serviceConfig.GetOrDefault("LOG_LEVEL", getEnv("LOG_LEVEL", "info")),

		// Business
		DefaultTaxRate:     serviceConfig.GetFloatOrDefault("DEFAULT_TAX_RATE", getEnvFloat("DEFAULT_TAX_RATE", 13.0)),
		DefaultServiceRate: serviceConfig.GetFloatOrDefault("DEFAULT_SERVICE_RATE", getEnvFloat("DEFAULT_SERVICE_RATE", 10.0)),
		OrderTimeout:       serviceConfig.GetIntOrDefault("ORDER_TIMEOUT", getEnvInt("ORDER_TIMEOUT", 30)),
	}

	logger.WithFields(logrus.Fields{
		"server_port":          config.ServerPort,
		"server_host":          config.ServerHost,
		"log_level":            config.LogLevel,
		"default_tax_rate":     config.DefaultTaxRate,
		"default_service_rate": config.DefaultServiceRate,
		"order_timeout":        config.OrderTimeout,
	}).Info("Orders service configuration loaded from database")

	return config, nil
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
