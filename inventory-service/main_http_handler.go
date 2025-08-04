package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
	SuppliersHandler   *suppliersHandlers.HttpHandler
	IngredientsHandler *ingredientsHandlers.HttpHandler

	// TODO: Add other entity handlers when implemented
	// ExistencesHandler  *existencesHandlers.HttpHandler
	// RecipesHandler     *recipesHandlers.HttpHandler
	// etc.
}

// NewMainHttpHandler creates a new main HTTP handler with all entity handlers
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger) *MainHttpHandler {
	// Initialize suppliers handlers
	suppliersDBHandler := suppliersHandlers.NewDBHandler(db, logger)
	suppliersHttpHandler := suppliersHandlers.NewHttpHandler(suppliersDBHandler, logger)

	// Initialize ingredients handlers
	ingredientsDBHandler := ingredientsHandlers.NewDBHandler(db, logger)
	ingredientsHttpHandler := ingredientsHandlers.NewHttpHandler(ingredientsDBHandler, logger)

	return &MainHttpHandler{
		db:                 db,
		logger:             logger,
		SuppliersHandler:   suppliersHttpHandler,
		IngredientsHandler: ingredientsHttpHandler,
		// TODO: Add other handlers when implemented
		// etc.
	}
}

// GetSuppliersHandler returns the suppliers HTTP handler
func (h *MainHttpHandler) GetSuppliersHandler() *suppliersHandlers.HttpHandler {
	return h.SuppliersHandler
}

// GetIngredientsHandler returns the ingredients HTTP handler
func (h *MainHttpHandler) GetIngredientsHandler() *ingredientsHandlers.HttpHandler {
	return h.IngredientsHandler
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
			"suppliers":   "ready",
			"ingredients": "ready",
			// TODO: Add other entities when implemented
			// "existences":  "ready",
			// "recipes":     "ready",
		},
	}
}
