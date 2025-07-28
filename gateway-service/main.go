package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

// Service configuration
type ServiceConfig struct {
	AuthServiceURL   string
	OrdersServiceURL string
}

func main() {
	// Service configuration
	config := &ServiceConfig{
		AuthServiceURL:   "http://localhost:8081",
		OrdersServiceURL: "http://localhost:8083",
	}

	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Gateway health check endpoint
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// Auth service proxy - route all /api/v1/auth/* to auth service
	authProxy := api.PathPrefix("/v1/auth").Subrouter()
	authProxy.PathPrefix("").HandlerFunc(createProxyHandler(config.AuthServiceURL, "/api/v1/auth"))

	// Orders service proxy - route all /api/v1/orders/* to orders service
	ordersProxy := api.PathPrefix("/v1/orders").Subrouter()
	ordersProxy.PathPrefix("").HandlerFunc(createProxyHandler(config.OrdersServiceURL, "/api/v1/orders"))

	// Example endpoints (keeping for demo)
	api.HandleFunc("/hello", helloHandler).Methods("GET")
	api.HandleFunc("/hello", createHelloHandler).Methods("POST")

	// CORS middleware
	//r.Use(corsMiddleware)

	// Static file serving (for client build)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/build/")))

	fmt.Println("🚀 Gateway Service starting on http://localhost:8082")
	fmt.Println("📡 API available at http://localhost:8082/api")
	fmt.Println("🔐 Auth endpoints: http://localhost:8082/api/v1/auth/*")
	fmt.Printf("   → Proxying to: %s\n", config.AuthServiceURL)
	fmt.Println("🛒 Orders endpoints: http://localhost:8082/api/v1/orders/*")
	fmt.Printf("   → Proxying to: %s\n", config.OrdersServiceURL)

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
			"message":   "The authentication service is currently unavailable",
			"timestamp": time.Now(),
			"service":   "auth-service",
		})
	}

	// Custom director to modify the request before forwarding
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Log the proxy request
		log.Printf("Proxying %s %s to %s%s", req.Method, req.URL.Path, target.String(), req.URL.Path)

		// Add any custom headers if needed
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check if services are healthy
	authHealthy := checkServiceHealth("http://localhost:8081/api/v1/auth/health")
	ordersHealthy := checkServiceHealth("http://localhost:8083/api/v1/orders/health")

	status := "healthy"
	if !authHealthy || !ordersHealthy {
		status = "degraded"
	}

	response := map[string]interface{}{
		"status":  status,
		"version": "1.0.0",
		"time":    time.Now(),
		"gateway": "operational",
		"services": map[string]string{
			"auth-service": func() string {
				if authHealthy {
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
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if !authHealthy || !ordersHealthy {
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
	var requestBody map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name := "World"
	if n, ok := requestBody["name"].(string); ok && n != "" {
		name = n
	}

	response := Response{
		Message:   fmt.Sprintf("Hello, %s! Message received.", name),
		Timestamp: time.Now(),
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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
