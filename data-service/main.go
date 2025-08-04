package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"data-service/pkg/database"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a logger with custom configuration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     true,
	})

	// Create database configuration
	config := &database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres123",
		DBName:   "icecream_store",
		SSLMode:  "disable",

		// Connection pool settings
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		// Timeout settings
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		// Retry settings
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}

	// Create database handler
	db := database.New(config, logger)

	// Connect to database
	fmt.Println("🍦 Connecting to Ice Cream Store Data Service...")
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Perform initial health check
	if err := db.HealthCheck(); err != nil {
		log.Fatalf("Initial database health check failed: %v", err)
	}

	fmt.Println("✅ Database connection established successfully")

	// Setup HTTP server
	router := setupRouter(db, logger)

	server := &http.Server{
		Addr:         ":8086", // Data service port
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", "8086").Info("Starting Data Service HTTP server")
		fmt.Println("🚀 Data Service HTTP server starting on :8086")
		fmt.Println("📡 Health endpoint available at: http://localhost:8086/health")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Data Service...")

	// Gracefully shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Data Service exited gracefully")
}

// setupRouter configures the HTTP routes
func setupRouter(db database.DatabaseHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCheck(w, r, db, logger)
	}).Methods("GET")

	// Stats endpoint (optional, for monitoring)
	router.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		statsEndpoint(w, r, db, logger)
	}).Methods("GET")

	return router
}

// healthCheck handles the health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request, db database.DatabaseHandler, logger *logrus.Logger) {
	response := map[string]interface{}{
		"service":   "data-service",
		"timestamp": time.Now(),
	}

	// Perform database health check
	if err := db.HealthCheck(); err != nil {
		logger.WithError(err).Error("Database health check failed")
		response["status"] = "unhealthy"
		response["message"] = "Database connection failed"
		response["error"] = err.Error()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Health check passed
	response["status"] = "healthy"
	response["message"] = "Database is operational"
	response["database"] = map[string]interface{}{
		"host":   "localhost",
		"port":   5432,
		"dbname": "icecream_store",
		"stats":  db.GetStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// statsEndpoint provides database connection statistics
func statsEndpoint(w http.ResponseWriter, r *http.Request, db database.DatabaseHandler, logger *logrus.Logger) {
	stats := db.GetStats()

	response := map[string]interface{}{
		"service":   "data-service",
		"timestamp": time.Now(),
		"database_stats": map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration":    stats.WaitDuration.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
