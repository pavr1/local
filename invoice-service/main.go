package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"invoice-service/config"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Invoice Service")

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create main HTTP handler with all entity handlers
	mainHandler := NewMainHttpHandler(db, logger)

	// Setup HTTP router
	router := setupRouter(mainHandler, logger)

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		return
	}

	logger.Info("Server exited")
}

// setupLogger configures the logger based on log level
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, defaulting to info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set JSON formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return logger
}

// connectToDatabase establishes a connection to the PostgreSQL database
func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to database")
	return db, nil
}

// setupRouter configures the HTTP router with all routes
func setupRouter(mainHandler *MainHttpHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(loggingMiddleware(logger))

	// CORS removed - gateway handles all CORS headers

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthData := mainHandler.HealthCheck()
		w.Header().Set("Content-Type", "application/json")

		// Check if service is unhealthy and set appropriate HTTP status
		status := http.StatusOK
		if healthData["status"] == "unhealthy" {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)

		// Use json.Marshal for proper JSON encoding
		jsonData, _ := json.Marshal(healthData)
		w.Write(jsonData)
	}).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public health endpoint (consistent with other services)
	api.HandleFunc("/invoices/p/health", func(w http.ResponseWriter, r *http.Request) {
		healthData := mainHandler.HealthCheck()
		w.Header().Set("Content-Type", "application/json")

		// Check if service is unhealthy and set appropriate HTTP status
		status := http.StatusOK
		if healthData["status"] == "unhealthy" {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)

		// Use json.Marshal for proper JSON encoding
		jsonData, _ := json.Marshal(healthData)
		w.Write(jsonData)
	}).Methods("GET")

	// Invoices routes (includes invoice details management)
	invoicesRouter := api.PathPrefix("/invoices").Subrouter()
	invoicesHandler := mainHandler.GetInvoicesHandler()

	// Main invoice operations
	invoicesRouter.HandleFunc("", invoicesHandler.CreateInvoiceWithDetails).Methods("POST")
	invoicesRouter.HandleFunc("", invoicesHandler.ListInvoices).Methods("GET")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.GetInvoiceByID).Methods("GET")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.UpdateInvoice).Methods("PUT")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.DeleteInvoice).Methods("DELETE")
	invoicesRouter.HandleFunc("/number/{number}", invoicesHandler.GetInvoiceByNumber).Methods("GET")

	// Invoice details are managed through the main invoice APIs

	// Expense Categories routes
	expenseCategoriesRouter := api.PathPrefix("/expense-categories").Subrouter()
	expenseCategoriesHandler := mainHandler.GetExpenseCategoriesHandler()

	// Main expense category operations
	expenseCategoriesRouter.HandleFunc("", expenseCategoriesHandler.CreateExpenseCategory).Methods("POST")
	expenseCategoriesRouter.HandleFunc("", expenseCategoriesHandler.ListExpenseCategories).Methods("GET")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.GetExpenseCategory).Methods("GET")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.UpdateExpenseCategory).Methods("PUT")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.DeleteExpenseCategory).Methods("DELETE")

	logger.Info("HTTP router configured successfully")
	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrappedWriter, r)

			// Log request
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"uri":         r.RequestURI,
				"status":      wrappedWriter.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Info("HTTP request processed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// corsMiddleware removed - gateway handles all CORS headers
