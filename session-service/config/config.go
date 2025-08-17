package config

import (
	"os"
	"strconv"
	"time"

	"github.com/pablovillalobos/local/data-service/pkg/settings"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// Server settings
	ServerPort string
	ServerHost string

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

		// JWT settings
		JWTSecret:         getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationTime: getEnvDuration("JWT_EXPIRATION_TIME", "30m"),

		// Logging
		LogLevel: getEnvString("LOG_LEVEL", "info"),
	}
}

// LoadConfigFromDatabase loads configuration from the database using the ServiceConfig
func LoadConfigFromDatabase(logger *logrus.Logger) (*Config, error) {
	// Create service config instance
	serviceConfig, err := settings.NewServiceConfig("Session", logger)
	if err != nil {
		return nil, err
	}

	// Create config with database values, falling back to environment variables
	config := &Config{
		// Server settings
		ServerPort: serviceConfig.GetOrDefault("SESSION_SERVER_PORT", getEnvString("SESSION_SERVER_PORT", "8081")),
		ServerHost: serviceConfig.GetOrDefault("SESSION_SERVER_HOST", getEnvString("SESSION_SERVER_HOST", "0.0.0.0")),

		// JWT settings
		JWTSecret:         serviceConfig.GetOrDefault("JWT_SECRET", getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")),
		JWTExpirationTime: getEnvDuration("JWT_EXPIRATION_TIME", "30m"), // Note: Duration parsing from DB would need custom implementation

		// Logging
		LogLevel: serviceConfig.GetOrDefault("LOG_LEVEL", getEnvString("LOG_LEVEL", "info")),
	}

	logger.WithFields(logrus.Fields{
		"server_port":    config.ServerPort,
		"server_host":    config.ServerHost,
		"log_level":      config.LogLevel,
		"jwt_expiration": config.JWTExpirationTime,
	}).Info("Session service configuration loaded from database")

	return config, nil
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
