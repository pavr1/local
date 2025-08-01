package handlers

import (
	"encoding/json"
	"net/http"

	"expense-service/entities/receipts/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HttpHandler handles HTTP requests for receipts and receipt items
type HttpHandler struct {
	dbHandler *DBHandler
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler for receipts and receipt items
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateReceiptWithItems handles POST /receipts (creates receipt with items in transaction)
func (h *HttpHandler) CreateReceiptWithItems(w http.ResponseWriter, r *http.Request) {
	var req models.CreateReceiptRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create receipt request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.ReceiptNumber == "" {
		h.writeErrorResponse(w, "Receipt number is required", http.StatusBadRequest)
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
		h.writeErrorResponse(w, "At least one item is required", http.StatusBadRequest)
		return
	}

	// Create receipt with items in transaction
	receipt, err := h.dbHandler.CreateReceiptWithItems(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create receipt with items")
		h.writeErrorResponse(w, "Failed to create receipt", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptResponse{
		Success: true,
		Data:    *receipt,
		Message: "Receipt created successfully",
	}

	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetReceiptByID handles GET /receipts/{id}
func (h *HttpHandler) GetReceiptByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	receipt, err := h.dbHandler.GetReceiptByID(id)
	if err != nil {
		if err.Error() == "receipt not found" {
			h.writeErrorResponse(w, "Receipt not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get receipt")
		h.writeErrorResponse(w, "Failed to get receipt", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptResponse{
		Success: true,
		Data:    *receipt,
		Message: "Receipt retrieved successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// GetReceiptWithItems handles GET /receipts/{id}/items
func (h *HttpHandler) GetReceiptWithItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	// Get receipt
	receipt, err := h.dbHandler.GetReceiptByID(id)
	if err != nil {
		if err.Error() == "receipt not found" {
			h.writeErrorResponse(w, "Receipt not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get receipt")
		h.writeErrorResponse(w, "Failed to get receipt", http.StatusInternalServerError)
		return
	}

	// Get receipt items using the unified dbHandler
	receiptItems, err := h.dbHandler.GetReceiptItemsByReceiptID(id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get receipt items")
		h.writeErrorResponse(w, "Failed to get receipt items", http.StatusInternalServerError)
		return
	}

	// Create response with receipt and items
	responseData := map[string]interface{}{
		"receipt": receipt,
		"items":   receiptItems,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    responseData,
		"message": "Receipt with items retrieved successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// GetReceiptByNumber handles GET /receipts/number/{receipt_number}
func (h *HttpHandler) GetReceiptByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptNumber := vars["receipt_number"]

	if receiptNumber == "" {
		h.writeErrorResponse(w, "Receipt number is required", http.StatusBadRequest)
		return
	}

	receipt, err := h.dbHandler.GetReceiptByNumber(receiptNumber)
	if err != nil {
		if err.Error() == "receipt not found" {
			h.writeErrorResponse(w, "Receipt not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get receipt by number")
		h.writeErrorResponse(w, "Failed to get receipt", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptResponse{
		Success: true,
		Data:    *receipt,
		Message: "Receipt retrieved successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListReceipts handles GET /receipts
func (h *HttpHandler) ListReceipts(w http.ResponseWriter, r *http.Request) {
	var req models.ListReceiptsRequest

	// Parse query parameters
	query := r.URL.Query()

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		// Simple conversion - in production you'd want proper validation
		limit := 50 // default
		req.Limit = &limit
	}

	// Parse offset
	if offsetStr := query.Get("offset"); offsetStr != "" {
		// Simple conversion - in production you'd want proper validation
		offset := 0 // default
		req.Offset = &offset
	}

	// Parse expense_category_id filter
	if categoryID := query.Get("expense_category_id"); categoryID != "" {
		req.ExpenseCategoryID = &categoryID
	}

	// Parse supplier_id filter
	if supplierID := query.Get("supplier_id"); supplierID != "" {
		req.SupplierID = &supplierID
	}

	receipts, count, err := h.dbHandler.ListReceipts(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list receipts")
		h.writeErrorResponse(w, "Failed to list receipts", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptsListResponse{
		Success: true,
		Data:    receipts,
		Count:   count,
		Message: "Receipts retrieved successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateReceipt handles PUT /receipts/{id}
func (h *HttpHandler) UpdateReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateReceiptRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode update receipt request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	receipt, err := h.dbHandler.UpdateReceipt(id, req)
	if err != nil {
		if err.Error() == "receipt not found" {
			h.writeErrorResponse(w, "Receipt not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to update receipt")
		h.writeErrorResponse(w, "Failed to update receipt", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptResponse{
		Success: true,
		Data:    *receipt,
		Message: "Receipt updated successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteReceipt handles DELETE /receipts/{id}
func (h *HttpHandler) DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteReceipt(id)
	if err != nil {
		if err.Error() == "receipt not found" {
			h.writeErrorResponse(w, "Receipt not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to delete receipt")
		h.writeErrorResponse(w, "Failed to delete receipt", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptDeleteResponse{
		Success: true,
		Message: "Receipt deleted successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// AddReceiptItem handles POST /receipts/{id}/items
func (h *HttpHandler) AddReceiptItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptID := vars["id"]

	if receiptID == "" {
		h.writeErrorResponse(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	var req models.CreateReceiptItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create receipt item request")
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the receipt ID from URL parameter
	req.ReceiptID = receiptID

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

	receiptItem, err := h.dbHandler.CreateReceiptItem(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create receipt item")
		h.writeErrorResponse(w, "Failed to create receipt item", http.StatusInternalServerError)
		return
	}

	response := models.ReceiptItemResponse{
		Success: true,
		Data:    *receiptItem,
		Message: "Receipt item added successfully",
	}

	h.writeJSONResponse(w, response, http.StatusCreated)
}

// Helper methods

// writeJSONResponse writes a JSON response
func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes an error response
func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := models.ErrorResponse{
		Success: false,
		Error:   message,
		Message: message,
	}

	h.writeJSONResponse(w, response, statusCode)
}
