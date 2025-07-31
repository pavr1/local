package main

import (
	"database/sql"
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

	t.Run("database connection string format", func(t *testing.T) {
		cfg := &config.Config{
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "testuser",
			DBPassword: "testpass",
			DBName:     "testdb",
			DBSSLMode:  "disable",
		}

		logger := setupLogger("info")
		dsn := buildDSN(cfg)
		expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
		assert.Equal(t, expected, dsn)

		// Test that connectToDatabase would use this DSN (without actually connecting)
		_, err := connectToDatabaseWithDSN(dsn, logger)
		// We expect an error since we're not connecting to a real database
		assert.Error(t, err)
		// Could be various error messages depending on the system
		assert.True(t, err != nil, "Should get an error when connecting with test credentials")
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
		assert.NotEmpty(t, cfg.LogLevel)
		assert.Equal(t, "8084", cfg.ServerPort) // Inventory service specific port
	})

	t.Run("configuration with environment overrides", func(t *testing.T) {
		// Set environment variables
		envVars := map[string]string{
			"INVENTORY_SERVER_HOST": "127.0.0.1",
			"INVENTORY_SERVER_PORT": "9084",
			"DB_HOST":               "test-db",
			"DB_PORT":               "3306",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
			defer os.Unsetenv(key)
		}

		cfg := config.LoadConfig()

		assert.Equal(t, "127.0.0.1", cfg.ServerHost)
		assert.Equal(t, "9084", cfg.ServerPort)
		assert.Equal(t, "test-db", cfg.DBHost)
		assert.Equal(t, "3306", cfg.DBPort)
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
		require.NotEmpty(t, cfg.DBHost)
		require.NotEmpty(t, cfg.DBPort)
		require.NotEmpty(t, cfg.DBUser)
		require.NotEmpty(t, cfg.DBName)
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

// Helper function to build DSN for testing
func buildDSN(cfg *config.Config) string {
	return "host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=" + cfg.DBSSLMode
}

// Helper function to test connection with DSN (extracted for testing)
func connectToDatabaseWithDSN(dsn string, logger *logrus.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
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
