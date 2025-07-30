package main

import (
	"database/sql"

	suppliersHandlers "inventory-service/entities/suppliers/handlers"

	"github.com/sirupsen/logrus"
)

// MainHttpHandler aggregates all entity HTTP handlers
type MainHttpHandler struct {
	// Database connection
	db     *sql.DB
	logger *logrus.Logger

	// Entity handlers
	SuppliersHandler *suppliersHandlers.HttpHandler

	// TODO: Add other entity handlers when implemented
	// IngredientsHandler *ingredientsHandlers.HttpHandler
	// ExistencesHandler  *existencesHandlers.HttpHandler
	// RecipesHandler     *recipesHandlers.HttpHandler
	// etc.
}

// NewMainHttpHandler creates a new main HTTP handler with all entity handlers
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger) *MainHttpHandler {
	// Initialize suppliers handlers
	suppliersDBHandler := suppliersHandlers.NewDBHandler(db, logger)
	suppliersHttpHandler := suppliersHandlers.NewHttpHandler(suppliersDBHandler, logger)

	// TODO: Initialize other entity handlers when implemented
	// ingredientsDBHandler := ingredientsHandlers.NewDBHandler(db, logger)
	// ingredientsHttpHandler := ingredientsHandlers.NewHttpHandler(ingredientsDBHandler, logger)

	return &MainHttpHandler{
		db:               db,
		logger:           logger,
		SuppliersHandler: suppliersHttpHandler,
		// TODO: Add other handlers when implemented
		// IngredientsHandler: ingredientsHttpHandler,
		// etc.
	}
}

// GetSuppliersHandler returns the suppliers HTTP handler
func (h *MainHttpHandler) GetSuppliersHandler() *suppliersHandlers.HttpHandler {
	return h.SuppliersHandler
}

// TODO: Add getter methods for other entity handlers when implemented
// func (h *MainHttpHandler) GetIngredientsHandler() *ingredientsHandlers.HttpHandler {
//     return h.IngredientsHandler
// }

// HealthCheck provides a health check endpoint for the entire service
func (h *MainHttpHandler) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service": "inventory-service",
		"status":  "healthy",
		"entities": map[string]string{
			"suppliers": "ready",
			// TODO: Add other entities when implemented
			// "ingredients": "ready",
			// "existences":  "ready",
			// "recipes":     "ready",
		},
	}
}
