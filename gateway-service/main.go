package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
)

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
type ServiceConfig struct {
	SessionServiceURL   string
	OrdersServiceURL    string
	InventoryServiceURL string
}

func main() {
	// Service configuration from environment variables
	config := &ServiceConfig{
		SessionServiceURL:   getEnv("SESSION_SERVICE_URL", "http://localhost:8081"),
		OrdersServiceURL:    getEnv("ORDERS_SERVICE_URL", "http://localhost:8083"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8084"),
	}

	log.Printf("Gateway configured with Session Service: %s", config.SessionServiceURL)
	log.Printf("Gateway configured with Orders Service: %s", config.OrdersServiceURL)
	log.Printf("Gateway configured with Inventory Service: %s", config.InventoryServiceURL)

	// Initialize session management
	sessionManager := NewSessionManager(config.SessionServiceURL)
	sessionMiddleware := NewSessionMiddleware(sessionManager)

	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// ==== GATEWAY ENDPOINTS ====

	// Gateway health check endpoint
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// ==== AUTH ENDPOINTS ====

	// Public auth endpoints (no session validation required)
	authPublicRouter := api.PathPrefix("/v1/auth").Subrouter()
	authPublicRouter.HandleFunc("/login", sessionMiddleware.SessionAwareLoginHandler(config.SessionServiceURL)).Methods("POST")
	authPublicRouter.HandleFunc("/health", createProxyHandler(config.SessionServiceURL, "/api/v1/auth/health")).Methods("GET")

	// Protected auth endpoints (require session validation)
	authProtectedRouter := api.PathPrefix("/v1/auth").Subrouter()
	authProtectedRouter.Use(sessionMiddleware.ValidateSession)
	authProtectedRouter.HandleFunc("/logout", sessionMiddleware.SessionAwareLogoutHandler(config.SessionServiceURL)).Methods("POST")
	authProtectedRouter.HandleFunc("/refresh", createProxyHandler(config.SessionServiceURL, "/api/v1/auth/refresh")).Methods("POST")
	authProtectedRouter.HandleFunc("/validate", createProxyHandler(config.SessionServiceURL, "/api/v1/auth/validate")).Methods("GET")
	authProtectedRouter.HandleFunc("/profile", createProxyHandler(config.SessionServiceURL, "/api/v1/auth/profile")).Methods("GET")
	authProtectedRouter.HandleFunc("/token-info", createProxyHandler(config.SessionServiceURL, "/api/v1/auth/token-info")).Methods("GET")

	// ==== PROTECTED BUSINESS SERVICE ROUTES ====

	// Orders service - all routes require session validation
	ordersRouter := api.PathPrefix("/v1/orders").Subrouter()
	ordersRouter.Use(sessionMiddleware.ValidateSession)
	ordersRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.OrdersServiceURL, "/api/v1/orders"))

	// Inventory service - all routes require session validation
	inventoryRouter := api.PathPrefix("/v1/inventory").Subrouter()
	inventoryRouter.Use(sessionMiddleware.ValidateSession)
	inventoryRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.InventoryServiceURL, "/api/v1/inventory"))

	// ==== SESSION MANAGEMENT ENDPOINTS ====

	// Direct access to session management APIs (protected)
	sessionRouter := api.PathPrefix("/v1/sessions").Subrouter()
	sessionRouter.Use(sessionMiddleware.ValidateSession)
	sessionRouter.PathPrefix("").HandlerFunc(createProxyHandler(config.SessionServiceURL, "/api/v1/sessions"))

	// ==== DEMO ENDPOINTS ====

	// Example endpoints (keeping for demo)
	api.HandleFunc("/hello", helloHandler).Methods("GET")
	api.HandleFunc("/hello", createHelloHandler).Methods("POST")

	// Apply CORS middleware to main router - gateway is single source of CORS
	r.Use(corsMiddleware)

	// Add explicit OPTIONS handling for CORS preflight
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers are already set by corsMiddleware
		w.WriteHeader(http.StatusOK)
	})

	// UI is now served by its own service on port 3000
	// Static file serving removed - UI runs independently

	fmt.Println("🚀 Gateway Service with Session Management starting on http://localhost:8082")
	fmt.Println("📡 API available at http://localhost:8082/api")
	fmt.Println("")
	fmt.Println("🔐 AUTH ENDPOINTS:")
	fmt.Println("   📂 Public:")
	fmt.Printf("      POST /api/v1/auth/login    → %s (+ session creation)\n", config.SessionServiceURL)
	fmt.Printf("      GET  /api/v1/auth/health   → %s\n", config.SessionServiceURL)
	fmt.Println("   🔒 Protected (require valid session):")
	fmt.Printf("      POST /api/v1/auth/logout   → %s (+ session revocation)\n", config.SessionServiceURL)
	fmt.Printf("      POST /api/v1/auth/refresh  → %s\n", config.SessionServiceURL)
	fmt.Printf("      GET  /api/v1/auth/validate → %s\n", config.SessionServiceURL)
	fmt.Printf("      GET  /api/v1/auth/profile  → %s\n", config.SessionServiceURL)
	fmt.Printf("      GET  /api/v1/auth/token-info → %s\n", config.SessionServiceURL)
	fmt.Println("")
	fmt.Println("🛒 BUSINESS SERVICE ENDPOINTS:")
	fmt.Printf("   🔒 /api/v1/orders/*          → %s (session validated)\n", config.OrdersServiceURL)
	fmt.Printf("   🔒 /api/v1/inventory/*       → %s (session validated)\n", config.InventoryServiceURL)
	fmt.Printf("       ├─ /suppliers/*          → Suppliers management\n")
	fmt.Printf("       ├─ /ingredients/*        → [Future] Ingredients management\n")
	fmt.Printf("       └─ /existences/*         → [Future] Stock management\n")
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

// createProxyHandler creates a reverse proxy handler for a specific service
func createProxyHandler(targetURL, stripPrefix string) http.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the proxy to handle errors and modify requests
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s %s: %v", r.Method, r.URL.Path, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "Service unavailable",
			"message":   "The session service is currently unavailable",
			"timestamp": time.Now(),
			"service":   "session-service",
		})
	}

	// Custom director to modify the request before forwarding
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Log the proxy request (only for important requests)
		if req.URL.Path != "/api/v1/auth/health" {
			log.Printf("Proxying %s %s to %s%s", req.Method, req.URL.Path, target.String(), req.URL.Path)
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check if services are healthy
	sessionHealthy := checkServiceHealth("http://localhost:8081/api/v1/auth/health")
	ordersHealthy := checkServiceHealth("http://localhost:8083/api/v1/orders/health")
	inventoryHealthy := checkServiceHealth("http://localhost:8084/api/v1/inventory/health")

	// Check session management health
	sessionMgmtHealthy := checkServiceHealth("http://localhost:8081/api/v1/sessions/health")

	status := "healthy"
	if !sessionHealthy || !ordersHealthy || !inventoryHealthy || !sessionMgmtHealthy {
		status = "degraded"
	}

	response := map[string]interface{}{
		"status":             status,
		"version":            "1.0.0",
		"time":               time.Now(),
		"gateway":            "operational",
		"session_management": "enabled",
		"services": map[string]string{
			"session-service": func() string {
				if sessionHealthy && sessionMgmtHealthy {
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
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if !sessionHealthy || !ordersHealthy || !inventoryHealthy || !sessionMgmtHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

// checkServiceHealth checks if a service is responding to health checks
func checkServiceHealth(healthURL string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message:   "Hello from the Go server!",
		Timestamp: time.Now(),
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createHelloHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message:   "Hello POST request received!",
		Timestamp: time.Now(),
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
