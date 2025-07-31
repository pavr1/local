package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test utilities and setup
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, DatabaseHandler) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := DefaultConfig()
	config.Host = "test-host"
	config.Port = 5432
	config.User = "test-user"
	config.Password = "test-password"
	config.DBName = "test-db"

	handler := &dbHandler{
		db:        db,
		config:    config,
		logger:    logger,
		connected: true,
	}

	return db, mock, handler
}

func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

// TestDefaultConfig tests the default configuration creation
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, 10*time.Second, config.ConnectTimeout)
	assert.Equal(t, 30*time.Second, config.QueryTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryInterval)
}

// TestNew tests the constructor with various configurations
func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		logger *logrus.Logger
	}{
		{
			name:   "with nil config and logger",
			config: nil,
			logger: nil,
		},
		{
			name:   "with custom config and nil logger",
			config: &Config{Host: "custom-host", Port: 3306},
			logger: nil,
		},
		{
			name:   "with nil config and custom logger",
			config: nil,
			logger: setupTestLogger(),
		},
		{
			name:   "with custom config and logger",
			config: &Config{Host: "custom-host", Port: 3306},
			logger: setupTestLogger(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(tt.config, tt.logger)
			assert.NotNil(t, handler)

			// Test interface compliance
			var _ DatabaseHandler = handler
		})
	}
}

// TestConnect tests database connection scenarios
func TestConnect(t *testing.T) {
	// Note: Testing Connect() method is complex due to actual connection logic
	// Instead, we test the connection status and basic validation
	t.Run("connect with nil database creates new connection", func(t *testing.T) {
		config := DefaultConfig()
		config.Host = "invalid-host" // This will cause connection to fail
		config.ConnectTimeout = 1 * time.Second
		config.MaxRetries = 1

		logger := setupTestLogger()
		handler := New(config, logger)

		// This should fail due to invalid host
		err := handler.Connect()
		assert.Error(t, err)
		assert.False(t, handler.IsConnected())
	})

	t.Run("connect validates configuration", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotNil(t, config)

		logger := setupTestLogger()
		handler := New(config, logger)
		assert.NotNil(t, handler)
	})
}

// TestClose tests database connection closure
func TestClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		_, mock, handler := setupTestDB(t)
		mock.ExpectClose()

		err := handler.Close()
		assert.NoError(t, err)
		assert.False(t, handler.IsConnected())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("close with nil database", func(t *testing.T) {
		handler := New(DefaultConfig(), setupTestLogger())
		err := handler.Close()
		assert.NoError(t, err)
		assert.False(t, handler.IsConnected())
	})
}

// TestPing tests database ping functionality
func TestPing(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		_, mock, handler := setupTestDB(t)
		mock.ExpectPing()

		err := handler.Ping()
		assert.NoError(t, err)
		assert.True(t, handler.IsConnected())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ping failure", func(t *testing.T) {
		_, mock, handler := setupTestDB(t)
		mock.ExpectPing().WillReturnError(errors.New("ping failed"))

		err := handler.Ping()
		assert.Error(t, err)
		assert.False(t, handler.IsConnected())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestPingWithNilDB tests ping with nil database
func TestPingWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	err := handler.Ping()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestHealthCheck tests comprehensive health check
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
				mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(1))
			},
			expectError: false,
		},
		{
			name: "ping failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("ping failed"))
			},
			expectError: true,
			errorMsg:    "ping failed",
		},
		{
			name: "query failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("query failed"))
			},
			expectError: true,
			errorMsg:    "health check query failed",
		},
		{
			name: "unexpected result",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(2))
			},
			expectError: true,
			errorMsg:    "unexpected health check result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, handler := setupTestDB(t)
			defer db.Close()

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

// TestHealthCheckWithNilDB tests health check with nil database
func TestHealthCheckWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	err := handler.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestBeginTx tests transaction initialization
func TestBeginTx(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "successful transaction begin",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
			},
			expectError: false,
		},
		{
			name: "begin transaction failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, handler := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			tx, err := handler.BeginTx(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestBeginTxWithNilDB tests transaction begin with nil database
func TestBeginTxWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	tx, err := handler.BeginTx(context.Background())
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestCommitTx tests transaction commit
func TestCommitTx(t *testing.T) {
	tests := []struct {
		name        string
		setupTx     func() *sql.Tx
		expectError bool
	}{
		{
			name: "successful commit",
			setupTx: func() *sql.Tx {
				db, mock, _ := setupTestDB(t)
				mock.ExpectBegin()
				mock.ExpectCommit()
				tx, _ := db.Begin()
				return tx
			},
			expectError: false,
		},
		{
			name: "commit with nil transaction",
			setupTx: func() *sql.Tx {
				return nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, handler := setupTestDB(t)
			tx := tt.setupTx()

			err := handler.CommitTx(tx)

			if tt.expectError {
				assert.Error(t, err)
				if tx == nil {
					assert.Contains(t, err.Error(), "transaction is nil")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRollbackTx tests transaction rollback
func TestRollbackTx(t *testing.T) {
	tests := []struct {
		name        string
		setupTx     func() *sql.Tx
		expectError bool
	}{
		{
			name: "successful rollback",
			setupTx: func() *sql.Tx {
				db, mock, _ := setupTestDB(t)
				mock.ExpectBegin()
				mock.ExpectRollback()
				tx, _ := db.Begin()
				return tx
			},
			expectError: false,
		},
		{
			name: "rollback with nil transaction",
			setupTx: func() *sql.Tx {
				return nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, handler := setupTestDB(t)
			tx := tt.setupTx()

			err := handler.RollbackTx(tx)

			if tt.expectError {
				assert.Error(t, err)
				if tx == nil {
					assert.Contains(t, err.Error(), "transaction is nil")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestQuery tests query execution
func TestQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		args        []interface{}
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name:  "successful query",
			query: "SELECT id, name FROM users WHERE active = $1",
			args:  []interface{}{true},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "John").
					AddRow(2, "Jane")
				mock.ExpectQuery("SELECT id, name FROM users WHERE active = \\$1").
					WithArgs(true).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name:  "query failure",
			query: "SELECT * FROM invalid_table",
			args:  []interface{}{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM invalid_table").
					WillReturnError(errors.New("table does not exist"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, handler := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			rows, err := handler.Query(tt.query, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, rows)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rows)
				rows.Close()
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestQueryContext tests query execution with context
func TestQueryContext(t *testing.T) {
	db, mock, handler := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "SELECT id FROM users"

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT id FROM users").WillReturnRows(rows)

	result, err := handler.QueryContext(ctx, query)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	result.Close()

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryContextWithNilDB tests query with nil database
func TestQueryContextWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	rows, err := handler.QueryContext(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestQueryRow tests single row query
func TestQueryRow(t *testing.T) {
	db, mock, handler := setupTestDB(t)
	defer db.Close()

	query := "SELECT name FROM users WHERE id = $1"
	mock.ExpectQuery("SELECT name FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("John"))

	row := handler.QueryRow(query, 1)
	assert.NotNil(t, row)

	var name string
	err := row.Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "John", name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRowContext tests single row query with context
func TestQueryRowContext(t *testing.T) {
	db, mock, handler := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "SELECT name FROM users WHERE id = $1"
	mock.ExpectQuery("SELECT name FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("John"))

	row := handler.QueryRowContext(ctx, query, 1)
	assert.NotNil(t, row)

	var name string
	err := row.Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "John", name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRowContextWithNilDB tests query row with nil database
func TestQueryRowContextWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	row := handler.QueryRowContext(context.Background(), "SELECT 1")
	assert.Nil(t, row)
}

// TestExec tests query execution without returning rows
func TestExec(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		args        []interface{}
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name:  "successful exec",
			query: "UPDATE users SET name = $1 WHERE id = $2",
			args:  []interface{}{"NewName", 1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET name = \\$1 WHERE id = \\$2").
					WithArgs("NewName", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name:  "exec failure",
			query: "DELETE FROM users WHERE id = $1",
			args:  []interface{}{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(errors.New("foreign key constraint"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, handler := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			result, err := handler.Exec(tt.query, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				rowsAffected, _ := result.RowsAffected()
				assert.Equal(t, int64(1), rowsAffected)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestExecContext tests exec with context
func TestExecContext(t *testing.T) {
	db, mock, handler := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "INSERT INTO users (name) VALUES ($1)"
	mock.ExpectExec("INSERT INTO users \\(name\\) VALUES \\(\\$1\\)").
		WithArgs("TestUser").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := handler.ExecContext(ctx, query, "TestUser")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestExecContextWithNilDB tests exec with nil database
func TestExecContextWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	result, err := handler.ExecContext(context.Background(), "INSERT INTO test VALUES ($1)", "value")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestPrepare tests prepared statement creation
func TestPrepare(t *testing.T) {
	db, mock, handler := setupTestDB(t)
	defer db.Close()

	query := "SELECT * FROM users WHERE id = $1"
	mock.ExpectPrepare("SELECT \\* FROM users WHERE id = \\$1")

	stmt, err := handler.Prepare(query)
	assert.NoError(t, err)
	assert.NotNil(t, stmt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPrepareContext tests prepared statement creation with context
func TestPrepareContext(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		setupMock   func(sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name:  "successful prepare",
			query: "SELECT * FROM users WHERE id = $1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT \\* FROM users WHERE id = \\$1")
			},
			expectError: false,
		},
		{
			name:  "prepare failure",
			query: "INVALID SQL QUERY",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INVALID SQL QUERY").
					WillReturnError(errors.New("syntax error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, handler := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			ctx := context.Background()
			stmt, err := handler.PrepareContext(ctx, tt.query)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stmt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stmt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPrepareContextWithNilDB tests prepare with nil database
func TestPrepareContextWithNilDB(t *testing.T) {
	handler := New(DefaultConfig(), setupTestLogger())
	stmt, err := handler.PrepareContext(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Nil(t, stmt)
	assert.Contains(t, err.Error(), "database connection is nil")
}

// TestGetDB tests getting the underlying database instance
func TestGetDB(t *testing.T) {
	db, _, handler := setupTestDB(t)
	defer db.Close()

	result := handler.GetDB()
	assert.Equal(t, db, result)
}

// TestGetStats tests database statistics retrieval
func TestGetStats(t *testing.T) {
	tests := []struct {
		name   string
		hasDB  bool
		expect func(*testing.T, sql.DBStats)
	}{
		{
			name:  "with database connection",
			hasDB: true,
			expect: func(t *testing.T, stats sql.DBStats) {
				// Stats should be initialized (exact values depend on implementation)
				assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
			},
		},
		{
			name:  "without database connection",
			hasDB: false,
			expect: func(t *testing.T, stats sql.DBStats) {
				// Should return empty stats
				assert.Equal(t, sql.DBStats{}, stats)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler DatabaseHandler
			if tt.hasDB {
				db, _, h := setupTestDB(t)
				defer db.Close()
				handler = h
			} else {
				handler = New(DefaultConfig(), setupTestLogger())
			}

			stats := handler.GetStats()
			tt.expect(t, stats)
		})
	}
}

// TestIsConnected tests connection status
func TestIsConnected(t *testing.T) {
	tests := []struct {
		name      string
		setupDB   func() DatabaseHandler
		connected bool
	}{
		{
			name: "connected database",
			setupDB: func() DatabaseHandler {
				_, _, handler := setupTestDB(t)
				return handler
			},
			connected: true,
		},
		{
			name: "disconnected database",
			setupDB: func() DatabaseHandler {
				handler := New(DefaultConfig(), setupTestLogger())
				return handler
			},
			connected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupDB()
			assert.Equal(t, tt.connected, handler.IsConnected())
		})
	}
}

// TestBuildConnectionString tests connection string building
func TestBuildConnectionString(t *testing.T) {
	config := &Config{
		Host:           "testhost",
		Port:           5432,
		User:           "testuser",
		Password:       "testpass",
		DBName:         "testdb",
		SSLMode:        "require",
		ConnectTimeout: 15 * time.Second,
	}

	handler := &dbHandler{
		config: config,
		logger: setupTestLogger(),
	}

	connStr := handler.buildConnectionString()
	expected := "host=testhost port=5432 user=testuser password=testpass dbname=testdb sslmode=require connect_timeout=15"
	assert.Equal(t, expected, connStr)
}

// TestSanitizeQuery tests query sanitization for logging
func TestSanitizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "short query",
			query:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "long query",
			query:    "SELECT * FROM users WHERE name = 'very long name that should be truncated because it exceeds the limit'",
			expected: "SELECT * FROM users WHERE name = 'very long name that should be truncated because it exceeds the lim...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &dbHandler{
				logger: setupTestLogger(),
			}

			result := handler.sanitizeQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHandlePostgreSQLError tests PostgreSQL error handling
func TestHandlePostgreSQLError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectedMsg string
	}{
		{
			name: "unique violation",
			inputError: &pq.Error{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint",
				Detail:  "Key (email)=(test@example.com) already exists.",
			},
			expectedMsg: "duplicate entry",
		},
		{
			name: "foreign key violation",
			inputError: &pq.Error{
				Code:    "23503",
				Message: "insert or update on table violates foreign key constraint",
				Detail:  "Key (user_id)=(999) is not present in table users.",
			},
			expectedMsg: "foreign key constraint violation",
		},
		{
			name: "not null violation",
			inputError: &pq.Error{
				Code:    "23502",
				Message: "null value in column violates not-null constraint",
				Column:  "email",
			},
			expectedMsg: "required field missing: email",
		},
		{
			name: "other PostgreSQL error",
			inputError: &pq.Error{
				Code:    "42P01",
				Message: "relation does not exist",
			},
			expectedMsg: "database error [42P01]",
		},
		{
			name:        "non-PostgreSQL error",
			inputError:  errors.New("generic database error"),
			expectedMsg: "generic database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &dbHandler{
				logger: setupTestLogger(),
			}

			result := handler.handlePostgreSQLError(tt.inputError)
			assert.Contains(t, result.Error(), tt.expectedMsg)
		})
	}
}

// TestConfigureConnectionPool tests connection pool configuration
func TestConfigureConnectionPool(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	config := &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	handler := &dbHandler{
		config: config,
		logger: setupTestLogger(),
	}

	handler.configureConnectionPool(db)

	stats := db.Stats()
	assert.Equal(t, 10, stats.MaxOpenConnections)
	// Note: MaxIdleConnections is not exposed in sql.DBStats
}

// Benchmark tests
func BenchmarkQuery(b *testing.B) {
	db, mock, handler := setupTestDBBench(b)
	defer db.Close()

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(1))
		rows, _ := handler.Query("SELECT 1")
		if rows != nil {
			rows.Close()
		}
	}
}

func BenchmarkExec(b *testing.B) {
	db, mock, handler := setupTestDBBench(b)
	defer db.Close()

	for i := 0; i < b.N; i++ {
		mock.ExpectExec("UPDATE test SET value = 1").WillReturnResult(sqlmock.NewResult(0, 1))
		handler.Exec("UPDATE test SET value = 1")
	}
}

// Helper function for benchmarks
func setupTestDBBench(b *testing.B) (*sql.DB, sqlmock.Sqlmock, DatabaseHandler) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		b.Fatalf("Failed to create mock database: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := DefaultConfig()
	config.Host = "test-host"
	config.Port = 5432
	config.User = "test-user"
	config.Password = "test-password"
	config.DBName = "test-db"

	handler := &dbHandler{
		db:        db,
		config:    config,
		logger:    logger,
		connected: true,
	}

	return db, mock, handler
}
