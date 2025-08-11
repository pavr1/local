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
	"session-service/middleware"
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

	// Set up database storage (always enabled)
	dbStorage, err := utils.NewDatabaseSessionStorage(db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database session storage")
	}

	// Create session manager with database storage
	sessionConfig := cfg.ToSessionConfig()
	sessionManager := utils.NewSessionManager(jwtManager, sessionConfig, dbStorage, logger)

	// Create handlers (auth handler now gets session manager for login integration)
	sessionHandler := handler.NewSessionHandler(sessionManager, jwtManager, logger)
	sessionAPI := handler.NewSessionAPI(sessionManager, jwtManager, db, logger)

	// Setup HTTP router
	router := setupRouter(sessionHandler, sessionAPI, logger)

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

func setupRouter(sessionHandler *handler.SessionHandler, sessionAPI *handler.SessionAPI, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware(logger))

	// Gateway validation middleware - block direct access
	gatewayMiddleware := middleware.NewGatewayMiddleware(logger)
	router.Use(gatewayMiddleware.ValidateGateway)

	// CORS removed - gateway handles all CORS headers

	// ==== SESSION MANAGEMENT API ROUTES ====

	// Single session router to avoid routing conflicts
	sessionRouter := router.PathPrefix("/api/v1/sessions").Subrouter()

	// Public endpoints (no authentication required) - /p/ prefix
	sessionRouter.HandleFunc("/p/health", sessionAPI.HealthCheck).Methods("GET")
	sessionRouter.HandleFunc("/p/login", sessionAPI.Login).Methods("POST")
	sessionRouter.HandleFunc("/p/validate", sessionAPI.ValidateSession).Methods("POST")
	sessionRouter.HandleFunc("/p/logout", sessionAPI.RevokeSessionByToken).Methods("POST")

	// Internal/Gateway endpoints
	sessionRouter.HandleFunc("", sessionAPI.CreateSession).Methods("POST")          // POST /api/v1/sessions
	sessionRouter.HandleFunc("/refresh", sessionAPI.RefreshSession).Methods("POST") // POST /api/v1/sessions/refresh
	sessionRouter.HandleFunc("/stats", sessionAPI.GetSessionStats).Methods("GET")   // GET /api/v1/sessions/stats

	// Protected endpoints (TODO: add auth middleware when available)
	sessionRouter.HandleFunc("/user/{userID}", sessionAPI.GetUserSessions).Methods("GET")          // GET /api/v1/sessions/user/{userID}
	sessionRouter.HandleFunc("/user/{userID}", sessionAPI.RevokeAllUserSessions).Methods("DELETE") // DELETE /api/v1/sessions/user/{userID}
	sessionRouter.HandleFunc("/{sessionID}", sessionAPI.RevokeSession).Methods("DELETE")           // DELETE /api/v1/sessions/{sessionID}
	// protectedRouter.HandleFunc("/auth/profile", sessionAPI.GetProfile).Methods("GET") // TODO: GetProfile method not available on SessionAPI

	// Admin only endpoints - TODO: Re-implement when methods are available
	// adminRouter := protectedRouter.PathPrefix("").Subrouter()
	// adminRouter.Use(authMiddleware.RequirePermission("admin-read")) // TODO: Re-enable when middleware available
	// adminRouter.HandleFunc("/auth/token-info", sessionAPI.GetTokenInfo).Methods("GET") // TODO: GetTokenInfo method not available on SessionAPI

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
