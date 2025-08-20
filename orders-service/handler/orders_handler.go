package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"orders-service/config"
	"orders-service/models"
	ordersql "orders-service/sql"

	// Removed utils import - gateway handles all auth

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type OrdersHandler interface {
	// Order operations
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrder(w http.ResponseWriter, r *http.Request)
	UpdateOrder(w http.ResponseWriter, r *http.Request)
	CancelOrder(w http.ResponseWriter, r *http.Request)
	ListOrders(w http.ResponseWriter, r *http.Request)
	GetOrdersByDateRange(w http.ResponseWriter, r *http.Request)

	// Statistics and reports
	GetOrderSummary(w http.ResponseWriter, r *http.Request)
	GetPaymentMethodStats(w http.ResponseWriter, r *http.Request)

	// Health check
	HealthCheck(w http.ResponseWriter, r *http.Request)

	// No longer needed - gateway handles all auth
	// GetJWTManager() *utils.JWTManager
}

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	CreateOrder(order *models.Order, items []models.OrderedRecipe) error
	GetOrderByID(id uuid.UUID) (*models.Order, error)
	GetOrderWithItems(id uuid.UUID) (*models.OrderWithItems, error)
	GetOrderedRecipesByOrderID(orderID uuid.UUID) ([]models.OrderedRecipe, error)
	UpdateOrder(id uuid.UUID, updates *models.UpdateOrderRequest) error
	CancelOrder(id uuid.UUID) error
	ListOrders(filter *models.OrderFilter) ([]models.Order, int, error)
	GetOrderSummary() (*models.OrderSummary, error)
	GetPaymentMethodStats() ([]models.PaymentMethodStats, error)
	HealthCheck() error
}

type ordersHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logrus.Logger
	// Removed jwtManager - gateway handles all auth
	repo OrderRepository
	// Invoice service client
	httpClient *http.Client
}

// New creates a new orders handler instance
func New(db *sql.DB, cfg *config.Config, logger *logrus.Logger) (OrdersHandler, error) {
	// Removed jwtManager creation - gateway handles all auth

	repo, err := ordersql.NewRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &ordersHandler{
		db:         db,
		config:     cfg,
		logger:     logger,
		// Removed jwtManager - gateway handles all auth
		repo:       repo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// === ORDER ENDPOINTS ===

// CreateOrder creates a new order
func (h *ordersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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

	// Calculate totals
	totalAmount := 0.0
	for _, item := range req.Items {
		totalAmount += float64(item.Quantity) * item.UnitPrice
	}

	// Calculate tax and service tax
	ivaAmount := totalAmount * (h.config.DefaultTaxRate / 100)
	serviceTaxAmount := totalAmount * 0.10 // 10% service tax
	subtotalAmount := totalAmount

	// Generate order number
	orderNumber := generateOrderNumber()

	// Create order
	order := &models.Order{
		ID:                    uuid.New(),
		OrderNumber:           orderNumber,
		CustomerID:            req.CustomerID,
		SalesRepresentativeID: nil, // Will be set from authenticated user
		Status:                models.OrderStatusPending,
		PaymentMethod:         req.PaymentMethod,
		TransactionReference:  req.TransactionReference,
		SinpeScreenshotURL:    req.SinpeScreenshotURL,
		SubtotalAmount:        subtotalAmount,
		DiscountAmount:        req.DiscountAmount,
		IvaAmount:             ivaAmount,
		ServiceTaxAmount:      serviceTaxAmount,
		TotalAmount:           totalAmount + ivaAmount + serviceTaxAmount - req.DiscountAmount,
		InvoiceNumber:         nil, // Will be set after invoice creation
		InvoiceURL:            nil, // Will be set after invoice creation
		TransactionTimestamp:  time.Now(),
		CompletedAt:           nil,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Create ordered recipes
	var items []models.OrderedRecipe
	for _, reqItem := range req.Items {
		subtotal := float64(reqItem.Quantity) * reqItem.UnitPrice
		item := models.OrderedRecipe{
			ID:           uuid.New(),
			OrderID:      order.ID,
			RecipeID:     reqItem.RecipeID,
			ProductName:  "", // Will be populated from recipe data
			Quantity:     reqItem.Quantity,
			ReceipePrice: reqItem.UnitPrice,
			Subtotal:     subtotal,
		}
		items = append(items, item)
	}

	// Save to database
	if err := h.repo.CreateOrder(order, items); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order", err)
		return
	}

	// Create income invoice
	invoiceData := map[string]interface{}{
		"invoice_type":        "income",
		"transaction_date":    order.TransactionTimestamp,
		"subtotal_amount":     order.SubtotalAmount,
		"discount_amount":     order.DiscountAmount,
		"iva_amount":          order.IvaAmount,
		"service_tax_amount":  order.ServiceTaxAmount,
		"total_amount":        order.TotalAmount,
		"payment_method":      order.PaymentMethod,
		"transaction_reference": order.TransactionReference,
		"sinpe_screenshot_url": order.SinpeScreenshotURL,
		"notes":               fmt.Sprintf("Invoice for order %s", order.OrderNumber),
		"expense_category_id": nil, // Income invoices don't have expense categories
	}

	// Call invoice service to create income invoice
	invoiceResp, err := h.createIncomeInvoice(invoiceData)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", order.ID).Error("Failed to create income invoice")
		// Don't fail the order creation if invoice creation fails
		// The order will be created without invoice details
	} else {
		// Update order with invoice details
		if invoiceResp != nil {
			order.InvoiceNumber = &invoiceResp.InvoiceNumber
			order.InvoiceURL = &invoiceResp.InvoiceURL
			
			// Update the order with invoice details
			updateReq := &models.UpdateOrderRequest{
				InvoiceNumber: order.InvoiceNumber,
				InvoiceURL:    order.InvoiceURL,
			}
			if err := h.repo.UpdateOrder(order.ID, updateReq); err != nil {
				h.logger.WithError(err).WithField("order_id", order.ID).Warn("Failed to update order with invoice details")
			}
		}
	}

	// Get the complete order with calculated final_amount
	createdOrder, err := h.repo.GetOrderWithItems(order.ID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve created order", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"total_amount": order.TotalAmount,
		"iva_amount":   order.IvaAmount,
		"service_tax":  order.ServiceTaxAmount,
	}).Info("Order created successfully")

	h.respondWithSuccess(w, http.StatusCreated, "Order created successfully", createdOrder)
}

// GetOrder retrieves an order by ID
func (h *ordersHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	order, err := h.repo.GetOrderWithItems(orderID)
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

// UpdateOrder updates an existing order
func (h *ordersHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.UpdateOrder(orderID, &req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(w, http.StatusNotFound, "Order not found", err)
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order", err)
		return
	}

	// Get updated order
	updatedOrder, err := h.repo.GetOrderWithItems(orderID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated order", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id": orderID,
	}).Info("Order updated successfully")

	h.respondWithSuccess(w, http.StatusOK, "Order updated successfully", updatedOrder)
}

// CancelOrder cancels an order
func (h *ordersHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}

	if err := h.repo.CancelOrder(orderID); err != nil {
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

// ListOrders retrieves orders with filtering and pagination
func (h *ordersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
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
	orders, totalCount, err := h.repo.ListOrders(filter)
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
func (h *ordersHandler) GetOrdersByDateRange(w http.ResponseWriter, r *http.Request) {
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
	orders, total, err := h.repo.ListOrders(filter)
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

// === STATISTICS ENDPOINTS ===

// GetOrderSummary retrieves order statistics
func (h *ordersHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetOrderSummary()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve order summary", err)
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Order summary retrieved successfully", summary)
}

// GetPaymentMethodStats retrieves payment method statistics
func (h *ordersHandler) GetPaymentMethodStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetPaymentMethodStats()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve payment method stats", err)
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Payment method stats retrieved successfully", stats)
}

// === HEALTH CHECK ===

// HealthCheck checks the health of the orders service
func (h *ordersHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://icecream_data_service:8086/api/v1/data/p/health")
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

// === HELPER METHODS ===

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", rand.Intn(10000))
	return fmt.Sprintf("ORD-%s-%s", timestamp, random)
}

// InvoiceResponse represents the response from invoice service
type InvoiceResponse struct {
	InvoiceNumber string `json:"invoice_number"`
	InvoiceURL    string `json:"invoice_url"`
}

// createIncomeInvoice calls the invoice service to create an income invoice
func (h *ordersHandler) createIncomeInvoice(invoiceData map[string]interface{}) (*InvoiceResponse, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(invoiceData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice data: %w", err)
	}

	// Call invoice service
	resp, err := h.httpClient.Post(
		"http://icecream_invoice_service:8084/api/v1/invoices",
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

func (h *ordersHandler) respondWithSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *ordersHandler) respondWithError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}

	if err != nil {
		h.logger.WithError(err).Error(message)
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
