package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service/config"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Inventory Service")

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
		logger.WithFields(logrus.Fields{
			"host": cfg.ServerHost,
			"port": cfg.ServerPort,
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Gracefully shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited gracefully")
}

// setupLogger configures the logrus logger
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

// connectToDatabase establishes connection to PostgreSQL database with retry logic
func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	var db *sql.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.WithError(err).Warnf("Failed to open database connection, attempt %d/%d", i+1, maxRetries)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			logger.WithError(err).Warnf("Failed to ping database, attempt %d/%d", i+1, maxRetries)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Connection successful
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.WithFields(logrus.Fields{
		"host":     cfg.DBHost,
		"port":     cfg.DBPort,
		"database": cfg.DBName,
	}).Info("Successfully connected to database")

	return db, nil
}

// setupRouter configures the HTTP routes
func setupRouter(mainHandler *MainHttpHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// API versioning
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	v1.HandleFunc("/inventory/health", func(w http.ResponseWriter, r *http.Request) {
		healthData := mainHandler.HealthCheck()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON encoding without importing encoding/json
		fmt.Fprintf(w, `{
			"service": "%s",
			"status": "%s",
			"timestamp": "%s",
			"entities": {
				"suppliers": "%s"
			}
		}`,
			healthData["service"],
			healthData["status"],
			time.Now().Format(time.RFC3339),
			healthData["entities"].(map[string]string)["suppliers"])
	}).Methods("GET")

	// Inventory module endpoints
	inventoryRouter := v1.PathPrefix("/inventory").Subrouter()

	// Suppliers endpoints under inventory
	suppliersRouter := inventoryRouter.PathPrefix("/suppliers").Subrouter()

	// GET /api/v1/inventory/suppliers - List all suppliers
	suppliersRouter.HandleFunc("", mainHandler.GetSuppliersHandler().ListSuppliers).Methods("GET")

	// POST /api/v1/inventory/suppliers - Create new supplier
	suppliersRouter.HandleFunc("", mainHandler.GetSuppliersHandler().CreateSupplier).Methods("POST")

	// GET /api/v1/inventory/suppliers/{id} - Get supplier by ID
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().GetSupplier).Methods("GET")

	// PUT /api/v1/inventory/suppliers/{id} - Update supplier
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().UpdateSupplier).Methods("PUT")

	// DELETE /api/v1/inventory/suppliers/{id} - Delete supplier
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().DeleteSupplier).Methods("DELETE")

	// TODO: Add other entity endpoints under inventory when implemented
	// ingredientsRouter := inventoryRouter.PathPrefix("/ingredients").Subrouter()
	// existencesRouter := inventoryRouter.PathPrefix("/existences").Subrouter()
	// recipesRouter := inventoryRouter.PathPrefix("/recipes").Subrouter()

	// Logging middleware
	router.Use(loggingMiddleware(logger))

	logger.Info("HTTP routes configured successfully")
	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Log the request
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrappedWriter.statusCode,
				"duration":   duration.String(),
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
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
