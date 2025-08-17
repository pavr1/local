package config

import (
	"os"
	"strconv"
	"time"

	"github.com/pablovillalobos/local/data-service/pkg/settings"
	"github.com/sirupsen/logrus"
)

// Config holds the configuration for the inventory service
type Config struct {
	ServerPort string
	ServerHost string
	LogLevel   string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		ServerPort: getEnvString("INVENTORY_SERVER_PORT", "8084"),
		ServerHost: getEnvString("INVENTORY_SERVER_HOST", "0.0.0.0"),
		LogLevel:   getEnvString("LOG_LEVEL", "info"),
	}
}

// LoadConfigFromDatabase loads configuration from the database using the ServiceConfig
func LoadConfigFromDatabase(logger *logrus.Logger) (*Config, error) {
	// Create service config instance
	serviceConfig, err := settings.NewServiceConfig("Inventory", logger)
	if err != nil {
		return nil, err
	}

	// Create config with database values, falling back to environment variables
	config := &Config{
		ServerPort: serviceConfig.GetOrDefault("INVENTORY_SERVER_PORT", getEnvString("INVENTORY_SERVER_PORT", "8084")),
		ServerHost: serviceConfig.GetOrDefault("INVENTORY_SERVER_HOST", getEnvString("INVENTORY_SERVER_HOST", "0.0.0.0")),
		LogLevel:   serviceConfig.GetOrDefault("LOG_LEVEL", getEnvString("LOG_LEVEL", "info")),
	}

	logger.WithFields(logrus.Fields{
		"server_port": config.ServerPort,
		"server_host": config.ServerHost,
		"log_level":   config.LogLevel,
	}).Info("Inventory service configuration loaded from database")

	return config, nil
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
