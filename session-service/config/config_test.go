package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test default values
	config := LoadConfig()

	if config.ServerPort != "8081" {
		t.Errorf("Expected ServerPort to be '8081', got '%s'", config.ServerPort)
	}

	if config.ServerHost != "0.0.0.0" {
		t.Errorf("Expected ServerHost to be '0.0.0.0', got '%s'", config.ServerHost)
	}

	if config.DatabaseHost != "localhost" {
		t.Errorf("Expected DatabaseHost to be 'localhost', got '%s'", config.DatabaseHost)
	}

	if config.DatabasePort != 5432 {
		t.Errorf("Expected DatabasePort to be 5432, got %d", config.DatabasePort)
	}

	if config.JWTExpirationTime != 30*time.Minute {
		t.Errorf("Expected JWTExpirationTime to be 30m, got %v", config.JWTExpirationTime)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SESSION_SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("JWT_EXPIRATION_TIME", "1h")

	config := LoadConfig()

	if config.ServerPort != "9090" {
		t.Errorf("Expected ServerPort to be '9090', got '%s'", config.ServerPort)
	}

	if config.DatabaseHost != "test-host" {
		t.Errorf("Expected DatabaseHost to be 'test-host', got '%s'", config.DatabaseHost)
	}

	if config.JWTExpirationTime != time.Hour {
		t.Errorf("Expected JWTExpirationTime to be 1h, got %v", config.JWTExpirationTime)
	}

	// Clean up
	os.Unsetenv("SESSION_SERVER_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("JWT_EXPIRATION_TIME")
}
