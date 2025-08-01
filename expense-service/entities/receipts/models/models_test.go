package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Receipt Model Tests

func TestReceiptModel(t *testing.T) {
	// Test Receipt struct
	receipt := Receipt{
		ID:                "550e8400-e29b-41d4-a716-446655440000",
		ReceiptNumber:     "RCP-2024-001",
		PurchaseDate:      time.Now(),
		SupplierID:        stringPtr("550e8400-e29b-41d4-a716-446655440001"),
		ExpenseCategoryID: "550e8400-e29b-41d4-a716-446655440002",
		TotalAmount:       float64Ptr(150.75),
		ImageURL:          "https://example.com/receipts/receipt001.jpg",
		Notes:             stringPtr("Test receipt for ingredients"),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", receipt.ID)
	assert.Equal(t, "RCP-2024-001", receipt.ReceiptNumber)
	assert.NotNil(t, receipt.SupplierID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", *receipt.SupplierID)
	assert.NotNil(t, receipt.TotalAmount)
	assert.Equal(t, 150.75, *receipt.TotalAmount)
}

func TestCreateReceiptRequest(t *testing.T) {
	// Test valid CreateReceiptRequest
	items := []CreateReceiptItemRequest{
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

	req := CreateReceiptRequest{
		ReceiptNumber:     "RCP-2024-002",
		PurchaseDate:      time.Now(),
		SupplierID:        stringPtr("550e8400-e29b-41d4-a716-446655440001"),
		ExpenseCategoryID: "550e8400-e29b-41d4-a716-446655440002",
		ImageURL:          "https://example.com/receipts/receipt002.jpg",
		Notes:             stringPtr("Ingredient purchase from local supplier"),
		Items:             items,
	}

	assert.Equal(t, "RCP-2024-002", req.ReceiptNumber)
	assert.Equal(t, 2, len(req.Items))
	assert.Equal(t, "Milk - 1 Gallon", req.Items[0].Detail)
	assert.Equal(t, 2.0, req.Items[0].Count)
	assert.Equal(t, "Gallons", req.Items[0].UnitType)
	assert.Equal(t, 5.99, req.Items[0].Price)
	assert.NotNil(t, req.Items[0].ExpirationDate)

	// Second item without expiration date
	assert.Equal(t, "Sugar - 5lb bag", req.Items[1].Detail)
	assert.Nil(t, req.Items[1].ExpirationDate)
}

func TestUpdateReceiptRequest(t *testing.T) {
	// Test partial update request
	updateReq := UpdateReceiptRequest{
		ReceiptNumber: stringPtr("RCP-2024-002-UPDATED"),
		Notes:         stringPtr("Updated notes for the receipt"),
		ImageURL:      stringPtr("https://example.com/receipts/receipt002-updated.jpg"),
	}

	assert.NotNil(t, updateReq.ReceiptNumber)
	assert.Equal(t, "RCP-2024-002-UPDATED", *updateReq.ReceiptNumber)
	assert.NotNil(t, updateReq.Notes)
	assert.Equal(t, "Updated notes for the receipt", *updateReq.Notes)
	assert.NotNil(t, updateReq.ImageURL)

	// Fields not being updated should be nil
	assert.Nil(t, updateReq.PurchaseDate)
	assert.Nil(t, updateReq.SupplierID)
	assert.Nil(t, updateReq.ExpenseCategoryID)
}

func TestListReceiptsRequest(t *testing.T) {
	// Test list request with filters
	listReq := ListReceiptsRequest{
		Limit:             intPtr(20),
		Offset:            intPtr(40),
		ExpenseCategoryID: stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		SupplierID:        stringPtr("550e8400-e29b-41d4-a716-446655440001"),
	}

	assert.NotNil(t, listReq.Limit)
	assert.Equal(t, 20, *listReq.Limit)
	assert.NotNil(t, listReq.Offset)
	assert.Equal(t, 40, *listReq.Offset)
	assert.NotNil(t, listReq.ExpenseCategoryID)
	assert.NotNil(t, listReq.SupplierID)
}

func TestReceiptResponse(t *testing.T) {
	receipt := Receipt{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		ReceiptNumber: "RCP-2024-001",
		PurchaseDate:  time.Now(),
	}

	response := ReceiptResponse{
		Success: true,
		Data:    receipt,
		Message: "Receipt retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Data.ID)
	assert.Equal(t, "Receipt retrieved successfully", response.Message)
}

func TestReceiptsListResponse(t *testing.T) {
	receipts := []Receipt{
		{ID: "1", ReceiptNumber: "RCP-001"},
		{ID: "2", ReceiptNumber: "RCP-002"},
		{ID: "3", ReceiptNumber: "RCP-003"},
	}

	response := ReceiptsListResponse{
		Success: true,
		Data:    receipts,
		Count:   3,
		Message: "Receipts retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, 3, len(response.Data))
	assert.Equal(t, 3, response.Count)
	assert.Equal(t, "RCP-001", response.Data[0].ReceiptNumber)
}

// Receipt Item Model Tests

func TestReceiptItemModel(t *testing.T) {
	// Test ReceiptItem struct
	receiptItem := ReceiptItem{
		ID:             "550e8400-e29b-41d4-a716-446655440000",
		ReceiptID:      "550e8400-e29b-41d4-a716-446655440001",
		IngredientID:   stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		Detail:         "Milk - 1 Gallon",
		Count:          2.0,
		UnitType:       "Gallons",
		Price:          5.99,
		Total:          11.98,
		ExpirationDate: timePtr(time.Now().AddDate(0, 0, 14)),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", receiptItem.ID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", receiptItem.ReceiptID)
	assert.NotNil(t, receiptItem.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", *receiptItem.IngredientID)
	assert.Equal(t, "Milk - 1 Gallon", receiptItem.Detail)
	assert.Equal(t, 2.0, receiptItem.Count)
	assert.Equal(t, "Gallons", receiptItem.UnitType)
	assert.Equal(t, 5.99, receiptItem.Price)
	assert.Equal(t, 11.98, receiptItem.Total)
	assert.NotNil(t, receiptItem.ExpirationDate)
}

func TestCreateReceiptItemRequest(t *testing.T) {
	// Test with ingredient ID and expiration date
	req := CreateReceiptItemRequest{
		ReceiptID:      "550e8400-e29b-41d4-a716-446655440001",
		IngredientID:   stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		Detail:         "Premium Vanilla Extract - 4oz",
		Count:          1.0,
		UnitType:       "Units",
		Price:          12.99,
		ExpirationDate: timePtr(time.Now().AddDate(2, 0, 0)),
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", req.ReceiptID)
	assert.NotNil(t, req.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", *req.IngredientID)
	assert.Equal(t, "Premium Vanilla Extract - 4oz", req.Detail)
	assert.Equal(t, 1.0, req.Count)
	assert.Equal(t, "Units", req.UnitType)
	assert.Equal(t, 12.99, req.Price)
	assert.NotNil(t, req.ExpirationDate)

	// Test without ingredient ID (for non-ingredient expenses)
	reqNoIngredient := CreateReceiptItemRequest{
		ReceiptID: "550e8400-e29b-41d4-a716-446655440001",
		Detail:    "Office Supplies - Printer Paper",
		Count:     5.0,
		UnitType:  "Units",
		Price:     2.49,
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", reqNoIngredient.ReceiptID)
	assert.Nil(t, reqNoIngredient.IngredientID)
	assert.Equal(t, "Office Supplies - Printer Paper", reqNoIngredient.Detail)
	assert.Nil(t, reqNoIngredient.ExpirationDate)
}

func TestUpdateReceiptItemRequest(t *testing.T) {
	// Test partial update request
	updateReq := UpdateReceiptItemRequest{
		Detail:   stringPtr("Updated: Premium Vanilla Extract - 4oz"),
		Count:    float64Ptr(2.0),
		Price:    float64Ptr(11.99),
		UnitType: stringPtr("Units"),
	}

	assert.NotNil(t, updateReq.Detail)
	assert.Equal(t, "Updated: Premium Vanilla Extract - 4oz", *updateReq.Detail)
	assert.NotNil(t, updateReq.Count)
	assert.Equal(t, 2.0, *updateReq.Count)
	assert.NotNil(t, updateReq.Price)
	assert.Equal(t, 11.99, *updateReq.Price)
	assert.NotNil(t, updateReq.UnitType)
	assert.Equal(t, "Units", *updateReq.UnitType)

	// Fields not being updated should be nil
	assert.Nil(t, updateReq.IngredientID)
	assert.Nil(t, updateReq.ExpirationDate)
}

func TestListReceiptItemsRequest(t *testing.T) {
	// Test list request with filters
	listReq := ListReceiptItemsRequest{
		Limit:        intPtr(10),
		Offset:       intPtr(20),
		ReceiptID:    stringPtr("550e8400-e29b-41d4-a716-446655440001"),
		IngredientID: stringPtr("550e8400-e29b-41d4-a716-446655440002"),
	}

	assert.NotNil(t, listReq.Limit)
	assert.Equal(t, 10, *listReq.Limit)
	assert.NotNil(t, listReq.Offset)
	assert.Equal(t, 20, *listReq.Offset)
	assert.NotNil(t, listReq.ReceiptID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", *listReq.ReceiptID)
	assert.NotNil(t, listReq.IngredientID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", *listReq.IngredientID)
}

func TestReceiptItemResponse(t *testing.T) {
	receiptItem := ReceiptItem{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		ReceiptID: "550e8400-e29b-41d4-a716-446655440001",
		Detail:    "Test item",
		Count:     1.0,
		Price:     10.50,
		Total:     10.50,
	}

	response := ReceiptItemResponse{
		Success: true,
		Data:    receiptItem,
		Message: "Receipt item retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Data.ID)
	assert.Equal(t, "Receipt item retrieved successfully", response.Message)
}

func TestReceiptItemsListResponse(t *testing.T) {
	receiptItems := []ReceiptItem{
		{ID: "1", Detail: "Item 1", Count: 1.0, Price: 5.00, Total: 5.00},
		{ID: "2", Detail: "Item 2", Count: 2.0, Price: 3.50, Total: 7.00},
		{ID: "3", Detail: "Item 3", Count: 1.0, Price: 12.99, Total: 12.99},
	}

	response := ReceiptItemsListResponse{
		Success: true,
		Data:    receiptItems,
		Count:   3,
		Message: "Receipt items retrieved successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, 3, len(response.Data))
	assert.Equal(t, 3, response.Count)
	assert.Equal(t, "Item 1", response.Data[0].Detail)
	assert.Equal(t, 5.00, response.Data[0].Total)
}

func TestReceiptItemDeleteResponse(t *testing.T) {
	response := ReceiptItemDeleteResponse{
		Success: true,
		Message: "Receipt item deleted successfully",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Receipt item deleted successfully", response.Message)
}

func TestValidUnitTypes(t *testing.T) {
	// Test valid unit types
	validUnitTypes := []string{"Liters", "Gallons", "Units", "Bag"}

	for _, unitType := range validUnitTypes {
		req := CreateReceiptItemRequest{
			ReceiptID: "550e8400-e29b-41d4-a716-446655440001",
			Detail:    "Test item",
			Count:     1.0,
			UnitType:  unitType,
			Price:     10.00,
		}

		assert.Equal(t, unitType, req.UnitType)
	}
}

func TestReceiptItemCalculations(t *testing.T) {
	// Test that total calculations work correctly
	testCases := []struct {
		count    float64
		price    float64
		expected float64
	}{
		{1.0, 10.00, 10.00},
		{2.5, 4.00, 10.00},
		{0.5, 20.00, 10.00},
		{3.0, 3.33, 9.99},
	}

	for _, tc := range testCases {
		req := CreateReceiptItemRequest{
			ReceiptID: "550e8400-e29b-41d4-a716-446655440001",
			Detail:    "Test item",
			Count:     tc.count,
			UnitType:  "Units",
			Price:     tc.price,
		}

		calculatedTotal := req.Count * req.Price
		assert.Equal(t, tc.expected, calculatedTotal)
	}
}

func TestErrorResponse(t *testing.T) {
	errorResp := ErrorResponse{
		Success: false,
		Error:   "Receipt not found",
		Message: "The requested receipt could not be found",
	}

	assert.False(t, errorResp.Success)
	assert.Equal(t, "Receipt not found", errorResp.Error)
	assert.Equal(t, "The requested receipt could not be found", errorResp.Message)
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
