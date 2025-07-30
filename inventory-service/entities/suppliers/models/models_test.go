package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSupplier_JSONMarshaling(t *testing.T) {
	// Test data
	contactNumber := "+1234567890"
	email := "test@example.com"
	address := "123 Test St"
	notes := "Test notes"

	supplier := Supplier{
		ID:            "123e4567-e89b-12d3-a456-426614174000",
		SupplierName:  "Test Supplier",
		ContactNumber: &contactNumber,
		Email:         &email,
		Address:       &address,
		Notes:         &notes,
		CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Test marshaling
	data, err := json.Marshal(supplier)
	if err != nil {
		t.Errorf("Failed to marshal supplier: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Supplier
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal supplier: %v", err)
	}

	// Assert values
	if unmarshaled.ID != supplier.ID {
		t.Errorf("Expected ID %s, got %s", supplier.ID, unmarshaled.ID)
	}
	if unmarshaled.SupplierName != supplier.SupplierName {
		t.Errorf("Expected supplier name %s, got %s", supplier.SupplierName, unmarshaled.SupplierName)
	}
	if *unmarshaled.ContactNumber != *supplier.ContactNumber {
		t.Errorf("Expected contact number %s, got %s", *supplier.ContactNumber, *unmarshaled.ContactNumber)
	}
}

func TestSupplier_NilFields(t *testing.T) {
	// Test supplier with nil optional fields
	supplier := Supplier{
		ID:            "123e4567-e89b-12d3-a456-426614174000",
		SupplierName:  "Test Supplier",
		ContactNumber: nil,
		Email:         nil,
		Address:       nil,
		Notes:         nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test marshaling
	data, err := json.Marshal(supplier)
	if err != nil {
		t.Errorf("Failed to marshal supplier with nil fields: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Supplier
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal supplier with nil fields: %v", err)
	}

	// Assert nil fields
	if unmarshaled.ContactNumber != nil {
		t.Errorf("Expected ContactNumber to be nil, got %v", unmarshaled.ContactNumber)
	}
	if unmarshaled.Email != nil {
		t.Errorf("Expected Email to be nil, got %v", unmarshaled.Email)
	}
	if unmarshaled.Address != nil {
		t.Errorf("Expected Address to be nil, got %v", unmarshaled.Address)
	}
	if unmarshaled.Notes != nil {
		t.Errorf("Expected Notes to be nil, got %v", unmarshaled.Notes)
	}
}

func TestCreateSupplierRequest_JSONMarshaling(t *testing.T) {
	contactNumber := "+1234567890"
	email := "test@example.com"
	address := "123 Test St"
	notes := "Test notes"

	request := CreateSupplierRequest{
		SupplierName:  "Test Supplier",
		ContactNumber: &contactNumber,
		Email:         &email,
		Address:       &address,
		Notes:         &notes,
	}

	// Test marshaling
	data, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal CreateSupplierRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled CreateSupplierRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal CreateSupplierRequest: %v", err)
	}

	// Assert values
	if unmarshaled.SupplierName != request.SupplierName {
		t.Errorf("Expected supplier name %s, got %s", request.SupplierName, unmarshaled.SupplierName)
	}
	if *unmarshaled.ContactNumber != *request.ContactNumber {
		t.Errorf("Expected contact number %s, got %s", *request.ContactNumber, *unmarshaled.ContactNumber)
	}
}

func TestUpdateSupplierRequest_JSONMarshaling(t *testing.T) {
	supplierName := "Updated Supplier"
	email := "updated@example.com"

	request := UpdateSupplierRequest{
		SupplierName: &supplierName,
		Email:        &email,
		// Intentionally leaving other fields nil
	}

	// Test marshaling
	data, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal UpdateSupplierRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled UpdateSupplierRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal UpdateSupplierRequest: %v", err)
	}

	// Assert values
	if *unmarshaled.SupplierName != *request.SupplierName {
		t.Errorf("Expected supplier name %s, got %s", *request.SupplierName, *unmarshaled.SupplierName)
	}
	if *unmarshaled.Email != *request.Email {
		t.Errorf("Expected email %s, got %s", *request.Email, *unmarshaled.Email)
	}

	// Assert nil fields
	if unmarshaled.ContactNumber != nil {
		t.Errorf("Expected ContactNumber to be nil, got %v", unmarshaled.ContactNumber)
	}
	if unmarshaled.Address != nil {
		t.Errorf("Expected Address to be nil, got %v", unmarshaled.Address)
	}
	if unmarshaled.Notes != nil {
		t.Errorf("Expected Notes to be nil, got %v", unmarshaled.Notes)
	}
}

func TestSupplierResponse_Success(t *testing.T) {
	supplier := Supplier{
		ID:           "123e4567-e89b-12d3-a456-426614174000",
		SupplierName: "Test Supplier",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	response := SupplierResponse{
		Success: true,
		Data:    supplier,
		Message: "Supplier retrieved successfully",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal SupplierResponse: %v", err)
	}

	// Test unmarshaling
	var unmarshaled SupplierResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SupplierResponse: %v", err)
	}

	// Assert values
	if unmarshaled.Success != response.Success {
		t.Errorf("Expected success %v, got %v", response.Success, unmarshaled.Success)
	}
	if unmarshaled.Data.ID != response.Data.ID {
		t.Errorf("Expected data ID %s, got %s", response.Data.ID, unmarshaled.Data.ID)
	}
	if unmarshaled.Message != response.Message {
		t.Errorf("Expected message %s, got %s", response.Message, unmarshaled.Message)
	}
}

func TestSuppliersListResponse_Success(t *testing.T) {
	suppliers := []Supplier{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174001",
			SupplierName: "Supplier 1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "123e4567-e89b-12d3-a456-426614174002",
			SupplierName: "Supplier 2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	response := SuppliersListResponse{
		Success: true,
		Data:    suppliers,
		Count:   len(suppliers),
		Message: "Suppliers retrieved successfully",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal SuppliersListResponse: %v", err)
	}

	// Test unmarshaling
	var unmarshaled SuppliersListResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SuppliersListResponse: %v", err)
	}

	// Assert values
	if unmarshaled.Success != response.Success {
		t.Errorf("Expected success %v, got %v", response.Success, unmarshaled.Success)
	}
	if len(unmarshaled.Data) != len(response.Data) {
		t.Errorf("Expected %d suppliers, got %d", len(response.Data), len(unmarshaled.Data))
	}
	if unmarshaled.Count != response.Count {
		t.Errorf("Expected count %d, got %d", response.Count, unmarshaled.Count)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	response := ErrorResponse{
		Success: false,
		Error:   "Bad Request",
		Message: "Invalid input provided",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal ErrorResponse: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ErrorResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal ErrorResponse: %v", err)
	}

	// Assert values
	if unmarshaled.Success != response.Success {
		t.Errorf("Expected success %v, got %v", response.Success, unmarshaled.Success)
	}
	if unmarshaled.Error != response.Error {
		t.Errorf("Expected error %s, got %s", response.Error, unmarshaled.Error)
	}
	if unmarshaled.Message != response.Message {
		t.Errorf("Expected message %s, got %s", response.Message, unmarshaled.Message)
	}
}

func TestSupplierDeleteResponse_Success(t *testing.T) {
	response := SupplierDeleteResponse{
		Success: true,
		Message: "Supplier deleted successfully",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal SupplierDeleteResponse: %v", err)
	}

	// Test unmarshaling
	var unmarshaled SupplierDeleteResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SupplierDeleteResponse: %v", err)
	}

	// Assert values
	if unmarshaled.Success != response.Success {
		t.Errorf("Expected success %v, got %v", response.Success, unmarshaled.Success)
	}
	if unmarshaled.Message != response.Message {
		t.Errorf("Expected message %s, got %s", response.Message, unmarshaled.Message)
	}
}
