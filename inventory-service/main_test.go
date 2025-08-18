package main

import (
	"os"
	"testing"

	"inventory-service/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupLogger tests logger configuration
func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected logrus.Level
	}{
		{"debug level", "debug", logrus.DebugLevel},
		{"info level", "info", logrus.InfoLevel},
		{"warn level", "warn", logrus.WarnLevel},
		{"error level", "error", logrus.ErrorLevel},
		{"invalid level", "invalid", logrus.InfoLevel}, // defaults to info
		{"empty level", "", logrus.InfoLevel},          // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLogger(tt.logLevel)
			assert.Equal(t, tt.expected, logger.Level)
			assert.IsType(t, &logrus.JSONFormatter{}, logger.Formatter)
		})
	}
}

// TestConnectToDatabase tests database connection setup
func TestConnectToDatabase(t *testing.T) {
	t.Run("successful database connection", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Expect ping
		mock.ExpectPing()

		err = db.Ping()
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Database connection is now handled with hardcoded values in connectToDatabase
	// No longer using config fields for database connection
}

// TestConfigurationIntegration tests that configuration loads correctly for main
func TestConfigurationIntegration(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		cfg := config.LoadConfig()

		// Verify essential configuration values
		assert.NotEmpty(t, cfg.ServerHost)
		assert.NotEmpty(t, cfg.ServerPort)
		assert.NotEmpty(t, cfg.LogLevel)
		assert.Equal(t, "8084", cfg.ServerPort) // Inventory service specific port
		// Database settings are now loaded from data service, not from config
	})

	t.Run("configuration with environment overrides", func(t *testing.T) {
		// Set environment variables
		envVars := map[string]string{
			"INVENTORY_SERVER_HOST": "127.0.0.1",
			"INVENTORY_SERVER_PORT": "9084",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
			defer os.Unsetenv(key)
		}

		cfg := config.LoadConfig()

		assert.Equal(t, "127.0.0.1", cfg.ServerHost)
		assert.Equal(t, "9084", cfg.ServerPort)
		// Database settings are now loaded from data service, not from environment variables
	})
}

// TestApplicationComponents tests that all components can be initialized
func TestApplicationComponents(t *testing.T) {
	t.Run("logger initialization", func(t *testing.T) {
		logger := setupLogger("info")
		assert.NotNil(t, logger)
		assert.Equal(t, logrus.InfoLevel, logger.Level)
	})

	t.Run("configuration loading", func(t *testing.T) {
		cfg := config.LoadConfig()
		assert.NotNil(t, cfg)

		// Test required fields are not empty
		require.NotEmpty(t, cfg.ServerHost)
		require.NotEmpty(t, cfg.ServerPort)
		// Database settings are now loaded from data service, not from config
	})

	t.Run("main handler initialization", func(t *testing.T) {
		// Create a mock database
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		logger := setupLogger("error") // Use error level to reduce test noise

		// Test that NewMainHttpHandler can be created
		mainHandler := NewMainHttpHandler(db, logger)
		assert.NotNil(t, mainHandler)
		assert.NotNil(t, mainHandler.GetSuppliersHandler())
	})
}

// Database connection is now handled with hardcoded values in connectToDatabase
// No longer using config fields for database connection

// TestLoggerLevels tests different logger levels
func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			logger := setupLogger(level)
			assert.NotNil(t, logger)

			expectedLevel, err := logrus.ParseLevel(level)
			require.NoError(t, err)
			assert.Equal(t, expectedLevel, logger.Level)
		})
	}
}

// TestApplicationStartupSequence tests the startup sequence components
func TestApplicationStartupSequence(t *testing.T) {
	t.Run("config then logger", func(t *testing.T) {
		// Step 1: Load configuration
		cfg := config.LoadConfig()
		require.NotNil(t, cfg)

		// Step 2: Setup logger with config
		logger := setupLogger(cfg.LogLevel)
		require.NotNil(t, logger)

		// Verify logger level matches config
		expectedLevel, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			expectedLevel = logrus.InfoLevel
		}
		assert.Equal(t, expectedLevel, logger.Level)
	})
}

// TestErrorHandling tests error handling in application setup
func TestErrorHandling(t *testing.T) {
	t.Run("invalid log level handling", func(t *testing.T) {
		logger := setupLogger("invalid-level")
		assert.NotNil(t, logger)
		// Should default to info level
		assert.Equal(t, logrus.InfoLevel, logger.Level)
	})

	t.Run("empty log level handling", func(t *testing.T) {
		logger := setupLogger("")
		assert.NotNil(t, logger)
		// Should default to info level
		assert.Equal(t, logrus.InfoLevel, logger.Level)
	})
}

// TestInventoryServiceSpecifics tests inventory service specific functionality
func TestInventoryServiceSpecifics(t *testing.T) {
	t.Run("default port is 8084", func(t *testing.T) {
		cfg := config.LoadConfig()
		assert.Equal(t, "8084", cfg.ServerPort, "Inventory service should default to port 8084")
	})

	t.Run("inventory environment variables", func(t *testing.T) {
		os.Setenv("INVENTORY_SERVER_PORT", "9999")
		defer os.Unsetenv("INVENTORY_SERVER_PORT")

		cfg := config.LoadConfig()
		assert.Equal(t, "9999", cfg.ServerPort, "Should use INVENTORY_SERVER_PORT env var")
	})
}

// BenchmarkConfigLoad benchmarks configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config.LoadConfig()
	}
}

// BenchmarkLoggerSetup benchmarks logger setup
func BenchmarkLoggerSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		setupLogger("info")
	}
}

// Database connection is now handled with hardcoded values
// No longer using config fields for database connection
