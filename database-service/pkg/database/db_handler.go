package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

// DatabaseHandler defines the interface for database operations
type DatabaseHandler interface {
	// Connection management
	Connect() error
	Close() error
	Ping() error
	HealthCheck() error

	// Transaction management
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error

	// Query operations
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Execute operations
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Prepared statements
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	// Utility methods
	GetDB() *sql.DB
	GetStats() sql.DBStats
	IsConnected() bool
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Timeout settings
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration

	// Retry settings
	MaxRetries    int
	RetryInterval time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    5432,
		SSLMode: "disable",

		// Connection pool defaults
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		// Timeout defaults
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		// Retry defaults
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}
}

// dbHandler implements the DatabaseHandler interface
type dbHandler struct {
	db        *sql.DB
	config    *Config
	logger    *logrus.Logger
	connected bool
}

// New creates a new database handler instance
func New(config *Config, logger *logrus.Logger) DatabaseHandler {
	if config == nil {
		config = DefaultConfig()
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	return &dbHandler{
		config:    config,
		logger:    logger,
		connected: false,
	}
}

// Connect establishes a connection to the database
func (h *dbHandler) Connect() error {
	h.logger.WithFields(logrus.Fields{
		"host":   h.config.Host,
		"port":   h.config.Port,
		"dbname": h.config.DBName,
		"user":   h.config.User,
	}).Info("Attempting to connect to database")

	// Build connection string
	connStr := h.buildConnectionString()

	var err error
	var db *sql.DB

	// Retry connection with exponential backoff
	for attempt := 1; attempt <= h.config.MaxRetries; attempt++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"error":   err.Error(),
			}).Warn("Failed to open database connection")

			if attempt < h.config.MaxRetries {
				time.Sleep(h.config.RetryInterval * time.Duration(attempt))
				continue
			}
			return fmt.Errorf("failed to open database after %d attempts: %w", h.config.MaxRetries, err)
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), h.config.ConnectTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"error":   err.Error(),
			}).Warn("Failed to ping database")

			db.Close()
			if attempt < h.config.MaxRetries {
				time.Sleep(h.config.RetryInterval * time.Duration(attempt))
				continue
			}
			return fmt.Errorf("failed to ping database after %d attempts: %w", h.config.MaxRetries, err)
		}

		break
	}

	// Configure connection pool
	h.configureConnectionPool(db)

	h.db = db
	h.connected = true

	h.logger.WithFields(logrus.Fields{
		"host":   h.config.Host,
		"port":   h.config.Port,
		"dbname": h.config.DBName,
	}).Info("Successfully connected to database")

	return nil
}

// Close closes the database connection
func (h *dbHandler) Close() error {
	if h.db == nil {
		return nil
	}

	h.logger.Info("Closing database connection")

	err := h.db.Close()
	if err != nil {
		h.logger.WithError(err).Error("Failed to close database connection")
		return err
	}

	h.connected = false
	h.logger.Info("Database connection closed successfully")
	return nil
}

// Ping tests the database connection
func (h *dbHandler) Ping() error {
	if h.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.ConnectTimeout)
	defer cancel()

	err := h.db.PingContext(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Database ping failed")
		h.connected = false
		return err
	}

	h.connected = true
	return nil
}

// HealthCheck performs a comprehensive health check
func (h *dbHandler) HealthCheck() error {
	if h.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	h.logger.Debug("Performing database health check")

	// Test basic connectivity
	if err := h.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test with a simple query
	ctx, cancel := context.WithTimeout(context.Background(), h.config.QueryTimeout)
	defer cancel()

	var result int
	err := h.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		h.logger.WithError(err).Error("Health check query failed")
		return fmt.Errorf("health check query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	h.logger.Debug("Database health check passed")
	return nil
}

// BeginTx starts a new transaction
func (h *dbHandler) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction")
		return nil, err
	}

	h.logger.Debug("Transaction started")
	return tx, nil
}

// CommitTx commits a transaction
func (h *dbHandler) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	err := tx.Commit()
	if err != nil {
		h.logger.WithError(err).Error("Failed to commit transaction")
		return err
	}

	h.logger.Debug("Transaction committed")
	return nil
}

// RollbackTx rolls back a transaction
func (h *dbHandler) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	err := tx.Rollback()
	if err != nil {
		h.logger.WithError(err).Error("Failed to rollback transaction")
		return err
	}

	h.logger.Debug("Transaction rolled back")
	return nil
}

// Query executes a query with logging
func (h *dbHandler) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return h.QueryContext(context.Background(), query, args...)
}

// QueryContext executes a query with context and logging
func (h *dbHandler) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	start := time.Now()
	rows, err := h.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	logEntry := h.logger.WithFields(logrus.Fields{
		"query":      h.sanitizeQuery(query),
		"duration":   duration,
		"args_count": len(args),
	})

	if err != nil {
		logEntry.WithError(err).Error("Query execution failed")
		return nil, h.handlePostgreSQLError(err)
	}

	logEntry.Debug("Query executed successfully")
	return rows, nil
}

// QueryRow executes a query that returns a single row
func (h *dbHandler) QueryRow(query string, args ...interface{}) *sql.Row {
	return h.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext executes a query that returns a single row with context
func (h *dbHandler) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if h.db == nil {
		h.logger.Error("Database connection is nil for QueryRow")
		return nil
	}

	start := time.Now()
	row := h.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	h.logger.WithFields(logrus.Fields{
		"query":      h.sanitizeQuery(query),
		"duration":   duration,
		"args_count": len(args),
	}).Debug("QueryRow executed")

	return row
}

// Exec executes a query without returning rows
func (h *dbHandler) Exec(query string, args ...interface{}) (sql.Result, error) {
	return h.ExecContext(context.Background(), query, args...)
}

// ExecContext executes a query without returning rows with context
func (h *dbHandler) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	start := time.Now()
	result, err := h.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	logEntry := h.logger.WithFields(logrus.Fields{
		"query":      h.sanitizeQuery(query),
		"duration":   duration,
		"args_count": len(args),
	})

	if err != nil {
		logEntry.WithError(err).Error("Exec execution failed")
		return nil, h.handlePostgreSQLError(err)
	}

	logEntry.Debug("Exec executed successfully")
	return result, nil
}

// Prepare creates a prepared statement
func (h *dbHandler) Prepare(query string) (*sql.Stmt, error) {
	return h.PrepareContext(context.Background(), query)
}

// PrepareContext creates a prepared statement with context
func (h *dbHandler) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	stmt, err := h.db.PrepareContext(ctx, query)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"query": h.sanitizeQuery(query),
		}).WithError(err).Error("Failed to prepare statement")
		return nil, h.handlePostgreSQLError(err)
	}

	h.logger.WithFields(logrus.Fields{
		"query": h.sanitizeQuery(query),
	}).Debug("Statement prepared successfully")

	return stmt, nil
}

// GetDB returns the underlying sql.DB instance
func (h *dbHandler) GetDB() *sql.DB {
	return h.db
}

// GetStats returns database connection statistics
func (h *dbHandler) GetStats() sql.DBStats {
	if h.db == nil {
		return sql.DBStats{}
	}
	return h.db.Stats()
}

// IsConnected returns the connection status
func (h *dbHandler) IsConnected() bool {
	return h.connected && h.db != nil
}

// buildConnectionString creates the PostgreSQL connection string
func (h *dbHandler) buildConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		h.config.Host,
		h.config.Port,
		h.config.User,
		h.config.Password,
		h.config.DBName,
		h.config.SSLMode,
		int(h.config.ConnectTimeout.Seconds()),
	)
}

// configureConnectionPool sets up the connection pool
func (h *dbHandler) configureConnectionPool(db *sql.DB) {
	db.SetMaxOpenConns(h.config.MaxOpenConns)
	db.SetMaxIdleConns(h.config.MaxIdleConns)
	db.SetConnMaxLifetime(h.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(h.config.ConnMaxIdleTime)

	h.logger.WithFields(logrus.Fields{
		"max_open_conns":     h.config.MaxOpenConns,
		"max_idle_conns":     h.config.MaxIdleConns,
		"conn_max_lifetime":  h.config.ConnMaxLifetime,
		"conn_max_idle_time": h.config.ConnMaxIdleTime,
	}).Info("Database connection pool configured")
}

// sanitizeQuery removes sensitive information from queries for logging
func (h *dbHandler) sanitizeQuery(query string) string {
	// Basic sanitization - remove potential passwords or sensitive data
	// This is a simple implementation; you might want to enhance it
	if len(query) > 100 {
		return query[:100] + "..."
	}
	return query
}

// handlePostgreSQLError handles PostgreSQL-specific errors
func (h *dbHandler) handlePostgreSQLError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		h.logger.WithFields(logrus.Fields{
			"code":       pqErr.Code,
			"constraint": pqErr.Constraint,
			"detail":     pqErr.Detail,
			"hint":       pqErr.Hint,
			"position":   pqErr.Position,
			"table":      pqErr.Table,
			"column":     pqErr.Column,
		}).Error("PostgreSQL error occurred")

		// Map PostgreSQL errors to custom errors if needed
		switch pqErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("duplicate entry: %s", pqErr.Detail)
		case "23503": // foreign_key_violation
			return fmt.Errorf("foreign key constraint violation: %s", pqErr.Detail)
		case "23502": // not_null_violation
			return fmt.Errorf("required field missing: %s", pqErr.Column)
		default:
			return fmt.Errorf("database error [%s]: %s", pqErr.Code, pqErr.Message)
		}
	}

	return err
}
