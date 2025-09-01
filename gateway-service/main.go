package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gateway-service/middleware"
	sessionmanager "gateway-service/middleware/session-manager"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	sharedConfig "shared/config"
	sharedLogger "shared/logger"
	sharedMiddleware "shared/middlewares"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Response structs
type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type HealthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}

// Service configuration - using shared config
type Config = sharedConfig.Config

func main() {
	logger := sharedLogger.GetRequestLogger(nil, sharedLogger.SERVICE_GATEWAY_SERVICE)
	logger.Info("Gateway service starting - logger initialized")

	dataServiceUrl := sharedConfig.DATA_SERVICE_URL

	logger.WithFields(logrus.Fields{
		"bootstrap_data_service_url": dataServiceUrl,
	}).Info("Created bootstrap configuration")

	// Load full configuration from data service
	logger.Info("Loading configuration from data service...")

	configLoader := sharedConfig.NewConfigLoader(dataServiceUrl)
	config, err := configLoader.LoadConfig("Gateway", logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration from data service")
	}

	gatewayServiceUrl := config.GetString("GATEWAY_SERVICE_URL")
	sessionServiceUrl := config.GetString("SESSION_SERVICE_URL")
	orderServiceUrl := config.GetString("ORDERS_SERVICE_URL")
	inventoryServiceUrl := config.GetString("INVENTORY_SERVICE_URL")
	invoiceServiceUrl := config.GetString("INVOICE_SERVICE_URL")
	dataServiceUrl = config.GetString("DATA_SERVICE_URL")

	logger.WithFields(logrus.Fields{
		"gateway_service":   gatewayServiceUrl,
		"session_service":   sessionServiceUrl,
		"orders_service":    orderServiceUrl,
		"inventory_service": inventoryServiceUrl,
		"invoice_service":   invoiceServiceUrl,
		"data_service":      dataServiceUrl,
	}).Info("Gateway service starting")

	// Create session manager for authentication with logger
	sessionManager := sessionmanager.NewSessionManager(sessionServiceUrl, logger.Logger)
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager, logger.Logger)

	r := mux.NewRouter()

	// Apply request ID middleware to main router (first)
	r.Use(sharedMiddleware.InjectRequestIDMiddleware(sharedLogger.SERVICE_GATEWAY_SERVICE))
	r.Use(sharedMiddleware.InjectGatewayHeadersMiddleware(sharedLogger.SERVICE_GATEWAY_SERVICE))

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/gateway/p/health", createHealthHandler(gatewayServiceUrl)).Methods("GET")

	// ==== SERVICE MANAGEMENT ENDPOINTS ====
	managementRouter := api.PathPrefix("/management").Subrouter()
	managementRouter.HandleFunc("/services/{service}/start", serviceStartHandler).Methods("POST")
	managementRouter.HandleFunc("/services/{service}/stop", serviceStopHandler).Methods("POST")
	managementRouter.HandleFunc("/services/{service}/restart", serviceRestartHandler).Methods("POST")

	// ==== PURE PROXY ROUTING TO SERVICES ====

	// Session service endpoints - pure proxy routing
	sessionRouter := api.PathPrefix("/v1/sessions").Subrouter()

	// Public session endpoints (no authentication required) - /p/ prefix
	sessionRouter.HandleFunc("/p/login", createProxyHandler(sessionServiceUrl, "/api/v1/sessions/p/login", logger.Logger)).Methods("POST")
	sessionRouter.HandleFunc("/p/validate", createProxyHandler(sessionServiceUrl, "/api/v1/sessions/p/validate", logger.Logger)).Methods("POST")
	// Protected session endpoints - session service handles authentication
	sessionRouter.HandleFunc("/logout", createProxyHandler(sessionServiceUrl, "/api/v1/sessions/logout", logger.Logger)).Methods("POST")

	// Public health endpoints (no authentication required)
	api.HandleFunc("/v1/sessions/p/health", createProxyHandler(sessionServiceUrl, "/api/v1/sessions/p/health", logger.Logger)).Methods("GET")
	api.HandleFunc("/v1/orders/p/health", createProxyHandler(orderServiceUrl, "/api/v1/orders/p/health", logger.Logger)).Methods("GET")
	api.HandleFunc("/v1/inventory/p/health", createProxyHandler(inventoryServiceUrl, "/api/v1/inventory/p/health", logger.Logger)).Methods("GET")
	api.HandleFunc("/v1/invoices/p/health", createProxyHandler(invoiceServiceUrl, "/api/v1/invoices/p/health", logger.Logger)).Methods("GET")
	api.HandleFunc("/v1/data/p/health", createProxyHandler(dataServiceUrl, "/api/v1/data/p/health", logger.Logger)).Methods("GET")

	// Public image serving endpoints (no authentication required) - MUST be defined BEFORE authenticated routes
	api.HandleFunc("/v1/data/images/{service}/{filename}", createProxyHandler(dataServiceUrl, "/api/v1/data/images/{service}/{filename}", logger.Logger)).Methods("GET")
	api.HandleFunc("/v1/data/images/{service}", createProxyHandler(dataServiceUrl, "/api/v1/data/images/{service}", logger.Logger)).Methods("POST")

	// Public logs endpoints (no authentication required) - for debugging
	api.HandleFunc("/v1/logs/{service}", createServiceLogsHandler(logger.Logger)).Methods("GET")

	// Orders service endpoints - with authentication middleware
	ordersRouter := api.PathPrefix("/v1/orders").Subrouter()
	ordersRouter.PathPrefix("").HandlerFunc(createProxyHandler(orderServiceUrl, "/api/v1/orders", logger.Logger))
	ordersRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Inventory service endpoints - with authentication middleware
	inventoryRouter := api.PathPrefix("/v1/inventory").Subrouter()
	inventoryRouter.PathPrefix("").HandlerFunc(createProxyHandler(inventoryServiceUrl, "/api/v1/inventory", logger.Logger))
	inventoryRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Invoice service routes - with authentication middleware
	invoiceRouter := api.PathPrefix("/v1/invoices").Subrouter()
	invoiceRouter.PathPrefix("").HandlerFunc(createProxyHandler(invoiceServiceUrl, "/api/v1", logger.Logger))
	invoiceRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Data service routes - with authentication middleware
	dataRouter := api.PathPrefix("/v1/data").Subrouter()

	// Settings endpoints (authenticated)
	dataRouter.HandleFunc("/settings/all", createProxyHandler(dataServiceUrl, "/api/v1/data/settings/all", logger.Logger)).Methods("GET")
	dataRouter.HandleFunc("/settings/by-service", createProxyHandler(dataServiceUrl, "/api/v1/data/settings/by-service", logger.Logger)).Methods("POST")
	dataRouter.HandleFunc("/settings/by-key", createProxyHandler(dataServiceUrl, "/api/v1/data/settings/by-key", logger.Logger)).Methods("POST")
	dataRouter.HandleFunc("/settings/reload", createProxyHandler(dataServiceUrl, "/api/v1/data/settings/reload", logger.Logger)).Methods("POST")
	dataRouter.HandleFunc("/settings/update-setting", createProxyHandler(dataServiceUrl, "/api/v1/data/settings/update-setting", logger.Logger)).Methods("POST")

	// Other data service endpoints (authenticated)
	dataRouter.PathPrefix("").HandlerFunc(createProxyHandler(dataServiceUrl, "/api/v1/data", logger.Logger))

	// Apply authentication middleware to data router
	dataRouter.Use(sessionMiddleware.ValidateSession)

	// Create CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware()

	// Apply CORS middleware to main router - gateway is single source of CORS
	r.Use(corsMiddleware.HandleCORS)

	// Add explicit OPTIONS handling for CORS preflight
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers are already set by corsMiddleware
		w.WriteHeader(http.StatusOK)
	})

	// UI is now served by its own service on port 3000
	// Static file serving removed - UI runs independently

	logger.Logger.WithFields(logrus.Fields{
		"url": gatewayServiceUrl,
	}).Info("Gateway service started successfully")

	logger.Logger.WithFields(logrus.Fields{
		"url": gatewayServiceUrl,
	}).Info("🚀 Gateway Service with Session Management starting on " + gatewayServiceUrl)
	logger.Logger.Info("📡 API available at http://localhost:8082/api")
	logger.Logger.Info("")
	logger.Logger.Info("🔐 SESSION MANAGEMENT ENDPOINTS:")
	logger.Logger.Info("   📂 Public:")
	logger.Logger.WithFields(logrus.Fields{
		"session_service_url": sessionServiceUrl,
	}).Info("      POST /api/v1/sessions/p/login    → " + sessionServiceUrl + "/api/v1/sessions/p/login (+ session creation)")
	logger.Logger.WithFields(logrus.Fields{
		"session_service_url": sessionServiceUrl,
	}).Info("      POST /api/v1/sessions/p/validate → " + sessionServiceUrl + "/api/v1/sessions/p/validate (+ session validation)")
	logger.Logger.WithFields(logrus.Fields{
		"session_service_url": sessionServiceUrl,
	}).Info("      GET  /api/v1/sessions/p/health   → " + sessionServiceUrl + "/api/v1/sessions/p/health")
	logger.Logger.Info("   🔒 Protected (require valid session):")
	logger.Logger.WithFields(logrus.Fields{
		"session_service_url": sessionServiceUrl,
	}).Info("      POST /api/v1/sessions/logout     → " + sessionServiceUrl + "/api/v1/sessions/logout (+ session revocation)")

	logger.Logger.Info("")
	logger.Logger.Info("🛒 BUSINESS SERVICE ENDPOINTS:")
	fmt.Println("   📂 Public Health Checks:")
	logger.Logger.WithFields(logrus.Fields{
		"orders_service_url": orderServiceUrl,
	}).Info("      GET  /api/v1/orders/p/health       → " + orderServiceUrl)
	logger.Logger.WithFields(logrus.Fields{
		"inventory_service_url": inventoryServiceUrl,
	}).Info("      GET  /api/v1/inventory/p/health    → " + inventoryServiceUrl)
	logger.Logger.WithFields(logrus.Fields{
		"invoice_service_url": invoiceServiceUrl,
	}).Info("      GET  /api/v1/invoices/p/health     → " + invoiceServiceUrl)
	logger.Logger.WithFields(logrus.Fields{
		"data_service_url": dataServiceUrl,
	}).Info("      GET  /api/v1/data/p/health         → " + dataServiceUrl)
	logger.Logger.Info("   🔒 Protected (require valid session):")
	logger.Logger.WithFields(logrus.Fields{
		"orders_service_url": orderServiceUrl,
	}).Info("      ALL  /api/v1/orders/*          → " + orderServiceUrl)
	logger.Logger.WithFields(logrus.Fields{
		"inventory_service_url": inventoryServiceUrl,
	}).Info("      ALL  /api/v1/inventory/*       → " + inventoryServiceUrl)
	logger.Logger.Info("           ├─ /suppliers/*          → Suppliers management")
	logger.Logger.Info("           ├─ /ingredients/*        → [Future] Ingredients management")
	logger.Logger.Info("           └─ /existences/*         → [Future] Stock management")
	logger.Logger.WithFields(logrus.Fields{
		"invoice_service_url": invoiceServiceUrl,
	}).Info("      ALL  /api/v1/invoices/*        → " + invoiceServiceUrl)
	logger.Logger.Info("           ├─ /invoices/*           → Invoice management")
	logger.Logger.Info("           └─ /invoices/{id}/details  → Invoice details management")
	logger.Logger.WithFields(logrus.Fields{
		"invoice_service_url": invoiceServiceUrl,
	}).Info("      ALL  /api/v1/expense-categories/* → " + invoiceServiceUrl)
	logger.Logger.Info("           └─ /expense-categories/*  → Expense categories management")
	logger.Logger.WithFields(logrus.Fields{
		"data_service_url": dataServiceUrl,
	}).Info("      ALL  /api/v1/data/*            → " + dataServiceUrl)
	logger.Logger.Info("           ├─ /settings/by-service   → Get settings by service")
	logger.Logger.Info("           ├─ /settings/by-name      → Get settings by name")
	logger.Logger.Info("           └─ /settings/grouped      → Get settings grouped by service")
	logger.Logger.Info("   📋 Service Management:")
	logger.Logger.Info("      GET  /api/v1/logs/{service}     → Service logs viewer")
	logger.Logger.Info("")
	logger.Logger.Info("📋 SESSION MANAGEMENT:")
	logger.Logger.WithFields(logrus.Fields{
		"session_service_url": sessionServiceUrl,
	}).Info("   🔒 /api/v1/sessions/*        → " + sessionServiceUrl + " (session validated)")
	logger.Logger.Info("")
	logger.Logger.Info("🔐 SESSION SECURITY FEATURES:")
	logger.Logger.Info("   ✅ Server-side token validation")
	logger.Logger.Info("   ✅ External token prevention")
	logger.Logger.Info("   ✅ Automatic token refresh")
	logger.Logger.Info("   ✅ Session revocation on logout")
	logger.Logger.Info("   ✅ User context injection")

	log.Fatal(http.ListenAndServe(":8082", r))
}

// createServiceLogsHandler creates a handler for service logs
func createServiceLogsHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serviceName := vars["service"]

		// Validate service name
		validServices := map[string]bool{
			"gateway-service":   true,
			"session-service":   true,
			"orders-service":    true,
			"inventory-service": true,
			"invoice-service":   true,
			"data-service":      true,
		}

		if !validServices[serviceName] {
			logger.WithFields(logrus.Fields{
				"service": serviceName,
			}).Error("Invalid service name")
			http.Error(w, "Invalid service name", http.StatusBadRequest)
			return
		}

		// Execute the make command to get logs
		cmd := exec.Command("make", "logs", "SERVICE="+serviceName)
		cmd.Dir = "." // Run from current directory (gateway-service)

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service": serviceName,
				"error":   err.Error(),
			}).Error("Failed to retrieve logs")
			http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
			return
		}

		// Return logs as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}
}

// createProxyHandler creates a reverse proxy handler for a specific service
func createProxyHandler(targetURL, stripPrefix string, logger *logrus.Logger) http.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"target_url": targetURL,
			"error":      err.Error(),
		}).Fatalf("Invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the proxy to handle errors and modify requests
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Determine which service is being called based on the URL path
		var serviceName string
		switch {
		case strings.Contains(r.URL.Path, "/sessions"):
			serviceName = "session-service"
		case strings.Contains(r.URL.Path, "/orders"):
			serviceName = "orders-service"
		case strings.Contains(r.URL.Path, "/inventory"):
			serviceName = "inventory-service"
		case strings.Contains(r.URL.Path, "/invoices"):
			serviceName = "invoice-service"
		default:
			serviceName = "unknown-service"
		}

		logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"target_url":  target.String(),
			"service":     serviceName,
			"error":       err.Error(),
			"remote_addr": r.RemoteAddr,
		}).Error("Proxy error - service unavailable")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "service_unavailable",
			"message":   fmt.Sprintf("The %s is currently unavailable", serviceName),
			"timestamp": time.Now(),
			"service":   serviceName,
			"path":      r.URL.Path,
		})
	}

	// Custom director to modify the request before forwarding
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Ensure request ID is set (should be set by middleware, but double-check)
		requestID := req.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a request ID if not present (fallback)
			requestID = generateRequestID()
			req.Header.Set("X-Request-ID", requestID)
			logger.WithField("request_id", requestID).Warn("Generated request ID in proxy director (fallback)")
		}

		// Add gateway headers
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Convert to hex string
	return hex.EncodeToString(bytes)
}

// createHealthHandler creates a health handler with config
func createHealthHandler(gatewayServiceUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)

		//pvillalobos - gateway should not be calling itself, it should be calling the downstream services here
		// Check all business services that appear on the dashboard + data service for UI monitoring
		// Use gateway's own proxy endpoints instead of direct service calls
		gatewayHealthy := true // Gateway is healthy if it's responding to this request
		sessionHealthy := checkServiceHealth(gatewayServiceUrl+"/api/v1/sessions/p/health", logger.Logger)
		ordersHealthy := checkServiceHealth(gatewayServiceUrl+"/api/v1/orders/p/health", logger.Logger)
		inventoryHealthy := checkServiceHealth(gatewayServiceUrl+"/api/v1/inventory/p/health", logger.Logger)
		invoiceHealthy := checkServiceHealth(gatewayServiceUrl+"/api/v1/invoices/p/health", logger.Logger)

		dataHealthy := checkServiceHealth(gatewayServiceUrl+"/api/v1/data/p/health", logger.Logger)
		status := "healthy"
		if !gatewayHealthy || !sessionHealthy || !ordersHealthy || !inventoryHealthy || !invoiceHealthy || !dataHealthy {
			status = "degraded"
		}

		response := map[string]interface{}{
			"status":             status,
			"version":            "1.0.0",
			"time":               time.Now(),
			"gateway":            "operational",
			"session_management": "enabled",
			"services": map[string]string{
				"gateway-service": func() string {
					if gatewayHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"session-service": func() string {
					if sessionHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"orders-service": func() string {
					if ordersHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"inventory-service": func() string {
					if inventoryHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"invoice-service": func() string {
					if invoiceHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"data-service": func() string {
					if dataHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		// Always return HTTP 200 - let the client decide how to handle degraded status
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// checkServiceHealth checks if a service is responding to health checks
func checkServiceHealth(healthURL string, logger *logrus.Logger) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create request with proper gateway headers
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"health_url": healthURL,
			"error":      err.Error(),
		}).Error("Failed to create health request")
		return false
	}

	// Add required gateway headers
	req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	req.Header.Set("X-Gateway-Session-Managed", "true")
	// Add user context headers for internal service communication
	req.Header.Set("X-User-ID", "system")
	req.Header.Set("X-Username", "system")
	req.Header.Set("X-User-Role", "admin")

	resp, err := client.Do(req)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"health_url": healthURL,
			"error":      err.Error(),
		}).Error("Health check failed")
		return false
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == http.StatusOK
	if !healthy {
		logger.WithFields(logrus.Fields{
			"health_url":  healthURL,
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		}).Warn("Service health check failed")
	}

	return healthy
}

// isServiceRunning checks if a service is currently running by checking its port
func isServiceRunning(serviceName string, logger *logrus.Logger) bool {
	// Map service names to their ports
	//pvillalobos - hardcoded values
	servicePorts := map[string]string{
		"gateway-service":   "8082",
		"session-service":   "8081",
		"orders-service":    "8083",
		"inventory-service": "8084",
		"invoice-service":   "8085",
		"data-service":      "8086",
	}

	port, exists := servicePorts[serviceName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Warn("Unknown service, cannot check running status")
		return false
	}

	// Check if the port is in use
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", port), 2*time.Second)
	if err != nil {
		// Port is not in use, service is not running
		return false
	}
	defer conn.Close()

	logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"port":         port,
	}).Info("Service is running")
	return true
}

// Service management handlers
func serviceStartHandler(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	logger.Logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"environment":  environment,
	}).Info("Starting service")

	// Check if service is already running
	isRunning := isServiceRunning(serviceName, logger.Logger)
	var finalOutput strings.Builder
	var finalSuccess bool = true
	var finalError error

	if isRunning {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Warn("Service is already running, stopping it first")

		finalOutput.WriteString(fmt.Sprintf("Service %s was already running, stopping first...\n", serviceName))

		// Stop the service first
		stopTarget := fmt.Sprintf("stop-%s", environment)
		stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget, logger.Logger)
		finalOutput.WriteString(fmt.Sprintf("Stop output: %s\n", stopOutput))

		if !stopSuccess || stopErr != nil {
			logger.Logger.WithError(stopErr).WithFields(logrus.Fields{
				"service_name": serviceName,
				"stop_output":  stopOutput,
			}).Error("Failed to stop running service")

			finalSuccess = false
			finalError = fmt.Errorf("failed to stop running service: %v", stopErr)
		} else {
			logger.Logger.WithFields(logrus.Fields{
				"service_name": serviceName,
				"stop_output":  stopOutput,
			}).Info("Successfully stopped running service")
			// Wait a moment for the service to fully stop
			time.Sleep(2 * time.Second)
		}
	}

	// Now start the service
	if finalSuccess {
		makeTarget := fmt.Sprintf("start-%s", environment)
		startSuccess, startOutput, startErr := executeServiceCommand(serviceName, makeTarget, logger.Logger)
		finalOutput.WriteString(fmt.Sprintf("Start output: %s", startOutput))

		if !startSuccess || startErr != nil {
			finalSuccess = false
			finalError = startErr
		}
	}

	message := fmt.Sprintf("Service %s start command executed", serviceName)
	if isRunning {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Info("Service was restarted (was already running)")
		message = fmt.Sprintf("Service %s was restarted (was already running)", serviceName)
	}

	response := map[string]interface{}{
		"service":     serviceName,
		"action":      "start",
		"environment": environment,
		"success":     finalSuccess,
		"message":     message,
		"output":      finalOutput.String(),
	}

	if finalError != nil {
		response["error"] = finalError.Error()
		logger.Logger.WithError(finalError).WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Error("Failed to start service")
	} else {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
			"environment":  environment,
		}).Info("Successfully executed start command for service")

		// If data-service was successfully started, automatically restart all dependent services
		if serviceName == "data-service" && finalSuccess {
			logger.Logger.WithFields(logrus.Fields{
				"service_name": serviceName,
				"environment":  environment,
			}).Info("Data service started successfully, auto-restarting dependent services")
			go restartDependentServices(environment, logger.Logger)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if finalSuccess {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}

// restartDependentServices automatically restarts all services that depend on the database
func restartDependentServices(environment string, logger *logrus.Logger) {
	// Services that depend on data-service (in dependency order)
	dependentServices := []string{
		"session-service",
		"orders-service",
		"inventory-service",
		"invoice-service",
		"gateway-service", // Gateway last to ensure all other services are ready
	}

	logger.WithFields(logrus.Fields{
		"environment": environment,
		"services":    dependentServices,
	}).Info("Starting automatic restart of dependent services")

	for _, serviceName := range dependentServices {
		logger.WithFields(logrus.Fields{
			"service_name": serviceName,
			"environment":  environment,
		}).Info("Auto-restarting service")

		// Check if service is running before attempting restart
		if isServiceRunning(serviceName, logger) {
			// Stop the service first
			stopTarget := fmt.Sprintf("stop-%s", environment)
			stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget, logger)

			if !stopSuccess || stopErr != nil {
				logger.WithError(stopErr).WithFields(logrus.Fields{
					"service_name": serviceName,
					"stop_output":  stopOutput,
				}).Error("Failed to stop service during auto-restart")
				continue // Skip to next service
			}

			logger.WithFields(logrus.Fields{
				"service_name": serviceName,
				"stop_output":  stopOutput,
			}).Info("Stopped service during auto-restart")

			// Wait for service to fully stop
			time.Sleep(2 * time.Second)
		}

		// Start the service
		startTarget := fmt.Sprintf("start-%s", environment)
		startSuccess, startOutput, startErr := executeServiceCommand(serviceName, startTarget, logger)

		if !startSuccess || startErr != nil {
			logger.WithError(startErr).WithFields(logrus.Fields{
				"service_name": serviceName,
				"start_output": startOutput,
			}).Error("Failed to start service during auto-restart")
		} else {
			logger.WithFields(logrus.Fields{
				"service_name": serviceName,
				"start_output": startOutput,
			}).Info("Successfully auto-restarted service")
		}

		// Wait before starting next service to avoid overwhelming the system
		time.Sleep(3 * time.Second)
	}

	logger.WithFields(logrus.Fields{
		"environment": environment,
		"services":    dependentServices,
	}).Info("Completed automatic restart of dependent services")
}

func serviceStopHandler(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	logger.Logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"environment":  environment,
	}).Info("Stopping service")

	// Check if service is already stopped
	isRunning := isServiceRunning(serviceName, logger.Logger)
	var success bool = true
	var output string
	var err error

	if !isRunning {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Info("Service is already stopped, ignoring stop request")
		output = fmt.Sprintf("Service %s was already stopped", serviceName)
		success = true
		err = nil
	} else {
		// Execute make command based on environment
		makeTarget := fmt.Sprintf("stop-%s", environment)
		success, output, err = executeServiceCommand(serviceName, makeTarget, logger.Logger)
	}

	message := fmt.Sprintf("Service %s stop command executed", serviceName)
	if !isRunning {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Info("Service was already stopped")
		message = fmt.Sprintf("Service %s was already stopped", serviceName)
	}

	response := map[string]interface{}{
		"service":     serviceName,
		"action":      "stop",
		"environment": environment,
		"success":     success,
		"message":     message,
		"output":      output,
	}

	if err != nil {
		response["error"] = err.Error()
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Error("Failed to stop service")
	} else {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Info("Successfully executed stop command for service")
	}

	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}

func serviceRestartHandler(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_GATEWAY_SERVICE)
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	logger.Logger.WithFields(logrus.Fields{
		"service_name": serviceName,
		"environment":  environment,
	}).Info("Restarting service")

	// For restart, we execute stop then start
	stopTarget := fmt.Sprintf("stop-%s", environment)
	startTarget := fmt.Sprintf("start-%s", environment)

	// First stop the service
	stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget, logger.Logger)

	// Wait a moment for graceful shutdown
	time.Sleep(2 * time.Second)

	// Then start the service
	startSuccess, startOutput, startErr := executeServiceCommand(serviceName, startTarget, logger.Logger)

	success := stopSuccess && startSuccess
	output := fmt.Sprintf("Stop output: %s\nStart output: %s", stopOutput, startOutput)

	response := map[string]interface{}{
		"service":     serviceName,
		"action":      "restart",
		"environment": environment,
		"success":     success,
		"message":     fmt.Sprintf("Service %s restart command executed", serviceName),
		"output":      output,
	}

	if stopErr != nil || startErr != nil {
		var errMsg string
		if stopErr != nil {
			errMsg += fmt.Sprintf("Stop error: %v ", stopErr)
		}
		if startErr != nil {
			errMsg += fmt.Sprintf("Start error: %v", startErr)
		}
		response["error"] = errMsg
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
			"stop_error":   stopErr,
			"start_error":  startErr,
			"error_msg":    errMsg,
		}).Error("Failed to restart service")
	} else {
		logger.Logger.WithFields(logrus.Fields{
			"service_name": serviceName,
			"environment":  environment,
		}).Info("Successfully executed restart command for service")
	}

	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}

// Execute service command using make in the appropriate directory
func executeServiceCommand(serviceName, makeTarget string, logger *logrus.Logger) (bool, string, error) {
	// Map service names to directories
	serviceDirectories := map[string]string{
		"data-service":      "data-service",
		"gateway-service":   "gateway-service",
		"session-service":   "session-service",
		"orders-service":    "orders-service",
		"inventory-service": "inventory-service",
		"invoice-service":   "invoice-service",
	}

	serviceDir, exists := serviceDirectories[serviceName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"service_name": serviceName,
		}).Error("Unknown service")
		return false, "", fmt.Errorf("unknown service: %s", serviceName)
	}

	// Build the command
	cmd := exec.Command("make", makeTarget)
	cmd.Dir = fmt.Sprintf("../%s", serviceDir) // Relative to gateway-service directory

	logger.WithFields(logrus.Fields{
		"service_dir": serviceDir,
		"make_target": makeTarget,
		"command":     fmt.Sprintf("cd %s && make %s", serviceDir, makeTarget),
	}).Info("Executing service command")

	// Capture output
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"service_dir": serviceDir,
			"make_target": makeTarget,
			"output":      string(output),
		}).Error("Command failed")
		return false, string(output), err
	}

	logger.WithFields(logrus.Fields{
		"service_dir": serviceDir,
		"make_target": makeTarget,
		"output":      string(output),
	}).Info("Command succeeded")
	return true, string(output), nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
