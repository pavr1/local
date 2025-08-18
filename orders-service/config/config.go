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

// LoadConfigFromDataService loads configuration from the data service API
func LoadConfigFromDataService(logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from data service")

	// Call data service to get settings
	settings, err := getSettingsFromDataService("Orders", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings from data service: %w", err)
	}

	// Set settings as environment variables
	setSettingsAsEnvVars(settings, logger)

	// Create config with environment variables (now populated from data service)
	config := &Config{
		// Server
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8083"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "icecream-super-secret-jwt-key-change-in-production-2024"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Business
		DefaultTaxRate:     getEnvFloat("DEFAULT_TAX_RATE", 13.0),
		DefaultServiceRate: getEnvFloat("DEFAULT_SERVICE_RATE", 10.0),
		OrderTimeout:       getEnvInt("ORDER_TIMEOUT", 30),
	}

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
	dataServiceURL := getEnv("DATA_SERVICE_URL", "http://localhost:8086")
	url := fmt.Sprintf("%s/api/v1/data/settings/by-service", dataServiceURL)

	// Prepare request
	reqBody := GetSettingsByServiceRequest{
		Service: serviceName,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
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
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data service returned status %d", resp.StatusCode)
	}

	// Parse response
	var settingsResponse SettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settingsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !settingsResponse.Success {
		return nil, fmt.Errorf("data service returned error: %s", settingsResponse.Message)
	}

	logger.WithFields(logrus.Fields{
		"service":        serviceName,
		"settings_count": len(settingsResponse.Data),
	}).Info("Successfully retrieved settings from data service")

	return settingsResponse.Data, nil
}

// setSettingsAsEnvVars sets the settings as environment variables
func setSettingsAsEnvVars(settings []Setting, logger *logrus.Logger) {
	for _, setting := range settings {
		// Set environment variable using the key as the variable name
		os.Setenv(setting.Key, setting.Value)

		logger.WithFields(logrus.Fields{
			"key":     setting.Key,
			"value":   setting.Value,
			"service": setting.Service,
		}).Debug("Set environment variable from data service")
	}

	logger.WithField("variables_set", len(settings)).Info("Environment variables set from data service settings")
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
