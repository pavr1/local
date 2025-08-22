package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"inventory-service/config"
	existencesHandlers "inventory-service/entities/existences/handlers"
	ingredientCategoriesHandlers "inventory-service/entities/ingredient_categories/handlers"
	ingredientsHandlers "inventory-service/entities/ingredients/handlers"
	recipeCategoriesHandlers "inventory-service/entities/recipe_categories/handlers"
	recipeIngredientsHandlers "inventory-service/entities/recipe_ingredients/handlers"
	recipesHandlers "inventory-service/entities/recipes/handlers"
	runoutIngredientsHandlers "inventory-service/entities/runout_ingredients/handlers"
	suppliersHandlers "inventory-service/entities/suppliers/handlers"
	"inventory-service/middleware"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MainHttpHandler aggregates all entity HTTP handlers
type MainHttpHandler struct {
	// Database connection
	db     *sql.DB
	logger *logrus.Logger
	config *config.Config



	// Entity handlers
	SuppliersHandler            *suppliersHandlers.HttpHandler
	IngredientCategoriesHandler *ingredientCategoriesHandlers.HttpHandler
	IngredientsHandler          *ingredientsHandlers.HttpHandler
	ExistencesHandler           *existencesHandlers.HttpHandler
	RunoutIngredientsHandler    *runoutIngredientsHandlers.RunoutIngredientHTTPHandler
	RecipeCategoriesHandler     *recipeCategoriesHandlers.RecipeCategoryHTTPHandler
	RecipesHandler              *recipesHandlers.RecipeHTTPHandler
	RecipeIngredientsHandler    *recipeIngredientsHandlers.RecipeIngredientHTTPHandler
}

// NewMainHttpHandler creates a new main HTTP handler with all entity handlers
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger, cfg *config.Config) *MainHttpHandler {
	// Initialize suppliers handlers
	suppliersDBHandler := suppliersHandlers.NewDBHandler(db, logger)
	suppliersHttpHandler := suppliersHandlers.NewHttpHandler(suppliersDBHandler, logger)

	// Initialize ingredient categories handlers
	ingredientCategoriesDBHandler := ingredientCategoriesHandlers.NewDBHandler(db, logger)
	ingredientCategoriesHttpHandler := ingredientCategoriesHandlers.NewHttpHandler(ingredientCategoriesDBHandler, logger)

	// Initialize ingredients handlers
	ingredientsDBHandler := ingredientsHandlers.NewDBHandler(db, logger)
	ingredientsHttpHandler := ingredientsHandlers.NewHttpHandler(ingredientsDBHandler, logger)

	// Initialize existences handlers
	existencesDBHandler := existencesHandlers.NewDBHandler(db, logger)
	existencesHttpHandler := existencesHandlers.NewHttpHandler(existencesDBHandler, logger)

	// Initialize runout ingredients handlers
	runoutIngredientsHttpHandler := runoutIngredientsHandlers.NewRunoutIngredientHTTPHandler(db, logger)

	// Initialize recipe categories handlers
	recipeCategoriesHttpHandler := recipeCategoriesHandlers.NewRecipeCategoryHTTPHandler(db, logger)

	// Initialize recipes handlers
	recipesHttpHandler := recipesHandlers.NewRecipeHTTPHandler(db, logger, cfg)

	// Initialize recipe ingredients handlers
	recipeIngredientsHttpHandler := recipeIngredientsHandlers.NewRecipeIngredientHTTPHandler(db, logger)

	return &MainHttpHandler{
		db:                          db,
		logger:                      logger,
		config:                      cfg,
		SuppliersHandler:            suppliersHttpHandler,
		IngredientCategoriesHandler: ingredientCategoriesHttpHandler,
		IngredientsHandler:          ingredientsHttpHandler,
		ExistencesHandler:           existencesHttpHandler,
		RunoutIngredientsHandler:    runoutIngredientsHttpHandler,
		RecipeCategoriesHandler:     recipeCategoriesHttpHandler,
		RecipesHandler:              recipesHttpHandler,
		RecipeIngredientsHandler:    recipeIngredientsHttpHandler,
	}
}

// GetSuppliersHandler returns the suppliers HTTP handler
func (h *MainHttpHandler) GetSuppliersHandler() *suppliersHandlers.HttpHandler {
	return h.SuppliersHandler
}

// GetIngredientCategoriesHandler returns the ingredient categories HTTP handler
func (h *MainHttpHandler) GetIngredientCategoriesHandler() *ingredientCategoriesHandlers.HttpHandler {
	return h.IngredientCategoriesHandler
}

// GetIngredientsHandler returns the ingredients HTTP handler
func (h *MainHttpHandler) GetIngredientsHandler() *ingredientsHandlers.HttpHandler {
	return h.IngredientsHandler
}

// GetExistencesHandler returns the existences HTTP handler
func (h *MainHttpHandler) GetExistencesHandler() *existencesHandlers.HttpHandler {
	return h.ExistencesHandler
}

// GetRunoutIngredientsHandler returns the runout ingredients HTTP handler
func (h *MainHttpHandler) GetRunoutIngredientsHandler() *runoutIngredientsHandlers.RunoutIngredientHTTPHandler {
	return h.RunoutIngredientsHandler
}

// GetRecipeCategoriesHandler returns the recipe categories HTTP handler
func (h *MainHttpHandler) GetRecipeCategoriesHandler() *recipeCategoriesHandlers.RecipeCategoryHTTPHandler {
	return h.RecipeCategoriesHandler
}

// GetRecipesHandler returns the recipes HTTP handler
func (h *MainHttpHandler) GetRecipesHandler() *recipesHandlers.RecipeHTTPHandler {
	return h.RecipesHandler
}

// GetRecipeIngredientsHandler returns the recipe ingredients HTTP handler
func (h *MainHttpHandler) GetRecipeIngredientsHandler() *recipeIngredientsHandlers.RecipeIngredientHTTPHandler {
	return h.RecipeIngredientsHandler
}

// SetupRoutes sets up all the routes for the service
func (h *MainHttpHandler) SetupRoutes(router *mux.Router) {
	// Public router for endpoints that don't require gateway validation
	publicRouter := router.PathPrefix("/api/v1/inventory").Subrouter()
	publicRouter.HandleFunc("/p/health", h.HealthCheck).Methods("GET")

	// Protected endpoints (require gateway validation)
	protectedRouter := router.PathPrefix("/api/v1/inventory").Subrouter()
	protectedRouter.Use(middleware.NewGatewayMiddleware(h.logger).ValidateGateway)
	protectedRouter.Use(h.loggingMiddleware())

	// Suppliers endpoints under inventory
	suppliersRouter := protectedRouter.PathPrefix("/suppliers").Subrouter()
	suppliersRouter.HandleFunc("", h.GetSuppliersHandler().ListSuppliers).Methods("GET")
	suppliersRouter.HandleFunc("", h.GetSuppliersHandler().CreateSupplier).Methods("POST")
	suppliersRouter.HandleFunc("/{id}", h.GetSuppliersHandler().GetSupplier).Methods("GET")
	suppliersRouter.HandleFunc("/{id}", h.GetSuppliersHandler().UpdateSupplier).Methods("PUT")
	suppliersRouter.HandleFunc("/{id}", h.GetSuppliersHandler().DeleteSupplier).Methods("DELETE")

	// Ingredient Categories endpoints under inventory
	categoriesRouter := protectedRouter.PathPrefix("/ingredient-categories").Subrouter()
	categoriesRouter.HandleFunc("", h.GetIngredientCategoriesHandler().ListIngredientCategories).Methods("GET")
	categoriesRouter.HandleFunc("", h.GetIngredientCategoriesHandler().CreateIngredientCategory).Methods("POST")
	categoriesRouter.HandleFunc("/{id}", h.GetIngredientCategoriesHandler().GetIngredientCategory).Methods("GET")
	categoriesRouter.HandleFunc("/{id}", h.GetIngredientCategoriesHandler().UpdateIngredientCategory).Methods("PUT")
	categoriesRouter.HandleFunc("/{id}", h.GetIngredientCategoriesHandler().DeleteIngredientCategory).Methods("DELETE")

	// Ingredients endpoints under inventory
	ingredientsRouter := protectedRouter.PathPrefix("/ingredients").Subrouter()
	ingredientsRouter.HandleFunc("", h.GetIngredientsHandler().ListIngredients).Methods("GET")
	ingredientsRouter.HandleFunc("", h.GetIngredientsHandler().CreateIngredient).Methods("POST")
	ingredientsRouter.HandleFunc("/{id}", h.GetIngredientsHandler().GetIngredient).Methods("GET")
	ingredientsRouter.HandleFunc("/{id}", h.GetIngredientsHandler().UpdateIngredient).Methods("PUT")
	ingredientsRouter.HandleFunc("/{id}", h.GetIngredientsHandler().DeleteIngredient).Methods("DELETE")

	// Existences endpoints under inventory
	existencesRouter := protectedRouter.PathPrefix("/existences").Subrouter()
	existencesRouter.HandleFunc("", h.GetExistencesHandler().ListExistences).Methods("GET")
	existencesRouter.HandleFunc("", h.GetExistencesHandler().CreateExistence).Methods("POST")
	existencesRouter.HandleFunc("/{id}", h.GetExistencesHandler().GetExistence).Methods("GET")
	existencesRouter.HandleFunc("/ingredient/{ingredientId}/unit-type/{unitType}", h.GetExistencesHandler().GetMostRecentExistenceByIngredientAndUnitType).Methods("GET")
	existencesRouter.HandleFunc("/{id}", h.GetExistencesHandler().UpdateExistence).Methods("PUT")
	existencesRouter.HandleFunc("/{id}", h.GetExistencesHandler().DeleteExistence).Methods("DELETE")

	// Runout Ingredients endpoints under inventory
	runoutIngredientsRouter := protectedRouter.PathPrefix("/runout-ingredients").Subrouter()
	runoutIngredientsRouter.HandleFunc("", h.GetRunoutIngredientsHandler().ListRunoutIngredients).Methods("GET")
	runoutIngredientsRouter.HandleFunc("", h.GetRunoutIngredientsHandler().CreateRunoutIngredient).Methods("POST")
	runoutIngredientsRouter.HandleFunc("/{id}", h.GetRunoutIngredientsHandler().GetRunoutIngredient).Methods("GET")
	runoutIngredientsRouter.HandleFunc("/{id}", h.GetRunoutIngredientsHandler().UpdateRunoutIngredient).Methods("PUT")
	runoutIngredientsRouter.HandleFunc("/{id}", h.GetRunoutIngredientsHandler().DeleteRunoutIngredient).Methods("DELETE")

	// Recipe Categories endpoints under inventory
	recipeCategoriesRouter := protectedRouter.PathPrefix("/recipe-categories").Subrouter()
	recipeCategoriesRouter.HandleFunc("", h.GetRecipeCategoriesHandler().ListRecipeCategories).Methods("GET")
	recipeCategoriesRouter.HandleFunc("", h.GetRecipeCategoriesHandler().CreateRecipeCategory).Methods("POST")
	recipeCategoriesRouter.HandleFunc("/{id}", h.GetRecipeCategoriesHandler().GetRecipeCategory).Methods("GET")
	recipeCategoriesRouter.HandleFunc("/{id}", h.GetRecipeCategoriesHandler().UpdateRecipeCategory).Methods("PUT")
	recipeCategoriesRouter.HandleFunc("/{id}", h.GetRecipeCategoriesHandler().DeleteRecipeCategory).Methods("DELETE")

	// Recipes endpoints under inventory
	recipesRouter := protectedRouter.PathPrefix("/recipes").Subrouter()
	recipesRouter.HandleFunc("", h.GetRecipesHandler().ListRecipes).Methods("GET")
	recipesRouter.HandleFunc("", h.GetRecipesHandler().CreateRecipe).Methods("POST")
	recipesRouter.HandleFunc("/{id}", h.GetRecipesHandler().GetRecipe).Methods("GET")
	recipesRouter.HandleFunc("/{id}", h.GetRecipesHandler().UpdateRecipe).Methods("PUT")
	recipesRouter.HandleFunc("/{id}", h.GetRecipesHandler().DeleteRecipe).Methods("DELETE")

	// Recipe Ingredients endpoints under inventory
	recipeIngredientsRouter := protectedRouter.PathPrefix("/recipe-ingredients").Subrouter()
	recipeIngredientsRouter.HandleFunc("", h.GetRecipeIngredientsHandler().ListRecipeIngredients).Methods("GET")
	recipeIngredientsRouter.HandleFunc("", h.GetRecipeIngredientsHandler().CreateRecipeIngredient).Methods("POST")
	recipeIngredientsRouter.HandleFunc("/{id}", h.GetRecipeIngredientsHandler().GetRecipeIngredient).Methods("GET")
	recipeIngredientsRouter.HandleFunc("/{id}", h.GetRecipeIngredientsHandler().UpdateRecipeIngredient).Methods("PUT")
	recipeIngredientsRouter.HandleFunc("/{id}", h.GetRecipeIngredientsHandler().DeleteRecipeIngredient).Methods("DELETE")

	h.logger.Info("HTTP routes configured successfully")
}



// HealthCheck handles health check requests
func (h *MainHttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/api/v1/inventory/p/health",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Health check requested")

	// Check data-service health
	dataServiceHealthy := h.checkDataServiceHealth()

	if !dataServiceHealthy {
		h.logger.Error("Data-service health check failed")
		h.writeJSONResponse(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "unhealthy",
			"service": "inventory-service",
			"message": "Data-service is not healthy",
		})
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "inventory-service",
		"message": "Inventory service is operational",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// checkDataServiceHealth checks if the data-service is healthy
func (h *MainHttpHandler) checkDataServiceHealth() bool {
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
func (h *MainHttpHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// loggingMiddleware logs HTTP requests
func (h *MainHttpHandler) loggingMiddleware() mux.MiddlewareFunc {
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
			h.logger.WithFields(logrus.Fields{
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
