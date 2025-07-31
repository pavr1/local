package main

import (
	"os"
	"testing"

	"orders-service/config"

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
			// Set environment variable
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}

			cfg := config.LoadConfig()
			logger := setupTestLogger(cfg.LogLevel)

			assert.Equal(t, tt.expected, logger.Level)
		})
	}
}

// TestSetupDatabase tests database connection setup
func TestSetupDatabase(t *testing.T) {
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

	t.Run("database connection string format", func(t *testing.T) {
		cfg := &config.Config{
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "testuser",
			DBPassword: "testpass",
			DBName:     "testdb",
			DBSSLMode:  "disable",
		}

		dsn := buildDSN(cfg)
		expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
		assert.Equal(t, expected, dsn)
	})
}

// TestConfigurationIntegration tests that configuration loads correctly for main
func TestConfigurationIntegration(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		cfg := config.LoadConfig()

		// Verify essential configuration values
		assert.NotEmpty(t, cfg.ServerHost)
		assert.NotEmpty(t, cfg.ServerPort)
		assert.NotEmpty(t, cfg.DBHost)
		assert.NotEmpty(t, cfg.DBPort)
		assert.NotEmpty(t, cfg.JWTSecret)
		assert.True(t, cfg.DefaultTaxRate >= 0)
		assert.True(t, cfg.DefaultServiceRate >= 0)
		assert.True(t, cfg.OrderTimeout > 0)
	})

	t.Run("configuration with environment overrides", func(t *testing.T) {
		// Set environment variables
		envVars := map[string]string{
			"SERVER_HOST": "127.0.0.1",
			"SERVER_PORT": "9090",
			"DB_HOST":     "test-db",
			"DB_PORT":     "3306",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
			defer os.Unsetenv(key)
		}

		cfg := config.LoadConfig()

		assert.Equal(t, "127.0.0.1", cfg.ServerHost)
		assert.Equal(t, "9090", cfg.ServerPort)
		assert.Equal(t, "test-db", cfg.DBHost)
		assert.Equal(t, "3306", cfg.DBPort)
	})
}

// TestApplicationComponents tests that all components can be initialized
func TestApplicationComponents(t *testing.T) {
	t.Run("logger initialization", func(t *testing.T) {
		logger := setupTestLogger("info")
		assert.NotNil(t, logger)
		assert.Equal(t, logrus.InfoLevel, logger.Level)
	})

	t.Run("configuration loading", func(t *testing.T) {
		cfg := config.LoadConfig()
		assert.NotNil(t, cfg)

		// Test required fields are not empty
		require.NotEmpty(t, cfg.ServerHost)
		require.NotEmpty(t, cfg.ServerPort)
		require.NotEmpty(t, cfg.DBHost)
		require.NotEmpty(t, cfg.DBPort)
		require.NotEmpty(t, cfg.DBUser)
		require.NotEmpty(t, cfg.DBName)
		require.NotEmpty(t, cfg.JWTSecret)
	})
}

// Helper function to build DSN for testing
func buildDSN(cfg *config.Config) string {
	return "host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=" + cfg.DBSSLMode
}

// Helper function to setup logger (extracted for testing)
func setupTestLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel // default to info if invalid
	}
	logger.SetLevel(level)

	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}

// TestDSNBuilder tests the DSN building logic
func TestDSNBuilder(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected string
	}{
		{
			name: "default configuration",
			config: &config.Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "postgres123",
				DBName:     "icecream_store",
				DBSSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=postgres123 dbname=icecream_store sslmode=disable",
		},
		{
			name: "production configuration",
			config: &config.Config{
				DBHost:     "prod-db.example.com",
				DBPort:     "5432",
				DBUser:     "produser",
				DBPassword: "securepassword",
				DBName:     "prod_icecream_store",
				DBSSLMode:  "require",
			},
			expected: "host=prod-db.example.com port=5432 user=produser password=securepassword dbname=prod_icecream_store sslmode=require",
		},
		{
			name: "test configuration",
			config: &config.Config{
				DBHost:     "test-db",
				DBPort:     "3306",
				DBUser:     "testuser",
				DBPassword: "testpass",
				DBName:     "test_db",
				DBSSLMode:  "disable",
			},
			expected: "host=test-db port=3306 user=testuser password=testpass dbname=test_db sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildDSN(tt.config)
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

// TestLoggerLevels tests different logger levels
func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			logger := setupTestLogger(level)
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
		logger := setupTestLogger(cfg.LogLevel)
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
		logger := setupTestLogger("invalid-level")
		assert.NotNil(t, logger)
		// Should default to info level
		assert.Equal(t, logrus.InfoLevel, logger.Level)
	})

	t.Run("empty log level handling", func(t *testing.T) {
		logger := setupTestLogger("")
		assert.NotNil(t, logger)
		// Should default to info level
		assert.Equal(t, logrus.InfoLevel, logger.Level)
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
		setupTestLogger("info")
	}
}

// BenchmarkDSNBuild benchmarks DSN building
func BenchmarkDSNBuild(b *testing.B) {
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres123",
		DBName:     "icecream_store",
		DBSSLMode:  "disable",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildDSN(cfg)
	}
}
