package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"data-service/pkg/database"
	"data-service/pkg/images"
	"data-service/pkg/settings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a logger with custom configuration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     true,
	})

	// Create database configuration
	config := &database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres123"),
		DBName:   getEnv("DB_NAME", "icecream_store"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Connection pool settings
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		// Timeout settings
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		// Retry settings
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}

	// Create database handler
	db := database.New(config, logger)

	// Connect to database
	fmt.Println("🍦 Connecting to Ice Cream Store Data Service...")
	if err := db.Connect(); err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Perform initial health check
	if err := db.HealthCheck(); err != nil {
		logger.WithError(err).Fatal("Initial database health check failed")
	}

	fmt.Println("✅ Database connection established successfully")

	// Create settings service
	settingsDBHandler := settings.NewSettingsDBHandler(db.GetDB(), logger)
	settingsService := settings.NewSettingsService(settingsDBHandler, logger)

	// Create settings handler
	settingsHandler, err := settings.NewSettingsHandler(settingsService, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize settings handler")
	}

	// Create image handler
	imageHandler := images.NewImageHandler("/app")

	// Setup HTTP server
	router := setupRouter(db, logger, settingsHandler, imageHandler)

	server := &http.Server{
		Addr:         "0.0.0.0:8086", // Data service port - bind to all interfaces
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", "8086").Info("Starting Data Service HTTP server")
		logger.WithField("port", "8086").Info("🚀 Data Service HTTP server starting on :8086")
		logger.WithField("port", "8086").Info("📡 Health endpoint available at: http://localhost:8086/api/v1/data/p/health")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Data Service...")

	// Gracefully shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Data Service exited gracefully")
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// setupRouter configures the HTTP routes
func setupRouter(db database.DatabaseHandler, logger *logrus.Logger, settingsHandler *settings.SettingsHandler, imageHandler *images.ImageHandler) *mux.Router {
	router := mux.NewRouter()

	// Root endpoint to test if router is working
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Data service is running"}`))
	}).Methods("GET")

	// Public health check endpoint (no authentication required)
	router.HandleFunc("/api/v1/data/p/health", func(w http.ResponseWriter, r *http.Request) {
		healthCheck(w, r, db, logger)
	}).Methods("GET")

	// Test endpoint to verify router is working
	router.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"API v1 router is working"}`))
	}).Methods("GET")

	// Stats endpoint (optional, for monitoring)
	router.HandleFunc("/api/v1/data/p/stats", func(w http.ResponseWriter, r *http.Request) {
		statsEndpoint(w, r, db, logger)
	}).Methods("GET")

	// Settings endpoints
	router.HandleFunc("/api/v1/data/settings/all", settingsHandler.GetAllSettings).Methods("GET")
	router.HandleFunc("/api/v1/data/settings/by-service", settingsHandler.GetByService).Methods("POST")
	router.HandleFunc("/api/v1/data/settings/by-key", settingsHandler.GetByKey).Methods("POST")
	router.HandleFunc("/api/v1/data/settings/reload", settingsHandler.Reload).Methods("POST")
	router.HandleFunc("/api/v1/data/settings/update-setting", settingsHandler.UpdateSetting).Methods("POST")

	// Image endpoints
	router.HandleFunc("/api/v1/data/images/{service}/{filename}", func(w http.ResponseWriter, r *http.Request) {
		serveImage(w, r, imageHandler, logger)
	}).Methods("GET")
	router.HandleFunc("/api/v1/data/images/{service}", func(w http.ResponseWriter, r *http.Request) {
		storeImage(w, r, imageHandler, logger)
	}).Methods("POST")
	router.HandleFunc("/api/v1/data/images/{service}/{filename}", func(w http.ResponseWriter, r *http.Request) {
		deleteImage(w, r, imageHandler, logger)
	}).Methods("DELETE")

	return router
}

// healthCheck handles the health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request, db database.DatabaseHandler, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"endpoint": "/api/v1/data/p/health",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Health check requested")

	response := map[string]interface{}{
		"service":   "data-service",
		"timestamp": time.Now(),
	}

	// Perform database health check
	if err := db.HealthCheck(); err != nil {
		logger.WithError(err).Error("Database health check failed")
		response["status"] = "unhealthy"
		response["message"] = "Database connection failed"
		response["error"] = err.Error()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Health check passed
	response["status"] = "healthy"
	response["message"] = "Database is operational"
	response["database"] = map[string]interface{}{
		"host":   "localhost",
		"port":   5432,
		"dbname": "icecream_store",
		"stats":  db.GetStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// statsEndpoint provides database connection statistics
func statsEndpoint(w http.ResponseWriter, _ *http.Request, db database.DatabaseHandler, _ *logrus.Logger) {
	stats := db.GetStats()

	response := map[string]interface{}{
		"service":   "data-service",
		"timestamp": time.Now(),
		"database_stats": map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration":    stats.WaitDuration.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// serveImage serves images from the file system
func serveImage(w http.ResponseWriter, r *http.Request, imageHandler *images.ImageHandler, logger *logrus.Logger) {
	vars := mux.Vars(r)
	service := vars["service"]
	filename := vars["filename"]

	if service == "" || filename == "" {
		logger.WithFields(logrus.Fields{
			"service":  service,
			"filename": filename,
		}).Error("Invalid image path")

		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Retrieve image data
	imageData, err := imageHandler.RetrieveImage(service, filename)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"service":  service,
			"filename": filename,
		}).Error("Failed to retrieve image")
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Determine content type based on file extension
	contentType := "image/jpeg" // default
	switch {
	case filename[len(filename)-4:] == ".png":
		contentType = "image/png"
	case filename[len(filename)-4:] == ".gif":
		contentType = "image/gif"
	case filename[len(filename)-5:] == ".webp":
		contentType = "image/webp"
	}

	// Set headers and serve image
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(imageData)
}

// storeImage stores an image in the file system
func storeImage(w http.ResponseWriter, r *http.Request, imageHandler *images.ImageHandler, logger *logrus.Logger) {
	vars := mux.Vars(r)
	service := vars["service"]

	if service == "" {
		logger.Error("Service parameter is required")
		http.Error(w, "Service parameter is required", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		logger.WithError(err).Error("Failed to parse form")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		logger.WithError(err).Error("No image file provided")
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	imageData := make([]byte, header.Size)
	_, err = file.Read(imageData)
	if err != nil {
		logger.WithError(err).Error("Failed to read uploaded image")
		http.Error(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	// Store the image
	err = imageHandler.AddImage(service, header.Filename, imageData)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"service":  service,
			"filename": header.Filename,
		}).Error("Failed to store image")
		http.Error(w, "Failed to store image", http.StatusInternalServerError)
		return
	}

	// Generate the image URL
	imageURL := imageHandler.GetImageURL(service, header.Filename)

	// Return success response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Image stored successfully",
		"filename":  header.Filename,
		"image_url": imageURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// deleteImage deletes an image from the file system
func deleteImage(w http.ResponseWriter, r *http.Request, imageHandler *images.ImageHandler, logger *logrus.Logger) {
	vars := mux.Vars(r)
	service := vars["service"]
	filename := vars["filename"]

	if service == "" || filename == "" {
		logger.WithFields(logrus.Fields{
			"service":  service,
			"filename": filename,
		}).Error("Invalid image path for deletion")

		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Delete the image
	err := imageHandler.DeleteImage(service, filename)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"service":  service,
			"filename": filename,
		}).Error("Failed to delete image")
		http.Error(w, "Failed to delete image", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success":  true,
		"message":  "Image deleted successfully",
		"filename": filename,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
