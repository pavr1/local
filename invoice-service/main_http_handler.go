package main

import (
	"database/sql"

	invoicesHandlers "invoice-service/entities/invoices/handlers"

	"github.com/sirupsen/logrus"
)

// MainHttpHandler aggregates all entity HTTP handlers
type MainHttpHandler struct {
	// Database connection
	db     *sql.DB
	logger *logrus.Logger

	// Entity handlers
	InvoicesHandler *invoicesHandlers.HttpHandler
}

// NewMainHttpHandler creates a new main HTTP handler with all entity handlers
func NewMainHttpHandler(db *sql.DB, logger *logrus.Logger) *MainHttpHandler {
	// Initialize invoices handlers (now includes invoice details functionality)
	invoicesDBHandler := invoicesHandlers.NewDBHandler(db, logger)
	invoicesHttpHandler := invoicesHandlers.NewHttpHandler(invoicesDBHandler, logger)

	return &MainHttpHandler{
		db:              db,
		logger:          logger,
		InvoicesHandler: invoicesHttpHandler,
	}
}

// GetInvoicesHandler returns the invoices HTTP handler
func (h *MainHttpHandler) GetInvoicesHandler() *invoicesHandlers.HttpHandler {
	return h.InvoicesHandler
}

// HealthCheck provides a health check endpoint for the entire service
func (h *MainHttpHandler) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service": "invoice-service",
		"status":  "healthy",
		"entities": map[string]string{
			"invoices": "ready",
		},
	}
}
