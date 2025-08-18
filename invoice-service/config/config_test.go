package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnvVars()
	defer clearEnvVars()

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "8085", cfg.ServerPort)
	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	// Database settings are now loaded from data service
	// No longer stored in config struct
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadConfigFromEnvironment(t *testing.T) {
	// Clear environment variables first
	clearEnvVars()
	defer clearEnvVars()

	// Set custom environment variables
	os.Setenv("INVOICE_SERVER_PORT", "9090")
	os.Setenv("INVOICE_SERVER_HOST", "127.0.0.1")
	os.Setenv("LOG_LEVEL", "debug")

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "127.0.0.1", cfg.ServerHost)
	// Database settings are now loaded from data service, not from environment variables
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadConfigPartialEnvironment(t *testing.T) {
	// Clear environment variables first
	clearEnvVars()
	defer clearEnvVars()

	// Set only some environment variables
	os.Setenv("INVOICE_SERVER_PORT", "8086")
	os.Setenv("LOG_LEVEL", "warn")

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	// Custom values from environment
	assert.Equal(t, "8086", cfg.ServerPort)
	assert.Equal(t, "warn", cfg.LogLevel)

	// Default values for unset variables
	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	// Database settings are now loaded from data service, not from environment variables
}

func TestGetEnvString(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_STRING_VAR", "test_value")
	defer os.Unsetenv("TEST_STRING_VAR")

	result := getEnvString("TEST_STRING_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	// Test with non-existing environment variable
	result = getEnvString("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)

	// Test with empty environment variable
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = getEnvString("EMPTY_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid integer environment variable
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 10)
	assert.Equal(t, 42, result)

	// Test with invalid integer environment variable
	os.Setenv("INVALID_INT_VAR", "not_a_number")
	defer os.Unsetenv("INVALID_INT_VAR")

	result = getEnvInt("INVALID_INT_VAR", 10)
	assert.Equal(t, 10, result)

	// Test with non-existing environment variable
	result = getEnvInt("NON_EXISTING_INT_VAR", 10)
	assert.Equal(t, 10, result)

	// Test with empty environment variable
	os.Setenv("EMPTY_INT_VAR", "")
	defer os.Unsetenv("EMPTY_INT_VAR")

	result = getEnvInt("EMPTY_INT_VAR", 10)
	assert.Equal(t, 10, result)
}

func TestGetEnvDuration(t *testing.T) {
	// Test with valid duration environment variable
	os.Setenv("TEST_DURATION_VAR", "30s")
	defer os.Unsetenv("TEST_DURATION_VAR")

	result := getEnvDuration("TEST_DURATION_VAR", 60*time.Second)
	assert.Equal(t, 30, int(result.Seconds()))

	// Test with invalid duration environment variable
	os.Setenv("INVALID_DURATION_VAR", "not_a_duration")
	defer os.Unsetenv("INVALID_DURATION_VAR")

	result = getEnvDuration("INVALID_DURATION_VAR", 60*time.Second)
	assert.Equal(t, 60, int(result.Seconds()))

	// Test with non-existing environment variable
	result = getEnvDuration("NON_EXISTING_DURATION_VAR", 60*time.Second)
	assert.Equal(t, 60, int(result.Seconds()))

	// Test with complex duration
	os.Setenv("COMPLEX_DURATION_VAR", "1h30m")
	defer os.Unsetenv("COMPLEX_DURATION_VAR")

	result = getEnvDuration("COMPLEX_DURATION_VAR", 60*time.Second)
	assert.Equal(t, 5400, int(result.Seconds())) // 1.5 hours = 5400 seconds
}

func TestConfigStructFields(t *testing.T) {
	cfg := &Config{
		ServerPort: "8085",
		ServerHost: "0.0.0.0",
		LogLevel:   "info",
	}

	// Test that all required fields are present
	assert.NotEmpty(t, cfg.ServerPort)
	assert.NotEmpty(t, cfg.ServerHost)
	assert.NotEmpty(t, cfg.LogLevel)
	// Database settings are now loaded from data service, not from config
}

func TestInvoiceServiceSpecificPort(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnvVars()
	defer clearEnvVars()

	cfg := LoadConfig()

	// Invoice service should default to port 8085 (different from inventory service 8084)
	assert.Equal(t, "8085", cfg.ServerPort)
}

func TestConfigConsistencyAcrossServices(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnvVars()
	defer clearEnvVars()

	_ = LoadConfig()

	// Database settings are now loaded from data service, not from config
	// All services use the same data service for configuration
}

// Helper function to clear all relevant environment variables
func clearEnvVars() {
	vars := []string{
		"INVOICE_SERVER_PORT",
		"INVOICE_SERVER_HOST",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SSLMODE",
		"LOG_LEVEL",
		"TEST_STRING_VAR",
		"NON_EXISTING_VAR",
		"EMPTY_VAR",
		"TEST_INT_VAR",
		"INVALID_INT_VAR",
		"NON_EXISTING_INT_VAR",
		"EMPTY_INT_VAR",
		"TEST_DURATION_VAR",
		"INVALID_DURATION_VAR",
		"NON_EXISTING_DURATION_VAR",
		"COMPLEX_DURATION_VAR",
	}

	for _, v := range vars {
		os.Unsetenv(v)
	}
}
