package config

import (
	"os"
	"testing"
	"time"

	"session-service/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig tests the default configuration loading
func TestLoadConfig(t *testing.T) {
	config := LoadConfig()

	// Server settings
	assert.Equal(t, "8081", config.ServerPort)
	assert.Equal(t, "0.0.0.0", config.ServerHost)

	// JWT settings
	assert.Equal(t, "your-super-secret-jwt-key-change-in-production", config.JWTSecret)
	assert.Equal(t, 30*time.Minute, config.JWTExpirationTime)
	assert.Equal(t, 5*time.Minute, config.JWTRefreshThreshold)

	// Session settings
	assert.Equal(t, 30*time.Minute, config.SessionDefaultExpiration)
	assert.Equal(t, 168*time.Hour, config.SessionRememberMeExpiration) // 7 days
	assert.Equal(t, 10*time.Minute, config.SessionCleanupInterval)
	assert.Equal(t, 5, config.SessionMaxConcurrent)
	assert.Equal(t, "memory", config.SessionStorageType)

	// Security settings
	assert.Equal(t, 12, config.BcryptCost)
	assert.Equal(t, 5, config.MaxLoginAttempts)
	assert.Equal(t, 15*time.Minute, config.LoginCooldownTime)

	// Database settings
	assert.Equal(t, "localhost", config.DatabaseHost)
	assert.Equal(t, 5432, config.DatabasePort)
	assert.Equal(t, "postgres", config.DatabaseUser)
	assert.Equal(t, "postgres123", config.DatabasePassword)
	assert.Equal(t, "icecream_store", config.DatabaseName)
	assert.Equal(t, "disable", config.DatabaseSSLMode)

	// Logging
	assert.Equal(t, "info", config.LogLevel)
}

// TestLoadConfigWithEnvironmentVariables tests configuration loading with environment variables
func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"SESSION_SERVER_PORT":            "9090",
		"SESSION_SERVER_HOST":            "127.0.0.1",
		"JWT_SECRET":                     "test-secret-key",
		"JWT_EXPIRATION_TIME":            "60m",
		"JWT_REFRESH_THRESHOLD":          "10m",
		"SESSION_DEFAULT_EXPIRATION":     "45m",
		"SESSION_REMEMBER_ME_EXPIRATION": "240h", // 10 days
		"SESSION_CLEANUP_INTERVAL":       "15m",
		"SESSION_MAX_CONCURRENT":         "10",
		"SESSION_STORAGE_TYPE":           "redis",
		"BCRYPT_COST":                    "14",
		"MAX_LOGIN_ATTEMPTS":             "3",
		"LOGIN_COOLDOWN_TIME":            "30m",
		"DB_HOST":                        "db.example.com",
		"DB_PORT":                        "3306",
		"DB_USER":                        "testuser",
		"DB_PASSWORD":                    "testpass",
		"DB_NAME":                        "testdb",
		"DB_SSLMODE":                     "require",
		"LOG_LEVEL":                      "debug",
	}

	// Set environment variables and defer cleanup
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config := LoadConfig()

	// Verify environment variables were used
	assert.Equal(t, "9090", config.ServerPort)
	assert.Equal(t, "127.0.0.1", config.ServerHost)
	assert.Equal(t, "test-secret-key", config.JWTSecret)
	assert.Equal(t, 60*time.Minute, config.JWTExpirationTime)
	assert.Equal(t, 10*time.Minute, config.JWTRefreshThreshold)
	assert.Equal(t, 45*time.Minute, config.SessionDefaultExpiration)
	assert.Equal(t, 240*time.Hour, config.SessionRememberMeExpiration)
	assert.Equal(t, 15*time.Minute, config.SessionCleanupInterval)
	assert.Equal(t, 10, config.SessionMaxConcurrent)
	assert.Equal(t, "redis", config.SessionStorageType)
	assert.Equal(t, 14, config.BcryptCost)
	assert.Equal(t, 3, config.MaxLoginAttempts)
	assert.Equal(t, 30*time.Minute, config.LoginCooldownTime)
	assert.Equal(t, "db.example.com", config.DatabaseHost)
	assert.Equal(t, 3306, config.DatabasePort)
	assert.Equal(t, "testuser", config.DatabaseUser)
	assert.Equal(t, "testpass", config.DatabasePassword)
	assert.Equal(t, "testdb", config.DatabaseName)
	assert.Equal(t, "require", config.DatabaseSSLMode)
	assert.Equal(t, "debug", config.LogLevel)
}

// TestToSessionConfig tests the conversion to session-specific config
func TestToSessionConfig(t *testing.T) {
	config := &Config{
		SessionDefaultExpiration:    45 * time.Minute,
		SessionRememberMeExpiration: 72 * time.Hour,
		JWTRefreshThreshold:         10 * time.Minute,
		SessionCleanupInterval:      20 * time.Minute,
		SessionMaxConcurrent:        8,
		SessionStorageType:          "database",
	}

	sessionConfig := config.ToSessionConfig()

	assert.Equal(t, 45*time.Minute, sessionConfig.DefaultExpiration)
	assert.Equal(t, 72*time.Hour, sessionConfig.RememberMeExpiration)
	assert.Equal(t, 10*time.Minute, sessionConfig.RefreshThreshold)
	assert.Equal(t, 20*time.Minute, sessionConfig.CleanupInterval)
	assert.Equal(t, 8, sessionConfig.MaxConcurrentSessions)
	assert.Equal(t, "database", sessionConfig.StorageType)

	// Verify it implements the SessionConfig interface correctly
	var _ *models.SessionConfig = sessionConfig
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
		defaultValue string
		expected     time.Duration
	}{
		{
			name:         "with valid duration environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "1h30m",
			defaultValue: "10m",
			expected:     1*time.Hour + 30*time.Minute,
		},
		{
			name:         "with invalid duration environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "invalid_duration",
			defaultValue: "10m",
			expected:     10 * time.Minute,
		},
		{
			name:         "with empty environment variable",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "",
			defaultValue: "15m",
			expected:     15 * time.Minute,
		},
		{
			name:         "without environment variable",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "20m",
			expected:     20 * time.Minute,
		},
		{
			name:         "with invalid default value",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "invalid",
			defaultValue: "invalid_default",
			expected:     10 * time.Minute, // Ultimate fallback
		},
		{
			name:         "with seconds duration",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "30s",
			defaultValue: "10m",
			expected:     30 * time.Second,
		},
		{
			name:         "with nanoseconds duration",
			envKey:       "TEST_DURATION_VAR",
			envValue:     "500ms",
			defaultValue: "10m",
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

	// Test that JWT expiration is reasonable
	assert.True(t, config.JWTExpirationTime > 0)
	assert.True(t, config.JWTRefreshThreshold > 0)
	assert.True(t, config.JWTRefreshThreshold < config.JWTExpirationTime)

	// Test that session expirations are reasonable
	assert.True(t, config.SessionDefaultExpiration > 0)
	assert.True(t, config.SessionRememberMeExpiration > config.SessionDefaultExpiration)
	assert.True(t, config.SessionCleanupInterval > 0)

	// Test that security settings are reasonable
	assert.True(t, config.BcryptCost >= 4)  // Minimum bcrypt cost
	assert.True(t, config.BcryptCost <= 31) // Maximum bcrypt cost
	assert.True(t, config.MaxLoginAttempts > 0)
	assert.True(t, config.LoginCooldownTime > 0)

	// Test that database port is valid
	assert.True(t, config.DatabasePort > 0)
	assert.True(t, config.DatabasePort <= 65535)

	// Test that session max concurrent is reasonable
	assert.True(t, config.SessionMaxConcurrent > 0)
}

// TestConfigEnvironmentVariableOverrides tests specific environment variable scenarios
func TestConfigEnvironmentVariableOverrides(t *testing.T) {
	t.Run("JWT_SECRET override", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "production-secret-key")
		defer os.Unsetenv("JWT_SECRET")

		config := LoadConfig()
		assert.Equal(t, "production-secret-key", config.JWTSecret)
	})

	t.Run("Database configuration override", func(t *testing.T) {
		os.Setenv("DB_HOST", "prod-db.example.com")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_NAME", "prod_icecream_store")
		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_NAME")
		}()

		config := LoadConfig()
		assert.Equal(t, "prod-db.example.com", config.DatabaseHost)
		assert.Equal(t, 5433, config.DatabasePort)
		assert.Equal(t, "prod_icecream_store", config.DatabaseName)
	})

	t.Run("Session storage type override", func(t *testing.T) {
		os.Setenv("SESSION_STORAGE_TYPE", "redis")
		defer os.Unsetenv("SESSION_STORAGE_TYPE")

		config := LoadConfig()
		assert.Equal(t, "redis", config.SessionStorageType)
	})
}

// BenchmarkLoadConfig benchmarks the configuration loading process
func BenchmarkLoadConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadConfig()
	}
}

// BenchmarkToSessionConfig benchmarks the session config conversion
func BenchmarkToSessionConfig(b *testing.B) {
	config := LoadConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.ToSessionConfig()
	}
}

// TestConfigStructFieldsAreParsedCorrectly tests that all struct fields are properly populated
func TestConfigStructFieldsAreParsedCorrectly(t *testing.T) {
	config := LoadConfig()

	// Use reflection to verify no zero values for required fields
	require.NotEmpty(t, config.ServerPort)
	require.NotEmpty(t, config.ServerHost)
	require.NotEmpty(t, config.JWTSecret)
	require.NotZero(t, config.JWTExpirationTime)
	require.NotZero(t, config.JWTRefreshThreshold)
	require.NotZero(t, config.SessionDefaultExpiration)
	require.NotZero(t, config.SessionRememberMeExpiration)
	require.NotZero(t, config.SessionCleanupInterval)
	require.NotZero(t, config.SessionMaxConcurrent)
	require.NotEmpty(t, config.SessionStorageType)
	require.NotZero(t, config.BcryptCost)
	require.NotZero(t, config.MaxLoginAttempts)
	require.NotZero(t, config.LoginCooldownTime)
	require.NotEmpty(t, config.DatabaseHost)
	require.NotZero(t, config.DatabasePort)
	require.NotEmpty(t, config.DatabaseUser)
	require.NotEmpty(t, config.DatabasePassword)
	require.NotEmpty(t, config.DatabaseName)
	require.NotEmpty(t, config.DatabaseSSLMode)
	require.NotEmpty(t, config.LogLevel)
}

// TestEdgeCases tests edge cases for environment variable parsing
func TestEdgeCases(t *testing.T) {
	t.Run("very large numbers", func(t *testing.T) {
		os.Setenv("TEST_LARGE_INT", "999999999")
		defer os.Unsetenv("TEST_LARGE_INT")

		result := getEnvInt("TEST_LARGE_INT", 10)
		assert.Equal(t, 999999999, result)
	})

	t.Run("very long duration", func(t *testing.T) {
		os.Setenv("TEST_LONG_DURATION", "8760h") // 1 year
		defer os.Unsetenv("TEST_LONG_DURATION")

		result := getEnvDuration("TEST_LONG_DURATION", "10m")
		assert.Equal(t, 8760*time.Hour, result)
	})

	t.Run("duration with mixed units", func(t *testing.T) {
		os.Setenv("TEST_MIXED_DURATION", "1h30m45s")
		defer os.Unsetenv("TEST_MIXED_DURATION")

		result := getEnvDuration("TEST_MIXED_DURATION", "10m")
		expected := 1*time.Hour + 30*time.Minute + 45*time.Second
		assert.Equal(t, expected, result)
	})
}
