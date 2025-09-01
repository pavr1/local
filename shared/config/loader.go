package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	sharedModels "shared/models"

	"github.com/sirupsen/logrus"
)

// ConfigLoader provides functionality to load configuration from the data service
type ConfigLoader struct {
	dataServiceURL string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(dataServiceURL string) *ConfigLoader {
	return &ConfigLoader{
		dataServiceURL: dataServiceURL,
	}
}

// Config is a generic configuration structure that can be used by all services
type Config struct {
	// Store all configuration as key-value pairs
	Values map[string]string
}

// NewConfig creates a new config with default values
func NewConfig() *Config {
	return &Config{
		Values: make(map[string]string),
	}
}

// GetString returns a string value from config
func (c *Config) GetString(key, defaultValue string) string {
	if value, exists := c.Values[key]; exists {
		return value
	}
	return defaultValue
}

// GetInt returns an int value from config
func (c *Config) GetInt(key string, defaultValue int) int {
	if value, exists := c.Values[key]; exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetFloat returns a float64 value from config
func (c *Config) GetFloat(key string, defaultValue float64) float64 {
	if value, exists := c.Values[key]; exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetDuration returns a time.Duration value from config
func (c *Config) GetDuration(key, defaultValue string) time.Duration {
	if value, exists := c.Values[key]; exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 30 * time.Minute // fallback
}

// Set sets a key-value pair in the config
func (c *Config) Set(key, value string) {
	c.Values[key] = value
}

// LoadSettingsFromDataService calls the data service API to get settings
func (cl *ConfigLoader) LoadSettingsFromDataService(serviceName string, logger *logrus.Logger) ([]sharedModels.Setting, error) {
	url := fmt.Sprintf("%s/api/v1/data/settings/by-service", cl.dataServiceURL)

	// Prepare request
	reqBody := sharedModels.GetSettingsByServiceRequest{
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
	req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
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
	var settingsResponse sharedModels.SettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settingsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if settingsResponse.Code != http.StatusOK {
		return nil, fmt.Errorf("data service returned error: %s", settingsResponse.Message)
	}

	// Convert interface{} to []Setting using JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(settingsResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	var settings []sharedModels.Setting
	if err := json.Unmarshal(jsonData, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"service":        serviceName,
		"settings_count": len(settings),
	}).Info("Successfully retrieved settings from data service")

	return settings, nil
}

// LoadMultipleServicesSettings loads settings for multiple services
func (cl *ConfigLoader) LoadMultipleServicesSettings(serviceNames []string, logger *logrus.Logger) ([]sharedModels.Setting, error) {
	var allSettings []sharedModels.Setting

	for _, serviceName := range serviceNames {
		settings, err := cl.LoadSettingsFromDataService(serviceName, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to get settings for service %s: %w", serviceName, err)
		}
		allSettings = append(allSettings, settings...)
	}

	return allSettings, nil
}

// LoadConfig loads configuration for any service
func (cl *ConfigLoader) LoadConfig(serviceName string, logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from data service")

	settings, err := cl.LoadSettingsFromDataService(serviceName, logger)
	if err != nil {
		return nil, err
	}

	config := NewConfig()

	// Set default values based on service
	setDefaultValues(config, serviceName)

	// Populate from settings
	populateConfigFromSettings(config, settings, logger)

	logger.WithFields(logrus.Fields{
		"service":        serviceName,
		"settings_count": len(settings),
	}).Info("Configuration loaded from data service")

	return config, nil
}

// LoadMultipleServicesConfig loads configuration for multiple services
func (cl *ConfigLoader) LoadMultipleServicesConfig(serviceNames []string, logger *logrus.Logger) (*Config, error) {
	logger.Info("Loading configuration from multiple services")

	settings, err := cl.LoadMultipleServicesSettings(serviceNames, logger)
	if err != nil {
		return nil, err
	}

	config := NewConfig()

	// Set default values for the primary service
	if len(serviceNames) > 0 {
		setDefaultValues(config, serviceNames[0])
	}

	// Populate from settings
	populateConfigFromSettings(config, settings, logger)

	logger.WithFields(logrus.Fields{
		"services":       serviceNames,
		"settings_count": len(settings),
	}).Info("Configuration loaded from multiple services")

	return config, nil
}

// setDefaultValues sets default values based on service name
func setDefaultValues(config *Config, serviceName string) {
	switch serviceName {
	case "Session":
		config.Set("SERVER_PORT", "8081")
		config.Set("SERVER_HOST", "0.0.0.0")
		config.Set("DB_HOST", "localhost")
		config.Set("DB_PORT", "5432")
		config.Set("DB_USER", "postgres")
		config.Set("DB_PASSWORD", "postgres123")
		config.Set("DB_NAME", "icecream_store")
		config.Set("DB_SSL_MODE", "disable")
		config.Set("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
		config.Set("JWT_EXPIRATION_TIME", "30m")
		config.Set("LOG_LEVEL", "info")
	case "Orders":
		config.Set("SERVER_PORT", "8083")
		config.Set("SERVER_HOST", "0.0.0.0")
		config.Set("DB_HOST", "localhost")
		config.Set("DB_PORT", "5432")
		config.Set("DB_USER", "postgres")
		config.Set("DB_PASSWORD", "postgres123")
		config.Set("DB_NAME", "icecream_store")
		config.Set("DB_SSL_MODE", "disable")
		config.Set("JWT_SECRET", "icecream-super-secret-jwt-key-change-in-production-2024")
		config.Set("LOG_LEVEL", "info")
		config.Set("DEFAULT_TAX_RATE", "13.0")
		config.Set("DEFAULT_SERVICE_RATE", "10.0")
		config.Set("ORDER_TIMEOUT", "30")
		config.Set("INVOICE_SERVICE_URL", "http://localhost:8085")
		config.Set("DATA_SERVICE_URL", "http://icecream_data_service:8086")
	case "Inventory":
		config.Set("SERVER_PORT", "8084")
		config.Set("SERVER_HOST", "0.0.0.0")
		config.Set("DB_HOST", "localhost")
		config.Set("DB_PORT", "5432")
		config.Set("DB_USER", "postgres")
		config.Set("DB_PASSWORD", "postgres123")
		config.Set("DB_NAME", "icecream_store")
		config.Set("DB_SSL_MODE", "disable")
		config.Set("LOG_LEVEL", "info")
		config.Set("INVENTORY_IMAGES_BASE_PATH", ".")
		config.Set("GATEWAY_URL", "http://localhost:8082")
	case "Invoice":
		config.Set("SERVER_PORT", "8085")
		config.Set("SERVER_HOST", "0.0.0.0")
		config.Set("DB_HOST", "localhost")
		config.Set("DB_PORT", "5432")
		config.Set("DB_USER", "postgres")
		config.Set("DB_PASSWORD", "postgres123")
		config.Set("DB_NAME", "icecream_store")
		config.Set("DB_SSL_MODE", "disable")
		config.Set("LOG_LEVEL", "info")
		config.Set("INVENTORY_SERVICE_URL", "http://localhost:8084")
	case "Gateway":
		config.Set("GATEWAY_SERVICE_URL", "http://localhost:8082")
		config.Set("SESSION_SERVICE_URL", "http://localhost:8081")
		config.Set("ORDERS_SERVICE_URL", "http://localhost:8083")
		config.Set("INVENTORY_SERVICE_URL", "http://localhost:8084")
		config.Set("INVOICE_SERVICE_URL", "http://localhost:8085")
		config.Set("DATA_SERVICE_URL", "http://icecream_data_service:8086")
	}
}

// populateConfigFromSettings populates the config from settings
func populateConfigFromSettings(config *Config, settings []sharedModels.Setting, logger *logrus.Logger) {
	for _, setting := range settings {
		config.Set(setting.Key, setting.Value)

		logger.WithFields(logrus.Fields{
			"key":     setting.Key,
			"value":   setting.Value,
			"service": setting.Service,
		}).Debug("Populated config from data service setting")
	}

	logger.WithField("settings_processed", len(settings)).Info("Config populated from data service settings")
}

// ===== ENVIRONMENT UTILITY FUNCTIONS =====

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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
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
