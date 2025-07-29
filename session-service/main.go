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

	"session-service/config"
	"session-service/handler"
	"session-service/utils"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Session Service")

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create JWT manager
	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpirationTime, logger)

	// Create session manager
	sessionConfig := cfg.ToSessionConfig()
	sessionManager := utils.NewSessionManager(jwtManager, sessionConfig, logger)

	// Create handlers (auth handler now gets session manager for login integration)
	authHandler := handler.New(db, cfg, logger, sessionManager)
	sessionAPI := handler.NewSessionAPI(sessionManager, jwtManager, logger)

	// Setup HTTP router
	router := setupRouter(authHandler, sessionAPI, logger)

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

func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseUser,
		cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabaseSSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Database connection established successfully")
	return db, nil
}

func setupRouter(authHandler handler.AuthHandler, sessionAPI *handler.SessionAPI, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware(logger))
	// CORS removed - gateway handles all CORS headers

	// Get auth middleware instance
	authMiddleware := authHandler.GetMiddleware()

	// ==== SESSION MANAGEMENT API ROUTES ====

	// Public session management routes (no authentication required)
	sessionPublicRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	sessionPublicRouter.HandleFunc("/health", sessionAPI.HealthCheck).Methods("GET")
	sessionPublicRouter.HandleFunc("/validate", sessionAPI.ValidateSession).Methods("POST")

	// Internal routes (for gateway use - could add API key protection later)
	sessionInternalRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	sessionInternalRouter.HandleFunc("", sessionAPI.CreateSession).Methods("POST")
	sessionInternalRouter.HandleFunc("/refresh", sessionAPI.RefreshSession).Methods("POST")
	sessionInternalRouter.HandleFunc("/logout", sessionAPI.RevokeSessionByToken).Methods("POST")
	sessionInternalRouter.HandleFunc("/stats", sessionAPI.GetSessionStats).Methods("GET")

	// Protected session management routes (authentication required)
	sessionProtectedRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	sessionProtectedRouter.Use(authMiddleware.Authenticate)
	sessionProtectedRouter.HandleFunc("/user/{userID}", sessionAPI.GetUserSessions).Methods("GET")
	sessionProtectedRouter.HandleFunc("/user/{userID}", sessionAPI.RevokeAllUserSessions).Methods("DELETE")
	sessionProtectedRouter.HandleFunc("/{sessionID}", sessionAPI.RevokeSession).Methods("DELETE")

	// ==== LEGACY AUTH API ROUTES (for backward compatibility) ====

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
			"service": "ice-cream-session-service",
			"version": "1.0.0",
			"status": "running",
			"time": "%s",
			"session_management": "enabled"
		}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	logger.Info("HTTP routes configured successfully with session management API")
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
