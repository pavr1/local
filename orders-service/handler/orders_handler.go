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
	InvoiceNumber string `json:"invoice_number"`
	InvoiceURL    string `json:"invoice_url"`
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", rand.Intn(10000))
	return fmt.Sprintf("ORD-%s-%s", timestamp, random)
}
