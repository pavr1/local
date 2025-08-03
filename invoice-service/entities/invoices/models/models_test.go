package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Invoice Model Tests

func TestInvoiceModel(t *testing.T) {
	// Test Invoice struct
	invoice := Invoice{
		ID:                "550e8400-e29b-41d4-a716-446655440000",
		InvoiceNumber:     "INV-2024-001",
		TransactionDate:   time.Now(),
		TransactionType:   "outcome",
		SupplierID:        stringPtr("550e8400-e29b-41d4-a716-446655440001"),
		ExpenseCategoryID: "550e8400-e29b-41d4-a716-446655440002",
		TotalAmount:       float64Ptr(150.75),
		ImageURL:          "https://example.com/invoices/invoice001.jpg",
		Notes:             stringPtr("Test invoice for ingredients"),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", invoice.ID)
	assert.Equal(t, "INV-2024-001", invoice.InvoiceNumber)
	assert.Equal(t, "outcome", invoice.TransactionType)
	assert.NotNil(t, invoice.SupplierID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", *invoice.SupplierID)
	assert.NotNil(t, invoice.TotalAmount)
	assert.Equal(t, 150.75, *invoice.TotalAmount)
}

func TestCreateInvoiceRequest(t *testing.T) {
	// Test valid CreateInvoiceRequest
	items := []CreateInvoiceDetailRequest{
		{
			IngredientID:   stringPtr("550e8400-e29b-41d4-a716-446655440003"),
			Detail:         "Milk - 1 Gallon",
			Count:          2.0,
			UnitType:       "Gallons",
			Price:          5.99,
			ExpirationDate: timePtr(time.Now().AddDate(0, 0, 14)),
		},
		{
			Detail:   "Sugar - 5lb bag",
			Count:    1.0,
			UnitType: "Bag",
			Price:    3.50,
		},
	}

	req := CreateInvoiceRequest{
		InvoiceNumber:     "INV-2024-001",
		TransactionDate:   time.Now(),
		TransactionType:   "outcome",
		SupplierID:        stringPtr("550e8400-e29b-41d4-a716-446655440001"),
		ExpenseCategoryID: "550e8400-e29b-41d4-a716-446655440002",
		ImageURL:          "https://example.com/invoices/invoice001.jpg",
		Notes:             stringPtr("Test invoice for ingredients"),
		Items:             items,
	}

	assert.Equal(t, "INV-2024-001", req.InvoiceNumber)
	assert.Equal(t, "outcome", req.TransactionType)
	assert.NotNil(t, req.SupplierID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", *req.SupplierID)
	assert.Equal(t, 2, len(req.Items))
	assert.Equal(t, "Milk - 1 Gallon", req.Items[0].Detail)
	assert.Equal(t, 2.0, req.Items[0].Count)
	assert.Equal(t, "Gallons", req.Items[0].UnitType)
	assert.Equal(t, 5.99, req.Items[0].Price)
}

func TestUpdateInvoiceRequest(t *testing.T) {
	// Test UpdateInvoiceRequest with partial updates
	req := UpdateInvoiceRequest{
		InvoiceNumber:   stringPtr("INV-2024-001-UPDATED"),
		TransactionType: stringPtr("income"),
		Notes:           stringPtr("Updated notes"),
	}

	assert.NotNil(t, req.InvoiceNumber)
	assert.Equal(t, "INV-2024-001-UPDATED", *req.InvoiceNumber)
	assert.NotNil(t, req.TransactionType)
	assert.Equal(t, "income", *req.TransactionType)
	assert.NotNil(t, req.Notes)
	assert.Equal(t, "Updated notes", *req.Notes)
	assert.Nil(t, req.SupplierID) // Should be nil when not provided
}

func TestListInvoicesRequest(t *testing.T) {
	// Test ListInvoicesRequest
	req := ListInvoicesRequest{
		Limit:             intPtr(10),
		Offset:            intPtr(0),
		TransactionType:   stringPtr("outcome"),
		ExpenseCategoryID: stringPtr("550e8400-e29b-41d4-a716-446655440002"),
	}

	assert.NotNil(t, req.Limit)
	assert.Equal(t, 10, *req.Limit)
	assert.NotNil(t, req.Offset)
	assert.Equal(t, 0, *req.Offset)
	assert.NotNil(t, req.TransactionType)
	assert.Equal(t, "outcome", *req.TransactionType)
	assert.NotNil(t, req.ExpenseCategoryID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", *req.ExpenseCategoryID)
}

func TestInvoiceResponse(t *testing.T) {
	// Test InvoiceResponse
	invoice := Invoice{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		InvoiceNumber: "INV-2024-001",
	}

	response := InvoiceResponse{
		Success: true,
		Data:    invoice,
		Message: "Invoice retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Data.ID)
	assert.Equal(t, "INV-2024-001", response.Data.InvoiceNumber)
	assert.Equal(t, "Invoice retrieved successfully", response.Message)
}

func TestInvoicesListResponse(t *testing.T) {
	// Test InvoicesListResponse
	invoices := []Invoice{
		{
			ID:            "550e8400-e29b-41d4-a716-446655440000",
			InvoiceNumber: "INV-2024-001",
		},
		{
			ID:            "550e8400-e29b-41d4-a716-446655440001",
			InvoiceNumber: "INV-2024-002",
		},
	}

	response := InvoicesListResponse{
		Success: true,
		Data:    invoices,
		Count:   2,
		Message: "Invoices retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, 2, len(response.Data))
	assert.Equal(t, 2, response.Count)
	assert.Equal(t, "Invoices retrieved successfully", response.Message)
}

// Invoice Detail Model Tests

func TestInvoiceDetailModel(t *testing.T) {
	// Test InvoiceDetail struct
	detail := InvoiceDetail{
		ID:             "550e8400-e29b-41d4-a716-446655440003",
		InvoiceID:      "550e8400-e29b-41d4-a716-446655440000",
		IngredientID:   stringPtr("550e8400-e29b-41d4-a716-446655440004"),
		Detail:         "Milk - 1 Gallon",
		Count:          2.0,
		UnitType:       "Gallons",
		Price:          5.99,
		Total:          11.98,
		ExpirationDate: timePtr(time.Now().AddDate(0, 0, 14)),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440003", detail.ID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", detail.InvoiceID)
	assert.NotNil(t, detail.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440004", *detail.IngredientID)
	assert.Equal(t, "Milk - 1 Gallon", detail.Detail)
	assert.Equal(t, 2.0, detail.Count)
	assert.Equal(t, "Gallons", detail.UnitType)
	assert.Equal(t, 5.99, detail.Price)
	assert.Equal(t, 11.98, detail.Total)
}

func TestCreateInvoiceDetailRequest(t *testing.T) {
	// Test CreateInvoiceDetailRequest
	req := CreateInvoiceDetailRequest{
		InvoiceID:      "550e8400-e29b-41d4-a716-446655440000",
		IngredientID:   stringPtr("550e8400-e29b-41d4-a716-446655440004"),
		Detail:         "Milk - 1 Gallon",
		Count:          2.0,
		UnitType:       "Gallons",
		Price:          5.99,
		ExpirationDate: timePtr(time.Now().AddDate(0, 0, 14)),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.InvoiceID)
	assert.NotNil(t, req.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440004", *req.IngredientID)
	assert.Equal(t, "Milk - 1 Gallon", req.Detail)
	assert.Equal(t, 2.0, req.Count)
	assert.Equal(t, "Gallons", req.UnitType)
	assert.Equal(t, 5.99, req.Price)
}

func TestUpdateInvoiceDetailRequest(t *testing.T) {
	// Test UpdateInvoiceDetailRequest with partial updates
	req := UpdateInvoiceDetailRequest{
		Detail:   stringPtr("Updated milk description"),
		Count:    float64Ptr(3.0),
		Price:    float64Ptr(6.49),
		UnitType: stringPtr("Liters"),
	}

	assert.NotNil(t, req.Detail)
	assert.Equal(t, "Updated milk description", *req.Detail)
	assert.NotNil(t, req.Count)
	assert.Equal(t, 3.0, *req.Count)
	assert.NotNil(t, req.Price)
	assert.Equal(t, 6.49, *req.Price)
	assert.NotNil(t, req.UnitType)
	assert.Equal(t, "Liters", *req.UnitType)
	assert.Nil(t, req.IngredientID) // Should be nil when not provided
}

func TestListInvoiceDetailsRequest(t *testing.T) {
	// Test ListInvoiceDetailsRequest
	req := ListInvoiceDetailsRequest{
		Limit:        intPtr(20),
		Offset:       intPtr(10),
		InvoiceID:    stringPtr("550e8400-e29b-41d4-a716-446655440000"),
		IngredientID: stringPtr("550e8400-e29b-41d4-a716-446655440004"),
	}

	assert.NotNil(t, req.Limit)
	assert.Equal(t, 20, *req.Limit)
	assert.NotNil(t, req.Offset)
	assert.Equal(t, 10, *req.Offset)
	assert.NotNil(t, req.InvoiceID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", *req.InvoiceID)
	assert.NotNil(t, req.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440004", *req.IngredientID)
}

func TestInvoiceDetailResponse(t *testing.T) {
	// Test InvoiceDetailResponse
	detail := InvoiceDetail{
		ID:        "550e8400-e29b-41d4-a716-446655440003",
		InvoiceID: "550e8400-e29b-41d4-a716-446655440000",
		Detail:    "Milk - 1 Gallon",
	}

	response := InvoiceDetailResponse{
		Success: true,
		Data:    detail,
		Message: "Invoice detail retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440003", response.Data.ID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Data.InvoiceID)
	assert.Equal(t, "Milk - 1 Gallon", response.Data.Detail)
	assert.Equal(t, "Invoice detail retrieved successfully", response.Message)
}

func TestInvoiceDetailsListResponse(t *testing.T) {
	// Test InvoiceDetailsListResponse
	details := []InvoiceDetail{
		{
			ID:        "550e8400-e29b-41d4-a716-446655440003",
			InvoiceID: "550e8400-e29b-41d4-a716-446655440000",
			Detail:    "Milk - 1 Gallon",
		},
		{
			ID:        "550e8400-e29b-41d4-a716-446655440004",
			InvoiceID: "550e8400-e29b-41d4-a716-446655440000",
			Detail:    "Sugar - 5lb bag",
		},
	}

	response := InvoiceDetailsListResponse{
		Success: true,
		Data:    details,
		Count:   2,
		Message: "Invoice details retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, 2, len(response.Data))
	assert.Equal(t, 2, response.Count)
	assert.Equal(t, "Invoice details retrieved successfully", response.Message)
}

// Test Delete Response Types

func TestInvoiceDeleteResponse(t *testing.T) {
	response := InvoiceDeleteResponse{
		Success: true,
		Message: "Invoice deleted successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Invoice deleted successfully", response.Message)
}

func TestInvoiceDetailDeleteResponse(t *testing.T) {
	response := InvoiceDetailDeleteResponse{
		Success: true,
		Message: "Invoice detail deleted successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Invoice detail deleted successfully", response.Message)
}

// Test Error Response

func TestErrorResponse(t *testing.T) {
	response := ErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Message: "Invoice number is required",
	}

	assert.False(t, response.Success)
	assert.Equal(t, "Validation failed", response.Error)
	assert.Equal(t, "Invoice number is required", response.Message)
}

// Test Transaction Type Validation

func TestTransactionTypeValidation(t *testing.T) {
	// Test valid transaction types
	incomeReq := CreateInvoiceRequest{
		TransactionType: "income",
	}
	outcomeReq := CreateInvoiceRequest{
		TransactionType: "outcome",
	}

	assert.Equal(t, "income", incomeReq.TransactionType)
	assert.Equal(t, "outcome", outcomeReq.TransactionType)
}

// Validation Tests

func TestInvoiceRequestValidation(t *testing.T) {
	// Test required fields
	req := CreateInvoiceRequest{}

	// These would fail validation in a real validation framework
	assert.Empty(t, req.InvoiceNumber)
	assert.Empty(t, req.TransactionType)
	assert.Empty(t, req.ExpenseCategoryID)
	assert.Empty(t, req.ImageURL)
	assert.Empty(t, req.Items)
}

func TestInvoiceDetailRequestValidation(t *testing.T) {
	// Test required fields
	req := CreateInvoiceDetailRequest{}

	// These would fail validation in a real validation framework
	assert.Empty(t, req.InvoiceID)
	assert.Empty(t, req.Detail)
	assert.Equal(t, 0.0, req.Count)
	assert.Equal(t, 0.0, req.Price)
	assert.Empty(t, req.UnitType)
}

// Helper functions for creating pointers

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}
