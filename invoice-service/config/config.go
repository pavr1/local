package config

import (
	"os"
	"strconv"
	"time"

	"github.com/pablovillalobos/local/data-service/pkg/settings"
	"github.com/sirupsen/logrus"
)

// Config holds the configuration for the invoice service
type Config struct {
	ServerPort          string
	ServerHost          string
	LogLevel            string
	InventoryServiceURL string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		ServerPort:          getEnvString("INVOICE_SERVER_PORT", "8085"),
		ServerHost:          getEnvString("INVOICE_SERVER_HOST", "0.0.0.0"),
		LogLevel:            getEnvString("LOG_LEVEL", "info"),
		InventoryServiceURL: getEnvString("INVENTORY_SERVICE_URL", "http://localhost:8084"),
	}
}

// LoadConfigFromDatabase loads configuration from the database using the ServiceConfig
func LoadConfigFromDatabase(logger *logrus.Logger) (*Config, error) {
	// Create service config instance
	serviceConfig, err := settings.NewServiceConfig("Invoice", logger)
	if err != nil {
		return nil, err
	}

	// Create config with database values, falling back to environment variables
	config := &Config{
		ServerPort:          serviceConfig.GetOrDefault("INVOICE_SERVER_PORT", getEnvString("INVOICE_SERVER_PORT", "8085")),
		ServerHost:          serviceConfig.GetOrDefault("INVOICE_SERVER_HOST", getEnvString("INVOICE_SERVER_HOST", "0.0.0.0")),
		LogLevel:            serviceConfig.GetOrDefault("LOG_LEVEL", getEnvString("LOG_LEVEL", "info")),
		InventoryServiceURL: serviceConfig.GetOrDefault("INVENTORY_SERVICE_URL", getEnvString("INVENTORY_SERVICE_URL", "http://localhost:8084")),
	}

	logger.WithFields(logrus.Fields{
		"server_port":           config.ServerPort,
		"server_host":           config.ServerHost,
		"log_level":             config.LogLevel,
		"inventory_service_url": config.InventoryServiceURL,
	}).Info("Invoice service configuration loaded from database")

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
