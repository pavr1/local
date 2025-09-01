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

	"orders-service/models"
	sharedConfig "shared/config"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the interface for database operations
type DBHandlerInterface interface {
	CreateOrder(req models.CreateOrderRequest, logger *logrus.Logger) (*models.OrderWithItems, error)
	GetOrder(id uuid.UUID, logger *logrus.Logger) (*models.OrderWithItems, error)
	UpdateOrder(id uuid.UUID, req *models.UpdateOrderRequest, logger *logrus.Logger) (*models.OrderWithItems, error)
	CancelOrder(id uuid.UUID, logger *logrus.Logger) error
	ListOrders(filter *models.OrderFilter, logger *logrus.Logger) ([]models.Order, int, error)
	GetOrderSummary(logger *logrus.Logger) (*models.OrderSummary, error)
	GetPaymentMethodStats(logger *logrus.Logger) ([]models.PaymentMethodStats, error)
	HealthCheck() error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for existence operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	config    *sharedConfig.Config
	// Invoice service client
	httpClient *http.Client
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface, config *sharedConfig.Config) *HttpHandler {
	return &HttpHandler{
		dbHandler:  dbHandler,
		config:     config,
		httpClient: http.DefaultClient,
	}
}

// CreateOrder handles POST /orders
func (h *HttpHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create order request")

		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON payload",
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		logger.WithError(err).Error("Validation failed for create order request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed",
		})
		return
	}

	// Create order in database
	createdOrder, err := h.dbHandler.CreateOrder(req, logger.Logger)
	if err != nil {
		logger.WithError(err).Error("Failed to create order")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create order",
		})
		return
	}

	// Create income invoice
	// Generate invoice number for income invoice
	invoiceNumber := fmt.Sprintf("INV-%s", createdOrder.Order.OrderNumber)

	// Create invoice request using proper struct
	notes := fmt.Sprintf("Income invoice for order %s", createdOrder.Order.OrderNumber)

	// Create invoice items from order items
	var invoiceItems []map[string]interface{}
	for _, orderItem := range createdOrder.Items {
		invoiceItem := map[string]interface{}{
			"ingredient_id":  nil, // Income invoices don't have ingredients
			"detail":         fmt.Sprintf("Order %s - %s", createdOrder.Order.OrderNumber, orderItem.ProductName),
			"count":          float64(orderItem.Quantity),
			"unit_type":      "Units",
			"items_per_unit": 1,
			"price":          orderItem.ReceipePrice,
		}
		invoiceItems = append(invoiceItems, invoiceItem)
	}

	// Build invoice request, explicitly setting expense_category_id to null
	invoiceReq := map[string]interface{}{
		"id":                  uuid.New().String(),
		"invoice_number":      invoiceNumber,
		"transaction_date":    createdOrder.Order.TransactionTimestamp,
		"transaction_type":    "income",
		"supplier_id":         nil,
		"expense_category_id": nil,                                      // Explicitly set to null for income invoices
		"image_url":           "https://example.com/income-invoice.jpg", // Required field, using placeholder
		"notes":               &notes,
		"items":               invoiceItems,
	}

	// Call invoice service to create income invoice
	invoiceResp, err := h.createIncomeInvoice(invoiceReq, logger.Logger)
	if err != nil {
		logger.WithError(err).WithField("order_id", createdOrder.Order.ID).Error("Failed to create income invoice")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create invoice",
		})
		return
	}

	// Update order with invoice details
	if invoiceResp != nil && invoiceResp.Success {
		updateReq := &models.UpdateOrderRequest{
			InvoiceNumber: &invoiceResp.Data.InvoiceNumber,
			InvoiceURL:    &invoiceResp.Data.ImageURL, // Use image_url as invoice URL
		}

		updatedOrder, err := h.dbHandler.UpdateOrder(createdOrder.Order.ID, updateReq, logger.Logger)
		if err != nil {
			logger.WithError(err).WithField("order_id", createdOrder.Order.ID).Error("Failed to update order with invoice details")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusInternalServerError,
				Message: "Failed to update order with invoice details",
			})
			return
		}
		createdOrder = updatedOrder
	}

	logger.WithFields(logrus.Fields{
		"order_id":     createdOrder.Order.ID,
		"order_number": createdOrder.Order.OrderNumber,
		"total_amount": createdOrder.Order.TotalAmount,
		"iva_amount":   createdOrder.Order.IvaAmount,
		"service_tax":  createdOrder.Order.ServiceTaxAmount,
	}).Info("Order created successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusCreated,
		Message: "Order created successfully",
		Data:    createdOrder,
	})
}

// GetOrder handles GET /orders/{id}
func (h *HttpHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		logger.WithError(err).Error("Invalid order ID")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid order ID",
		})
		return
	}

	order, err := h.dbHandler.GetOrder(orderID, logger.Logger)
	if err != nil {
		//pvillalobos - hardcoded error message
		if strings.Contains(err.Error(), "not found") {
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Order not found",
			})
			return
		}
		logger.WithError(err).Error("Failed to retrieve order")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve order",
		})
		return
	}

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Order retrieved successfully",
		Data:    order,
	})
}

// UpdateOrder handles PUT /orders/{id}
func (h *HttpHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		logger.WithError(err).Error("Invalid order ID")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid order ID",
		})
		return
	}

	var req models.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update order request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON payload",
		})
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
			logger.WithError(err).Error("Invalid payment method")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid payment method",
			})
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
			logger.WithError(err).Error("Invalid order status")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid order status",
			})
			return
		}
	}

	// Update order
	updatedOrder, err := h.dbHandler.UpdateOrder(orderID, &req, logger.Logger)
	if err != nil {
		//pvillalobos - hardcoded error message
		if strings.Contains(err.Error(), "not found") {
			logger.WithError(err).Error("Order not found")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Order not found",
			})
			return
		}
		logger.WithError(err).Error("Failed to update order")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update order",
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"order_id": orderID,
	}).Info("Order updated successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Order updated successfully",
		Data:    updatedOrder,
	})
}

// CancelOrder handles DELETE /orders/{id}
func (h *HttpHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		logger.WithError(err).Error("Invalid order ID")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid order ID",
		})
		return
	}

	if err := h.dbHandler.CancelOrder(orderID, logger.Logger); err != nil {
		//pvillalobos - hardcoded error message
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "completed") {
			logger.WithError(err).Error("Order cannot be cancelled")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Order cannot be cancelled",
			})
			return
		}
		logger.WithError(err).Error("Failed to cancel order")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to cancel order",
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"order_id": orderID,
	}).Info("Order cancelled successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Order cancelled successfully",
		Data: map[string]interface{}{
			"order_id": orderID,
			"status":   "cancelled",
		},
	})
}

// ListOrders handles GET /orders
func (h *HttpHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	filter := &models.OrderFilter{}

	// Parse query parameters
	query := r.URL.Query()

	// Customer ID filter
	if customerIDStr := query.Get("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			logger.WithError(err).Error("Invalid customer ID")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid customer ID",
			})
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
	//pvillalobos - hardcoded error message
	if dateFromStr := query.Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			logger.WithError(err).Error("Invalid date_from format, use YYYY-MM-DD")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid date_from format, use YYYY-MM-DD",
			})
			return
		}
		filter.DateFrom = &dateFrom
	}

	//pvillalobos - hardcoded error message
	if dateToStr := query.Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			logger.WithError(err).Error("Invalid date_to format, use YYYY-MM-DD")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid date_to format, use YYYY-MM-DD",
			})
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
			logger.WithError(err).Error("Invalid min_amount")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid min_amount",
			})
			return
		}
		filter.MinAmount = &minAmount
	}

	if maxAmountStr := query.Get("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseFloat(maxAmountStr, 64)
		if err != nil {
			logger.WithError(err).Error("Invalid max_amount")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid max_amount",
			})
			return
		}
		filter.MaxAmount = &maxAmount
	}

	//pvillalobos - not sure we will implement limits and pagination
	// Pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			logger.WithError(err).Error("Invalid limit")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid limit",
			})
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // default
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			logger.WithError(err).Error("Invalid offset")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid offset",
			})
			return
		}
		filter.Offset = offset
	}

	//pvillalobos - this doesnt look too dynamic
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
			logger.Error("Invalid sort_by field")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid sort_by field",
			})
			return
		}
		filter.SortBy = sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		if sortOrder != "asc" && sortOrder != "desc" {
			logger.Error("Invalid sort_order, use 'asc' or 'desc'")
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid sort_order, use 'asc' or 'desc'",
			})
			return
		}
		filter.SortOrder = sortOrder
	}

	// Get orders
	orders, totalCount, err := h.dbHandler.ListOrders(filter, logger.Logger)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve orders")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve orders",
		})
		return
	}

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": totalCount,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
		"has_more":    filter.Offset+len(orders) < totalCount,
	}

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Orders retrieved successfully",
		Data:    response,
	})
}

// GetOrdersByDateRange handles GET /orders/by-date-range
func (h *HttpHandler) GetOrdersByDateRange(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	logger.WithFields(logrus.Fields{
		"endpoint": "/orders/by-date-range",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Get orders by date range requested")

	// Parse query parameters
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" || dateToStr == "" {
		logger.Error("Both date_from and date_to parameters are required")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Both date_from and date_to parameters are required",
		})
		return
	}

	// Parse dates
	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		logger.WithError(err).WithField("date_from", dateFromStr).Warn("Invalid date_from format")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid date_from format. Use YYYY-MM-DD",
		})
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		logger.WithError(err).WithField("date_to", dateToStr).Warn("Invalid date_to format")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid date_to format. Use YYYY-MM-DD",
		})
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
	orders, total, err := h.dbHandler.ListOrders(filter, logger.Logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get orders by date range")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get orders",
		})
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

	logger.WithFields(logrus.Fields{
		"orders_count": len(orders),
		"total":        total,
		"date_from":    dateFromStr,
		"date_to":      dateToStr,
	}).Info("Successfully retrieved orders by date range")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Orders retrieved successfully",
		Data:    response,
	})
}

// GetOrderSummary handles GET /orders/summary
func (h *HttpHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	logger.WithFields(logrus.Fields{
		"endpoint": "/orders/summary",
		"method":   r.Method,
		"remote":   r.RemoteAddr,
	}).Info("Get order summary requested")

	summary, err := h.dbHandler.GetOrderSummary(logger.Logger)
	if err != nil {
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve order summary",
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"total_orders":     summary.TotalOrders,
		"pending_orders":   summary.PendingOrders,
		"completed_orders": summary.CompletedOrders,
		"total_revenue":    summary.TotalRevenue,
	}).Info("Order summary retrieved successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Order summary retrieved successfully",
		Data:    summary,
	})
}

// GetPaymentMethodStats handles GET /orders/payment-method-stats
func (h *HttpHandler) GetPaymentMethodStats(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	stats, err := h.dbHandler.GetPaymentMethodStats(logger.Logger)
	if err != nil {
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve payment method stats",
		})
		return
	}

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Payment method stats retrieved successfully",
		Data:    stats,
	})
}

// HealthCheck handles GET /health
func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_ORDERS_SERVICE)

	// Check database health
	if err := h.dbHandler.HealthCheck(); err != nil {
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusServiceUnavailable,
			Message: "Database connection failed",
		})
		return
	}

	// Check data-service health (which checks database connectivity)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	dataServiceHealthURL := fmt.Sprintf("%s/api/v1/data/p/health", h.config.GetString("DATA_SERVICE_URL"))
	resp, err := client.Get(dataServiceHealthURL)
	if err != nil {
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusServiceUnavailable,
			Message: "Data service connection failed",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithField("status_code", resp.StatusCode).Error("Data service health check failed")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
			Code:    http.StatusServiceUnavailable,
			Message: "Data service is unhealthy",
		})
		return
	}

	response := map[string]interface{}{
		"service": "orders-service",
		"status":  "healthy",
		"time":    time.Now(),
		"version": "1.0.0",
	}

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_ORDERS_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Orders service is healthy",
		Data:    response,
	})
}

// createIncomeInvoice calls the invoice service to create an income invoice
func (h *HttpHandler) createIncomeInvoice(invoiceReq map[string]interface{}, logger *logrus.Logger) (*models.InvoiceResponse, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice data: %w", err)
	}

	// Debug: Log the JSON being sent
	logger.WithFields(logrus.Fields{
		"invoice_data": string(jsonData),
		"invoice_url":  fmt.Sprintf("%s/api/v1/invoices", h.config.GetString("INVOICE_SERVICE_URL")),
	}).Info("Sending invoice request to invoice service")

	// Call invoice service using configured URL
	invoiceURL := fmt.Sprintf("%s/api/v1/invoices", h.config.GetString("INVOICE_SERVICE_URL"))
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
	var invoiceResp models.InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to decode invoice response: %w", err)
	}

	return &invoiceResp, nil
}
