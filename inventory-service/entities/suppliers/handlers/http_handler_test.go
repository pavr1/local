package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory-service/entities/suppliers/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TestMockDBHandler implements DBHandlerInterface for testing (renamed to avoid conflict)
type TestMockDBHandler struct {
	CreateSupplierFunc  func(req models.CreateSupplierRequest) (*models.Supplier, error)
	GetSupplierByIDFunc func(id string) (*models.Supplier, error)
	ListSuppliersFunc   func() ([]models.Supplier, error)
	UpdateSupplierFunc  func(id string, req models.UpdateSupplierRequest) (*models.Supplier, error)
	DeleteSupplierFunc  func(id string) error
}

// Ensure TestMockDBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*TestMockDBHandler)(nil)

func (m *TestMockDBHandler) CreateSupplier(req models.CreateSupplierRequest) (*models.Supplier, error) {
	if m.CreateSupplierFunc != nil {
		return m.CreateSupplierFunc(req)
	}
	return nil, nil
}

func (m *TestMockDBHandler) GetSupplierByID(id string) (*models.Supplier, error) {
	if m.GetSupplierByIDFunc != nil {
		return m.GetSupplierByIDFunc(id)
	}
	return nil, nil
}

func (m *TestMockDBHandler) ListSuppliers() ([]models.Supplier, error) {
	if m.ListSuppliersFunc != nil {
		return m.ListSuppliersFunc()
	}
	return nil, nil
}

func (m *TestMockDBHandler) UpdateSupplier(id string, req models.UpdateSupplierRequest) (*models.Supplier, error) {
	if m.UpdateSupplierFunc != nil {
		return m.UpdateSupplierFunc(id, req)
	}
	return nil, nil
}

func (m *TestMockDBHandler) DeleteSupplier(id string) error {
	if m.DeleteSupplierFunc != nil {
		return m.DeleteSupplierFunc(id)
	}
	return nil
}

func setupTestHttpHandler() (*HttpHandler, *TestMockDBHandler) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing

	mockDB := &TestMockDBHandler{}
	handler := NewHttpHandlerWithInterface(mockDB, logger)

	return handler, mockDB
}

func TestHttpHandler_CreateSupplier_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	// Test data
	requestData := models.CreateSupplierRequest{
		SupplierName: "Test Supplier",
		Email:        stringPtrForTest("test@example.com"),
	}

	expectedSupplier := &models.Supplier{
		ID:           "123e4567-e89b-12d3-a456-426614174000",
		SupplierName: "Test Supplier",
		Email:        stringPtrForTest("test@example.com"),
	}

	// Setup mock
	mockDB.CreateSupplierFunc = func(req models.CreateSupplierRequest) (*models.Supplier, error) {
		if req.SupplierName != requestData.SupplierName {
			t.Errorf("Expected supplier name %s, got %s", requestData.SupplierName, req.SupplierName)
		}
		return expectedSupplier, nil
	}

	// Create request
	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/suppliers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.CreateSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}

	// Assert response body
	var response models.SupplierResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if response.Data.ID != expectedSupplier.ID {
		t.Errorf("Expected supplier ID %s, got %s", expectedSupplier.ID, response.Data.ID)
	}
}

func TestHttpHandler_CreateSupplier_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHttpHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/suppliers", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.CreateSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Assert response body contains error
	var response models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success to be false, got true")
	}
}

func TestHttpHandler_CreateSupplier_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	requestData := models.CreateSupplierRequest{
		SupplierName: "Test Supplier",
	}

	// Setup mock to return database error
	mockDB.CreateSupplierFunc = func(req models.CreateSupplierRequest) (*models.Supplier, error) {
		return nil, fmt.Errorf("database connection failed")
	}

	// Create request
	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/suppliers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.CreateSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Assert response body
	var response models.SupplierResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success to be false, got true")
	}
}

func TestHttpHandler_GetSupplier_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"
	expectedSupplier := &models.Supplier{
		ID:           supplierID,
		SupplierName: "Test Supplier",
	}

	// Setup mock
	mockDB.GetSupplierByIDFunc = func(id string) (*models.Supplier, error) {
		if id != supplierID {
			t.Errorf("Expected supplier ID %s, got %s", supplierID, id)
		}
		return expectedSupplier, nil
	}

	// Create request with URL parameter
	req := httptest.NewRequest("GET", "/suppliers/"+supplierID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": supplierID})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Assert response body
	var response models.SupplierResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if response.Data.ID != expectedSupplier.ID {
		t.Errorf("Expected supplier ID %s, got %s", expectedSupplier.ID, response.Data.ID)
	}
}

func TestHttpHandler_GetSupplier_NotFound(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	supplierID := "nonexistent-id"

	// Setup mock to return not found error
	mockDB.GetSupplierByIDFunc = func(id string) (*models.Supplier, error) {
		return nil, sql.ErrNoRows
	}

	// Create request
	req := httptest.NewRequest("GET", "/suppliers/"+supplierID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": supplierID})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rr.Code)
	}

	// Assert response body
	var response models.SupplierResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success to be false, got true")
	}
}

func TestHttpHandler_ListSuppliers_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	expectedSuppliers := []models.Supplier{
		{ID: "id1", SupplierName: "Supplier 1"},
		{ID: "id2", SupplierName: "Supplier 2"},
	}

	// Setup mock
	mockDB.ListSuppliersFunc = func() ([]models.Supplier, error) {
		return expectedSuppliers, nil
	}

	// Create request
	req := httptest.NewRequest("GET", "/suppliers", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ListSuppliers(rr, req)

	// Assert response code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Assert response body
	var response models.SuppliersListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 suppliers, got %d", len(response.Data))
	}

	if response.Count != 2 {
		t.Errorf("Expected count to be 2, got %d", response.Count)
	}
}

func TestHttpHandler_DeleteSupplier_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"

	// Setup mock
	mockDB.DeleteSupplierFunc = func(id string) error {
		if id != supplierID {
			t.Errorf("Expected supplier ID %s, got %s", supplierID, id)
		}
		return nil
	}

	// Create request
	req := httptest.NewRequest("DELETE", "/suppliers/"+supplierID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": supplierID})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.DeleteSupplier(rr, req)

	// Assert response code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Assert response body
	var response models.SupplierDeleteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}
}

// Helper function to create string pointers
func stringPtrForTest(s string) *string {
	return &s
}
