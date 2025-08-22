package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Image storage configuration
	ImagesBasePath string
}

// LoadConfigFromDataService loads configuration from the data service API
func LoadConfigFromDataService(logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from data service")

	// Call data service to get settings for both Inventory and General services
	inventorySettings, err := getSettingsFromDataService("Inventory", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory settings from data service: %w", err)
	}

	generalSettings, err := getSettingsFromDataService("General", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get general settings from data service: %w", err)
	}

	// Combine settings
	settings := append(inventorySettings, generalSettings...)

	// Create config and populate from settings
	config := &Config{
		// Default values
		ServerPort: "8084",
		ServerHost: "0.0.0.0",
		LogLevel:   "info",

		// Database defaults
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres123",
		DBName:     "icecream_store",
		DBSSLMode:  "disable",

		// Image storage defaults
		ImagesBasePath: ".",
	}

	// Populate config from settings
	populateConfigFromSettings(config, settings, logger)

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
	dataServiceURL := "http://localhost:8086"
	if value := os.Getenv("DATA_SERVICE_URL"); value != "" {
		dataServiceURL = value
	}
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

// populateConfigFromSettings populates the config struct from settings
func populateConfigFromSettings(config *Config, settings []Setting, logger *logrus.Logger) {
	for _, setting := range settings {
		switch setting.Key {
		case "INVENTORY_SERVER_PORT":
			config.ServerPort = setting.Value
		case "INVENTORY_SERVER_HOST":
			config.ServerHost = setting.Value
		case "LOG_LEVEL":
			config.LogLevel = setting.Value
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
		case "INVENTORY_IMAGES_BASE_PATH":
			config.ImagesBasePath = setting.Value
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

// Note: Environment variable helpers removed - all configuration should come from data service settings
