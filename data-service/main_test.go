package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"data-service/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test setup helpers
func setupTestHandler(t *testing.T) (database.DatabaseHandler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := &database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "test-user",
		Password: "test-password",
		DBName:   "test-db",
		SSLMode:  "disable",

		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}

	// Create a mock handler that wraps the mock database
	handler := &mockHandler{
		db:     db,
		config: config,
		logger: logger,
		mock:   mock,
	}

	return handler, mock
}

// mockHandler implements database.DatabaseHandler for testing
type mockHandler struct {
	db     *sql.DB
	config *database.Config
	logger *logrus.Logger
	mock   sqlmock.Sqlmock
}

func (m *mockHandler) Connect() error                               { return nil }
func (m *mockHandler) Close() error                                 { return m.db.Close() }
func (m *mockHandler) Ping() error                                  { return nil }
func (m *mockHandler) HealthCheck() error                           { return nil }
func (m *mockHandler) BeginTx(ctx context.Context) (*sql.Tx, error) { return m.db.BeginTx(ctx, nil) }
func (m *mockHandler) CommitTx(tx *sql.Tx) error                    { return tx.Commit() }
func (m *mockHandler) RollbackTx(tx *sql.Tx) error                  { return tx.Rollback() }
func (m *mockHandler) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}
func (m *mockHandler) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}
func (m *mockHandler) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}
func (m *mockHandler) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}
func (m *mockHandler) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}
func (m *mockHandler) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}
func (m *mockHandler) Prepare(query string) (*sql.Stmt, error) { return m.db.Prepare(query) }
func (m *mockHandler) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return m.db.PrepareContext(ctx, query)
}
func (m *mockHandler) GetDB() *sql.DB        { return m.db }
func (m *mockHandler) GetStats() sql.DBStats { return m.db.Stats() }
func (m *mockHandler) IsConnected() bool     { return true }

// TestHealthCheck tests the health check functionality
func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful health check",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			expectError: false,
		},
		{
			name: "health check failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("connection failed"))
			},
			expectError: true,
			errorMsg:    "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := handler.HealthCheck()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDatabaseConnection tests the database connection functionality
func TestDatabaseConnection(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer handler.Close()

	// Test successful connection
	mock.ExpectPing()
	err := handler.Connect()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test connection stats
	stats := handler.GetStats()
	assert.NotNil(t, stats)
}

// TestDatabaseOperations tests basic database operations
func TestDatabaseOperations(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer handler.Close()

	// Test query
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(1))
	rows, err := handler.Query("SELECT 1")
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	rows.Close()

	// Test exec
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	result, err := handler.Exec("CREATE TABLE test")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}
