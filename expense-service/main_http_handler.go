package main

import (
	"database/sql"

	receiptsHandlers "expense-service/entities/receipts/handlers"

	"github.com/sirupsen/logrus"
)

// MainHttpHandler aggregates all entity HTTP handlers
type MainHttpHandler struct {
	// Database connection
	db     *sql.DB
	logger *logrus.Logger

	// Entity handlers
	ReceiptsHandler *receiptsHandlers.HttpHandler
}

// NewMainHttpHandler creates a new main HTTP handler with all entity handlers
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger) *MainHttpHandler {
	// Initialize receipts handlers (now includes receipt items functionality)
	receiptsDBHandler := receiptsHandlers.NewDBHandler(db, logger)
	receiptsHttpHandler := receiptsHandlers.NewHttpHandler(receiptsDBHandler, logger)

	return &MainHttpHandler{
		db:              db,
		logger:          logger,
		ReceiptsHandler: receiptsHttpHandler,
	}
}

// GetReceiptsHandler returns the receipts HTTP handler
func (h *MainHttpHandler) GetReceiptsHandler() *receiptsHandlers.HttpHandler {
	return h.ReceiptsHandler
}

// HealthCheck provides a health check endpoint for the entire service
func (h *MainHttpHandler) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service": "expense-service",
		"status":  "healthy",
		"entities": map[string]string{
			"receipts": "ready",
		},
	}
}
