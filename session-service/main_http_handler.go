package main

import (
	"net/http"
	"session-service/config"
	"session-service/entities/sessions/handlers"
	"session-service/middleware"
	"time"

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
	publicRouter.HandleFunc("/p/validate", h.sessionsHandler.ValidateSession).Methods("POST")

	// Protected endpoints (require gateway validation)
	protectedRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	protectedRouter.Use(h.gatewayMiddleware.ValidateGateway)
	protectedRouter.HandleFunc("/logout", h.sessionsHandler.LogoutSession).Methods("POST")
}

// HealthCheck handles health check requests
func (h *MainHTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/api/v1/sessions/p/health",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Health check requested")

	// Check data-service health
	dataServiceHealthy := h.checkDataServiceHealth()

	if !dataServiceHealthy {
		h.logger.Error("Data-service health check failed")
		h.writeJSONResponse(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "unhealthy",
			"service": "session-service",
			"message": "Data-service is not healthy",
		})
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "session-service",
		"message": "Session service is operational",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// checkDataServiceHealth checks if the data-service is healthy
func (h *MainHTTPHandler) checkDataServiceHealth() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://icecream_data_service:8086/api/v1/data/p/health")
	if err != nil {
		h.logger.WithError(err).Error("Failed to connect to data-service")
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// writeJSONResponse writes a JSON response
func (h *MainHTTPHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}
