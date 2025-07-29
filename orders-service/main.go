package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orders-service/config"
	"orders-service/handler"

	// Removed middleware import - gateway handles all auth

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Orders Service")

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create orders handler
	ordersHandler, err := handler.New(db, cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create orders handler")
	}

	// Setup HTTP router
	router := setupRouter(ordersHandler, logger)

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
		}).Info("Orders service starting on")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down orders service...")

	// Graceful shutdown
	if err := server.Close(); err != nil {
		logger.WithError(err).Error("Error during server shutdown")
	}

	logger.Info("Orders service shutdown complete")
}

// setupLogger configures the logrus logger
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return logger
}

// connectToDatabase establishes database connection
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
			logger.WithError(err).Warnf("Database connection attempt %d failed", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			logger.WithError(err).Warnf("Database ping attempt %d failed", i+1)
			db.Close()
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
	db.SetConnMaxLifetime(30 * time.Minute)

	logger.Info("Database connection established successfully")
	return db, nil
}

// setupRouter configures the HTTP routes
func setupRouter(ordersHandler handler.OrdersHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Create middleware
	// Removed authMiddleware - gateway handles all auth

	// Add global middleware
	// Removed authMiddleware.LoggingMiddleware - gateway handles all logging
	// router.Use(authMiddleware.CORS) // Disabled: Gateway handles CORS for all services

	// Public routes (no authentication required)
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	publicRouter.HandleFunc("/orders/health", ordersHandler.HealthCheck).Methods("GET")

	// Protected routes (authentication required)
	protectedRouter := router.PathPrefix("/api/v1").Subrouter()
	// Removed protectedRouter.Use(authMiddleware.Authenticate) - gateway handles all auth

	// Order endpoints with permission checks
	// Create order - requires orders-write permission
	protectedRouter.Handle("/orders",
		// Removed authMiddleware.RequireOrdersPermission("write") - gateway handles all auth
		http.HandlerFunc(ordersHandler.CreateOrder)).Methods("POST")

	// Get order - requires orders-read permission
	protectedRouter.Handle("/orders/{id}",
		// Removed authMiddleware.RequireOrdersPermission("read") - gateway handles all auth
		http.HandlerFunc(ordersHandler.GetOrder)).Methods("GET")

	// Update order - requires orders-write permission
	protectedRouter.Handle("/orders/{id}",
		// Removed authMiddleware.RequireOrdersPermission("write") - gateway handles all auth
		http.HandlerFunc(ordersHandler.UpdateOrder)).Methods("PUT")

	// Cancel order - requires orders-write permission
	protectedRouter.Handle("/orders/{id}/cancel",
		// Removed authMiddleware.RequireOrdersPermission("write") - gateway handles all auth
		http.HandlerFunc(ordersHandler.CancelOrder)).Methods("POST")

	// List orders - requires orders-read permission
	protectedRouter.Handle("/orders",
		// Removed authMiddleware.RequireOrdersPermission("read") - gateway handles all auth
		http.HandlerFunc(ordersHandler.ListOrders)).Methods("GET")

	// Statistics endpoints - admin only
	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	// Removed adminRouter.Use(authMiddleware.AdminOnly) - gateway handles all auth

	adminRouter.HandleFunc("/orders/summary", ordersHandler.GetOrderSummary).Methods("GET")
	adminRouter.HandleFunc("/orders/stats/payment-methods", ordersHandler.GetPaymentMethodStats).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"service": "ice-cream-orders-service",
			"version": "1.0.0",
			"status": "healthy",
			"timestamp": "%s",
			"endpoints": {
				"health": "/api/v1/orders/health",
				"orders": "/api/v1/orders",
				"statistics": "/api/v1/orders/summary"
			}
		}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	return router
}
