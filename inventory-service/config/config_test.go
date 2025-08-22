package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfigFromDataService tests loading configuration from data service
func TestLoadConfigFromDataService(t *testing.T) {
	// Create a mock data service server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response for settings
		response := `{
			"success": true,
			"data": [
				{
					"setting_id": "test-1",
					"service": "Inventory",
					"key": "INVENTORY_SERVER_PORT",
					"value": "8084",
					"description": "Port for inventory service"
				},
				{
					"setting_id": "test-2",
					"service": "Inventory",
					"key": "INVENTORY_SERVER_HOST",
					"value": "0.0.0.0",
					"description": "Host for inventory service"
				},
				{
					"setting_id": "test-3",
					"service": "Inventory",
					"key": "LOG_LEVEL",
					"value": "info",
					"description": "Log level"
				},
				{
					"setting_id": "test-4",
					"service": "Inventory",
					"key": "INVENTORY_IMAGES_BASE_PATH",
					"value": ".",
					"description": "Base path for images"
				}
			],
			"total": 4,
			"message": "Settings retrieved successfully"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Set the data service URL to our mock server
	originalURL := os.Getenv("DATA_SERVICE_URL")
	os.Setenv("DATA_SERVICE_URL", server.URL)
	defer func() {
		if originalURL != "" {
			os.Setenv("DATA_SERVICE_URL", originalURL)
		} else {
			os.Unsetenv("DATA_SERVICE_URL")
		}
	}()

	logger := logrus.New()
	config, err := LoadConfigFromDataService(logger)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Test that settings were loaded correctly
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, "8084", config.ServerPort)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, ".", config.ImagesBasePath)
}

// TestLoadConfigFromDataServiceError tests error handling when data service is unavailable
func TestLoadConfigFromDataServiceError(t *testing.T) {
	// Set an invalid data service URL
	originalURL := os.Getenv("DATA_SERVICE_URL")
	os.Setenv("DATA_SERVICE_URL", "http://invalid-url-that-does-not-exist:9999")
	defer func() {
		if originalURL != "" {
			os.Setenv("DATA_SERVICE_URL", originalURL)
		} else {
			os.Unsetenv("DATA_SERVICE_URL")
		}
	}()

	logger := logrus.New()
	config, err := LoadConfigFromDataService(logger)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to get settings from data service")
}

// TestLoadConfigFromDataServiceInvalidResponse tests handling of invalid JSON response
func TestLoadConfigFromDataServiceInvalidResponse(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	// Set the data service URL to our mock server
	originalURL := os.Getenv("DATA_SERVICE_URL")
	os.Setenv("DATA_SERVICE_URL", server.URL)
	defer func() {
		if originalURL != "" {
			os.Setenv("DATA_SERVICE_URL", originalURL)
		} else {
			os.Unsetenv("DATA_SERVICE_URL")
		}
	}()

	logger := logrus.New()
	config, err := LoadConfigFromDataService(logger)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to decode response")
}

// TestLoadConfigFromDataServiceFailureResponse tests handling of failure response from data service
func TestLoadConfigFromDataServiceFailureResponse(t *testing.T) {
	// Create a mock server that returns a failure response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"success": false,
			"data": [],
			"total": 0,
			"message": "Service not found"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Set the data service URL to our mock server
	originalURL := os.Getenv("DATA_SERVICE_URL")
	os.Setenv("DATA_SERVICE_URL", server.URL)
	defer func() {
		if originalURL != "" {
			os.Setenv("DATA_SERVICE_URL", originalURL)
		} else {
			os.Unsetenv("DATA_SERVICE_URL")
		}
	}()

	logger := logrus.New()
	config, err := LoadConfigFromDataService(logger)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "data service returned error")
}

// TestConfigStruct tests the Config struct fields
func TestConfigStruct(t *testing.T) {
	config := &Config{
		ServerPort:     "8084",
		ServerHost:     "0.0.0.0",
		LogLevel:       "info",
		ImagesBasePath: ".",
	}

	// Test that all fields are properly set
	assert.Equal(t, "8084", config.ServerPort)
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, ".", config.ImagesBasePath)
}

// TestConfigValidation tests that config values are valid
func TestConfigValidation(t *testing.T) {
	config := &Config{
		ServerPort:     "8084",
		ServerHost:     "0.0.0.0",
		LogLevel:       "info",
		ImagesBasePath: ".",
	}

	// Test that required fields are not empty
	require.NotEmpty(t, config.ServerHost)
	require.NotEmpty(t, config.ServerPort)
	require.NotEmpty(t, config.LogLevel)
	require.NotEmpty(t, config.ImagesBasePath)

	// Test that server port is a valid number
	assert.Regexp(t, `^\d+$`, config.ServerPort, "Server port should be numeric")

	// Test valid log levels
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	assert.Contains(t, validLogLevels, config.LogLevel, "Log level should be valid")
}

// BenchmarkLoadConfigFromDataService benchmarks the configuration loading process
func BenchmarkLoadConfigFromDataService(b *testing.B) {
	// Create a mock data service server for benchmarking
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"success": true,
			"data": [
				{
					"setting_id": "test-1",
					"service": "Inventory",
					"key": "INVENTORY_SERVER_PORT",
					"value": "8084",
					"description": "Port for inventory service"
				}
			],
			"total": 1,
			"message": "Settings retrieved successfully"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Set the data service URL to our mock server
	originalURL := os.Getenv("DATA_SERVICE_URL")
	os.Setenv("DATA_SERVICE_URL", server.URL)
	defer func() {
		if originalURL != "" {
			os.Setenv("DATA_SERVICE_URL", originalURL)
		} else {
			os.Unsetenv("DATA_SERVICE_URL")
		}
	}()

	logger := logrus.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfigFromDataService(logger)
		if err != nil {
			b.Fatal(err)
		}
	}
}
