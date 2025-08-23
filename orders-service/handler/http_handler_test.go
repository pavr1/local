package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"orders-service/config"
	"orders-service/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOrderDBHandler implements OrderDBHandlerInterface for testing
type mockOrderDBHandler struct {
	shouldError  bool
	errorMessage string
	orders       map[uuid.UUID]*models.OrderWithItems
}

func newMockOrderDBHandler() *mockOrderDBHandler {
	return &mockOrderDBHandler{
		orders: make(map[uuid.UUID]*models.OrderWithItems),
	}
}

func (m *mockOrderDBHandler) CreateOrder(req models.CreateOrderRequest) (*models.OrderWithItems, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	orderID := uuid.New()
	customerID := uuid.New()
	order := &models.OrderWithItems{
		Order: models.Order{
			ID:                    orderID,
			OrderNumber:           "ORD-20240101-0001",
			CustomerID:            &customerID,
			Status:                models.OrderStatusPending,
			PaymentMethod:         req.PaymentMethod,
			SubtotalAmount:        20.0,
			IvaAmount:             2.6,
			ServiceTaxAmount:      2.0,
			TotalAmount:           24.6,
			TransactionTimestamp:  time.Now(),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		},
		Items: []models.OrderedRecipe{},
	}
	
	m.orders[orderID] = order
	return order, nil
}

func (m *mockOrderDBHandler) GetOrder(id uuid.UUID) (*models.OrderWithItems, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	order, exists := m.orders[id]
	if !exists {
		return nil, assert.AnError
	}
	return order, nil
}

func (m *mockOrderDBHandler) UpdateOrder(id uuid.UUID, req *models.UpdateOrderRequest) (*models.OrderWithItems, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	order, exists := m.orders[id]
	if !exists {
		return nil, assert.AnError
	}
	
	if req.Status != nil {
		order.Order.Status = *req.Status
	}
	if req.PaymentMethod != nil {
		order.Order.PaymentMethod = *req.PaymentMethod
	}
	
	return order, nil
}

func (m *mockOrderDBHandler) CancelOrder(id uuid.UUID) error {
	if m.shouldError {
		return assert.AnError
	}
	
	order, exists := m.orders[id]
	if !exists {
		return assert.AnError
	}
	
	order.Order.Status = models.OrderStatusCancelled
	return nil
}

func (m *mockOrderDBHandler) ListOrders(filter *models.OrderFilter) ([]models.Order, int, error) {
	if m.shouldError {
		return nil, 0, assert.AnError
	}
	
	var orders []models.Order
	for _, order := range m.orders {
		orders = append(orders, order.Order)
	}
	return orders, len(orders), nil
}

func (m *mockOrderDBHandler) GetOrderSummary() (*models.OrderSummary, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	return &models.OrderSummary{
		TotalOrders:     2,
		PendingOrders:   1,
		CompletedOrders: 1,
		TotalRevenue:    100.0,
	}, nil
}

func (m *mockOrderDBHandler) GetPaymentMethodStats() ([]models.PaymentMethodStats, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	return []models.PaymentMethodStats{
		{PaymentMethod: "cash", Count: 5, TotalAmount: 100.0},
		{PaymentMethod: "card", Count: 3, TotalAmount: 75.0},
	}, nil
}

func (m *mockOrderDBHandler) HealthCheck() error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

// Test helper functions
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

func createTestConfig() *config.Config {
	return &config.Config{
		DefaultTaxRate: 13.0,
	}
}

// TestNewOrderHTTPHandler tests the creation of a new OrderHTTPHandler
func TestNewOrderHTTPHandler(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)
	require.NotNil(t, handler)
	assert.Equal(t, dbHandler, handler.dbHandler)
	assert.Equal(t, cfg, handler.config)
	assert.Equal(t, logger, handler.logger)
}

// TestOrderHTTPHandler_CreateOrder tests order creation via HTTP
func TestOrderHTTPHandler_CreateOrder(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	customerID := uuid.New()
	reqBody := models.CreateOrderRequest{
		CustomerID:    &customerID,
		PaymentMethod: "cash",
		Items: []models.CreateOrderedRecipeRequest{
			{
				RecipeID:  uuid.New(),
				Quantity:  2,
				UnitPrice: 10.0,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Order created successfully", response["message"])
}

// TestOrderHTTPHandler_GetOrder tests order retrieval via HTTP
func TestOrderHTTPHandler_GetOrder(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	// Create a test order first
	orderID := uuid.New()
	customerID := uuid.New()
	testOrder := &models.OrderWithItems{
		Order: models.Order{
			ID:          orderID,
			OrderNumber: "ORD-20240101-0001",
			CustomerID:  &customerID,
			Status:      models.OrderStatusPending,
		},
		Items: []models.OrderedRecipe{},
	}
	dbHandler.orders[orderID] = testOrder

	req := httptest.NewRequest("GET", "/orders/"+orderID.String(), nil)
	w := httptest.NewRecorder()

	// Set up router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/orders/{id}", handler.GetOrder).Methods("GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Order retrieved successfully", response["message"])
}

// TestOrderHTTPHandler_UpdateOrder tests order update via HTTP
func TestOrderHTTPHandler_UpdateOrder(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	// Create a test order first
	orderID := uuid.New()
	customerID := uuid.New()
	testOrder := &models.OrderWithItems{
		Order: models.Order{
			ID:          orderID,
			OrderNumber: "ORD-20240101-0001",
			CustomerID:  &customerID,
			Status:      models.OrderStatusPending,
		},
		Items: []models.OrderedRecipe{},
	}
	dbHandler.orders[orderID] = testOrder

	updateReq := models.UpdateOrderRequest{
		Status: &[]string{models.OrderStatusCompleted}[0],
	}

	jsonBody, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/orders/"+orderID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Set up router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/orders/{id}", handler.UpdateOrder).Methods("PUT")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Order updated successfully", response["message"])
}

// TestOrderHTTPHandler_CancelOrder tests order cancellation via HTTP
func TestOrderHTTPHandler_CancelOrder(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	// Create a test order first
	orderID := uuid.New()
	customerID := uuid.New()
	testOrder := &models.OrderWithItems{
		Order: models.Order{
			ID:          orderID,
			OrderNumber: "ORD-20240101-0001",
			CustomerID:  &customerID,
			Status:      models.OrderStatusPending,
		},
		Items: []models.OrderedRecipe{},
	}
	dbHandler.orders[orderID] = testOrder

	req := httptest.NewRequest("DELETE", "/orders/"+orderID.String(), nil)
	w := httptest.NewRecorder()

	// Set up router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/orders/{id}", handler.CancelOrder).Methods("DELETE")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Order cancelled successfully", response["message"])
}

// TestOrderHTTPHandler_ListOrders tests order listing via HTTP
func TestOrderHTTPHandler_ListOrders(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Orders retrieved successfully", response["message"])
}

// TestOrderHTTPHandler_GetOrderSummary tests order summary via HTTP
func TestOrderHTTPHandler_GetOrderSummary(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("GET", "/orders/summary", nil)
	w := httptest.NewRecorder()

	handler.GetOrderSummary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Order summary retrieved successfully", response["message"])
}

// TestOrderHTTPHandler_GetPaymentMethodStats tests payment method stats via HTTP
func TestOrderHTTPHandler_GetPaymentMethodStats(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("GET", "/orders/payment-method-stats", nil)
	w := httptest.NewRecorder()

	handler.GetPaymentMethodStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Payment method stats retrieved successfully", response["message"])
}

// TestOrderHTTPHandler_HealthCheck tests health check via HTTP
func TestOrderHTTPHandler_HealthCheck(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Orders service is healthy", response["message"])
}

// TestOrderHTTPHandler_InvalidJSON tests invalid JSON handling
func TestOrderHTTPHandler_InvalidJSON(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Invalid JSON payload", response["message"])
}

// TestOrderHTTPHandler_InvalidOrderID tests invalid order ID handling
func TestOrderHTTPHandler_InvalidOrderID(t *testing.T) {
	dbHandler := newMockOrderDBHandler()
	cfg := createTestConfig()
	logger := createTestLogger()

	handler := NewOrderHTTPHandler(dbHandler, cfg, logger)

	req := httptest.NewRequest("GET", "/orders/invalid-uuid", nil)
	w := httptest.NewRecorder()

	// Set up router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/orders/{id}", handler.GetOrder).Methods("GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Invalid order ID", response["message"])
}
