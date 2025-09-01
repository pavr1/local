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

	sharedConfig "shared/config"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup initial logger
	logger := sharedLogger.GetRequestLogger(nil, sharedLogger.SERVICE_INVENTORY_SERVICE) // Default log level for initial setup
	logger.Info("Starting Ice Cream Store Inventory Service")

	// Load configuration from data service
	logger.Info("Loading configuration from data service...")
	dataServiceUrl := sharedConfig.DATA_SERVICE_URL
	configLoader := sharedConfig.NewConfigLoader(dataServiceUrl)

	cfg, err := configLoader.LoadConfig("Inventory", logger.Logger)
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
	router := mux.NewRouter()
	mainHandler.SetupRoutes(router)
	serverHost := cfg.GetString("SERVER_HOST")
	serverPort := cfg.GetString("SERVER_PORT")

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", serverHost, serverPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"host": serverHost,
			"port": serverPort,
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

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// connectToDatabase establishes connection to PostgreSQL database using config
func connectToDatabase(cfg *sharedConfig.Config, logger *logrus.Logger) (*sql.DB, error) {
	dbHost := cfg.GetString("DB_HOST")
	dbPort := cfg.GetString("DB_PORT")
	dbUser := cfg.GetString("DB_USER")
	dbPassword := cfg.GetString("DB_PASSWORD")
	dbName := cfg.GetString("DB_NAME")
	dbSslMode := cfg.GetString("DB_SSL_MODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		dbSslMode)

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
	//pvillalobos - hardcoded values
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.WithFields(logrus.Fields{
		"host":     "localhost",
		"port":     "5432",
		"database": "icecream_store",
	}).Info("Successfully connected to database")

	return db, nil
}
