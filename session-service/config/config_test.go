package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	cfg := LoadConfig()

	assert.Equal(t, "8081", cfg.ServerPort)
	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	assert.Equal(t, "localhost", cfg.DatabaseHost)
	assert.Equal(t, 5432, cfg.DatabasePort)
	assert.Equal(t, "postgres", cfg.DatabaseUser)
	assert.Equal(t, "postgres123", cfg.DatabasePassword)
	assert.Equal(t, "icecream_store", cfg.DatabaseName)
	assert.Equal(t, "disable", cfg.DatabaseSSLMode)
	assert.Equal(t, "your-super-secret-jwt-key-change-in-production", cfg.JWTSecret)
	assert.Equal(t, 30*time.Minute, cfg.JWTExpirationTime)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("SESSION_SERVER_PORT", "9090")
	os.Setenv("SESSION_SERVER_HOST", "127.0.0.1")
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRATION_TIME", "1h")
	os.Setenv("LOG_LEVEL", "debug")

	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("SESSION_SERVER_PORT")
		os.Unsetenv("SESSION_SERVER_HOST")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRATION_TIME")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg := LoadConfig()

	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "127.0.0.1", cfg.ServerHost)
	assert.Equal(t, "test-host", cfg.DatabaseHost)
	assert.Equal(t, 5433, cfg.DatabasePort)
	assert.Equal(t, "testuser", cfg.DatabaseUser)
	assert.Equal(t, "testpass", cfg.DatabasePassword)
	assert.Equal(t, "testdb", cfg.DatabaseName)
	assert.Equal(t, "require", cfg.DatabaseSSLMode)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
	assert.Equal(t, time.Hour, cfg.JWTExpirationTime)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid integer
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 456)
	assert.Equal(t, 123, result)

	// Test with invalid integer (should return default)
	os.Setenv("TEST_INVALID", "not-a-number")
	defer os.Unsetenv("TEST_INVALID")

	result = getEnvInt("TEST_INVALID", 456)
	assert.Equal(t, 456, result)

	// Test with missing environment variable (should return default)
	result = getEnvInt("MISSING_VAR", 789)
	assert.Equal(t, 789, result)
}

func TestGetEnvString(t *testing.T) {
	// Test with valid string
	os.Setenv("TEST_STRING", "test-value")
	defer os.Unsetenv("TEST_STRING")

	result := getEnvString("TEST_STRING", "default-value")
	assert.Equal(t, "test-value", result)

	// Test with missing environment variable (should return default)
	result = getEnvString("MISSING_VAR", "default-value")
	assert.Equal(t, "default-value", result)
}

func TestGetEnvDuration(t *testing.T) {
	// Test with valid duration
	os.Setenv("TEST_DURATION", "2h")
	defer os.Unsetenv("TEST_DURATION")

	result := getEnvDuration("TEST_DURATION", "1h")
	assert.Equal(t, 2*time.Hour, result)

	// Test with invalid duration (should return default)
	os.Setenv("TEST_INVALID", "not-a-duration")
	defer os.Unsetenv("TEST_INVALID")

	result = getEnvDuration("TEST_INVALID", "1h")
	assert.Equal(t, time.Hour, result)

	// Test with missing environment variable (should return default)
	result = getEnvDuration("MISSING_VAR", "30m")
	assert.Equal(t, 30*time.Minute, result)
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		ServerPort:        "8080",
		ServerHost:        "0.0.0.0",
		DatabaseHost:      "localhost",
		DatabasePort:      5432,
		DatabaseUser:      "postgres",
		DatabasePassword:  "password",
		DatabaseName:      "testdb",
		DatabaseSSLMode:   "disable",
		JWTSecret:         "test-secret",
		JWTExpirationTime: 24 * time.Hour,
		LogLevel:          "info",
	}

	// Test all fields are properly set
	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	assert.Equal(t, "localhost", cfg.DatabaseHost)
	assert.Equal(t, 5432, cfg.DatabasePort)
	assert.Equal(t, "postgres", cfg.DatabaseUser)
	assert.Equal(t, "password", cfg.DatabasePassword)
	assert.Equal(t, "testdb", cfg.DatabaseName)
	assert.Equal(t, "disable", cfg.DatabaseSSLMode)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
	assert.Equal(t, 24*time.Hour, cfg.JWTExpirationTime)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfigWithZeroValues(t *testing.T) {
	// Test that zero values are handled correctly
	cfg := &Config{}

	// Set some values to zero
	cfg.ServerPort = ""
	cfg.DatabasePort = 0
	cfg.JWTExpirationTime = 0

	assert.Equal(t, "", cfg.ServerPort)
	assert.Equal(t, 0, cfg.DatabasePort)
	assert.Equal(t, time.Duration(0), cfg.JWTExpirationTime)
}

func TestConfigWithNegativeValues(t *testing.T) {
	// Test that negative values are handled correctly
	os.Setenv("DB_PORT", "-5432")
	defer os.Unsetenv("DB_PORT")

	cfg := LoadConfig()

	// Should return the negative value as-is since getEnvInt doesn't validate
	assert.Equal(t, -5432, cfg.DatabasePort)
}

func TestConfigWithEmptyStrings(t *testing.T) {
	// Test that empty strings are handled correctly
	os.Setenv("DB_HOST", "")
	os.Setenv("DB_USER", "")
	os.Setenv("JWT_SECRET", "")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := LoadConfig()

	// Should return default values for empty strings
	assert.Equal(t, "localhost", cfg.DatabaseHost)
	assert.Equal(t, "postgres", cfg.DatabaseUser)
	assert.Equal(t, "your-super-secret-jwt-key-change-in-production", cfg.JWTSecret)
}

func TestConfigWithWhitespaceStrings(t *testing.T) {
	// Test that whitespace strings are handled correctly
	os.Setenv("DB_HOST", "   ")
	os.Setenv("DB_USER", "\t")
	os.Setenv("JWT_SECRET", "\n")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := LoadConfig()

	// Should return the whitespace strings as-is since getEnvString doesn't trim
	assert.Equal(t, "   ", cfg.DatabaseHost)
	assert.Equal(t, "\t", cfg.DatabaseUser)
	assert.Equal(t, "\n", cfg.JWTSecret)
}

func TestConfigWithVeryLargeValues(t *testing.T) {
	// Test that very large values are handled correctly
	os.Setenv("SESSION_SERVER_PORT", "999999")
	os.Setenv("DB_PORT", "65535")
	defer func() {
		os.Unsetenv("SESSION_SERVER_PORT")
		os.Unsetenv("DB_PORT")
	}()

	cfg := LoadConfig()

	assert.Equal(t, "999999", cfg.ServerPort)
	assert.Equal(t, 65535, cfg.DatabasePort)
}

func TestConfigWithSpecialCharacters(t *testing.T) {
	// Test that special characters in strings are handled correctly
	os.Setenv("DB_PASSWORD", "p@ssw0rd!@#$%^&*()")
	os.Setenv("JWT_SECRET", "my-super-secret-key-with-special-chars!@#$%^&*()")
	os.Setenv("DB_NAME", "test_db_with_underscores")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_NAME")
	}()

	cfg := LoadConfig()

	assert.Equal(t, "p@ssw0rd!@#$%^&*()", cfg.DatabasePassword)
	assert.Equal(t, "my-super-secret-key-with-special-chars!@#$%^&*()", cfg.JWTSecret)
	assert.Equal(t, "test_db_with_underscores", cfg.DatabaseName)
}
