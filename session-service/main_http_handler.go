package main

import (
	"net/http"
	"session-service/entities/sessions/handlers"
	sharedConfig "shared/config"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"
	sharedMiddleware "shared/middlewares"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MainHTTPHandler handles all HTTP requests for the session service
type MainHTTPHandler struct {
	sessionsHandler *handlers.HTTPHandler
	logger          *logrus.Logger
}

// NewMainHTTPHandler creates a new main HTTP handler
func NewMainHTTPHandler(cfg *sharedConfig.Config, logger *logrus.Logger) (*MainHTTPHandler, error) {
	// Create JWT handler
	jwtHandler := handlers.NewJWTHandler(cfg.GetString("JWT_SECRET"), cfg.GetDuration("JWT_EXPIRATION_TIME"), logger)

	// Create database handler
	dbHandler, err := handlers.NewDBHandler(cfg, jwtHandler, logger)
	if err != nil {
		return nil, err
	}

	// Create HTTP handler
	sessionsHandler := handlers.NewHTTPHandler(dbHandler, logger)

	return &MainHTTPHandler{
		sessionsHandler: sessionsHandler,
		logger:          logger,
	}, nil
}

// SetupRoutes sets up all the routes for the service
func (h *MainHTTPHandler) SetupRoutes(router *mux.Router) {
	// Add request ID middleware to all routes
	requestIDMiddleware := sharedMiddleware.NewRequestIDMiddleware(h.logger)
	gatewayMiddleware := sharedMiddleware.NewGatewayMiddleware(h.logger)
	router.Use(requestIDMiddleware.InjectRequestIDMiddleware)

	// Public router for endpoints that don't require gateway validation
	publicRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	publicRouter.HandleFunc("/p/health", h.HealthCheck).Methods("GET")
	publicRouter.HandleFunc("/p/login", h.sessionsHandler.CreateSession).Methods("POST")
	publicRouter.HandleFunc("/p/validate", h.sessionsHandler.ValidateSession).Methods("POST")

	// Protected endpoints (require gateway validation)
	protectedRouter := router.PathPrefix("/api/v1/sessions").Subrouter()
	protectedRouter.Use(gatewayMiddleware.CheckGatewayMiddleware)
	protectedRouter.Use(requestIDMiddleware.CheckRequestIDMiddleware(sharedLogger.SERVICE_SESSION_SERVICE, "/api/v1/sessions/p/health"))
	protectedRouter.HandleFunc("/logout", h.sessionsHandler.LogoutSession).Methods("POST")
}

// HealthCheck handles health check requests
func (h *MainHTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check data-service health
	dataServiceHealthy := h.checkDataServiceHealth(r)

	if !dataServiceHealthy {
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code: http.StatusServiceUnavailable,
			Data: map[string]interface{}{
				"status":  "unhealthy",
				"service": "session-service",
				"message": "Data-service is not healthy",
			},
		})
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "session-service",
		"message": "Session service is operational",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
		Code: http.StatusOK,
		Data: response,
	})
}

// checkDataServiceHealth checks if the data-service is healthy
func (h *MainHTTPHandler) checkDataServiceHealth(r *http.Request) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Direct call to data service (internal service communication)
	req, err := http.NewRequest("GET", "http://icecream_data_service:8086/api/v1/data/p/health", nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create data service health check request")

		return false
	}

	// Add gateway headers for internal service communication
	req.Header.Set("X-Gateway-Service", "gateway")
	req.Header.Set("X-Gateway-Session-Managed", "true")
	req.Header.Set("X-User-ID", "system")
	req.Header.Set("X-User-Role", "admin")

	// Forward the existing X-Request-ID from the current request
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to connect to data-service")

		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
