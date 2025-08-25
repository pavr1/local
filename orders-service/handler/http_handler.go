package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	invoiceModels "invoice-service/entities/invoices/models"
	"orders-service/config"
	"orders-service/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// OrderDBHandlerInterface defines the interface for database operations
type OrderDBHandlerInterface interface {
	CreateOrder(req models.CreateOrderRequest) (*models.OrderWithItems, error)
	GetOrder(id uuid.UUID) (*models.OrderWithItems, error)
	UpdateOrder(id uuid.UUID, req *models.UpdateOrderRequest) (*models.OrderWithItems, error)
	CancelOrder(id uuid.UUID) error
	ListOrders(filter *models.OrderFilter) ([]models.Order, int, error)
	GetOrderSummary() (*models.OrderSummary, error)
	GetPaymentMethodStats() ([]models.PaymentMethodStats, error)
	HealthCheck() error
}

// OrderHTTPHandler handles HTTP requests for orders
type OrderHTTPHandler struct {
	dbHandler OrderDBHandlerInterface
	config    *config.Config
	logger    *logrus.Logger
	// Invoice service client
	httpClient *http.Client
}

// NewOrderHTTPHandler creates a new order HTTP handler
func NewOrderHTTPHandler(dbHandler OrderDBHandlerInterface, cfg *config.Config, logger *logrus.Logger) *OrderHTTPHandler {
	return &OrderHTTPHandler{
		dbHandler:  dbHandler,
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateOrder handles POST /orders
func (h *OrderHTTPHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Create order in database
	createdOrder, err := h.dbHandler.CreateOrder(req)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order", err)
		return
	}

	// Create income invoice
	// Generate invoice number for income invoice
	invoiceNumber := fmt.Sprintf("INV-ORD-%s", createdOrder.Order.OrderNumber)

	// Create invoice request using proper struct
	notes := fmt.Sprintf("Income invoice for order %s", createdOrder.Order.OrderNumber)
	
	// Create invoice items from order items
	var invoiceItems []invoiceModels.CreateInvoiceItemRequest
	for _, orderItem := range createdOrder.Items {
		invoiceItem := invoiceModels.CreateInvoiceItemRequest{
			Detail:       fmt.Sprintf("Order %s - %s", createdOrder.Order.OrderNumber, orderItem.ProductName),
			Count:        float64(orderItem.Quantity),
			UnitType:     "Units",
			ItemsPerUnit: 1,
			Price:        orderItem.ReceipePrice,
		}
		invoiceItems = append(invoiceItems, invoiceItem)
	}
	
	invoiceReq := invoiceModels.CreateInvoiceRequest{
		InvoiceNumber:     invoiceNumber,
		TransactionDate:   createdOrder.Order.TransactionTimestamp,
		TransactionType:   "income", //pvillalobos - this should eventually be handled in the db (hardcoded)
		ExpenseCategoryID: nil,      // Income invoices don't have expense categories
		ImageURL:          "",       // Income invoices don't have images
		Notes:             &notes,
		Items:             invoiceItems,
	}

	// Call invoice service to create income invoice
	invoiceResp, err := h.createIncomeInvoice(invoiceReq)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", createdOrder.Order.ID).Error("Failed to create income invoice")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create invoice", err)
		return
	}

	// Update order with invoice details
	if invoiceResp != nil {
		updateReq := &models.UpdateOrderRequest{
			InvoiceNumber: &invoiceResp.InvoiceNumber,
			InvoiceURL:    &invoiceResp.InvoiceURL,
		}

		updatedOrder, err := h.dbHandler.UpdateOrder(createdOrder.Order.ID, updateReq)
		if err != nil {
			h.logger.WithError(err).WithField("order_id", createdOrder.Order.ID).Error("Failed to update order with invoice details")
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update order with invoice details", err)
			return
		}
		createdOrder = updatedOrder
	}

	h.logger.WithFields(logrus.Fields{
		"order_id":     createdOrder.Order.ID,
		"order_number": createdOrder.Order.OrderNumber,
		"total_amount": createdOrder.Order.TotalAmount,
		"iva_amount":   createdOrder.Order.IvaAmount,
		"service_tax":  createdOrder.Order.ServiceTaxAmount,
	}).Info("Order created successfully")

	h.respondWithSuccess(w, http.StatusCreated, "Order created successfully", createdOrder)
}

// GetOrder handles GET /orders/{id}
func (h *OrderHTTPHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	order, err := h.dbHandler.GetOrder(orderID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(w, http.StatusNotFound, "Order not found", err)
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve order", err)
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Order retrieved successfully", order)
}

// UpdateOrder handles PUT /orders/{id}
func (h *OrderHTTPHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	var req models.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Validate payment method if provided
	if req.PaymentMethod != nil {
		validMethods := []string{models.PaymentMethodCash, models.PaymentMethodCard, models.PaymentMethodSinpe}
		valid := false
		for _, method := range validMethods {
			if *req.PaymentMethod == method {
				valid = true
				break
			}
		}
		if !valid {
			h.respondWithError(w, http.StatusBadRequest, "Invalid payment method", nil)
			return
		}
	}

	// Validate order status if provided
	if req.Status != nil {
		validStatuses := []string{models.OrderStatusPending, models.OrderStatusCompleted, models.OrderStatusCancelled, models.OrderStatusSinpePending}
		valid := false
		for _, status := range validStatuses {
			if *req.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			h.respondWithError(w, http.StatusBadRequest, "Invalid order status", nil)
			return
		}
	}

	// Update order
	updatedOrder, err := h.dbHandler.UpdateOrder(orderID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(w, http.StatusNotFound, "Order not found", err)
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id": orderID,
	}).Info("Order updated successfully")

	h.respondWithSuccess(w, http.StatusOK, "Order updated successfully", updatedOrder)
}

// CancelOrder handles DELETE /orders/{id}
func (h *OrderHTTPHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	if err := h.dbHandler.CancelOrder(orderID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "completed") {
			h.respondWithError(w, http.StatusBadRequest, "Order cannot be cancelled", err)
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Failed to cancel order", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id": orderID,
	}).Info("Order cancelled successfully")

	h.respondWithSuccess(w, http.StatusOK, "Order cancelled successfully", map[string]interface{}{
		"order_id": orderID,
		"status":   "cancelled",
	})
}

// ListOrders handles GET /orders
func (h *OrderHTTPHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	filter := &models.OrderFilter{}

	// Parse query parameters
	query := r.URL.Query()

	// Customer ID filter
	if customerIDStr := query.Get("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid customer_id", err)
			return
		}
		filter.CustomerID = &customerID
	}

	// Order status filter
	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	// Payment method filter
	if method := query.Get("payment_method"); method != "" {
		filter.PaymentMethod = &method
	}

	// Date filters
	if dateFromStr := query.Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid date_from format, use YYYY-MM-DD", err)
			return
		}
		filter.DateFrom = &dateFrom
	}

	if dateToStr := query.Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid date_to format, use YYYY-MM-DD", err)
			return
		}
		// Set to end of day
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.DateTo = &dateTo
	}

	// Amount filters
	if minAmountStr := query.Get("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseFloat(minAmountStr, 64)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid min_amount", err)
			return
		}
		filter.MinAmount = &minAmount
	}

	if maxAmountStr := query.Get("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseFloat(maxAmountStr, 64)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid max_amount", err)
			return
		}
		filter.MaxAmount = &maxAmount
	}

	// Pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			h.respondWithError(w, http.StatusBadRequest, "Invalid limit", err)
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // default
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			h.respondWithError(w, http.StatusBadRequest, "Invalid offset", err)
			return
		}
		filter.Offset = offset
	}

	// Sorting
	if sortBy := query.Get("sort_by"); sortBy != "" {
		validSortFields := []string{"order_date", "total_amount", "final_amount", "order_status", "payment_method"}
		valid := false
		for _, field := range validSortFields {
			if sortBy == field {
				valid = true
				break
			}
		}
		if !valid {
			h.respondWithError(w, http.StatusBadRequest, "Invalid sort_by field", nil)
			return
		}
		filter.SortBy = sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		if sortOrder != "asc" && sortOrder != "desc" {
			h.respondWithError(w, http.StatusBadRequest, "Invalid sort_order, use 'asc' or 'desc'", nil)
			return
		}
		filter.SortOrder = sortOrder
	}

	// Get orders
	orders, totalCount, err := h.dbHandler.ListOrders(filter)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve orders", err)
		return
	}

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": totalCount,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
		"has_more":    filter.Offset+len(orders) < totalCount,
	}

	h.respondWithSuccess(w, http.StatusOK, "Orders retrieved successfully", response)
}

// GetOrdersByDateRange handles GET /orders/by-date-range
func (h *OrderHTTPHandler) GetOrdersByDateRange(w http.ResponseWriter, r *http.Request) {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/orders/by-date-range",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Get orders by date range requested")

	// Parse query parameters
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" || dateToStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Both date_from and date_to parameters are required", nil)
		return
	}

	// Parse dates
	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		h.logger.WithError(err).WithField("date_from", dateFromStr).Warn("Invalid date_from format")
		h.respondWithError(w, http.StatusBadRequest, "Invalid date_from format. Use YYYY-MM-DD", err)
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		h.logger.WithError(err).WithField("date_to", dateToStr).Warn("Invalid date_to format")
		h.respondWithError(w, http.StatusBadRequest, "Invalid date_to format. Use YYYY-MM-DD", err)
		return
	}

	// Set time to start and end of day for inclusive range
	dateFrom = time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateTo = time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)

	// Create filter
	filter := &models.OrderFilter{
		DateFrom:  &dateFrom,
		DateTo:    &dateTo,
		Limit:     1000, // Large limit for date range queries
		SortBy:    "transaction_timestamp",
		SortOrder: "DESC",
	}

	// Get orders
	orders, total, err := h.dbHandler.ListOrders(filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get orders by date range")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get orders", err)
		return
	}

	// Create response
	response := map[string]interface{}{
		"success": true,
		"data":    orders,
		"total":   total,
		"filters": map[string]interface{}{
			"date_from": dateFromStr,
			"date_to":   dateToStr,
		},
	}

	h.logger.WithFields(logrus.Fields{
		"orders_count": len(orders),
		"total":        total,
		"date_from":    dateFromStr,
		"date_to":      dateToStr,
	}).Info("Successfully retrieved orders by date range")

	h.respondWithSuccess(w, http.StatusOK, "Orders retrieved successfully", response)
}

// GetOrderSummary handles GET /orders/summary
func (h *OrderHTTPHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/orders/summary",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Get order summary requested")

	summary, err := h.dbHandler.GetOrderSummary()
	if err != nil {
		h.logger.WithError(err).Error("Failed to retrieve order summary")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve order summary", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"total_orders":     summary.TotalOrders,
		"pending_orders":   summary.PendingOrders,
		"completed_orders": summary.CompletedOrders,
		"total_revenue":    summary.TotalRevenue,
	}).Info("Order summary retrieved successfully")

	h.respondWithSuccess(w, http.StatusOK, "Order summary retrieved successfully", summary)
}

// GetPaymentMethodStats handles GET /orders/payment-method-stats
func (h *OrderHTTPHandler) GetPaymentMethodStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dbHandler.GetPaymentMethodStats()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve payment method stats", err)
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Payment method stats retrieved successfully", stats)
}

// HealthCheck handles GET /health
func (h *OrderHTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database health
	if err := h.dbHandler.HealthCheck(); err != nil {
		h.respondWithError(w, http.StatusServiceUnavailable, "Database connection failed", err)
		return
	}

	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	dataServiceHealthURL := fmt.Sprintf("%s/api/v1/data/p/health", h.config.DataServiceURL)
	resp, err := client.Get(dataServiceHealthURL)
	if err != nil {
		h.respondWithError(w, http.StatusServiceUnavailable, "Data service connection failed", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("data service returned status %d", resp.StatusCode)
		h.respondWithError(w, http.StatusServiceUnavailable, "Data service is unhealthy", err)
		return
	}

	response := map[string]interface{}{
		"service": "orders-service",
		"status":  "healthy",
		"time":    time.Now(),
		"version": "1.0.0",
	}

	h.respondWithSuccess(w, http.StatusOK, "Orders service is healthy", response)
}

// createIncomeInvoice calls the invoice service to create an income invoice
func (h *OrderHTTPHandler) createIncomeInvoice(invoiceReq invoiceModels.CreateInvoiceRequest) (*InvoiceResponse, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice data: %w", err)
	}

	// Call invoice service using configured URL
	invoiceURL := fmt.Sprintf("%s/api/v1/invoices", h.config.InvoiceServiceURL)
	resp, err := h.httpClient.Post(
		invoiceURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call invoice service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invoice service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var invoiceResp InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to decode invoice response: %w", err)
	}

	return &invoiceResp, nil
}

// Helper methods for HTTP responses
func (h *OrderHTTPHandler) respondWithSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHTTPHandler) respondWithError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}

	// Always log the error message
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"status_code": status,
			"endpoint":    "orders-service",
		}).Error(message)
		response["error"] = err.Error()
	} else {
		h.logger.WithFields(logrus.Fields{
			"status_code": status,
			"endpoint":    "orders-service",
		}).Error(message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
