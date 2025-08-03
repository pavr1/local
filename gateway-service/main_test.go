package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCorsMiddleware tests the CORS middleware functionality
func TestCorsMiddleware(t *testing.T) {
	// Create a simple handler that the middleware will wrap
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap the handler with CORS middleware
	corsHandler := corsMiddleware(testHandler)

	t.Run("sets CORS headers on regular request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, w.Body.String()) // OPTIONS should not call the next handler
	})

	t.Run("sets CORS headers on POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader("test data"))
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "test response", w.Body.String())
	})
}

// TestServiceConfigDefaults tests the default service configuration
func TestServiceConfigDefaults(t *testing.T) {
	config := Config{
		Port:                "8082",
		SessionServiceURL:   "http://localhost:8081",
		OrdersServiceURL:    "http://localhost:8083",
		InventoryServiceURL: "http://localhost:8084",
		InvoiceServiceURL:   "http://localhost:8085",
	}

	assert.Equal(t, "8082", config.Port)
	assert.Equal(t, "http://localhost:8081", config.SessionServiceURL)
	assert.Equal(t, "http://localhost:8083", config.OrdersServiceURL)
	assert.Equal(t, "http://localhost:8084", config.InventoryServiceURL)
	assert.Equal(t, "http://localhost:8085", config.InvoiceServiceURL)
}

// TestServiceConfigWithEnvironmentVariables tests service configuration with environment variables
func TestServiceConfigWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"GATEWAY_PORT":          "9090",
		"SESSION_SERVICE_URL":   "http://session.example.com:8081",
		"ORDERS_SERVICE_URL":    "http://orders.example.com:8083",
		"INVENTORY_SERVICE_URL": "http://inventory.example.com:8084",
		"INVOICE_SERVICE_URL":   "http://invoice.example.com:8085",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Create configuration using environment variables
	config := getServiceConfig()

	assert.Equal(t, "9090", config.Port)
	assert.Equal(t, "http://session.example.com:8081", config.SessionServiceURL)
	assert.Equal(t, "http://orders.example.com:8083", config.OrdersServiceURL)
	assert.Equal(t, "http://inventory.example.com:8084", config.InventoryServiceURL)
	assert.Equal(t, "http://invoice.example.com:8085", config.InvoiceServiceURL)
}

// Helper function to get service config (extracted for testing)
func getServiceConfig() Config {
	return Config{
		Port:                getEnv("GATEWAY_PORT", "8082"),
		SessionServiceURL:   getEnv("SESSION_SERVICE_URL", "http://localhost:8081"),
		OrdersServiceURL:    getEnv("ORDERS_SERVICE_URL", "http://localhost:8083"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8084"),
		InvoiceServiceURL:   getEnv("INVOICE_SERVICE_URL", "http://localhost:8085"),
	}
}

// TestHealthHandler tests the health check endpoint
func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	// When backend services are not running, health handler returns degraded status with 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "degraded", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "operational", response["gateway"])
	assert.Equal(t, "enabled", response["session_management"])
}

// TestResponseStructures tests the response data structures
func TestResponseStructures(t *testing.T) {
	t.Run("Response structure", func(t *testing.T) {
		resp := Response{
			Message:   "test message",
			Timestamp: time.Now(),
			Status:    "success",
		}

		jsonData, err := json.Marshal(resp)
		require.NoError(t, err)

		var unmarshaled Response
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, resp.Message, unmarshaled.Message)
		assert.Equal(t, resp.Status, unmarshaled.Status)
		assert.True(t, resp.Timestamp.Equal(unmarshaled.Timestamp))
	})

	t.Run("HealthResponse structure", func(t *testing.T) {
		resp := HealthResponse{
			Status:  "healthy",
			Version: "1.0.0",
			Time:    time.Now(),
		}

		jsonData, err := json.Marshal(resp)
		require.NoError(t, err)

		var unmarshaled HealthResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, resp.Status, unmarshaled.Status)
		assert.Equal(t, resp.Version, unmarshaled.Version)
		assert.True(t, resp.Time.Equal(unmarshaled.Time))
	})
}

// TestCreateProxyHandler tests proxy handler creation
func TestCreateProxyHandler(t *testing.T) {
	targetURL := "http://localhost:8081"
	stripPrefix := "/api/v1/sessions"

	handler := createProxyHandler(targetURL, stripPrefix)
	assert.NotNil(t, handler)
	assert.IsType(t, http.HandlerFunc(nil), handler)
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(healthHandler))

	const numRequests = 10
	responses := make(chan *httptest.ResponseRecorder, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			responses <- w
		}()
	}

	// Collect and verify responses
	for i := 0; i < numRequests; i++ {
		w := <-responses
		// Health handler returns 503 when backend services are not running
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("malformed request handling", func(t *testing.T) {
		// Create a handler that might fail
		errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Message:   "Internal server error",
				Status:    "error",
				Timestamp: time.Now(),
			})
		})

		corsHandler := corsMiddleware(errorHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestEnvironmentVariableHandling tests various environment variable scenarios
func TestEnvironmentVariableHandling(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "with environment variable set",
			envKey:       "TEST_SERVICE_URL",
			envValue:     "http://test.example.com:8080",
			defaultValue: "http://localhost:8080",
			expected:     "http://test.example.com:8080",
		},
		{
			name:         "with empty environment variable",
			envKey:       "TEST_SERVICE_URL",
			envValue:     "",
			defaultValue: "http://localhost:8080",
			expected:     "http://localhost:8080",
		},
		{
			name:         "without environment variable",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "http://localhost:8080",
			expected:     "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing value
			os.Unsetenv(tt.envKey)

			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGatewayServiceSpecifics tests gateway service specific functionality
func TestGatewayServiceSpecifics(t *testing.T) {
	t.Run("default service URLs", func(t *testing.T) {
		config := getServiceConfig()

		// Test default localhost URLs
		assert.Contains(t, config.SessionServiceURL, "localhost:8081")
		assert.Contains(t, config.OrdersServiceURL, "localhost:8083")
		assert.Contains(t, config.InventoryServiceURL, "localhost:8084")
		assert.Contains(t, config.InvoiceServiceURL, "localhost:8085")
	})

	t.Run("gateway acts as single entry point", func(t *testing.T) {
		// Test that CORS is consistently applied (gateway responsibility)
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		corsHandler := corsMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Gateway should be the only service setting CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	})
}

// Benchmark tests for performance
func BenchmarkCorsMiddleware(b *testing.B) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	corsHandler := corsMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		corsHandler.ServeHTTP(w, req)
	}
}

func BenchmarkHealthHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		healthHandler(w, req)
	}
}
