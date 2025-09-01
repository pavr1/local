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

	sharedConfig "shared/config"
	sharedLogger "shared/logger"
	sharedMiddleware "shared/middlewares"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup initial logger
	logger := sharedLogger.GetRequestLogger(nil, sharedLogger.SERVICE_INVOICE_SERVICE)
	logger.Info("Starting Ice Cream Store Invoice Service")

	// Load configuration from data service
	logger.Info("Loading configuration from data service...")
	dataServiceURL := getEnvString("DATA_SERVICE_URL", "http://icecream_data_service:8086")
	configLoader := sharedConfig.NewConfigLoader(dataServiceURL)

	cfg, err := configLoader.LoadConfig("Invoice", logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration from data service")
	}

	// Connect to database using config
	db, err := connectToDatabase(cfg, logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create main HTTP handler with all entity handlers
	mainHandler := NewMainHttpHandler(db, logger.Logger, cfg)

	// Setup HTTP router
	router := setupRouter(mainHandler, logger.Logger)

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.GetString("SERVER_HOST", "0.0.0.0"), cfg.GetString("SERVER_PORT", "8085")),
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

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// connectToDatabase establishes a connection to the PostgreSQL database using config
func connectToDatabase(cfg *sharedConfig.Config, logger *logrus.Logger) (*sql.DB, error) {
	// Build connection string using config
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.GetString("DB_HOST", "localhost"),
		cfg.GetString("DB_PORT", "5432"),
		cfg.GetString("DB_USER", "postgres"),
		cfg.GetString("DB_PASSWORD", "postgres123"),
		cfg.GetString("DB_NAME", "icecream_store"),
		cfg.GetString("DB_SSL_MODE", "disable"))

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

	//pvillalobos - use this public/protected router pattern for all services
	// publicRouter := router.PathPrefix("/api/v1/inventory").Subrouter()
	// publicRouter.HandleFunc("/p/health", h.HealthCheck).Methods("GET")

	// Health check endpoint
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Public health check endpoint
	v1.HandleFunc("/invoices/p/health", func(w http.ResponseWriter, r *http.Request) {
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
	// Protected health check endpoint (for internal use)
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
	invoicesRouter.Use(sharedMiddleware.CheckGatewayMiddleware(sharedLogger.SERVICE_INVOICE_SERVICE))
	invoicesRouter.Use(sharedMiddleware.CheckRequestIDMiddleware(sharedLogger.SERVICE_INVOICE_SERVICE, "/api/v1/invoices/p/health"))
	invoicesHandler := mainHandler.GetInvoicesHandler()

	// Expense Categories routes - under invoices (MUST be before generic {id} routes)
	expenseCategoriesRouter := invoicesRouter.PathPrefix("/expense-categories").Subrouter()
	expenseCategoriesHandler := mainHandler.GetExpenseCategoriesHandler()

	// Main expense category operations
	expenseCategoriesRouter.HandleFunc("", expenseCategoriesHandler.CreateExpenseCategory).Methods("POST")
	expenseCategoriesRouter.HandleFunc("", expenseCategoriesHandler.ListExpenseCategories).Methods("GET")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.GetExpenseCategory).Methods("GET")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.UpdateExpenseCategory).Methods("PUT")
	expenseCategoriesRouter.HandleFunc("/{id}", expenseCategoriesHandler.DeleteExpenseCategory).Methods("DELETE")

	// Main invoice operations (MUST be after specific routes)
	invoicesRouter.HandleFunc("", invoicesHandler.CreateInvoice).Methods("POST")
	invoicesRouter.HandleFunc("", invoicesHandler.ListInvoices).Methods("GET")
	invoicesRouter.HandleFunc("/number/{number}", invoicesHandler.GetInvoiceByNumber).Methods("GET")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.GetInvoiceByID).Methods("GET")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.UpdateInvoice).Methods("PUT")
	invoicesRouter.HandleFunc("/{id}", invoicesHandler.DeleteInvoice).Methods("DELETE")

	logger.Info("HTTP router configured successfully")
	return router
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

// getEnv returns the environment variable value or default if not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
