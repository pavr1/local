package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"orders-service/config"
	"orders-service/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOrderRepository implements OrderRepository for testing
type mockOrderRepository struct {
	orders         map[uuid.UUID]*models.Order
	orderedRecipes map[uuid.UUID][]models.OrderedRecipe
	shouldError    bool
	errorMessage   string
}

func newMockRepository() *mockOrderRepository {
	return &mockOrderRepository{
		orders:         make(map[uuid.UUID]*models.Order),
		orderedRecipes: make(map[uuid.UUID][]models.OrderedRecipe),
		shouldError:    false,
	}
}

func (m *mockOrderRepository) CreateOrder(order *models.Order, items []models.OrderedRecipe) error {
	if m.shouldError {
		return fmt.Errorf(m.errorMessage)
	}
	m.orders[order.ID] = order
	m.orderedRecipes[order.ID] = items
	return nil
}

func (m *mockOrderRepository) GetOrderByID(id uuid.UUID) (*models.Order, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMessage)
	}
	order, exists := m.orders[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return order, nil
}

func (m *mockOrderRepository) GetOrderWithItems(id uuid.UUID) (*models.OrderWithItems, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMessage)
	}
	order, exists := m.orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	items := m.orderedRecipes[id]
	return &models.OrderWithItems{Order: *order, Items: items}, nil
}

func (m *mockOrderRepository) GetOrderedRecipesByOrderID(orderID uuid.UUID) ([]models.OrderedRecipe, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMessage)
	}
	items, exists := m.orderedRecipes[orderID]
	if !exists {
		return []models.OrderedRecipe{}, nil
	}
	return items, nil
}

func (m *mockOrderRepository) UpdateOrder(id uuid.UUID, updates *models.UpdateOrderRequest) error {
	if m.shouldError {
		return fmt.Errorf(m.errorMessage)
	}
	order, exists := m.orders[id]
	if !exists {
		return fmt.Errorf("order not found")
	}

	// Apply updates
	if updates.PaymentMethod != nil {
		order.PaymentMethod = *updates.PaymentMethod
	}
	if updates.OrderStatus != nil {
		order.OrderStatus = *updates.OrderStatus
	}
	if updates.Notes != nil {
		order.Notes = updates.Notes
	}
	if updates.DiscountAmount != nil {
		order.DiscountAmount = *updates.DiscountAmount
	}
	order.UpdatedAt = time.Now()

	return nil
}

func (m *mockOrderRepository) CancelOrder(id uuid.UUID) error {
	if m.shouldError {
		return fmt.Errorf(m.errorMessage)
	}
	order, exists := m.orders[id]
	if !exists {
		return fmt.Errorf("order not found")
	}
	order.OrderStatus = "cancelled"
	order.UpdatedAt = time.Now()
	return nil
}

func (m *mockOrderRepository) ListOrders(filter *models.OrderFilter) ([]models.Order, int, error) {
	if m.shouldError {
		return nil, 0, fmt.Errorf(m.errorMessage)
	}

	orders := make([]models.Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, *order)
	}

	return orders, len(orders), nil
}

func (m *mockOrderRepository) GetOrderSummary() (*models.OrderSummary, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMessage)
	}

	summary := &models.OrderSummary{
		TotalOrders:     len(m.orders),
		PendingOrders:   0,
		CompletedOrders: 0,
		CancelledOrders: 0,
		TotalRevenue:    0.0,
		AverageOrder:    0.0,
	}

	for _, order := range m.orders {
		switch order.OrderStatus {
		case "pending":
			summary.PendingOrders++
		case "completed":
			summary.CompletedOrders++
			summary.TotalRevenue += order.FinalAmount
		case "cancelled":
			summary.CancelledOrders++
		}
	}

	if summary.CompletedOrders > 0 {
		summary.AverageOrder = summary.TotalRevenue / float64(summary.CompletedOrders)
	}

	return summary, nil
}

func (m *mockOrderRepository) GetPaymentMethodStats() ([]models.PaymentMethodStats, error) {
	if m.shouldError {
		return nil, fmt.Errorf(m.errorMessage)
	}

	stats := []models.PaymentMethodStats{
		{PaymentMethod: "card", Count: 10, TotalAmount: 1500.0, Percentage: 60.0},
		{PaymentMethod: "cash", Count: 8, TotalAmount: 800.0, Percentage: 32.0},
		{PaymentMethod: "sinpe", Count: 2, TotalAmount: 200.0, Percentage: 8.0},
	}

	return stats, nil
}

func (m *mockOrderRepository) HealthCheck() error {
	if m.shouldError {
		return fmt.Errorf(m.errorMessage)
	}
	return nil
}

// setupTestHandler creates a test handler with mock dependencies
func setupTestHandler() (*ordersHandler, *mockOrderRepository) {
	db, _, _ := sqlmock.New()

	cfg := &config.Config{
		DefaultTaxRate:     13.0,
		DefaultServiceRate: 10.0,
		OrderTimeout:       30,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	mockRepo := newMockRepository()

	handler := &ordersHandler{
		db:     db,
		config: cfg,
		logger: logger,
		repo:   mockRepo,
	}

	return handler, mockRepo
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	t.Run("successful health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Orders service is healthy", response["message"])

		// Extract data and check status
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "healthy", data["status"])
	})

	t.Run("health check with database error", func(t *testing.T) {
		mockRepo.shouldError = true
		mockRepo.errorMessage = "database connection failed"

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestCreateOrder tests the create order endpoint
func TestCreateOrder(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	validRequest := models.CreateOrderRequest{
		PaymentMethod:  "cash",
		DiscountAmount: 0,
		Items: []models.CreateOrderedRecipeRequest{
			{
				RecipeID:  uuid.New(),
				Quantity:  2,
				UnitPrice: 25.0,
			},
		},
	}

	t.Run("successful order creation", func(t *testing.T) {
		jsonData, _ := json.Marshal(validRequest)
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrder(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "created successfully")

		// Extract data and verify order details
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)

		order, ok := data["order"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "cash", order["payment_method"])

		items, ok := data["items"].([]interface{})
		require.True(t, ok)
		assert.Len(t, items, 1)
	})

	t.Run("invalid JSON payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrder(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation failure", func(t *testing.T) {
		invalidRequest := models.CreateOrderRequest{
			PaymentMethod:  "bitcoin", // Invalid payment method
			DiscountAmount: 0,
			Items: []models.CreateOrderedRecipeRequest{
				{
					RecipeID:  uuid.New(),
					Quantity:  2,
					UnitPrice: 25.0,
				},
			},
		}

		jsonData, _ := json.Marshal(invalidRequest)
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrder(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.shouldError = true
		mockRepo.errorMessage = "database error"

		jsonData, _ := json.Marshal(validRequest)
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrder(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestGetOrder tests the get order endpoint
func TestGetOrder(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	// Create a test order
	orderID := uuid.New()
	testOrder := &models.Order{
		ID:            orderID,
		OrderDate:     time.Now(),
		TotalAmount:   100.0,
		PaymentMethod: "card",
		OrderStatus:   "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockRepo.orders[orderID] = testOrder

	t.Run("successful order retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orders/"+orderID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"id": orderID.String()})
		w := httptest.NewRecorder()

		handler.GetOrder(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "retrieved successfully")

		// Extract data and verify order details
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)

		order, ok := data["order"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, orderID.String(), order["id"])
	})

	t.Run("invalid order ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orders/invalid-id", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})
		w := httptest.NewRecorder()

		handler.GetOrder(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("order not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/orders/"+nonExistentID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"id": nonExistentID.String()})
		w := httptest.NewRecorder()

		handler.GetOrder(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestUpdateOrder tests the update order endpoint
func TestUpdateOrder(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	// Create a test order
	orderID := uuid.New()
	testOrder := &models.Order{
		ID:            orderID,
		OrderDate:     time.Now(),
		TotalAmount:   100.0,
		PaymentMethod: "cash",
		OrderStatus:   "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockRepo.orders[orderID] = testOrder

	t.Run("successful order update", func(t *testing.T) {
		newPaymentMethod := "card"
		updateRequest := models.UpdateOrderRequest{
			PaymentMethod: &newPaymentMethod,
		}

		jsonData, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest("PUT", "/orders/"+orderID.String(), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": orderID.String()})
		w := httptest.NewRecorder()

		handler.UpdateOrder(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "updated successfully")

		// Extract data and verify order details
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)

		// UpdateOrder returns OrderWithItems structure
		order, ok := data["order"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "card", order["payment_method"])
	})

	t.Run("invalid JSON payload", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/orders/"+orderID.String(), bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": orderID.String()})
		w := httptest.NewRecorder()

		handler.UpdateOrder(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestCancelOrder tests the cancel order endpoint
func TestCancelOrder(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	// Create a test order
	orderID := uuid.New()
	testOrder := &models.Order{
		ID:            orderID,
		OrderDate:     time.Now(),
		TotalAmount:   100.0,
		PaymentMethod: "cash",
		OrderStatus:   "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockRepo.orders[orderID] = testOrder

	t.Run("successful order cancellation", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/orders/"+orderID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"id": orderID.String()})
		w := httptest.NewRecorder()

		handler.CancelOrder(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "cancelled", testOrder.OrderStatus)
	})

	t.Run("invalid order ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/orders/invalid-id", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})
		w := httptest.NewRecorder()

		handler.CancelOrder(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestListOrders tests the list orders endpoint
func TestListOrders(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	// Create test orders
	for i := 0; i < 3; i++ {
		orderID := uuid.New()
		testOrder := &models.Order{
			ID:            orderID,
			OrderDate:     time.Now(),
			TotalAmount:   float64(100 * (i + 1)),
			PaymentMethod: "cash",
			OrderStatus:   "pending",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		mockRepo.orders[orderID] = testOrder
	}

	t.Run("successful orders listing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orders", nil)
		w := httptest.NewRecorder()

		handler.ListOrders(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "retrieved successfully")

		// Extract data and verify orders
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)

		orders, ok := data["orders"].([]interface{})
		require.True(t, ok)
		assert.Len(t, orders, 3)
	})
}

// TestGetOrderSummary tests the order summary endpoint
func TestGetOrderSummary(t *testing.T) {
	handler, mockRepo := setupTestHandler()

	t.Run("successful summary retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orders/summary", nil)
		w := httptest.NewRecorder()

		handler.GetOrderSummary(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "retrieved successfully")

		// Extract data and verify summary
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok)

		totalOrders, ok := data["total_orders"].(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, int(totalOrders), 0)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.shouldError = true
		mockRepo.errorMessage = "database error"

		req := httptest.NewRequest("GET", "/orders/summary", nil)
		w := httptest.NewRecorder()

		handler.GetOrderSummary(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestGetPaymentMethodStats tests the payment method statistics endpoint
func TestGetPaymentMethodStats(t *testing.T) {
	handler, _ := setupTestHandler()

	t.Run("successful stats retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orders/payment-stats", nil)
		w := httptest.NewRecorder()

		handler.GetPaymentMethodStats(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Payment method stats retrieved successfully", response["message"])

		// Extract data and verify it's the expected array
		data, ok := response["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 3) // card, cash, sinpe
	})
}

// BenchmarkCreateOrder benchmarks the create order endpoint
func BenchmarkCreateOrder(b *testing.B) {
	handler, _ := setupTestHandler()

	validRequest := models.CreateOrderRequest{
		PaymentMethod:  "cash",
		DiscountAmount: 0,
		Items: []models.CreateOrderedRecipeRequest{
			{
				RecipeID:  uuid.New(),
				Quantity:  2,
				UnitPrice: 25.0,
			},
		},
	}

	jsonData, _ := json.Marshal(validRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrder(w, req)
	}
}

// BenchmarkGetOrder benchmarks the get order endpoint
func BenchmarkGetOrder(b *testing.B) {
	handler, mockRepo := setupTestHandler()

	// Create a test order
	orderID := uuid.New()
	testOrder := &models.Order{
		ID:            orderID,
		OrderDate:     time.Now(),
		TotalAmount:   100.0,
		PaymentMethod: "card",
		OrderStatus:   "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockRepo.orders[orderID] = testOrder

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/orders/"+orderID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"id": orderID.String()})
		w := httptest.NewRecorder()

		handler.GetOrder(w, req)
	}
}
