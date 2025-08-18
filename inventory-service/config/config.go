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

// LoadConfigFromDataService loads configuration from the data service API
func LoadConfigFromDataService(logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from data service")

	// Call data service to get settings
	settings, err := getSettingsFromDataService("Inventory", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings from data service: %w", err)
	}

	// Set settings as environment variables
	setSettingsAsEnvVars(settings, logger)

	// Create config with environment variables (now populated from data service)
	config := &Config{
		ServerPort: getEnvString("INVENTORY_SERVER_PORT", "8084"),
		ServerHost: getEnvString("INVENTORY_SERVER_HOST", "0.0.0.0"),
		LogLevel:   getEnvString("LOG_LEVEL", "info"),
	}

	logger.WithFields(logrus.Fields{
		"server_port":    config.ServerPort,
		"server_host":    config.ServerHost,
		"log_level":      config.LogLevel,
		"settings_count": len(settings),
	}).Info("Inventory service configuration loaded from data service")

	return config, nil
}

// getSettingsFromDataService calls the data service API to get settings
func getSettingsFromDataService(serviceName string, logger *logrus.Logger) ([]Setting, error) {
	dataServiceURL := getEnvString("DATA_SERVICE_URL", "http://localhost:8086")
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
