package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig tests the default configuration loading
func TestLoadConfig(t *testing.T) {
	config := LoadConfig()

	// Server settings
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, "8084", config.ServerPort)

	// Database settings
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "5432", config.DBPort)
	assert.Equal(t, "postgres", config.DBUser)
	assert.Equal(t, "postgres123", config.DBPassword)
	assert.Equal(t, "icecream_store", config.DBName)
	assert.Equal(t, "disable", config.DBSSLMode)

	// Logging
	assert.Equal(t, "info", config.LogLevel)
}

// TestLoadConfigWithEnvironmentVariables tests configuration loading with environment variables
func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"INVENTORY_SERVER_HOST": "127.0.0.1",
		"INVENTORY_SERVER_PORT": "9084",
		"DB_HOST":               "db.example.com",
		"DB_PORT":               "3306",
		"DB_USER":               "testuser",
		"DB_PASSWORD":           "testpass",
		"DB_NAME":               "testdb",
		"DB_SSLMODE":            "require",
		"LOG_LEVEL":             "debug",
	}

	// Set environment variables and defer cleanup
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config := LoadConfig()

	// Verify environment variables were used
	assert.Equal(t, "127.0.0.1", config.ServerHost)
	assert.Equal(t, "9084", config.ServerPort)
	assert.Equal(t, "db.example.com", config.DBHost)
	assert.Equal(t, "3306", config.DBPort)
	assert.Equal(t, "testuser", config.DBUser)
	assert.Equal(t, "testpass", config.DBPassword)
	assert.Equal(t, "testdb", config.DBName)
	assert.Equal(t, "require", config.DBSSLMode)
	assert.Equal(t, "debug", config.LogLevel)
}

// TestGetEnvString tests the getEnvString helper function
func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "with environment variable set",
			envKey:       "TEST_STRING_VAR",
			envValue:     "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
		{
			name:         "with empty environment variable",
			envKey:       "TEST_STRING_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "without environment variable",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing value
			os.Unsetenv(tt.envKey)

			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvString(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvInt tests the getEnvInt helper function
func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "with valid integer environment variable",
			envKey:       "TEST_INT_VAR",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "with invalid integer environment variable",
			envKey:       "TEST_INT_VAR",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "with float environment variable",
			envKey:       "TEST_INT_VAR",
			envValue:     "42.5",
			defaultValue: 10,
			expected:     10, // Should fail to parse and use default
		},
		{
			name:         "with empty environment variable",
			envKey:       "TEST_INT_VAR",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "without environment variable",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "with negative integer",
			envKey:       "TEST_INT_VAR",
			envValue:     "-5",
			defaultValue: 10,
			expected:     -5,
		},
		{
			name:         "with zero value",
			envKey:       "TEST_INT_VAR",
			envValue:     "0",
			defaultValue: 10,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing value
			os.Unsetenv(tt.envKey)

			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvInt(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvDuration tests the getEnvDuration helper function
func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "with valid duration environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "30s",
			defaultValue: 10 * time.Second,
			expected:     30 * time.Second,
		},
		{
			name:         "with minutes duration",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "5m",
			defaultValue: 10 * time.Second,
			expected:     5 * time.Minute,
		},
		{
			name:         "with hours duration",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "2h",
			defaultValue: 10 * time.Second,
			expected:     2 * time.Hour,
		},
		{
			name:         "with invalid duration environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "not_a_duration",
			defaultValue: 10 * time.Second,
			expected:     10 * time.Second,
		},
		{
			name:         "with empty environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "",
			defaultValue: 10 * time.Second,
			expected:     10 * time.Second,
		},
		{
			name:         "without environment variable",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: 10 * time.Second,
			expected:     10 * time.Second,
		},
		{
			name:         "with milliseconds duration",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "500ms",
			defaultValue: 10 * time.Second,
			expected:     500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing value
			os.Unsetenv(tt.envKey)

			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvDuration(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConfigConsistency tests that all configurations are consistent
func TestConfigConsistency(t *testing.T) {
	config := LoadConfig()

	// Test that required fields are not empty
	require.NotEmpty(t, config.ServerHost)
	require.NotEmpty(t, config.ServerPort)
	require.NotEmpty(t, config.DBHost)
	require.NotEmpty(t, config.DBPort)
	require.NotEmpty(t, config.DBUser)
	require.NotEmpty(t, config.DBName)
	require.NotEmpty(t, config.LogLevel)

	// Test that server port is a valid number
	assert.Regexp(t, `^\d+$`, config.ServerPort, "Server port should be numeric")

	// Test that DB port is a valid number
	assert.Regexp(t, `^\d+$`, config.DBPort, "DB port should be numeric")

	// Test valid SSL modes
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	assert.Contains(t, validSSLModes, config.DBSSLMode, "DB SSL mode should be valid")

	// Test valid log levels
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	assert.Contains(t, validLogLevels, config.LogLevel, "Log level should be valid")
}

// TestInventorySpecificConfig tests inventory service specific configuration
func TestInventorySpecificConfig(t *testing.T) {
	config := LoadConfig()

	// Test default inventory service port
	assert.Equal(t, "8084", config.ServerPort, "Default inventory service port should be 8084")

	// Test that it uses inventory-specific environment variables
	os.Setenv("INVENTORY_SERVER_PORT", "9999")
	defer os.Unsetenv("INVENTORY_SERVER_PORT")

	config2 := LoadConfig()
	assert.Equal(t, "9999", config2.ServerPort, "Should use INVENTORY_SERVER_PORT env var")
}

// TestEnvironmentVariableOverrides tests specific environment variable scenarios
func TestEnvironmentVariableOverrides(t *testing.T) {
	t.Run("server configuration override", func(t *testing.T) {
		os.Setenv("INVENTORY_SERVER_HOST", "inventory.example.com")
		os.Setenv("INVENTORY_SERVER_PORT", "8585")
		defer func() {
			os.Unsetenv("INVENTORY_SERVER_HOST")
			os.Unsetenv("INVENTORY_SERVER_PORT")
		}()

		config := LoadConfig()
		assert.Equal(t, "inventory.example.com", config.ServerHost)
		assert.Equal(t, "8585", config.ServerPort)
	})

	t.Run("database configuration override", func(t *testing.T) {
		os.Setenv("DB_HOST", "prod-db.example.com")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_NAME", "prod_icecream_store")
		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_NAME")
		}()

		config := LoadConfig()
		assert.Equal(t, "prod-db.example.com", config.DBHost)
		assert.Equal(t, "5433", config.DBPort)
		assert.Equal(t, "prod_icecream_store", config.DBName)
	})
}

// TestEdgeCases tests edge cases for environment variable parsing
func TestEdgeCases(t *testing.T) {
	t.Run("very large numbers", func(t *testing.T) {
		os.Setenv("TEST_LARGE_INT", "999999999")
		defer os.Unsetenv("TEST_LARGE_INT")

		result := getEnvInt("TEST_LARGE_INT", 10)
		assert.Equal(t, 999999999, result)
	})

	t.Run("complex duration", func(t *testing.T) {
		os.Setenv("TEST_COMPLEX_DURATION", "1h30m45s")
		defer os.Unsetenv("TEST_COMPLEX_DURATION")

		expected := 1*time.Hour + 30*time.Minute + 45*time.Second
		result := getEnvDuration("TEST_COMPLEX_DURATION", 10*time.Second)
		assert.Equal(t, expected, result)
	})

	t.Run("string with spaces", func(t *testing.T) {
		os.Setenv("TEST_SPACES", " value with spaces ")
		defer os.Unsetenv("TEST_SPACES")

		result := getEnvString("TEST_SPACES", "default")
		assert.Equal(t, " value with spaces ", result)
	})
}

// BenchmarkLoadConfig benchmarks the configuration loading process
func BenchmarkLoadConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadConfig()
	}
}

// BenchmarkGetEnvString benchmarks string environment variable retrieval
func BenchmarkGetEnvString(b *testing.B) {
	os.Setenv("BENCH_TEST_VAR", "test_value")
	defer os.Unsetenv("BENCH_TEST_VAR")

	for i := 0; i < b.N; i++ {
		getEnvString("BENCH_TEST_VAR", "default")
	}
}
