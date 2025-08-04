package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost:8086/health")
	if err != nil {
		h.logger.WithError(err).Error("Failed to connect to data-service during health check")
		return map[string]interface{}{
			"service": "invoice-service",
			"status":  "unhealthy",
			"message": "Data service connection failed",
			"error":   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.WithField("status_code", resp.StatusCode).Error("Data service health check failed")
		return map[string]interface{}{
			"service": "invoice-service",
			"status":  "unhealthy",
			"message": "Data service is unhealthy",
			"error":   fmt.Sprintf("Data service returned status %d", resp.StatusCode),
		}
	}

	return map[string]interface{}{
		"service": "invoice-service",
		"status":  "healthy",
		"entities": map[string]string{
			"invoices": "ready",
		},
	}
}
