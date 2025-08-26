package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// OrdersHandler interface defines the HTTP handlers for orders
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
}

// InvoiceResponse represents the response from invoice service
type InvoiceResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID                string   `json:"id"`
		InvoiceNumber     string   `json:"invoice_number"`
		TransactionDate   string   `json:"transaction_date"`
		TransactionType   string   `json:"transaction_type"`
		SupplierID        *string  `json:"supplier_id"`
		ExpenseCategoryID *string  `json:"expense_category_id"`
		TotalAmount       *float64 `json:"total_amount"`
		ImageURL          string   `json:"image_url"`
		Notes             *string  `json:"notes"`
		CreatedAt         string   `json:"created_at"`
		UpdatedAt         string   `json:"updated_at"`
	} `json:"data"`
	Message string `json:"message"`
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", rand.Intn(10000))
	return fmt.Sprintf("ORD-%s-%s", timestamp, random)
}
