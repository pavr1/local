package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	existencesHandlers "inventory-service/entities/existences/handlers"
	ingredientCategoriesHandlers "inventory-service/entities/ingredient_categories/handlers"
	ingredientsHandlers "inventory-service/entities/ingredients/handlers"
	recipeCategoriesHandlers "inventory-service/entities/recipe_categories/handlers"
	recipeIngredientsHandlers "inventory-service/entities/recipe_ingredients/handlers"
	recipesHandlers "inventory-service/entities/recipes/handlers"
	runoutIngredientsHandlers "inventory-service/entities/runout_ingredients/handlers"
	suppliersHandlers "inventory-service/entities/suppliers/handlers"

	"github.com/sirupsen/logrus"
)

// MainHttpHandler aggregates all entity HTTP handlers
type MainHttpHandler struct {
	// Database connection
	db     *sql.DB
	logger *logrus.Logger

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
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger) *MainHttpHandler {
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
	recipesHttpHandler := recipesHandlers.NewRecipeHTTPHandler(db, logger)

	// Initialize recipe ingredients handlers
	recipeIngredientsHttpHandler := recipeIngredientsHandlers.NewRecipeIngredientHTTPHandler(db, logger)

	return &MainHttpHandler{
		db:                          db,
		logger:                      logger,
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

// HealthCheck provides a health check endpoint for the entire service
func (h *MainHttpHandler) HealthCheck() map[string]interface{} {
	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost:8086/health")
	if err != nil {
		h.logger.WithError(err).Error("Failed to connect to data-service during health check")
		return map[string]interface{}{
			"service": "inventory-service",
			"status":  "unhealthy",
			"message": "Data service connection failed",
			"error":   err.Error(),
			"time":    time.Now().Format(time.RFC3339),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.WithField("status_code", resp.StatusCode).Error("Data service returned non-OK status during health check")
		return map[string]interface{}{
			"service": "inventory-service",
			"status":  "unhealthy",
			"message": fmt.Sprintf("Data service returned status %d", resp.StatusCode),
			"time":    time.Now().Format(time.RFC3339),
		}
	}

	return map[string]interface{}{
		"service": "inventory-service",
		"status":  "healthy",
		"message": "Service is running normally",
		"time":    time.Now().Format(time.RFC3339),
	}
}
