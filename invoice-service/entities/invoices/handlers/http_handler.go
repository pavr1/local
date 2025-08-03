package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"invoice-service/entities/invoices/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HttpHandler handles HTTP requests for invoices and invoice details
type HttpHandler struct {
	dbHandler *DBHandler
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler for invoices and invoice details
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateInvoiceWithDetails handles POST /invoices (creates invoice with details in transaction)
func (h *HttpHandler) CreateInvoiceWithDetails(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInvoiceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create invoice request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.InvoiceNumber == "" {
		h.writeErrorResponse(w, "Invoice number is required", http.StatusBadRequest)
		return
	}

	if req.TransactionType == "" {
		h.writeErrorResponse(w, "Transaction type is required", http.StatusBadRequest)
		return
	}

	if req.TransactionType != "income" && req.TransactionType != "outcome" {
		h.writeErrorResponse(w, "Transaction type must be 'income' or 'outcome'", http.StatusBadRequest)
		return
	}

	if req.ExpenseCategoryID == "" {
		h.writeErrorResponse(w, "Expense category ID is required", http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		h.writeErrorResponse(w, "Image URL is required", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		h.writeErrorResponse(w, "At least one invoice item is required", http.StatusBadRequest)
		return
	}

	// Validate each item
	for i, item := range req.Items {
		if item.Detail == "" {
			h.writeErrorResponse(w, "Item detail is required", http.StatusBadRequest)
			return
		}
		if item.Count <= 0 {
			h.writeErrorResponse(w, "Item count must be greater than 0", http.StatusBadRequest)
			return
		}
		if item.Price <= 0 {
			h.writeErrorResponse(w, "Item price must be greater than 0", http.StatusBadRequest)
			return
		}
		if item.UnitType == "" {
			h.writeErrorResponse(w, "Item unit type is required", http.StatusBadRequest)
			return
		}

		// Set invoice ID for each item (will be set by DB handler)
		req.Items[i].InvoiceID = "" // This will be set by the DB handler
	}

	// Create invoice with details
	invoice, err := h.dbHandler.CreateInvoiceWithDetails(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create invoice with details")
		h.writeErrorResponse(w, "Failed to create invoice", http.StatusInternalServerError)
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
	id, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByID(id)
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Error("Failed to get invoice")
		h.writeErrorResponse(w, "Invoice not found", http.StatusNotFound)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// GetInvoiceByNumber handles GET /invoices/number/{number}
func (h *HttpHandler) GetInvoiceByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number, ok := vars["number"]
	if !ok {
		h.writeErrorResponse(w, "Invoice number is required", http.StatusBadRequest)
		return
	}

	invoice, err := h.dbHandler.GetInvoiceByNumber(number)
	if err != nil {
		h.logger.WithError(err).WithField("invoice_number", number).Error("Failed to get invoice by number")
		h.writeErrorResponse(w, "Invoice not found", http.StatusNotFound)
		return
	}

	response := models.InvoiceResponse{
		Success: true,
		Data:    *invoice,
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListInvoices handles GET /invoices
func (h *HttpHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	var req models.ListInvoicesRequest

	// Parse query parameters
	queryParams := r.URL.Query()

	// Parse limit
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		var limit int
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && limit > 0 && limit <= 100 {
			req.Limit = &limit
		}
	}

	// Parse offset
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && offset >= 0 {
			req.Offset = &offset
		}
	}

	// Parse transaction_type
	if transactionType := queryParams.Get("transaction_type"); transactionType != "" {
		if transactionType == "income" || transactionType == "outcome" {
			req.TransactionType = &transactionType
		}
	}

	// Parse expense_category_id
	if categoryID := queryParams.Get("expense_category_id"); categoryID != "" {
		req.ExpenseCategoryID = &categoryID
	}

	// Parse supplier_id
	if supplierID := queryParams.Get("supplier_id"); supplierID != "" {
		req.SupplierID = &supplierID
	}

	invoices, totalCount, err := h.dbHandler.ListInvoices(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list invoices")
		h.writeErrorResponse(w, "Failed to list invoices", http.StatusInternalServerError)
		return
	}

	response := models.InvoicesListResponse{
		Success: true,
		Data:    invoices,
		Count:   totalCount,
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateInvoice handles PUT /invoices/{id}
func (h *HttpHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode update invoice request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate transaction type if provided
	if req.TransactionType != nil {
		if *req.TransactionType != "income" && *req.TransactionType != "outcome" {
			h.writeErrorResponse(w, "Transaction type must be 'income' or 'outcome'", http.StatusBadRequest)
			return
		}
	}

	invoice, err := h.dbHandler.UpdateInvoice(id, req)
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Error("Failed to update invoice")
		h.writeErrorResponse(w, "Failed to update invoice", http.StatusInternalServerError)
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
	id, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteInvoice(id)
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Error("Failed to delete invoice")
		h.writeErrorResponse(w, "Failed to delete invoice", http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDeleteResponse{
		Success: true,
		Message: "Invoice deleted successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// Invoice Detail Methods

// CreateInvoiceDetail handles POST /invoices/{id}/details
func (h *HttpHandler) CreateInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	var req models.CreateInvoiceDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create invoice detail request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the invoice ID from URL
	req.InvoiceID = invoiceID

	// Basic validation
	if req.Detail == "" {
		h.writeErrorResponse(w, "Detail is required", http.StatusBadRequest)
		return
	}
	if req.Count <= 0 {
		h.writeErrorResponse(w, "Count must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Price <= 0 {
		h.writeErrorResponse(w, "Price must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.UnitType == "" {
		h.writeErrorResponse(w, "Unit type is required", http.StatusBadRequest)
		return
	}

	invoiceDetail, err := h.dbHandler.CreateInvoiceDetail(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create invoice detail")
		h.writeErrorResponse(w, "Failed to create invoice detail", http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailResponse{
		Success: true,
		Data:    *invoiceDetail,
		Message: "Invoice detail created successfully",
	}

	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetInvoiceDetailsByInvoiceID handles GET /invoices/{id}/details
func (h *HttpHandler) GetInvoiceDetailsByInvoiceID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, "Invoice ID is required", http.StatusBadRequest)
		return
	}

	req := models.ListInvoiceDetailsRequest{
		InvoiceID: &invoiceID,
	}

	invoiceDetails, totalCount, err := h.dbHandler.ListInvoiceDetails(req)
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", invoiceID).Error("Failed to get invoice details")
		h.writeErrorResponse(w, "Failed to get invoice details", http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailsListResponse{
		Success: true,
		Data:    invoiceDetails,
		Count:   totalCount,
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListInvoiceDetails handles GET /invoice-details
func (h *HttpHandler) ListInvoiceDetails(w http.ResponseWriter, r *http.Request) {
	var req models.ListInvoiceDetailsRequest

	// Parse query parameters (similar to ListInvoices but for invoice details)
	queryParams := r.URL.Query()

	// Parse limit
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		var limit int
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && limit > 0 && limit <= 100 {
			req.Limit = &limit
		}
	}

	// Parse offset
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && offset >= 0 {
			req.Offset = &offset
		}
	}

	// Parse invoice_id
	if invoiceID := queryParams.Get("invoice_id"); invoiceID != "" {
		req.InvoiceID = &invoiceID
	}

	// Parse ingredient_id
	if ingredientID := queryParams.Get("ingredient_id"); ingredientID != "" {
		req.IngredientID = &ingredientID
	}

	invoiceDetails, totalCount, err := h.dbHandler.ListInvoiceDetails(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list invoice details")
		h.writeErrorResponse(w, "Failed to list invoice details", http.StatusInternalServerError)
		return
	}

	response := models.InvoiceDetailsListResponse{
		Success: true,
		Data:    invoiceDetails,
		Count:   totalCount,
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods

func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := models.ErrorResponse{
		Success: false,
		Error:   message,
		Message: message,
	}
	h.writeJSONResponse(w, errorResponse, statusCode)
}
