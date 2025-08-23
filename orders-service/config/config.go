package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// Setting represents a configuration setting from the data service
type Setting struct {
	SettingID   string    `json:"setting_id"`
	Service     string    `json:"service"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SettingsResponse represents the response from the data service
type SettingsResponse struct {
	Success bool      `json:"success"`
	Data    []Setting `json:"data"`
	Total   int       `json:"total"`
	Message string    `json:"message"`
}

// GetSettingsByServiceRequest represents the request to get settings by service
type GetSettingsByServiceRequest struct {
	Service string `json:"service"`
}

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

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func LoadConfig() *Config {
	return &Config{
		// Server
		//pvillalobos - hardcoded values env variables not used
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

// LoadConfigFromDataService loads configuration from the data service API
func LoadConfigFromDataService(logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from data service")

	// Call data service to get settings
	settings, err := getSettingsFromDataService("Orders", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings from data service: %w", err)
	}

	// Create config and populate it directly from settings
	//pvillalobos - hardcoded values
	config := &Config{
		// Server
		ServerHost: "0.0.0.0",
		ServerPort: "8083",

		// JWT
		JWTSecret: "icecream-super-secret-jwt-key-change-in-production-2024",

		// Logging
		LogLevel: "info",

		// Business
		DefaultTaxRate:     13.0,
		DefaultServiceRate: 10.0,
		OrderTimeout:       30,

		// Database
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres123",
		DBName:     "icecream_store",
		DBSSLMode:  "disable",
	}

	// Populate config from settings
	populateConfigFromSettings(config, settings, logger)

	logger.WithFields(logrus.Fields{
		"server_port":          config.ServerPort,
		"server_host":          config.ServerHost,
		"log_level":            config.LogLevel,
		"default_tax_rate":     config.DefaultTaxRate,
		"default_service_rate": config.DefaultServiceRate,
		"order_timeout":        config.OrderTimeout,
		"settings_count":       len(settings),
	}).Info("Orders service configuration loaded from data service")

	return config, nil
}

// getSettingsFromDataService calls the data service API to get settings
func getSettingsFromDataService(serviceName string, logger *logrus.Logger) ([]Setting, error) {
	//pvillalobos - hardcoded values env variables not used
	dataServiceURL := getEnv("DATA_SERVICE_URL", "http://localhost:8086")
	url := fmt.Sprintf("%s/api/v1/data/settings/by-service", dataServiceURL)

	// Prepare request
	reqBody := GetSettingsByServiceRequest{
		Service: serviceName,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal request body")
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.WithError(err).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add gateway headers for internal service communication
	req.Header.Set("X-Gateway-Service", "gateway")
	req.Header.Set("X-Gateway-Session-Managed", "true")
	req.Header.Set("X-User-ID", "system")
	req.Header.Set("X-Username", "system")
	req.Header.Set("X-User-Role", "admin")

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("Failed to make HTTP request")
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(logrus.Fields{
			"service": serviceName,
			"status":  resp.StatusCode,
		}).Error("Data service returned non-OK status")
		return nil, fmt.Errorf("data service returned status %d", resp.StatusCode)
	}

	// Parse response
	var settingsResponse SettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settingsResponse); err != nil {
		logger.WithError(err).Error("Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !settingsResponse.Success {
		logger.WithFields(logrus.Fields{
			"service": serviceName,
			"message": settingsResponse.Message,
		}).Error("Data service returned error")
		return nil, fmt.Errorf("data service returned error: %s", settingsResponse.Message)
	}

	logger.WithFields(logrus.Fields{
		"service":        serviceName,
		"settings_count": len(settingsResponse.Data),
	}).Info("Successfully retrieved settings from data service")

	return settingsResponse.Data, nil
}

// populateConfigFromSettings populates the config struct from settings
func populateConfigFromSettings(config *Config, settings []Setting, logger *logrus.Logger) {
	//pvillalobos - hardcoded values
	for _, setting := range settings {
		switch setting.Key {
		case "SERVER_PORT":
			config.ServerPort = setting.Value
		case "SERVER_HOST":
			config.ServerHost = setting.Value
		case "JWT_SECRET":
			config.JWTSecret = setting.Value
		case "LOG_LEVEL":
			config.LogLevel = setting.Value
		case "DEFAULT_TAX_RATE":
			if rate, err := strconv.ParseFloat(setting.Value, 64); err == nil {
				config.DefaultTaxRate = rate
			} else {
				logger.WithError(err).Warnf("Failed to parse DEFAULT_TAX_RATE: %s", setting.Value)
			}
		case "DEFAULT_SERVICE_RATE":
			if rate, err := strconv.ParseFloat(setting.Value, 64); err == nil {
				config.DefaultServiceRate = rate
			} else {
				logger.WithError(err).Warnf("Failed to parse DEFAULT_SERVICE_RATE: %s", setting.Value)
			}
		case "ORDER_TIMEOUT":
			if timeout, err := strconv.Atoi(setting.Value); err == nil {
				config.OrderTimeout = timeout
			} else {
				logger.WithError(err).Warnf("Failed to parse ORDER_TIMEOUT: %s", setting.Value)
			}
		case "DB_HOST":
			config.DBHost = setting.Value
		case "DB_PORT":
			config.DBPort = setting.Value
		case "DB_USER":
			config.DBUser = setting.Value
		case "DB_PASSWORD":
			config.DBPassword = setting.Value
		case "DB_NAME":
			config.DBName = setting.Value
		case "DB_SSL_MODE":
			config.DBSSLMode = setting.Value
		default:
			logger.WithField("key", setting.Key).Debug("Setting not mapped to config struct")
		}

		logger.WithFields(logrus.Fields{
			"key":     setting.Key,
			"value":   setting.Value,
			"service": setting.Service,
		}).Debug("Populated config from data service setting")
	}

	logger.WithField("settings_processed", len(settings)).Info("Config populated from data service settings")
}

// pvillalobos - environment variables not used
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
