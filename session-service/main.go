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
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := sharedLogger.GetRequestLogger(nil, sharedLogger.SERVICE_SESSION_SERVICE)
	logger.Info("Starting Session Service")

	// Load configuration from data service
	logger.Info("Loading configuration from data service...")
	dataServiceURL := getEnvString("DATA_SERVICE_URL", "http://icecream_data_service:8086")
	configLoader := sharedConfig.NewConfigLoader(dataServiceURL)

	config, err := configLoader.LoadConfig("Session", logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration from data service")
	}

	logger.WithFields(logrus.Fields{
		"server_port": config.GetString("SERVER_PORT", "8081"),
		"server_host": config.GetString("SERVER_HOST", "0.0.0.0"),
		"log_level":   config.GetString("LOG_LEVEL", "info"),
	}).Info("Session service configuration loaded")

	// Connect to database
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.GetString("DB_HOST", "localhost"),
		config.GetString("DB_PORT", "5432"),
		config.GetString("DB_USER", "postgres"),
		config.GetString("DB_PASSWORD", "postgres123"),
		config.GetString("DB_NAME", "icecream_store"),
		config.GetString("DB_SSL_MODE", "disable")))
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}
	logger.Info("Database connection established")

	// Create main HTTP handler
	mainHandler, err := NewMainHTTPHandler(config, logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create HTTP handler")
	}

	// Setup router
	router := mux.NewRouter()
	mainHandler.SetupRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.GetString("SERVER_HOST", "0.0.0.0"), config.GetString("SERVER_PORT", "8081")),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"host": config.GetString("SERVER_HOST", "0.0.0.0"),
			"port": config.GetString("SERVER_PORT", "8081"),
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
