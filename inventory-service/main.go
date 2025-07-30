package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service/config"
	"inventory-service/handlers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	logger.Info("Starting Ice Cream Store Inventory Service")

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create handlers
	inventoryHandler := handlers.NewInventoryHandler(db, cfg, logger)

	// Setup HTTP router
	router := setupRouter(inventoryHandler, logger)

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
		}).Info("Inventory service starting on")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down inventory service...")

	// Graceful shutdown
	if err := server.Close(); err != nil {
		logger.WithError(err).Error("Error during server shutdown")
	}

	logger.Info("Inventory service shutdown complete")
}

// setupLogger configures the logrus logger
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return logger
}

// connectToDatabase establishes database connection
func connectToDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	var db *sql.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.WithError(err).Warnf("Database connection attempt %d failed", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			logger.WithError(err).Warnf("Database ping attempt %d failed", i+1)
			db.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Connection successful
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	logger.Info("Database connection established successfully")
	return db, nil
}

// setupRouter configures the HTTP routes
func setupRouter(inventoryHandler *handlers.InventoryHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Public routes (health check)
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	publicRouter.HandleFunc("/inventory/health", inventoryHandler.HealthCheck).Methods("GET")

	// Protected routes (all inventory endpoints - gateway handles auth)
	protectedRouter := router.PathPrefix("/api/v1").Subrouter()

	// Suppliers endpoints
	protectedRouter.HandleFunc("/suppliers", inventoryHandler.CreateSupplier).Methods("POST")
	protectedRouter.HandleFunc("/suppliers", inventoryHandler.ListSuppliers).Methods("GET")
	protectedRouter.HandleFunc("/suppliers/{id}", inventoryHandler.GetSupplier).Methods("GET")
	protectedRouter.HandleFunc("/suppliers/{id}", inventoryHandler.UpdateSupplier).Methods("PUT")
	protectedRouter.HandleFunc("/suppliers/{id}", inventoryHandler.DeleteSupplier).Methods("DELETE")

	// Ingredients endpoints
	protectedRouter.HandleFunc("/ingredients", inventoryHandler.CreateIngredient).Methods("POST")
	protectedRouter.HandleFunc("/ingredients", inventoryHandler.ListIngredients).Methods("GET")
	protectedRouter.HandleFunc("/ingredients/{id}", inventoryHandler.GetIngredient).Methods("GET")
	protectedRouter.HandleFunc("/ingredients/{id}", inventoryHandler.UpdateIngredient).Methods("PUT")
	protectedRouter.HandleFunc("/ingredients/{id}", inventoryHandler.DeleteIngredient).Methods("DELETE")

	// Existences endpoints
	protectedRouter.HandleFunc("/existences", inventoryHandler.CreateExistence).Methods("POST")
	protectedRouter.HandleFunc("/existences", inventoryHandler.ListExistences).Methods("GET")
	protectedRouter.HandleFunc("/existences/{id}", inventoryHandler.GetExistence).Methods("GET")
	protectedRouter.HandleFunc("/existences/{id}", inventoryHandler.UpdateExistence).Methods("PUT")
	protectedRouter.HandleFunc("/existences/{id}", inventoryHandler.DeleteExistence).Methods("DELETE")
	protectedRouter.HandleFunc("/existences/low-stock", inventoryHandler.ListLowStock).Methods("GET")
	protectedRouter.HandleFunc("/existences/expiring-soon", inventoryHandler.ListExpiringSoon).Methods("GET")

	// Runout Reports endpoints
	protectedRouter.HandleFunc("/runout-reports", inventoryHandler.CreateRunoutReport).Methods("POST")
	protectedRouter.HandleFunc("/runout-reports", inventoryHandler.ListRunoutReports).Methods("GET")
	protectedRouter.HandleFunc("/runout-reports/{id}", inventoryHandler.GetRunoutReport).Methods("GET")

	// Recipe Categories endpoints
	protectedRouter.HandleFunc("/recipe-categories", inventoryHandler.CreateRecipeCategory).Methods("POST")
	protectedRouter.HandleFunc("/recipe-categories", inventoryHandler.ListRecipeCategories).Methods("GET")
	protectedRouter.HandleFunc("/recipe-categories/{id}", inventoryHandler.GetRecipeCategory).Methods("GET")
	protectedRouter.HandleFunc("/recipe-categories/{id}", inventoryHandler.UpdateRecipeCategory).Methods("PUT")
	protectedRouter.HandleFunc("/recipe-categories/{id}", inventoryHandler.DeleteRecipeCategory).Methods("DELETE")

	// Recipes endpoints
	protectedRouter.HandleFunc("/recipes", inventoryHandler.CreateRecipe).Methods("POST")
	protectedRouter.HandleFunc("/recipes", inventoryHandler.ListRecipes).Methods("GET")
	protectedRouter.HandleFunc("/recipes/{id}", inventoryHandler.GetRecipe).Methods("GET")
	protectedRouter.HandleFunc("/recipes/{id}", inventoryHandler.UpdateRecipe).Methods("PUT")
	protectedRouter.HandleFunc("/recipes/{id}", inventoryHandler.DeleteRecipe).Methods("DELETE")

	// Recipe Ingredients endpoints
	protectedRouter.HandleFunc("/recipe-ingredients", inventoryHandler.CreateRecipeIngredient).Methods("POST")
	protectedRouter.HandleFunc("/recipe-ingredients", inventoryHandler.ListRecipeIngredients).Methods("GET")
	protectedRouter.HandleFunc("/recipe-ingredients/{id}", inventoryHandler.GetRecipeIngredient).Methods("GET")
	protectedRouter.HandleFunc("/recipe-ingredients/{id}", inventoryHandler.UpdateRecipeIngredient).Methods("PUT")
	protectedRouter.HandleFunc("/recipe-ingredients/{id}", inventoryHandler.DeleteRecipeIngredient).Methods("DELETE")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"service": "ice-cream-inventory-service",
			"version": "1.0.0",
			"status": "healthy",
			"timestamp": "%s",
			"endpoints": {
				"health": "/api/v1/inventory/health",
				"suppliers": "/api/v1/suppliers",
				"ingredients": "/api/v1/ingredients",
				"existences": "/api/v1/existences",
				"runout_reports": "/api/v1/runout-reports",
				"recipe_categories": "/api/v1/recipe-categories",
				"recipes": "/api/v1/recipes",
				"recipe_ingredients": "/api/v1/recipe-ingredients"
			}
		}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	return router
}
