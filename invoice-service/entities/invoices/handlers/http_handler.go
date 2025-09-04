package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"invoice-service/entities/invoices/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateInvoice(req models.CreateInvoiceRequest, logger *logrus.Logger) (*models.Invoice, error)
	GetInvoiceByID(id string, logger *logrus.Logger) (*models.Invoice, error)
	GetInvoiceByNumber(number string, logger *logrus.Logger) (*models.Invoice, error)
	ListInvoices(logger *logrus.Logger) ([]models.Invoice, error)
	UpdateInvoice(id string, req models.UpdateInvoiceRequest, logger *logrus.Logger) (*models.Invoice, error)
	DeleteInvoice(id string, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for existence operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateInvoice handles POST /invoices - creates invoice with all details
func (h *HttpHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create invoice request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateInvoiceRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number":   req.InvoiceNumber,
			"transaction_type": req.TransactionType,
		}).Error("Validation failed for create invoice request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	invoice, err := h.dbHandler.CreateInvoice(req, h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Invoice{},
			Message: "Failed to create invoice: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *invoice,
		Message: "Invoice created successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":       invoice.ID,
		"invoice_number":   invoice.InvoiceNumber,
		"transaction_type": invoice.TransactionType,
	}).Info("Invoice created successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
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
		if err := h.validateInvoiceItem(item); err != nil {
			return fmt.Errorf("item %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateInvoiceItem validates a single invoice item
func (h *HttpHandler) validateInvoiceItem(item models.CreateInvoiceItemRequest) error {
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
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		})
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByID(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Invoice{},
			Message: "Failed to retrieve invoice: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *invoice,
		Message: "Invoice retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// GetInvoiceByNumber handles GET /invoices/number/{number}
func (h *HttpHandler) GetInvoiceByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number := vars["number"]

	if number == "" {
		h.logger.Warn("Missing invoice number in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invoice number is required",
		})
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByNumber(number, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Invoice{},
			Message: "Failed to retrieve invoice: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *invoice,
		Message: "Invoice retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// ListInvoices handles GET /invoices
func (h *HttpHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.dbHandler.ListInvoices(h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.Invoice{},
			Message: "Failed to list invoices: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    invoices,
		Message: "Invoices listed successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// UpdateInvoice handles PUT /invoices/{id}
func (h *HttpHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing invoice ID in update request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		})
		return
	}

	var req models.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update invoice request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	invoice, err := h.dbHandler.UpdateInvoice(id, req, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Invoice{},
				Message: "Invoice not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Invoice{},
			Message: "Failed to update invoice: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *invoice,
		Message: "Invoice updated successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":       invoice.ID,
		"invoice_number":   invoice.InvoiceNumber,
		"transaction_type": invoice.TransactionType,
	}).Info("Invoice updated successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// DeleteInvoice handles DELETE /invoices/{id}
func (h *HttpHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing invoice ID in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		})
		return
	}

	err := h.dbHandler.DeleteInvoice(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Invoice not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete invoice: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Invoice deleted successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id": id,
	}).Info("Invoice deleted successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}
