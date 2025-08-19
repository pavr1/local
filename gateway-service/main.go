package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gateway-service/middleware"
	sessionmanager "gateway-service/middleware/session-manager"
	"gateway-service/pkg/logger"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Global logger instance
var logrusLogger *logrus.Logger

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

// Initialize logger with structured format
func initLogger() *logrus.Logger {
	log := logrus.New()

	// Set log level from environment or default to info
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set structured JSON formatter for better parsing
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Set output to stdout for containerized environments
	log.SetOutput(os.Stdout)

	return log
}

// corsMiddleware handles CORS for all services - gateway is the single source of truth
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers - only the gateway sets these
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Service configuration
type Config struct {
	Port                string
	SessionServiceURL   string
	OrdersServiceURL    string
	InventoryServiceURL string
	InvoiceServiceURL   string
	DataServiceURL      string
}

func main() {
	// Initialize structured logger
	logrusLogger = initLogger()

	// Bootstrap config with hardcoded data service URL for initial config loading
	bootstrapConfig := Config{
		DataServiceURL: "http://icecream_data_service:8086", // Hardcoded for bootstrap - Docker service name
	}

	// Load full configuration from data service
	config, err := loadConfigFromDataService(&bootstrapConfig, logrusLogger)
	if err != nil {
		logrusLogger.WithError(err).Fatal("Failed to load configuration from data service")
	}

	// Initialize centralized logging
	fluentdHost := getEnv("FLUENTD_HOST", "localhost")
	fluentdPort := 24224
	if port := getEnv("FLUENTD_PORT", ""); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			fluentdPort = p
		}
	}

	// Initialize shared logger
	centralLogger := logger.InitLogger("gateway-service", fluentdHost, fluentdPort)
	defer centralLogger.Close()

	logrusLogger.WithFields(logrus.Fields{
		"port":              config.Port,
		"session_service":   config.SessionServiceURL,
		"orders_service":    config.OrdersServiceURL,
		"inventory_service": config.InventoryServiceURL,
		"invoice_service":   config.InvoiceServiceURL,
		"data_service":      config.DataServiceURL,
		"fluentd_host":      fluentdHost,
		"fluentd_port":      fluentdPort,
	}).Info("Gateway service starting")

	centralLogger.Info("Gateway service starting", map[string]interface{}{
		"port":              config.Port,
		"session_service":   config.SessionServiceURL,
		"orders_service":    config.OrdersServiceURL,
		"inventory_service": config.InventoryServiceURL,
		"invoice_service":   config.InvoiceServiceURL,
		"data_service":      config.DataServiceURL,
	})

	log.Printf("Gateway configured with Invoice Service: %s", config.InvoiceServiceURL)
	log.Printf("Gateway configured with Session Service: %s", config.SessionServiceURL)
	log.Printf("Gateway configured with Orders Service: %s", config.OrdersServiceURL)
	log.Printf("Gateway configured with Inventory Service: %s", config.InventoryServiceURL)
	log.Printf("Gateway configured with Data Service: %s", config.DataServiceURL)

	// Create session manager for authentication with logger
	sessionManager := sessionmanager.NewSessionManager(config.SessionServiceURL, logrusLogger)
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)

	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// ==== GATEWAY ENDPOINTS ====

	// Gateway health check endpoint
	// API versioning
	v1 := api.PathPrefix("/v1").Subrouter()

	// Public health check endpoint
	v1.HandleFunc("/gateway/p/health", createHealthHandler(config, logrusLogger)).Methods("GET")

	// ==== SERVICE MANAGEMENT ENDPOINTS ====
	managementRouter := api.PathPrefix("/management").Subrouter()
	managementRouter.HandleFunc("/services/{service}/start", serviceStartHandler).Methods("POST")
	managementRouter.HandleFunc("/services/{service}/stop", serviceStopHandler).Methods("POST")
	managementRouter.HandleFunc("/services/{service}/restart", serviceRestartHandler).Methods("POST")

	// ==== PURE PROXY ROUTING TO SERVICES ====

	// Session service endpoints - pure proxy routing
	sessionRouter := api.PathPrefix("/v1/sessions").Subrouter()

	// Public session endpoints (no authentication required) - /p/ prefix
	sessionRouter.HandleFunc("/p/login", createProxyHandler(config.SessionServiceURL, "/api/v1/sessions/p/login", logrusLogger)).Methods("POST")
	sessionRouter.HandleFunc("/p/validate", createProxyHandler(config.SessionServiceURL, "/api/v1/sessions/p/validate", logrusLogger)).Methods("POST")
	// Protected session endpoints - session service handles authentication
	sessionRouter.HandleFunc("/logout", createProxyHandler(config.SessionServiceURL, "/api/v1/sessions/logout", logrusLogger)).Methods("POST")

	// Public health endpoints (no authentication required)
	api.HandleFunc("/v1/sessions/p/health", createProxyHandler(config.SessionServiceURL, "/api/v1/sessions/p/health", logrusLogger)).Methods("GET")
	api.HandleFunc("/v1/orders/p/health", createProxyHandler(config.OrdersServiceURL, "/api/v1/orders/p/health", logrusLogger)).Methods("GET")
	api.HandleFunc("/v1/inventory/p/health", createProxyHandler(config.InventoryServiceURL, "/api/v1/inventory/p/health", logrusLogger)).Methods("GET")
	api.HandleFunc("/v1/invoices/p/health", createInvoiceHealthHandler(config.InvoiceServiceURL, logrusLogger)).Methods("GET")
	api.HandleFunc("/v1/data/p/health", createProxyHandler(config.DataServiceURL, "/api/v1/data/p/health", logrusLogger)).Methods("GET")

	// Public logs endpoints (no authentication required) - for debugging
	api.HandleFunc("/v1/logs/{service}", createServiceLogsHandler(logrusLogger)).Methods("GET")

	// Orders service endpoints - with authentication middleware
	ordersRouter := api.PathPrefix("/v1/orders").Subrouter()
	ordersRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.OrdersServiceURL, "/api/v1/orders", logrusLogger))
	ordersRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Inventory service endpoints - with authentication middleware
	inventoryRouter := api.PathPrefix("/v1/inventory").Subrouter()
	inventoryRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.InventoryServiceURL, "/api/v1/inventory", logrusLogger))
	inventoryRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Invoice service routes - with authentication middleware
	invoiceRouter := api.PathPrefix("/v1/invoices").Subrouter()
	invoiceRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.InvoiceServiceURL, "/api/v1", logrusLogger))
	invoiceRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Data service routes - with authentication middleware
	dataRouter := api.PathPrefix("/v1/data").Subrouter()
	
	// Settings endpoints (authenticated)
	dataRouter.HandleFunc("/settings/all", createProxyHandler(config.DataServiceURL, "/api/v1/data/settings/all", logrusLogger)).Methods("GET")
	dataRouter.HandleFunc("/settings/by-service", createProxyHandler(config.DataServiceURL, "/api/v1/data/settings/by-service", logrusLogger)).Methods("POST")
	dataRouter.HandleFunc("/settings/by-key", createProxyHandler(config.DataServiceURL, "/api/v1/data/settings/by-key", logrusLogger)).Methods("POST")
	dataRouter.HandleFunc("/settings/reload", createProxyHandler(config.DataServiceURL, "/api/v1/data/settings/reload", logrusLogger)).Methods("POST")
	dataRouter.HandleFunc("/settings/update-setting", createProxyHandler(config.DataServiceURL, "/api/v1/data/settings/update-setting", logrusLogger)).Methods("POST")
	
	// Other data service endpoints
	dataRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.DataServiceURL, "/api/v1/data", logrusLogger))
	dataRouter.Use(sessionMiddleware.ValidateSession) // Add authentication for business endpoints

	// Apply CORS middleware to main router - gateway is single source of CORS
	r.Use(corsMiddleware)

	// Add explicit OPTIONS handling for CORS preflight
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers are already set by corsMiddleware
		w.WriteHeader(http.StatusOK)
	})

	// UI is now served by its own service on port 3000
	// Static file serving removed - UI runs independently

	logrusLogger.WithFields(logrus.Fields{
		"port": config.Port,
		"url":  fmt.Sprintf("http://localhost:%s", config.Port),
	}).Info("Gateway service started successfully")

	fmt.Println("🚀 Gateway Service with Session Management starting on http://localhost:8082")
	fmt.Println("📡 API available at http://localhost:8082/api")
	fmt.Println("")
	fmt.Println("🔐 SESSION MANAGEMENT ENDPOINTS:")
	fmt.Println("   📂 Public:")
	fmt.Printf("      POST /api/v1/sessions/p/login    → %s/api/v1/sessions/p/login (+ session creation)\n", config.SessionServiceURL)
	fmt.Printf("      POST /api/v1/sessions/p/validate → %s/api/v1/sessions/p/validate (+ session validation)\n", config.SessionServiceURL)
	fmt.Printf("      GET  /api/v1/sessions/p/health   → %s/api/v1/sessions/p/health\n", config.SessionServiceURL)
	fmt.Println("   🔒 Protected (require valid session):")
	fmt.Printf("      POST /api/v1/sessions/logout     → %s/api/v1/sessions/logout (+ session revocation)\n", config.SessionServiceURL)

	fmt.Println("")
	fmt.Println("🛒 BUSINESS SERVICE ENDPOINTS:")
	fmt.Println("   📂 Public Health Checks:")
	fmt.Printf("      GET  /api/v1/orders/p/health       → %s\n", config.OrdersServiceURL)
	fmt.Printf("      GET  /api/v1/inventory/p/health    → %s\n", config.InventoryServiceURL)
	fmt.Printf("      GET  /api/v1/invoices/p/health     → %s\n", config.InvoiceServiceURL)
	fmt.Printf("      GET  /api/v1/data/p/health         → %s\n", config.DataServiceURL)
	fmt.Println("   🔒 Protected (require valid session):")
	fmt.Printf("      ALL  /api/v1/orders/*          → %s\n", config.OrdersServiceURL)
	fmt.Printf("      ALL  /api/v1/inventory/*       → %s\n", config.InventoryServiceURL)
	fmt.Printf("           ├─ /suppliers/*          → Suppliers management\n")
	fmt.Printf("           ├─ /ingredients/*        → [Future] Ingredients management\n")
	fmt.Printf("           └─ /existences/*         → [Future] Stock management\n")
	fmt.Printf("      ALL  /api/v1/invoices/*        → %s\n", config.InvoiceServiceURL)
	fmt.Printf("           ├─ /invoices/*           → Invoice management\n")
	fmt.Printf("           └─ /invoices/{id}/details  → Invoice details management\n")
	fmt.Printf("      ALL  /api/v1/expense-categories/* → %s\n", config.InvoiceServiceURL)
	fmt.Printf("           └─ /expense-categories/*  → Expense categories management\n")
	fmt.Printf("      ALL  /api/v1/data/*            → %s\n", config.DataServiceURL)
	fmt.Printf("           ├─ /settings/by-service   → Get settings by service\n")
	fmt.Printf("           ├─ /settings/by-name      → Get settings by name\n")
	fmt.Printf("           └─ /settings/grouped      → Get settings grouped by service\n")
	fmt.Println("   📋 Service Management:")
	fmt.Printf("      GET  /api/v1/logs/{service}     → Service logs viewer\n")
	fmt.Println("")
	fmt.Println("📋 SESSION MANAGEMENT:")
	fmt.Printf("   🔒 /api/v1/sessions/*        → %s (session validated)\n", config.SessionServiceURL)
	fmt.Println("")
	fmt.Println("🔐 SESSION SECURITY FEATURES:")
	fmt.Println("   ✅ Server-side token validation")
	fmt.Println("   ✅ External token prevention")
	fmt.Println("   ✅ Automatic token refresh")
	fmt.Println("   ✅ Session revocation on logout")
	fmt.Println("   ✅ User context injection")

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

// createInvoiceHealthHandler creates a custom health handler for invoice service
func createInvoiceHealthHandler(invoiceServiceURL string, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a direct request to invoice service health endpoint
		healthURL := invoiceServiceURL + "/api/v1/invoices/p/health"

		logger.WithFields(logrus.Fields{
			"invoice_service_url": healthURL,
		}).Info("Proxying invoice health check")

		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest("GET", healthURL, nil)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Failed to create health request")
			http.Error(w, "Failed to create health request", http.StatusInternalServerError)
			return
		}

		// Add gateway headers
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
		req.Header.Set("X-Gateway-Session-Managed", "true")
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)

		resp, err := client.Do(req)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Invoice health check failed")
			http.Error(w, "Invoice service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Copy status code and body
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
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

		// Log the proxy request (only for important requests)
		if req.URL.Path != "/api/v1/sessions/p/health" {
			logger.WithFields(logrus.Fields{
				"method": req.Method,
				"path":   req.URL.Path,
				"target": target.String(),
			}).Info("Proxying request")
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

// createHealthHandler creates a health handler with config
func createHealthHandler(config *Config, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check all business services that appear on the dashboard + data service for UI monitoring
		gatewayHealthy := true // Gateway is healthy if it's responding to this request
		sessionHealthy := checkServiceHealth(config.SessionServiceURL+"/api/v1/sessions/p/health", logger)
		ordersHealthy := checkServiceHealth(config.OrdersServiceURL+"/api/v1/orders/p/health", logger)
		inventoryHealthy := checkServiceHealth(config.InventoryServiceURL+"/api/v1/inventory/p/health", logger)
		invoiceHealthy := checkServiceHealth(config.InvoiceServiceURL+"/api/v1/invoices/p/health", logger)
		//pvillalobos - gateway should not be hitting data service health endpoint, all business services do that already
		dataHealthy := checkServiceHealth(config.DataServiceURL+"/api/v1/data/p/health", logger) // For UI monitoring

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

	resp, err := client.Do(req)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"health_url": healthURL,
			"error":      err.Error(),
		}).Error("Health check failed")
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// isServiceRunning checks if a service is currently running by checking its port
func isServiceRunning(serviceName string) bool {
	// Map service names to their ports
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
		log.Printf("⚠️  Unknown service %s, cannot check running status", serviceName)
		return false
	}

	// Check if the port is in use
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", port), 2*time.Second)
	if err != nil {
		// Port is not in use, service is not running
		return false
	}
	defer conn.Close()

	log.Printf("🔍 Service %s is running on port %s", serviceName, port)
	return true
}

// Service management handlers
func serviceStartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	log.Printf("🔧 Starting %s service (environment: %s)", serviceName, environment)

	// Check if service is already running
	isRunning := isServiceRunning(serviceName)
	var finalOutput strings.Builder
	var finalSuccess bool = true
	var finalError error

	if isRunning {
		log.Printf("⚠️  Service %s is already running, stopping it first...", serviceName)
		finalOutput.WriteString(fmt.Sprintf("Service %s was already running, stopping first...\n", serviceName))

		// Stop the service first
		stopTarget := fmt.Sprintf("stop-%s", environment)
		stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget)
		finalOutput.WriteString(fmt.Sprintf("Stop output: %s\n", stopOutput))

		if !stopSuccess || stopErr != nil {
			log.Printf("❌ Failed to stop running service %s: %v", serviceName, stopErr)
			finalSuccess = false
			finalError = fmt.Errorf("failed to stop running service: %v", stopErr)
		} else {
			log.Printf("✅ Successfully stopped running service %s", serviceName)
			// Wait a moment for the service to fully stop
			time.Sleep(2 * time.Second)
		}
	}

	// Now start the service
	if finalSuccess {
		makeTarget := fmt.Sprintf("start-%s", environment)
		startSuccess, startOutput, startErr := executeServiceCommand(serviceName, makeTarget)
		finalOutput.WriteString(fmt.Sprintf("Start output: %s", startOutput))

		if !startSuccess || startErr != nil {
			finalSuccess = false
			finalError = startErr
		}
	}

	message := fmt.Sprintf("Service %s start command executed", serviceName)
	if isRunning {
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
		log.Printf("❌ Failed to start %s: %v", serviceName, finalError)
	} else {
		log.Printf("✅ Successfully executed start command for %s", serviceName)

		// If data-service was successfully started, automatically restart all dependent services
		if serviceName == "data-service" && finalSuccess {
			log.Printf("🔄 Data service started successfully, auto-restarting dependent services...")
			go restartDependentServices(environment)
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
func restartDependentServices(environment string) {
	// Services that depend on data-service (in dependency order)
	dependentServices := []string{
		"session-service",
		"orders-service",
		"inventory-service",
		"invoice-service",
		"gateway-service", // Gateway last to ensure all other services are ready
	}

	log.Printf("🔄 Starting automatic restart of dependent services...")

	for _, serviceName := range dependentServices {
		log.Printf("🔄 Auto-restarting %s...", serviceName)

		// Check if service is running before attempting restart
		if isServiceRunning(serviceName) {
			// Stop the service first
			stopTarget := fmt.Sprintf("stop-%s", environment)
			stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget)

			if !stopSuccess || stopErr != nil {
				log.Printf("❌ Failed to stop %s during auto-restart: %v", serviceName, stopErr)
				continue // Skip to next service
			}

			log.Printf("✅ Stopped %s, output: %s", serviceName, stopOutput)

			// Wait for service to fully stop
			time.Sleep(2 * time.Second)
		}

		// Start the service
		startTarget := fmt.Sprintf("start-%s", environment)
		startSuccess, startOutput, startErr := executeServiceCommand(serviceName, startTarget)

		if !startSuccess || startErr != nil {
			log.Printf("❌ Failed to start %s during auto-restart: %v", serviceName, startErr)
		} else {
			log.Printf("✅ Successfully auto-restarted %s, output: %s", serviceName, startOutput)
		}

		// Wait before starting next service to avoid overwhelming the system
		time.Sleep(3 * time.Second)
	}

	log.Printf("🎉 Completed automatic restart of dependent services!")
}

func serviceStopHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	log.Printf("🔧 Stopping %s service (environment: %s)", serviceName, environment)

	// Check if service is already stopped
	isRunning := isServiceRunning(serviceName)
	var success bool = true
	var output string
	var err error

	if !isRunning {
		log.Printf("ℹ️  Service %s is already stopped, ignoring stop request", serviceName)
		output = fmt.Sprintf("Service %s was already stopped", serviceName)
		success = true
		err = nil
	} else {
		// Execute make command based on environment
		makeTarget := fmt.Sprintf("stop-%s", environment)
		success, output, err = executeServiceCommand(serviceName, makeTarget)
	}

	message := fmt.Sprintf("Service %s stop command executed", serviceName)
	if !isRunning {
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
		log.Printf("❌ Failed to stop %s: %v", serviceName, err)
	} else {
		log.Printf("✅ Successfully executed stop command for %s", serviceName)
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
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var requestBody struct {
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	environment := requestBody.Environment
	if environment == "" {
		environment = "locally" // Default
	}

	log.Printf("🔧 Restarting %s service (environment: %s)", serviceName, environment)

	// For restart, we execute stop then start
	stopTarget := fmt.Sprintf("stop-%s", environment)
	startTarget := fmt.Sprintf("start-%s", environment)

	// First stop the service
	stopSuccess, stopOutput, stopErr := executeServiceCommand(serviceName, stopTarget)

	// Wait a moment for graceful shutdown
	time.Sleep(2 * time.Second)

	// Then start the service
	startSuccess, startOutput, startErr := executeServiceCommand(serviceName, startTarget)

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
		log.Printf("❌ Failed to restart %s: %s", serviceName, errMsg)
	} else {
		log.Printf("✅ Successfully executed restart command for %s", serviceName)
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
func executeServiceCommand(serviceName, makeTarget string) (bool, string, error) {
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
		return false, "", fmt.Errorf("unknown service: %s", serviceName)
	}

	// Build the command
	cmd := exec.Command("make", makeTarget)
	cmd.Dir = fmt.Sprintf("../%s", serviceDir) // Relative to gateway-service directory

	log.Printf("🔧 Executing: cd %s && make %s", serviceDir, makeTarget)

	// Capture output
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ Command failed: %v, output: %s", err, string(output))
		return false, string(output), err
	}

	log.Printf("✅ Command succeeded, output: %s", string(output))
	return true, string(output), nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// loadConfigFromDataService loads configuration from the data service API
func loadConfigFromDataService(bootstrapConfig *Config, logger *logrus.Logger) (*Config, error) {
	// Get settings from data service
	settings, err := getSettingsFromDataService("Gateway", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings from data service: %w", err)
	}

	// Create config with defaults
	config := &Config{
		Port:                "8082",
		SessionServiceURL:   "http://localhost:8081",
		OrdersServiceURL:    "http://localhost:8083",
		InventoryServiceURL: "http://localhost:8084",
		InvoiceServiceURL:   "http://localhost:8085",
		DataServiceURL:      bootstrapConfig.DataServiceURL, // Use bootstrap value
	}

	// Populate config from settings
	populateConfigFromSettings(config, settings, logger)

	return config, nil
}

// getSettingsFromDataService calls the data service API to get settings
func getSettingsFromDataService(serviceName string, logger *logrus.Logger) ([]Setting, error) {
	dataServiceURL := "http://icecream_data_service:8086" // Hardcoded for bootstrap - Docker service name

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request body
	requestBody := map[string]string{
		"service": serviceName,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make request to data service
	resp, err := client.Post(
		dataServiceURL+"/api/v1/data/settings/by-service",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data service returned status %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		Success bool      `json:"success"`
		Data    []Setting `json:"data"`
		Message string    `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("data service error: %s", response.Message)
	}

	return response.Data, nil
}

// Setting represents a setting from the data service
type Setting struct {
	Service     string `json:"service"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// populateConfigFromSettings populates the config struct from settings
func populateConfigFromSettings(config *Config, settings []Setting, logger *logrus.Logger) {
	for _, setting := range settings {
		switch setting.Key {
		case "SESSION_SERVICE_URL":
			config.SessionServiceURL = setting.Value
		case "ORDERS_SERVICE_URL":
			config.OrdersServiceURL = setting.Value
		case "INVENTORY_SERVICE_URL":
			config.InventoryServiceURL = setting.Value
		case "INVOICE_SERVICE_URL":
			config.InvoiceServiceURL = setting.Value
		case "DATA_SERVICE_URL":
			config.DataServiceURL = setting.Value
		case "GATEWAY_PORT":
			config.Port = setting.Value
		}
	}

	logger.WithFields(logrus.Fields{
		"session_service":   config.SessionServiceURL,
		"orders_service":    config.OrdersServiceURL,
		"inventory_service": config.InventoryServiceURL,
		"invoice_service":   config.InvoiceServiceURL,
		"data_service":      config.DataServiceURL,
		"port":              config.Port,
	}).Info("Configuration loaded from data service")
}
