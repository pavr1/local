package main

import (
	"net/http"
	"session-service/config"
	"session-service/entities/sessions/handlers"
	"session-service/middleware"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MainHTTPHandler handles all HTTP requests for the session service
type MainHTTPHandler struct {
	sessionsHandler   *handlers.HTTPHandler
	gatewayMiddleware *middleware.GatewayMiddleware
	logger            *logrus.Logger
}

// NewMainHTTPHandler creates a new main HTTP handler
func NewMainHTTPHandler(cfg *config.Config, logger *logrus.Logger) (*MainHTTPHandler, error) {
	// Create JWT handler
	jwtHandler := handlers.NewJWTHandler(cfg.JWTSecret, cfg.JWTExpirationTime, logger)

	// Create database handler
	dbHandler, err := handlers.NewDBHandler(cfg, jwtHandler, logger)
	if err != nil {
		return nil, err
	}

	// Create HTTP handler
	sessionsHandler := handlers.NewHTTPHandler(dbHandler, logger)

	// Create gateway middleware
	gatewayMiddleware := middleware.NewGatewayMiddleware()

	return &MainHTTPHandler{
		sessionsHandler:   sessionsHandler,
		gatewayMiddleware: gatewayMiddleware,
		logger:            logger,
	}, nil
}

// SetupRoutes sets up all the routes for the service
func (h *MainHTTPHandler) SetupRoutes(router *mux.Router) {
	// Public router for endpoints that don't require gateway validation
	publicRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	publicRouter.HandleFunc("/p/health", h.HealthCheck).Methods("GET")
	publicRouter.HandleFunc("/p/login", h.sessionsHandler.CreateSession).Methods("POST")

	// Protected endpoints (require gateway validation)
	protectedRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	protectedRouter.Use(h.gatewayMiddleware.ValidateGateway)
	protectedRouter.HandleFunc("/validate", h.sessionsHandler.ValidateSession).Methods("POST")
	protectedRouter.HandleFunc("/logout", h.sessionsHandler.LogoutSession).Methods("POST")
}

// HealthCheck handles health check requests
func (h *MainHTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "session-service",
		"message": "Session service is operational",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response
func (h *MainHTTPHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// For now, we'll just write a simple response
	// In a real implementation, you'd use json.NewEncoder
	w.Write([]byte(`{"status":"ok"}`))
}
