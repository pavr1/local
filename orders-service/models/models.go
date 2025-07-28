package models

import (
	"time"

	"github.com/google/uuid"
)

// Order represents an ice cream order
type Order struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	CustomerID     *uuid.UUID `json:"customer_id" db:"customer_id"`
	OrderDate      time.Time  `json:"order_date" db:"order_date"`
	TotalAmount    float64    `json:"total_amount" db:"total_amount"`
	TaxAmount      float64    `json:"tax_amount" db:"tax_amount"`
	DiscountAmount float64    `json:"discount_amount" db:"discount_amount"`
	FinalAmount    float64    `json:"final_amount" db:"final_amount"`
	PaymentMethod  string     `json:"payment_method" db:"payment_method"`
	OrderStatus    string     `json:"order_status" db:"order_status"`
	Notes          *string    `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// OrderedRecipe represents a recipe item within an order
type OrderedRecipe struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	OrderID             uuid.UUID `json:"order_id" db:"order_id"`
	RecipeID            uuid.UUID `json:"recipe_id" db:"recipe_id"`
	Quantity            int       `json:"quantity" db:"quantity"`
	UnitPrice           float64   `json:"unit_price" db:"unit_price"`
	TotalPrice          float64   `json:"total_price" db:"total_price"`
	SpecialInstructions *string   `json:"special_instructions" db:"special_instructions"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerID     *uuid.UUID                   `json:"customer_id"`
	PaymentMethod  string                       `json:"payment_method"`
	Notes          *string                      `json:"notes"`
	DiscountAmount float64                      `json:"discount_amount"`
	Items          []CreateOrderedRecipeRequest `json:"items"`
}

// CreateOrderedRecipeRequest represents a recipe item in the order creation request
type CreateOrderedRecipeRequest struct {
	RecipeID            uuid.UUID `json:"recipe_id"`
	Quantity            int       `json:"quantity"`
	UnitPrice           float64   `json:"unit_price"`
	SpecialInstructions *string   `json:"special_instructions"`
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	PaymentMethod  *string  `json:"payment_method"`
	OrderStatus    *string  `json:"order_status"`
	Notes          *string  `json:"notes"`
	DiscountAmount *float64 `json:"discount_amount"`
}

// OrderWithItems represents an order with its ordered recipes
type OrderWithItems struct {
	Order Order           `json:"order"`
	Items []OrderedRecipe `json:"items"`
}

// OrderSummary represents a summary of order statistics
type OrderSummary struct {
	TotalOrders     int     `json:"total_orders"`
	PendingOrders   int     `json:"pending_orders"`
	CompletedOrders int     `json:"completed_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageOrder    float64 `json:"average_order"`
}

// PaymentMethodStats represents payment method statistics
type PaymentMethodStats struct {
	PaymentMethod string  `json:"payment_method"`
	Count         int     `json:"count"`
	TotalAmount   float64 `json:"total_amount"`
	Percentage    float64 `json:"percentage"`
}

// OrderFilter represents filters for order queries
type OrderFilter struct {
	CustomerID    *uuid.UUID `json:"customer_id"`
	OrderStatus   *string    `json:"order_status"`
	PaymentMethod *string    `json:"payment_method"`
	DateFrom      *time.Time `json:"date_from"`
	DateTo        *time.Time `json:"date_to"`
	MinAmount     *float64   `json:"min_amount"`
	MaxAmount     *float64   `json:"max_amount"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order"`
}

// Validation methods

// ValidatePaymentMethod checks if payment method is valid
func (o *Order) ValidatePaymentMethod() bool {
	validMethods := []string{"cash", "card", "sinpe"}
	for _, method := range validMethods {
		if o.PaymentMethod == method {
			return true
		}
	}
	return false
}

// ValidateOrderStatus checks if order status is valid
func (o *Order) ValidateOrderStatus() bool {
	validStatuses := []string{"pending", "completed", "cancelled"}
	for _, status := range validStatuses {
		if o.OrderStatus == status {
			return true
		}
	}
	return false
}

// ValidateCreateRequest validates the create order request
func (req *CreateOrderRequest) Validate() error {
	if req.PaymentMethod == "" {
		return &ValidationError{Field: "payment_method", Message: "payment method is required"}
	}

	validMethods := []string{"cash", "card", "sinpe"}
	valid := false
	for _, method := range validMethods {
		if req.PaymentMethod == method {
			valid = true
			break
		}
	}
	if !valid {
		return &ValidationError{Field: "payment_method", Message: "invalid payment method"}
	}

	if len(req.Items) == 0 {
		return &ValidationError{Field: "items", Message: "at least one item is required"}
	}

	for i, item := range req.Items {
		if item.Quantity <= 0 {
			return &ValidationError{Field: "items", Message: "quantity must be greater than 0", Index: &i}
		}
		if item.UnitPrice < 0 {
			return &ValidationError{Field: "items", Message: "unit price cannot be negative", Index: &i}
		}
	}

	if req.DiscountAmount < 0 {
		return &ValidationError{Field: "discount_amount", Message: "discount amount cannot be negative"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Index   *int   `json:"index,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Index != nil {
		return "validation error in " + e.Field + "[" + string(rune(*e.Index)) + "]: " + e.Message
	}
	return "validation error in " + e.Field + ": " + e.Message
}

// Constants for order statuses and payment methods
const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"

	PaymentMethodCash  = "cash"
	PaymentMethodCard  = "card"
	PaymentMethodSinpe = "sinpe"
)
