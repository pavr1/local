package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	existencesHandlers "inventory-service/entities/existences/handlers"
	ingredientCategoriesHandlers "inventory-service/entities/ingredient_categories/handlers"
	ingredientsHandlers "inventory-service/entities/ingredients/handlers"
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

	// TODO: Add other entity handlers when implemented
	// RecipesHandler     *recipesHandlers.HttpHandler
	// etc.
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

	return &MainHttpHandler{
		db:                          db,
		logger:                      logger,
		SuppliersHandler:            suppliersHttpHandler,
		IngredientCategoriesHandler: ingredientCategoriesHttpHandler,
		IngredientsHandler:          ingredientsHttpHandler,
		ExistencesHandler:           existencesHttpHandler,
		// TODO: Add other handlers when implemented
		// etc.
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

// TODO: Add getter methods for other entity handlers when implemented

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
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.WithField("status_code", resp.StatusCode).Error("Data service health check failed")
		return map[string]interface{}{
			"service": "inventory-service",
			"status":  "unhealthy",
			"message": "Data service is unhealthy",
			"error":   fmt.Sprintf("Data service returned status %d", resp.StatusCode),
		}
	}

	return map[string]interface{}{
		"service": "inventory-service",
		"status":  "healthy",
		"entities": map[string]string{
			"suppliers":             "ready",
			"ingredient_categories": "ready",
			"ingredients":           "ready",
			"existences":            "ready",
			// TODO: Add other entities when implemented
			// "recipes":     "ready",
		},
	}
}
