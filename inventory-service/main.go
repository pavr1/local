package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service/config"

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

	// Create main HTTP handler with all entity handlers
	mainHandler := NewMainHttpHandler(db, logger)

	// Setup HTTP router
	router := setupRouter(mainHandler, logger)

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
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Gracefully shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited gracefully")
}

// setupLogger configures the logrus logger
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

// connectToDatabase establishes connection to PostgreSQL database with retry logic
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
			logger.WithError(err).Warnf("Failed to open database connection, attempt %d/%d", i+1, maxRetries)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			logger.WithError(err).Warnf("Failed to ping database, attempt %d/%d", i+1, maxRetries)
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
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.WithFields(logrus.Fields{
		"host":     cfg.DBHost,
		"port":     cfg.DBPort,
		"database": cfg.DBName,
	}).Info("Successfully connected to database")

	return db, nil
}

// setupRouter configures the HTTP routes
func setupRouter(mainHandler *MainHttpHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// API versioning
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	v1.HandleFunc("/inventory/p/health", func(w http.ResponseWriter, r *http.Request) {
		healthData := mainHandler.HealthCheck()
		w.Header().Set("Content-Type", "application/json")

		// Check if service is unhealthy and set appropriate HTTP status
		status := http.StatusOK
		if healthData["status"] == "unhealthy" {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)

		// Use json.Marshal for proper JSON encoding
		jsonData, _ := json.Marshal(healthData)
		w.Write(jsonData)
	}).Methods("GET")

	// Inventory module endpoints
	inventoryRouter := v1.PathPrefix("/inventory").Subrouter()

	// Suppliers endpoints under inventory
	suppliersRouter := inventoryRouter.PathPrefix("/suppliers").Subrouter()

	// GET /api/v1/inventory/suppliers - List all suppliers
	suppliersRouter.HandleFunc("", mainHandler.GetSuppliersHandler().ListSuppliers).Methods("GET")

	// POST /api/v1/inventory/suppliers - Create new supplier
	suppliersRouter.HandleFunc("", mainHandler.GetSuppliersHandler().CreateSupplier).Methods("POST")

	// GET /api/v1/inventory/suppliers/{id} - Get supplier by ID
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().GetSupplier).Methods("GET")

	// PUT /api/v1/inventory/suppliers/{id} - Update supplier
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().UpdateSupplier).Methods("PUT")

	// DELETE /api/v1/inventory/suppliers/{id} - Delete supplier
	suppliersRouter.HandleFunc("/{id}", mainHandler.GetSuppliersHandler().DeleteSupplier).Methods("DELETE")

	// Ingredient Categories endpoints under inventory
	categoriesRouter := inventoryRouter.PathPrefix("/ingredient-categories").Subrouter()

	// GET /api/v1/inventory/ingredient-categories - List all ingredient categories
	categoriesRouter.HandleFunc("", mainHandler.GetIngredientCategoriesHandler().ListIngredientCategories).Methods("GET")

	// POST /api/v1/inventory/ingredient-categories - Create new ingredient category
	categoriesRouter.HandleFunc("", mainHandler.GetIngredientCategoriesHandler().CreateIngredientCategory).Methods("POST")

	// GET /api/v1/inventory/ingredient-categories/{id} - Get ingredient category by ID
	categoriesRouter.HandleFunc("/{id}", mainHandler.GetIngredientCategoriesHandler().GetIngredientCategory).Methods("GET")

	// PUT /api/v1/inventory/ingredient-categories/{id} - Update ingredient category
	categoriesRouter.HandleFunc("/{id}", mainHandler.GetIngredientCategoriesHandler().UpdateIngredientCategory).Methods("PUT")

	// DELETE /api/v1/inventory/ingredient-categories/{id} - Delete ingredient category
	categoriesRouter.HandleFunc("/{id}", mainHandler.GetIngredientCategoriesHandler().DeleteIngredientCategory).Methods("DELETE")

	// Ingredients endpoints under inventory
	ingredientsRouter := inventoryRouter.PathPrefix("/ingredients").Subrouter()

	// GET /api/v1/inventory/ingredients - List all ingredients
	ingredientsRouter.HandleFunc("", mainHandler.GetIngredientsHandler().ListIngredients).Methods("GET")

	// POST /api/v1/inventory/ingredients - Create new ingredient
	ingredientsRouter.HandleFunc("", mainHandler.GetIngredientsHandler().CreateIngredient).Methods("POST")

	// GET /api/v1/inventory/ingredients/{id} - Get ingredient by ID
	ingredientsRouter.HandleFunc("/{id}", mainHandler.GetIngredientsHandler().GetIngredient).Methods("GET")

	// PUT /api/v1/inventory/ingredients/{id} - Update ingredient
	ingredientsRouter.HandleFunc("/{id}", mainHandler.GetIngredientsHandler().UpdateIngredient).Methods("PUT")

	// DELETE /api/v1/inventory/ingredients/{id} - Delete ingredient
	ingredientsRouter.HandleFunc("/{id}", mainHandler.GetIngredientsHandler().DeleteIngredient).Methods("DELETE")

	// Existences endpoints under inventory
	existencesRouter := inventoryRouter.PathPrefix("/existences").Subrouter()

	// GET /api/v1/inventory/existences - List all existences
	existencesRouter.HandleFunc("", mainHandler.GetExistencesHandler().ListExistences).Methods("GET")

	// POST /api/v1/inventory/existences - Create new existence
	existencesRouter.HandleFunc("", mainHandler.GetExistencesHandler().CreateExistence).Methods("POST")

	// GET /api/v1/inventory/existences/{id} - Get existence by ID
	existencesRouter.HandleFunc("/{id}", mainHandler.GetExistencesHandler().GetExistence).Methods("GET")

	// PUT /api/v1/inventory/existences/{id} - Update existence
	existencesRouter.HandleFunc("/{id}", mainHandler.GetExistencesHandler().UpdateExistence).Methods("PUT")

	// DELETE /api/v1/inventory/existences/{id} - Delete existence
	existencesRouter.HandleFunc("/{id}", mainHandler.GetExistencesHandler().DeleteExistence).Methods("DELETE")

	// Runout Ingredients endpoints under inventory
	runoutIngredientsRouter := inventoryRouter.PathPrefix("/runout-ingredients").Subrouter()

	// GET /api/v1/inventory/runout-ingredients - List all runout ingredients
	runoutIngredientsRouter.HandleFunc("", mainHandler.GetRunoutIngredientsHandler().ListRunoutIngredients).Methods("GET")

	// POST /api/v1/inventory/runout-ingredients - Create new runout ingredient
	runoutIngredientsRouter.HandleFunc("", mainHandler.GetRunoutIngredientsHandler().CreateRunoutIngredient).Methods("POST")

	// GET /api/v1/inventory/runout-ingredients/{id} - Get runout ingredient by ID
	runoutIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRunoutIngredientsHandler().GetRunoutIngredient).Methods("GET")

	// PUT /api/v1/inventory/runout-ingredients/{id} - Update runout ingredient
	runoutIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRunoutIngredientsHandler().UpdateRunoutIngredient).Methods("PUT")

	// DELETE /api/v1/inventory/runout-ingredients/{id} - Delete runout ingredient
	runoutIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRunoutIngredientsHandler().DeleteRunoutIngredient).Methods("DELETE")

	// Recipe Categories endpoints under inventory
	recipeCategoriesRouter := inventoryRouter.PathPrefix("/recipe-categories").Subrouter()

	// GET /api/v1/inventory/recipe-categories - List all recipe categories
	recipeCategoriesRouter.HandleFunc("", mainHandler.GetRecipeCategoriesHandler().ListRecipeCategories).Methods("GET")

	// POST /api/v1/inventory/recipe-categories - Create new recipe category
	recipeCategoriesRouter.HandleFunc("", mainHandler.GetRecipeCategoriesHandler().CreateRecipeCategory).Methods("POST")

	// GET /api/v1/inventory/recipe-categories/{id} - Get recipe category by ID
	recipeCategoriesRouter.HandleFunc("/{id}", mainHandler.GetRecipeCategoriesHandler().GetRecipeCategory).Methods("GET")

	// PUT /api/v1/inventory/recipe-categories/{id} - Update recipe category
	recipeCategoriesRouter.HandleFunc("/{id}", mainHandler.GetRecipeCategoriesHandler().UpdateRecipeCategory).Methods("PUT")

	// DELETE /api/v1/inventory/recipe-categories/{id} - Delete recipe category
	recipeCategoriesRouter.HandleFunc("/{id}", mainHandler.GetRecipeCategoriesHandler().DeleteRecipeCategory).Methods("DELETE")

	// Recipes endpoints under inventory
	recipesRouter := inventoryRouter.PathPrefix("/recipes").Subrouter()

	// GET /api/v1/inventory/recipes - List all recipes
	recipesRouter.HandleFunc("", mainHandler.GetRecipesHandler().ListRecipes).Methods("GET")

	// POST /api/v1/inventory/recipes - Create new recipe
	recipesRouter.HandleFunc("", mainHandler.GetRecipesHandler().CreateRecipe).Methods("POST")

	// GET /api/v1/inventory/recipes/{id} - Get recipe by ID
	recipesRouter.HandleFunc("/{id}", mainHandler.GetRecipesHandler().GetRecipe).Methods("GET")

	// PUT /api/v1/inventory/recipes/{id} - Update recipe
	recipesRouter.HandleFunc("/{id}", mainHandler.GetRecipesHandler().UpdateRecipe).Methods("PUT")

	// DELETE /api/v1/inventory/recipes/{id} - Delete recipe
	recipesRouter.HandleFunc("/{id}", mainHandler.GetRecipesHandler().DeleteRecipe).Methods("DELETE")

	// Recipe Ingredients endpoints under inventory
	recipeIngredientsRouter := inventoryRouter.PathPrefix("/recipe-ingredients").Subrouter()

	// GET /api/v1/inventory/recipe-ingredients - List all recipe ingredients
	recipeIngredientsRouter.HandleFunc("", mainHandler.GetRecipeIngredientsHandler().ListRecipeIngredients).Methods("GET")

	// POST /api/v1/inventory/recipe-ingredients - Create new recipe ingredient
	recipeIngredientsRouter.HandleFunc("", mainHandler.GetRecipeIngredientsHandler().CreateRecipeIngredient).Methods("POST")

	// GET /api/v1/inventory/recipe-ingredients/{id} - Get recipe ingredient by ID
	recipeIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRecipeIngredientsHandler().GetRecipeIngredient).Methods("GET")

	// PUT /api/v1/inventory/recipe-ingredients/{id} - Update recipe ingredient
	recipeIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRecipeIngredientsHandler().UpdateRecipeIngredient).Methods("PUT")

	// DELETE /api/v1/inventory/recipe-ingredients/{id} - Delete recipe ingredient
	recipeIngredientsRouter.HandleFunc("/{id}", mainHandler.GetRecipeIngredientsHandler().DeleteRecipeIngredient).Methods("DELETE")

	// Logging middleware
	router.Use(loggingMiddleware(logger))

	logger.Info("HTTP routes configured successfully")
	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Log the request
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrappedWriter.statusCode,
				"duration":   duration.String(),
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
			}).Info("HTTP request processed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
