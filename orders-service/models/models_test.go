package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrder_ValidatePaymentMethod(t *testing.T) {
	tests := map[string]struct {
		paymentMethod string
		expected      bool
	}{
		"valid cash": {
			paymentMethod: "cash",
			expected:      true,
		},
		"valid card": {
			paymentMethod: "card",
			expected:      true,
		},
		"valid sinpe": {
			paymentMethod: "sinpe",
			expected:      true,
		},
		"invalid method": {
			paymentMethod: "invalid",
			expected:      false,
		},
		"empty method": {
			paymentMethod: "",
			expected:      false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			order := &Order{PaymentMethod: tc.paymentMethod}
			result := order.ValidatePaymentMethod()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOrder_ValidateOrderStatus(t *testing.T) {
	tests := map[string]struct {
		status   string
		expected bool
	}{
		"valid pending": {
			status:   "pending",
			expected: true,
		},
		"valid completed": {
			status:   "completed",
			expected: true,
		},
		"valid cancelled": {
			status:   "cancelled",
			expected: true,
		},
		"valid sinpe_pending": {
			status:   "sinpe_pending",
			expected: true,
		},
		"invalid status": {
			status:   "invalid",
			expected: false,
		},
		"empty status": {
			status:   "",
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			order := &Order{Status: tc.status}
			result := order.ValidateOrderStatus()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateOrderRequest_Validate(t *testing.T) {
	validUUID := uuid.New()

	tests := map[string]struct {
		request     *CreateOrderRequest
		expectError bool
		errorField  string
	}{
		"valid request": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  2,
						UnitPrice: 10.50,
					},
				},
				DiscountAmount: 0,
			},
			expectError: false,
		},
		"missing payment method": {
			request: &CreateOrderRequest{
				PaymentMethod: "",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  1,
						UnitPrice: 10.50,
					},
				},
			},
			expectError: true,
			errorField:  "payment_method",
		},
		"invalid payment method": {
			request: &CreateOrderRequest{
				PaymentMethod: "invalid",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  1,
						UnitPrice: 10.50,
					},
				},
			},
			expectError: true,
			errorField:  "payment_method",
		},
		"empty items": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items:         []CreateOrderedRecipeRequest{},
			},
			expectError: true,
			errorField:  "items",
		},
		"nil items": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items:         nil,
			},
			expectError: true,
			errorField:  "items",
		},
		"invalid quantity": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  0,
						UnitPrice: 10.50,
					},
				},
			},
			expectError: true,
			errorField:  "items",
		},
		"negative quantity": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  -1,
						UnitPrice: 10.50,
					},
				},
			},
			expectError: true,
			errorField:  "items",
		},
		"negative unit price": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  1,
						UnitPrice: -10.50,
					},
				},
			},
			expectError: true,
			errorField:  "items",
		},
		"negative discount amount": {
			request: &CreateOrderRequest{
				PaymentMethod: "cash",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  1,
						UnitPrice: 10.50,
					},
				},
				DiscountAmount: -5.0,
			},
			expectError: true,
			errorField:  "discount_amount",
		},
		"valid with discount": {
			request: &CreateOrderRequest{
				PaymentMethod: "card",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  3,
						UnitPrice: 15.00,
					},
				},
				DiscountAmount: 5.0,
			},
			expectError: false,
		},
		"valid sinpe payment": {
			request: &CreateOrderRequest{
				PaymentMethod: "sinpe",
				Items: []CreateOrderedRecipeRequest{
					{
						RecipeID:  validUUID,
						Quantity:  1,
						UnitPrice: 20.00,
					},
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectError {
				assert.Error(t, err)
				if validationErr, ok := err.(*ValidationError); ok {
					assert.Equal(t, tc.errorField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := map[string]struct {
		validationError *ValidationError
		expected        string
	}{
		"error without index": {
			validationError: &ValidationError{
				Field:   "payment_method",
				Message: "invalid payment method",
			},
			expected: "validation error in payment_method: invalid payment method",
		},
		"error with index": {
			validationError: &ValidationError{
				Field:   "items",
				Message: "quantity must be greater than 0",
				Index:   &[]int{0}[0],
			},
			expected: "validation error in items[\x00]: quantity must be greater than 0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.validationError.Error()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOrder_Struct(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	salesRepID := uuid.New()
	now := time.Now()
	transactionRef := "REF123"
	sinpeURL := "https://example.com/screenshot.jpg"
	invoiceNum := "INV001"
	invoiceURL := "https://example.com/invoice.pdf"
	completedAt := now.Add(time.Hour)

	order := Order{
		ID:                    orderID,
		OrderNumber:           "ORD-001",
		CustomerID:            &customerID,
		SalesRepresentativeID: &salesRepID,
		Status:                "pending",
		PaymentMethod:         "cash",
		TransactionReference:  &transactionRef,
		SinpeScreenshotURL:    &sinpeURL,
		SubtotalAmount:        100.00,
		DiscountAmount:        10.00,
		IvaAmount:             13.00,
		ServiceTaxAmount:      5.00,
		TotalAmount:           108.00,
		InvoiceNumber:         &invoiceNum,
		InvoiceURL:            &invoiceURL,
		TransactionTimestamp:  now,
		CompletedAt:           &completedAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, "ORD-001", order.OrderNumber)
	assert.Equal(t, &customerID, order.CustomerID)
	assert.Equal(t, &salesRepID, order.SalesRepresentativeID)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, "cash", order.PaymentMethod)
	assert.Equal(t, &transactionRef, order.TransactionReference)
	assert.Equal(t, &sinpeURL, order.SinpeScreenshotURL)
	assert.Equal(t, 100.00, order.SubtotalAmount)
	assert.Equal(t, 10.00, order.DiscountAmount)
	assert.Equal(t, 13.00, order.IvaAmount)
	assert.Equal(t, 5.00, order.ServiceTaxAmount)
	assert.Equal(t, 108.00, order.TotalAmount)
	assert.Equal(t, &invoiceNum, order.InvoiceNumber)
	assert.Equal(t, &invoiceURL, order.InvoiceURL)
	assert.Equal(t, now, order.TransactionTimestamp)
	assert.Equal(t, &completedAt, order.CompletedAt)
	assert.Equal(t, now, order.CreatedAt)
	assert.Equal(t, now, order.UpdatedAt)
}

func TestOrderedRecipe_Struct(t *testing.T) {
	orderID := uuid.New()
	recipeID := uuid.New()

	orderedRecipe := OrderedRecipe{
		ID:           uuid.New(),
		OrderID:      orderID,
		RecipeID:     recipeID,
		ProductName:  "Vanilla Ice Cream",
		Quantity:     2,
		ReceipePrice: 15.00,
		Subtotal:     30.00,
	}

	assert.Equal(t, orderID, orderedRecipe.OrderID)
	assert.Equal(t, recipeID, orderedRecipe.RecipeID)
	assert.Equal(t, "Vanilla Ice Cream", orderedRecipe.ProductName)
	assert.Equal(t, 2, orderedRecipe.Quantity)
	assert.Equal(t, 15.00, orderedRecipe.ReceipePrice)
	assert.Equal(t, 30.00, orderedRecipe.Subtotal)
}

func TestCreateOrderedRecipeRequest_Struct(t *testing.T) {
	recipeID := uuid.New()

	request := CreateOrderedRecipeRequest{
		RecipeID:  recipeID,
		Quantity:  3,
		UnitPrice: 12.50,
	}

	assert.Equal(t, recipeID, request.RecipeID)
	assert.Equal(t, 3, request.Quantity)
	assert.Equal(t, 12.50, request.UnitPrice)
}

func TestUpdateOrderRequest_Struct(t *testing.T) {
	paymentMethod := "card"
	status := "completed"
	notes := "Delivered to customer"
	discountAmount := 5.00
	invoiceNumber := "INV002"
	invoiceURL := "https://example.com/invoice2.pdf"

	request := UpdateOrderRequest{
		PaymentMethod:  &paymentMethod,
		Status:         &status,
		Notes:          &notes,
		DiscountAmount: &discountAmount,
		InvoiceNumber:  &invoiceNumber,
		InvoiceURL:     &invoiceURL,
	}

	assert.Equal(t, &paymentMethod, request.PaymentMethod)
	assert.Equal(t, &status, request.Status)
	assert.Equal(t, &notes, request.Notes)
	assert.Equal(t, &discountAmount, request.DiscountAmount)
	assert.Equal(t, &invoiceNumber, request.InvoiceNumber)
	assert.Equal(t, &invoiceURL, request.InvoiceURL)
}

func TestOrderWithItems_Struct(t *testing.T) {
	order := Order{
		ID:          uuid.New(),
		OrderNumber: "ORD-001",
		Status:      "pending",
	}

	items := []OrderedRecipe{
		{
			ID:           uuid.New(),
			OrderID:      order.ID,
			RecipeID:     uuid.New(),
			ProductName:  "Chocolate Ice Cream",
			Quantity:     2,
			ReceipePrice: 18.00,
			Subtotal:     36.00,
		},
	}

	orderWithItems := OrderWithItems{
		Order: order,
		Items: items,
	}

	assert.Equal(t, order, orderWithItems.Order)
	assert.Equal(t, items, orderWithItems.Items)
	assert.Len(t, orderWithItems.Items, 1)
}

func TestOrderSummary_Struct(t *testing.T) {
	summary := OrderSummary{
		TotalOrders:     100,
		PendingOrders:   20,
		CompletedOrders: 70,
		CancelledOrders: 10,
		TotalRevenue:    5000.00,
		AverageOrder:    50.00,
	}

	assert.Equal(t, 100, summary.TotalOrders)
	assert.Equal(t, 20, summary.PendingOrders)
	assert.Equal(t, 70, summary.CompletedOrders)
	assert.Equal(t, 10, summary.CancelledOrders)
	assert.Equal(t, 5000.00, summary.TotalRevenue)
	assert.Equal(t, 50.00, summary.AverageOrder)
}

func TestPaymentMethodStats_Struct(t *testing.T) {
	stats := PaymentMethodStats{
		PaymentMethod: "cash",
		Count:         50,
		TotalAmount:   2500.00,
		Percentage:    50.0,
	}

	assert.Equal(t, "cash", stats.PaymentMethod)
	assert.Equal(t, 50, stats.Count)
	assert.Equal(t, 2500.00, stats.TotalAmount)
	assert.Equal(t, 50.0, stats.Percentage)
}

func TestOrderFilter_Struct(t *testing.T) {
	customerID := uuid.New()
	status := "pending"
	paymentMethod := "card"
	dateFrom := time.Now().AddDate(0, -1, 0)
	dateTo := time.Now()
	minAmount := 10.00
	maxAmount := 100.00

	filter := OrderFilter{
		CustomerID:    &customerID,
		Status:        &status,
		PaymentMethod: &paymentMethod,
		DateFrom:      &dateFrom,
		DateTo:        &dateTo,
		MinAmount:     &minAmount,
		MaxAmount:     &maxAmount,
		Limit:         20,
		Offset:        0,
		SortBy:        "created_at",
		SortOrder:     "desc",
	}

	assert.Equal(t, &customerID, filter.CustomerID)
	assert.Equal(t, &status, filter.Status)
	assert.Equal(t, &paymentMethod, filter.PaymentMethod)
	assert.Equal(t, &dateFrom, filter.DateFrom)
	assert.Equal(t, &dateTo, filter.DateTo)
	assert.Equal(t, &minAmount, filter.MinAmount)
	assert.Equal(t, &maxAmount, filter.MaxAmount)
	assert.Equal(t, 20, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
	assert.Equal(t, "created_at", filter.SortBy)
	assert.Equal(t, "desc", filter.SortOrder)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "pending", OrderStatusPending)
	assert.Equal(t, "completed", OrderStatusCompleted)
	assert.Equal(t, "cancelled", OrderStatusCancelled)
	assert.Equal(t, "sinpe_pending", OrderStatusSinpePending)

	assert.Equal(t, "cash", PaymentMethodCash)
	assert.Equal(t, "card", PaymentMethodCard)
	assert.Equal(t, "sinpe", PaymentMethodSinpe)
}
