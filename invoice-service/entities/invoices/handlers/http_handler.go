package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"invoice-service/entities/invoices/models"
	"invoice-service/pkg/requestlogger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateInvoice(req models.CreateInvoiceRequest) (*models.Invoice, error)
	GetInvoiceByID(id string) (*models.Invoice, error)
	GetInvoiceByNumber(number string) (*models.Invoice, error)
	ListInvoices() ([]models.Invoice, error)
	UpdateInvoice(id string, req models.UpdateInvoiceRequest) (*models.Invoice, error)
	DeleteInvoice(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for invoice operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// NewHttpHandlerWithInterface creates a new HTTP handler with interface (for testing)
func NewHttpHandlerWithInterface(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateInvoice handles POST /invoices - creates invoice with all details
func (h *HttpHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	var req models.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create invoice request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateInvoiceRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number":   req.InvoiceNumber,
			"transaction_type": req.TransactionType,
		}).Error("Validation failed for create invoice request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.CreateInvoice(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceResponse{
			Success: false,
			Data:    models.Invoice{},
			Message: "Failed to create invoice: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
		Message: "Invoice created successfully",
	}

	logger.WithFields(logrus.Fields{
		"invoice_id":       invoice.ID,
		"invoice_number":   invoice.InvoiceNumber,
		"transaction_type": invoice.TransactionType,
	}).Info("Invoice created successfully")

	h.writeJSONResponse(w, response, http.StatusCreated)
}

// validateCreateInvoiceRequest validates all required fields for invoice creation
func (h *HttpHandler) validateCreateInvoiceRequest(req models.CreateInvoiceRequest) error {
	// Validate invoice_number (required, non-empty)
	if req.InvoiceNumber == "" {
		return fmt.Errorf("invoice_number is required and cannot be empty")
	}

	// Validate transaction_date (required, cannot be zero time)
	if req.TransactionDate.IsZero() {
		return fmt.Errorf("transaction_date is required and cannot be zero")
	}

	// Validate transaction_type (required, must be 'income' or 'outcome')
	if req.TransactionType == "" {
		return fmt.Errorf("transaction_type is required and cannot be empty")
	}
	if req.TransactionType != "income" && req.TransactionType != "outcome" {
		return fmt.Errorf("transaction_type must be either 'income' or 'outcome', got: %s", req.TransactionType)
	}

	// Validate expense_category_id based on transaction_type
	if req.TransactionType == "outcome" {
		// For outcome invoices, expense_category_id is required
		if req.ExpenseCategoryID == nil || *req.ExpenseCategoryID == "" {
			return fmt.Errorf("expense_category_id is required for outcome invoices")
		}
	} else {
		// For income invoices, expense_category_id should be nil or empty
		if req.ExpenseCategoryID != nil && *req.ExpenseCategoryID != "" {
			return fmt.Errorf("expense_category_id should not be provided for income invoices")
		}
	}

	// Validate image_url (required, non-empty)
	if req.ImageURL == "" {
		return fmt.Errorf("image_url is required and cannot be empty")
	}

	// Validate items (required, non-empty array)
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	// Validate each item
	for i, item := range req.Items {
		if err := h.validateInvoiceItem(item, i); err != nil {
			return fmt.Errorf("item %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateInvoiceItem validates a single invoice item
func (h *HttpHandler) validateInvoiceItem(item models.CreateInvoiceItemRequest, index int) error {
	// Validate detail (required, non-empty)
	if item.Detail == "" {
		return fmt.Errorf("detail is required and cannot be empty")
	}

	// Validate count (required, greater than 0)
	if item.Count <= 0 {
		return fmt.Errorf("count must be greater than 0, got: %f", item.Count)
	}

	// Validate unit_type (required, must be one of the allowed values)
	if item.UnitType == "" {
		return fmt.Errorf("unit_type is required and cannot be empty")
	}
	allowedUnitTypes := []string{"Liters", "Gallons", "Units", "Bag"}
	unitTypeValid := false
	for _, allowed := range allowedUnitTypes {
		if item.UnitType == allowed {
			unitTypeValid = true
			break
		}
	}
	if !unitTypeValid {
		return fmt.Errorf("unit_type must be one of %v, got: %s", allowedUnitTypes, item.UnitType)
	}

	// Validate items_per_unit (required, greater than 0)
	if item.ItemsPerUnit <= 0 {
		return fmt.Errorf("items_per_unit must be greater than 0, got: %d", item.ItemsPerUnit)
	}

	// Validate price (required, greater than 0)
	if item.Price <= 0 {
		return fmt.Errorf("price must be greater than 0, got: %f", item.Price)
	}

	return nil
}

// GetInvoiceByID handles GET /invoices/{id}
func (h *HttpHandler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing invoice ID in get request")
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.InvoiceResponse{
				Success: false,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceResponse{
			Success: false,
			Data:    models.Invoice{},
			Message: "Failed to retrieve invoice: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
		Message: "Invoice retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// GetInvoiceByNumber handles GET /invoices/number/{number}
func (h *HttpHandler) GetInvoiceByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number := vars["number"]

	if number == "" {
		h.logger.Warn("Missing invoice number in get request")
		h.writeErrorResponse(w, "Invoice number is required", http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByNumber(number)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.InvoiceResponse{
				Success: false,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceResponse{
			Success: false,
			Data:    models.Invoice{},
			Message: "Failed to retrieve invoice: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
		Message: "Invoice retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListInvoices handles GET /invoices
func (h *HttpHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.dbHandler.ListInvoices()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.InvoicesListResponse{
			Success: false,
			Data:    []models.Invoice{},
			Count:   0,
			Message: "Failed to list invoices: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoicesListResponse{
		Success: true,
		Data:    invoices,
		Count:   len(invoices),
		Message: "Invoices listed successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateInvoice handles PUT /invoices/{id}
func (h *HttpHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing invoice ID in update request")
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update invoice request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.UpdateInvoice(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.InvoiceResponse{
				Success: false,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceResponse{
			Success: false,
			Data:    models.Invoice{},
			Message: "Failed to update invoice: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
		Message: "Invoice updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"invoice_id":       invoice.ID,
		"invoice_number":   invoice.InvoiceNumber,
		"transaction_type": invoice.TransactionType,
	}).Info("Invoice updated successfully")

	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteInvoice handles DELETE /invoices/{id}
func (h *HttpHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing invoice ID in delete request")
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteInvoice(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.InvoiceDeleteResponse{
				Success: false,
				Message: "Invoice not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceDeleteResponse{
			Success: false,
			Message: "Failed to delete invoice: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDeleteResponse{
		Success: true,
		Message: "Invoice deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"invoice_id": id,
	}).Info("Invoice deleted successfully")

	h.writeJSONResponse(w, response, http.StatusOK)
}

// writeJSONResponse writes a JSON response with the given status code
func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// writeErrorResponse writes an error response with the given message and status code
func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := models.ErrorResponse{
		Success: false,
		Error:   message,
		Message: message,
	}
	h.writeJSONResponse(w, response, statusCode)
}
