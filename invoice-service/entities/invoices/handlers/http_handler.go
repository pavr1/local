package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"invoice-service/entities/invoices/models"

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
	CreateInvoiceDetail(req models.CreateInvoiceDetailRequest) (*models.InvoiceDetail, error)
	GetInvoiceDetailByID(id string) (*models.InvoiceDetail, error)
	GetInvoiceDetailsByInvoiceID(invoiceID string) ([]models.InvoiceDetail, error)
	ListInvoiceDetails() ([]models.InvoiceDetail, error)
	UpdateInvoiceDetail(id string, req models.UpdateInvoiceDetailRequest) (*models.InvoiceDetail, error)
	DeleteInvoiceDetail(id string) error
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

// CreateInvoiceWithDetails handles POST /invoices
func (h *HttpHandler) CreateInvoiceWithDetails(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create invoice request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Set current timestamp as default if no transaction date is provided
	if req.TransactionDate == nil {
		now := time.Now()
		req.TransactionDate = &now
		h.logger.WithField("invoice_number", req.InvoiceNumber).Info("Setting default transaction date to current timestamp")
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
	h.writeJSONResponse(w, response, http.StatusCreated)
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
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteInvoice handles DELETE /invoices/{id}
func (h *HttpHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
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
	h.writeJSONResponse(w, response, http.StatusOK)
}

// CreateInvoiceDetail handles POST /invoices/{id}/details
func (h *HttpHandler) CreateInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID := vars["id"]

	if invoiceID == "" {
		h.logger.Warn("Missing invoice ID in create detail request")
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	var req models.CreateInvoiceDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create invoice detail request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Set the invoice ID from the URL
	req.InvoiceID = invoiceID

	detail, err := h.dbHandler.CreateInvoiceDetail(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceDetailResponse{
			Success: false,
			Data:    models.InvoiceDetail{},
			Message: "Failed to create invoice detail: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailResponse{
		Success: true,
		Data:    *detail,
		Message: "Invoice detail created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetInvoiceDetailsByInvoiceID handles GET /invoices/{id}/details
func (h *HttpHandler) GetInvoiceDetailsByInvoiceID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID := vars["id"]

	if invoiceID == "" {
		h.logger.Warn("Missing invoice ID in get details request")
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	details, err := h.dbHandler.GetInvoiceDetailsByInvoiceID(invoiceID)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceDetailsListResponse{
			Success: false,
			Data:    []models.InvoiceDetail{},
			Count:   0,
			Message: "Failed to retrieve invoice details: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailsListResponse{
		Success: true,
		Data:    details,
		Count:   len(details),
		Message: "Invoice details retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListInvoiceDetails handles GET /invoice-details
func (h *HttpHandler) ListInvoiceDetails(w http.ResponseWriter, r *http.Request) {
	details, err := h.dbHandler.ListInvoiceDetails()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.InvoiceDetailsListResponse{
			Success: false,
			Data:    []models.InvoiceDetail{},
			Count:   0,
			Message: "Failed to list invoice details: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailsListResponse{
		Success: true,
		Data:    details,
		Count:   len(details),
		Message: "Invoice details listed successfully",
	}
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
