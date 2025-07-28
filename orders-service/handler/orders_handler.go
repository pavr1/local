package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"orders-service/config"
	"orders-service/models"
	ordersql "orders-service/sql"
	"orders-service/utils"

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

	// Statistics and reports
	GetOrderSummary(w http.ResponseWriter, r *http.Request)
	GetPaymentMethodStats(w http.ResponseWriter, r *http.Request)

	// Health check
	HealthCheck(w http.ResponseWriter, r *http.Request)

	// Middleware access
	GetJWTManager() *utils.JWTManager
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
	db         *sql.DB
	config     *config.Config
	logger     *logrus.Logger
	jwtManager *utils.JWTManager
	repo       OrderRepository
}

// New creates a new orders handler instance
func New(db *sql.DB, cfg *config.Config, logger *logrus.Logger) (OrdersHandler, error) {
	jwtManager := utils.NewJWTManager(cfg.JWTSecret)

	repo, err := ordersql.NewRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &ordersHandler{
		db:         db,
		config:     cfg,
		logger:     logger,
		jwtManager: jwtManager,
		repo:       repo,
	}, nil
}

func (h *ordersHandler) GetJWTManager() *utils.JWTManager {
	return h.jwtManager
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

	// Calculate tax
	taxAmount := totalAmount * (h.config.DefaultTaxRate / 100)

	// Create order
	order := &models.Order{
		ID:             uuid.New(),
		CustomerID:     req.CustomerID,
		OrderDate:      time.Now(),
		TotalAmount:    totalAmount,
		TaxAmount:      taxAmount,
		DiscountAmount: req.DiscountAmount,
		PaymentMethod:  req.PaymentMethod,
		OrderStatus:    models.OrderStatusPending,
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create ordered recipes
	var items []models.OrderedRecipe
	for _, reqItem := range req.Items {
		totalPrice := float64(reqItem.Quantity) * reqItem.UnitPrice
		item := models.OrderedRecipe{
			ID:                  uuid.New(),
			OrderID:             order.ID,
			RecipeID:            reqItem.RecipeID,
			Quantity:            reqItem.Quantity,
			UnitPrice:           reqItem.UnitPrice,
			TotalPrice:          totalPrice,
			SpecialInstructions: reqItem.SpecialInstructions,
			CreatedAt:           time.Now(),
		}
		items = append(items, item)
	}

	// Calculate final amount (total + tax - discount)
	order.FinalAmount = order.TotalAmount + order.TaxAmount - order.DiscountAmount

	// Save to database
	if err := h.repo.CreateOrder(order, items); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order", err)
		return
	}

	// Get the complete order with calculated final_amount
	createdOrder, err := h.repo.GetOrderWithItems(order.ID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve created order", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id":     order.ID,
		"total_amount": totalAmount,
		"tax_amount":   taxAmount,
		"final_amount": createdOrder.Order.FinalAmount,
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
	if req.OrderStatus != nil {
		validStatuses := []string{models.OrderStatusPending, models.OrderStatusCompleted, models.OrderStatusCancelled}
		valid := false
		for _, status := range validStatuses {
			if *req.OrderStatus == status {
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
		filter.OrderStatus = &status
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
	// Check database connectivity
	if err := h.repo.HealthCheck(); err != nil {
		h.respondWithError(w, http.StatusServiceUnavailable, "Database connection failed", err)
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
