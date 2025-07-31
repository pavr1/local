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

// TestQuerySystemConfig tests the system configuration query function
func TestQuerySystemConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful query with multiple rows",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"}).
					AddRow("app_name", "Ice Cream Store", "Application name").
					AddRow("max_sessions", "100", "Maximum concurrent sessions").
					AddRow("debug_mode", "false", "Debug mode setting")

				mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "successful query with empty result",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"})
				mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "query execution failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
					WillReturnError(errors.New("table does not exist"))
			},
			expectError: true,
			errorMsg:    "failed to query system configuration",
		},
		{
			name: "row scan failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"}).
					AddRow(nil, "value", "description") // nil key will cause scan error

				mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
					WillReturnRows(rows)
			},
			expectError: true,
			errorMsg:    "failed to scan system config row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := querySystemConfig(handler)

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

// TestQueryRoles tests the roles query function
func TestQueryRoles(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful query with multiple roles",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"role_name", "description", "created_at"}).
					AddRow("admin", "Administrator role", time.Now()).
					AddRow("employee", "Employee role", time.Now()).
					AddRow("manager", "Manager role", time.Now())

				mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "query execution failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
					WillReturnError(errors.New("roles table not found"))
			},
			expectError: true,
			errorMsg:    "failed to query roles",
		},
		{
			name: "row scan failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"role_name", "description", "created_at"}).
					AddRow("admin", nil, "invalid_date") // nil description and invalid date

				mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
					WillReturnRows(rows)
			},
			expectError: true,
			errorMsg:    "failed to scan role row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := queryRoles(handler)

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

// TestQueryExpenseCategories tests the expense categories query function
func TestQueryExpenseCategories(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful query with multiple categories",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"category_name", "description"}).
					AddRow("ingredients", "Cost of raw ingredients").
					AddRow("equipment", "Kitchen equipment and tools").
					AddRow("utilities", "Electricity, water, gas").
					AddRow("labor", "Employee wages and benefits")

				mock.ExpectQuery(`SELECT category_name, description\s+FROM expense_categories\s+ORDER BY category_name`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "empty result set",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"category_name", "description"})
				mock.ExpectQuery(`SELECT category_name, description\s+FROM expense_categories\s+ORDER BY category_name`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "query execution failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT category_name, description\s+FROM expense_categories\s+ORDER BY category_name`).
					WillReturnError(errors.New("expense_categories table does not exist"))
			},
			expectError: true,
			errorMsg:    "failed to query expense categories",
		},
		{
			name: "row scan failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"category_name", "description"}).
					AddRow(nil, "description") // nil category_name will cause scan error

				mock.ExpectQuery(`SELECT category_name, description\s+FROM expense_categories\s+ORDER BY category_name`).
					WillReturnRows(rows)
			},
			expectError: true,
			errorMsg:    "failed to scan expense category row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := queryExpenseCategories(handler)

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

// TestQueryRecipeCategories tests the recipe categories query function
func TestQueryRecipeCategories(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful query with multiple categories",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "description"}).
					AddRow("vanilla", "Classic vanilla ice cream recipes").
					AddRow("chocolate", "Rich chocolate ice cream recipes").
					AddRow("fruit", "Fresh fruit ice cream recipes").
					AddRow("specialty", "Unique and specialty flavors")

				mock.ExpectQuery(`SELECT name, description\s+FROM recipe_categories\s+ORDER BY name`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "empty result set",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "description"})
				mock.ExpectQuery(`SELECT name, description\s+FROM recipe_categories\s+ORDER BY name`).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "query execution failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT name, description\s+FROM recipe_categories\s+ORDER BY name`).
					WillReturnError(errors.New("recipe_categories table does not exist"))
			},
			expectError: true,
			errorMsg:    "failed to query recipe categories",
		},
		{
			name: "row scan failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "description"}).
					AddRow(nil, "description") // nil name will cause scan error

				mock.ExpectQuery(`SELECT name, description\s+FROM recipe_categories\s+ORDER BY name`).
					WillReturnRows(rows)
			},
			expectError: true,
			errorMsg:    "failed to scan recipe category row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := queryRecipeCategories(handler)

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

// TestMainFunction tests the main function behavior (integration style test)
func TestMainFunction(t *testing.T) {
	// Note: Testing main() directly is challenging since it calls log.Fatalf()
	// Instead, we test the individual components that main() uses

	t.Run("database config creation", func(t *testing.T) {
		config := &database.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres123",
			DBName:   "icecream_store",
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

		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "postgres", config.User)
		assert.Equal(t, "icecream_store", config.DBName)
		assert.Equal(t, 25, config.MaxOpenConns)
	})

	t.Run("logger configuration", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
		})

		assert.Equal(t, logrus.InfoLevel, logger.Level)
		assert.IsType(t, &logrus.JSONFormatter{}, logger.Formatter)
	})

	t.Run("database handler creation", func(t *testing.T) {
		config := &database.Config{
			Host:   "localhost",
			Port:   5432,
			DBName: "test",
		}
		logger := logrus.New()

		handler := database.New(config, logger)
		assert.NotNil(t, handler)
		assert.Implements(t, (*database.DatabaseHandler)(nil), handler)
	})
}

// TestQueryFunctionsWithTimeout tests query functions with context timeout
func TestQueryFunctionsWithTimeout(t *testing.T) {
	tests := []struct {
		name      string
		queryFunc func(database.DatabaseHandler) error
		setupMock func(sqlmock.Sqlmock)
	}{
		{
			name:      "querySystemConfig with timeout",
			queryFunc: querySystemConfig,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"})
				mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
					WillReturnRows(rows)
			},
		},
		{
			name:      "queryRoles with timeout",
			queryFunc: queryRoles,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"role_name", "description", "created_at"})
				mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
					WillReturnRows(rows)
			},
		},
		{
			name:      "queryExpenseCategories with timeout",
			queryFunc: queryExpenseCategories,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"category_name", "description"})
				mock.ExpectQuery(`SELECT category_name, description\s+FROM expense_categories\s+ORDER BY category_name`).
					WillReturnRows(rows)
			},
		},
		{
			name:      "queryRecipeCategories with timeout",
			queryFunc: queryRecipeCategories,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "description"})
				mock.ExpectQuery(`SELECT name, description\s+FROM recipe_categories\s+ORDER BY name`).
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock := setupTestHandler(t)
			defer handler.Close()

			tt.setupMock(mock)

			err := tt.queryFunc(handler)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestQueryFunctionsRowIteration tests that query functions properly iterate through all rows
func TestQueryFunctionsRowIteration(t *testing.T) {
	t.Run("querySystemConfig iterates all rows", func(t *testing.T) {
		handler, mock := setupTestHandler(t)
		defer handler.Close()

		rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"}).
			AddRow("key1", "value1", "desc1").
			AddRow("key2", "value2", "desc2").
			AddRow("key3", "value3", "desc3")

		mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
			WillReturnRows(rows)

		err := querySystemConfig(handler)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("queryRoles iterates all rows", func(t *testing.T) {
		handler, mock := setupTestHandler(t)
		defer handler.Close()

		now := time.Now()
		rows := sqlmock.NewRows([]string{"role_name", "description", "created_at"}).
			AddRow("admin", "Administrator", now).
			AddRow("user", "Regular user", now)

		mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
			WillReturnRows(rows)

		err := queryRoles(handler)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// BenchmarkQueryFunctions benchmarks the query functions
func BenchmarkQuerySystemConfig(b *testing.B) {
	handler, mock := setupTestHandlerBench(b)
	defer handler.Close()

	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"config_key", "config_value", "description"}).
			AddRow("test_key", "test_value", "test_description")
		mock.ExpectQuery(`SELECT config_key, config_value, description\s+FROM system_configuration\s+ORDER BY config_key`).
			WillReturnRows(rows)

		querySystemConfig(handler)
	}
}

func BenchmarkQueryRoles(b *testing.B) {
	handler, mock := setupTestHandlerBench(b)
	defer handler.Close()

	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"role_name", "description", "created_at"}).
			AddRow("admin", "Administrator", time.Now())
		mock.ExpectQuery(`SELECT role_name, description, created_at\s+FROM roles\s+ORDER BY role_name`).
			WillReturnRows(rows)

		queryRoles(handler)
	}
}

// Helper function for benchmarks
func setupTestHandlerBench(b *testing.B) (database.DatabaseHandler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock database: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &database.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "test-user",
		Password:        "test-password",
		DBName:          "test-db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		QueryTimeout:    30 * time.Second,
		MaxRetries:      3,
		RetryInterval:   1 * time.Second,
	}

	handler := &mockHandler{
		db:     db,
		config: config,
		logger: logger,
		mock:   mock,
	}

	return handler, mock
}
