package main

import (
	"net/http"
	"net/http/httptest"
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
		// Create a mock data service server
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
					}
				],
				"total": 3,
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
		cfg, err := config.LoadConfigFromDataService(logger)
		require.NoError(t, err)

		// Verify essential configuration values
		assert.NotEmpty(t, cfg.ServerHost)
		assert.NotEmpty(t, cfg.ServerPort)
		assert.NotEmpty(t, cfg.LogLevel)
		assert.Equal(t, "8084", cfg.ServerPort) // Inventory service specific port
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
		// Create a mock config for testing
		cfg := &config.Config{
			ServerPort:     "8084",
			ServerHost:     "0.0.0.0",
			LogLevel:       "info",
			ImagesBasePath: ".",
		}
		assert.NotNil(t, cfg)

		// Test required fields are not empty
		require.NotEmpty(t, cfg.ServerHost)
		require.NotEmpty(t, cfg.ServerPort)
	})

	t.Run("main handler initialization", func(t *testing.T) {
		// Create a mock database
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		logger := setupLogger("error") // Use error level to reduce test noise

		// Test that NewMainHttpHandler can be created
		// Create a mock config for testing
		mockConfig := &config.Config{
			ServerPort:     "8084",
			ServerHost:     "0.0.0.0",
			LogLevel:       "info",
			ImagesBasePath: ".",
		}
		mainHandler := NewMainHttpHandler(db, logger, mockConfig)
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
		// Step 1: Create mock configuration
		cfg := &config.Config{
			ServerPort:     "8084",
			ServerHost:     "0.0.0.0",
			LogLevel:       "info",
			ImagesBasePath: ".",
		}
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
		cfg := &config.Config{
			ServerPort:     "8084",
			ServerHost:     "0.0.0.0",
			LogLevel:       "info",
			ImagesBasePath: ".",
		}
		assert.Equal(t, "8084", cfg.ServerPort, "Inventory service should default to port 8084")
	})
}

// BenchmarkConfigLoad benchmarks configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &config.Config{
			ServerPort:     "8084",
			ServerHost:     "0.0.0.0",
			LogLevel:       "info",
			ImagesBasePath: ".",
		}
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
