package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfig tests the default configuration loading
func TestLoadConfig(t *testing.T) {
	config := LoadConfig()

	// Server settings
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, "8083", config.ServerPort)

	// Database settings
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "5432", config.DBPort)
	assert.Equal(t, "postgres", config.DBUser)
	assert.Equal(t, "postgres123", config.DBPassword)
	assert.Equal(t, "icecream_store", config.DBName)
	assert.Equal(t, "disable", config.DBSSLMode)

	// Business settings
	assert.Equal(t, 13.0, config.DefaultTaxRate)
	assert.Equal(t, 10.0, config.DefaultServiceRate)
	assert.Equal(t, 30, config.OrderTimeout)
}

// TestGetEnv tests the getEnv helper function
func TestGetEnv(t *testing.T) {
	// Clean up any existing value
	os.Unsetenv("TEST_VAR")

	// Test with environment variable set
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", result)

	// Test with environment variable not set
	os.Unsetenv("TEST_VAR")
	result = getEnv("TEST_VAR", "default")
	assert.Equal(t, "default", result)
}
