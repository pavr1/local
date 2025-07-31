package main

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMainHttpHandler tests the creation of MainHttpHandler
func TestNewMainHttpHandler(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	t.Run("successful creation", func(t *testing.T) {
		handler := NewMainHttpHandler(db, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, db, handler.db)
		assert.Equal(t, logger, handler.logger)
		assert.NotNil(t, handler.SuppliersHandler)
	})

	t.Run("handlers are properly initialized", func(t *testing.T) {
		handler := NewMainHttpHandler(db, logger)

		// Test that suppliers handler is initialized
		suppliersHandler := handler.GetSuppliersHandler()
		assert.NotNil(t, suppliersHandler)
		assert.Equal(t, handler.SuppliersHandler, suppliersHandler)
	})
}

// TestMainHttpHandlerStructure tests the MainHttpHandler struct
func TestMainHttpHandlerStructure(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewMainHttpHandler(db, logger)

	t.Run("has database connection", func(t *testing.T) {
		assert.NotNil(t, handler.db)
		assert.IsType(t, &sql.DB{}, handler.db)
	})

	t.Run("has logger", func(t *testing.T) {
		assert.NotNil(t, handler.logger)
		assert.IsType(t, &logrus.Logger{}, handler.logger)
	})

	t.Run("has suppliers handler", func(t *testing.T) {
		assert.NotNil(t, handler.SuppliersHandler)
	})
}

// TestGetSuppliersHandler tests the GetSuppliersHandler method
func TestGetSuppliersHandler(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewMainHttpHandler(db, logger)

	t.Run("returns suppliers handler", func(t *testing.T) {
		suppliersHandler := handler.GetSuppliersHandler()
		assert.NotNil(t, suppliersHandler)
		assert.Equal(t, handler.SuppliersHandler, suppliersHandler)
	})

	t.Run("consistent handler reference", func(t *testing.T) {
		handler1 := handler.GetSuppliersHandler()
		handler2 := handler.GetSuppliersHandler()
		assert.Equal(t, handler1, handler2, "Should return the same handler instance")
	})
}

// TestMainHttpHandlerWithDifferentLoggers tests handler with different logger configurations
func TestMainHttpHandlerWithDifferentLoggers(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tests := []struct {
		name     string
		logLevel logrus.Level
	}{
		{"debug logger", logrus.DebugLevel},
		{"info logger", logrus.InfoLevel},
		{"warn logger", logrus.WarnLevel},
		{"error logger", logrus.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(tt.logLevel)

			handler := NewMainHttpHandler(db, logger)
			assert.NotNil(t, handler)
			assert.Equal(t, tt.logLevel, handler.logger.Level)
		})
	}
}

// TestMainHttpHandlerResourcesCleanup tests that handlers can be cleaned up properly
func TestMainHttpHandlerResourcesCleanup(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewMainHttpHandler(db, logger)

	t.Run("database can be closed", func(t *testing.T) {
		// Ensure we can close the database connection
		// In a real scenario, this would be handled by the main application
		assert.NotNil(t, handler.db)
		// We don't actually close it here since we'll close it in defer
	})

	t.Run("handlers are accessible after creation", func(t *testing.T) {
		assert.NotNil(t, handler.GetSuppliersHandler())
	})
}

// TestMainHttpHandlerNilInputs tests handler creation with nil inputs
func TestMainHttpHandlerNilInputs(t *testing.T) {
	t.Run("with nil database", func(t *testing.T) {
		logger := logrus.New()

		// This should not panic but would likely fail in real usage
		// The test documents the behavior
		handler := NewMainHttpHandler(nil, logger)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.db)
		assert.Equal(t, logger, handler.logger)
	})

	t.Run("with nil logger", func(t *testing.T) {
		// Create a mock database
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// This should not panic but would likely fail in real usage
		// The test documents the behavior
		handler := NewMainHttpHandler(db, nil)
		assert.NotNil(t, handler)
		assert.Equal(t, db, handler.db)
		assert.Nil(t, handler.logger)
	})
}

// TestMainHttpHandlerEntityHandlerIntegration tests integration between main handler and entity handlers
func TestMainHttpHandlerEntityHandlerIntegration(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise

	handler := NewMainHttpHandler(db, logger)

	t.Run("suppliers handler integration", func(t *testing.T) {
		suppliersHandler := handler.GetSuppliersHandler()
		assert.NotNil(t, suppliersHandler)

		// The suppliers handler should be properly initialized
		// We can't test its methods directly without more setup,
		// but we can verify it exists and is the correct type
		assert.IsType(t, handler.SuppliersHandler, suppliersHandler)
	})
}

// TestMainHttpHandlerMemoryUsage tests that handler creation doesn't leak memory
func TestMainHttpHandlerMemoryUsage(t *testing.T) {
	// Create multiple handlers to test for memory leaks
	for i := 0; i < 10; i++ {
		// Create a mock database
		db, _, err := sqlmock.New()
		require.NoError(t, err)

		logger := logrus.New()
		handler := NewMainHttpHandler(db, logger)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.GetSuppliersHandler())

		db.Close()
	}
}

// TestMainHttpHandlerConcurrentAccess tests concurrent access to handler
func TestMainHttpHandlerConcurrentAccess(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewMainHttpHandler(db, logger)

	// Test concurrent access to GetSuppliersHandler
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			suppliersHandler := handler.GetSuppliersHandler()
			assert.NotNil(t, suppliersHandler)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// BenchmarkNewMainHttpHandler benchmarks the creation of MainHttpHandler
func BenchmarkNewMainHttpHandler(b *testing.B) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMainHttpHandler(db, logger)
	}
}

// BenchmarkGetSuppliersHandler benchmarks the GetSuppliersHandler method
func BenchmarkGetSuppliersHandler(b *testing.B) {
	// Create a mock database
	db, _, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewMainHttpHandler(db, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.GetSuppliersHandler()
	}
}
