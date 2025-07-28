package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/config"
	"auth-service/handler"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Auth Service")

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create auth handler
	authHandler := handler.New(db, cfg, logger)

	// Setup HTTP router
	router := setupRouter(authHandler, logger)

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
		}).Info("Auth service starting on")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down auth service...")

	// Graceful shutdown
	if err := server.Close(); err != nil {
		logger.WithError(err).Error("Error during server shutdown")
	}

	logger.Info("Auth service shutdown complete")
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

	// Set formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     false,
	})

	// Log to stdout
	logger.SetOutput(os.Stdout)

	return logger
}

// connectToDatabase establishes a connection to the PostgreSQL database
func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)

	logger.WithFields(logrus.Fields{
		"host":   cfg.DatabaseHost,
		"port":   cfg.DatabasePort,
		"dbname": cfg.DatabaseName,
		"user":   cfg.DatabaseUser,
	}).Info("Connecting to database")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

// setupRouter configures the HTTP routes
func setupRouter(authHandler handler.AuthHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Add request logging middleware
	router.Use(loggingMiddleware(logger))

	// Add CORS middleware
	// router.Use(corsMiddleware) // Disabled: Gateway handles CORS for all services

	// Get auth middleware instance
	authMiddleware := authHandler.GetMiddleware()

	// Public routes (no authentication required)
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	publicRouter.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	publicRouter.HandleFunc("/auth/health", authHandler.HealthCheck).Methods("GET")

	// Protected routes (authentication required)
	protectedRouter := router.PathPrefix("/api/v1").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	// Auth endpoints
	protectedRouter.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	protectedRouter.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")
	protectedRouter.HandleFunc("/auth/validate", authHandler.ValidateToken).Methods("GET")
	protectedRouter.HandleFunc("/auth/profile", authHandler.GetProfile).Methods("GET")

	// Admin only endpoints
	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	adminRouter.Use(authMiddleware.RequirePermission("admin-read"))
	adminRouter.HandleFunc("/auth/token-info", authHandler.GetTokenInfo).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Simple JSON response without encoding issues
		fmt.Fprintf(w, `{
			"service": "ice-cream-auth-service",
			"version": "1.0.0",
			"status": "running",
			"time": "%s"
		}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	logger.Info("HTTP routes configured successfully")
	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"url":        r.URL.Path,
				"status":     wrapped.statusCode,
				"duration":   time.Since(start),
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
			}).Info("HTTP request")
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

// // corsMiddleware adds CORS headers
// func corsMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }
