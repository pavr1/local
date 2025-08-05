package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-service/entities/existences/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestMockDBHandler implements DBHandlerInterface for testing
type TestMockDBHandler struct {
	CreateExistenceFunc  func(req models.CreateExistenceRequest) (*models.Existence, error)
	GetExistenceByIDFunc func(id string) (*models.Existence, error)
	ListExistencesFunc   func(req models.ListExistencesRequest) ([]models.Existence, error)
	UpdateExistenceFunc  func(id string, req models.UpdateExistenceRequest) (*models.Existence, error)
	DeleteExistenceFunc  func(id string) error
}

// Ensure TestMockDBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*TestMockDBHandler)(nil)

func (m *TestMockDBHandler) CreateExistence(req models.CreateExistenceRequest) (*models.Existence, error) {
	if m.CreateExistenceFunc != nil {
		return m.CreateExistenceFunc(req)
	}
	return nil, nil
}

func (m *TestMockDBHandler) GetExistenceByID(id string) (*models.Existence, error) {
	if m.GetExistenceByIDFunc != nil {
		return m.GetExistenceByIDFunc(id)
	}
	return nil, nil
}

func (m *TestMockDBHandler) ListExistences(req models.ListExistencesRequest) ([]models.Existence, error) {
	if m.ListExistencesFunc != nil {
		return m.ListExistencesFunc(req)
	}
	return nil, nil
}

func (m *TestMockDBHandler) UpdateExistence(id string, req models.UpdateExistenceRequest) (*models.Existence, error) {
	if m.UpdateExistenceFunc != nil {
		return m.UpdateExistenceFunc(id, req)
	}
	return nil, nil
}

func (m *TestMockDBHandler) DeleteExistence(id string) error {
	if m.DeleteExistenceFunc != nil {
		return m.DeleteExistenceFunc(id)
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

// Helper functions are defined in db_handler_test.go

func TestHttpHandler_CreateExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	// Test data
	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	reqBody := models.CreateExistenceRequest{
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         10.0,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerUnit:            12000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: float64Ptr(30.0),
		IvaPercentage:          float64Ptr(13.0),
		ServiceTaxPercentage:   float64Ptr(10.0),
		FinalPrice:             float64Ptr(15000.00),
	}

	expectedExistence := models.Existence{
		ID:                     "existence-id-123",
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         10.0,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         120000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		IvaPercentage:          13.0,
		IvaAmount:              20280.00,
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       15600.00,
		CalculatedPrice:        191880.00,
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Mock setup
	mockDB.CreateExistenceFunc = func(req models.CreateExistenceRequest) (*models.Existence, error) {
		return &expectedExistence, nil
	}

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/existences", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.ExistenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, expectedExistence.ID, response.Data.ID)
	assert.Equal(t, expectedExistence.UnitType, response.Data.UnitType)
	assert.Contains(t, response.Message, "created successfully")
}

func TestHttpHandler_CreateExistence_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHttpHandler()

	// Prepare invalid request
	req := httptest.NewRequest(http.MethodPost, "/existences", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHttpHandler_CreateExistence_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	reqBody := models.CreateExistenceRequest{
		IngredientID:    "ingredient-id-123",
		InvoiceDetailID: "invoice-detail-id-123",
		UnitsPurchased:  10.0,
		UnitsAvailable:  10.0,
		UnitType:        "Liters",
		ItemsPerUnit:    31,
		CostPerUnit:     12000.00,
	}

	// Mock setup
	mockDB.CreateExistenceFunc = func(req models.CreateExistenceRequest) (*models.Existence, error) {
		return nil, fmt.Errorf("database error")
	}

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/existences", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHttpHandler_GetExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "existence-id-123"
	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         8.5,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         102000.00,
		ExpirationDate:         timePtr(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		IvaPercentage:          13.0,
		IvaAmount:              20280.00,
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       15600.00,
		CalculatedPrice:        191880.00,
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Mock setup
	mockDB.GetExistenceByIDFunc = func(id string) (*models.Existence, error) {
		if id == existenceID {
			return &expectedExistence, nil
		}
		return nil, sql.ErrNoRows
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.GetExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ExistenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, expectedExistence.ID, response.Data.ID)
	assert.Equal(t, expectedExistence.UnitsAvailable, response.Data.UnitsAvailable)
}

func TestHttpHandler_GetExistence_NotFound(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "nonexistent-id"

	// Mock setup
	mockDB.GetExistenceByIDFunc = func(id string) (*models.Existence, error) {
		return nil, sql.ErrNoRows
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.GetExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHttpHandler_GetExistence_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "existence-id-123"

	// Mock setup
	mockDB.GetExistenceByIDFunc = func(id string) (*models.Existence, error) {
		return nil, fmt.Errorf("database error")
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.GetExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHttpHandler_ListExistences_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	expectedExistences := []models.Existence{
		{
			ID:                     "existence-1",
			ExistenceReferenceCode: 1001,
			IngredientID:           "ingredient-id-123",
			InvoiceDetailID:        "invoice-detail-1",
			UnitsPurchased:         10.0,
			UnitsAvailable:         8.5,
			UnitType:               "Liters",
			ItemsPerUnit:           31,
			CostPerItem:            387.10,
			CostPerUnit:            12000.00,
			TotalPurchaseCost:      120000.00,
			RemainingValue:         102000.00,
			ExpirationDate:         timePtr(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
			IncomeMarginPercentage: 30.0,
			IncomeMarginAmount:     36000.00,
			IvaPercentage:          13.0,
			IvaAmount:              20280.00,
			ServiceTaxPercentage:   10.0,
			ServiceTaxAmount:       15600.00,
			CalculatedPrice:        191880.00,
			FinalPrice:             float64Ptr(15000.00),
			CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	// Mock setup
	mockDB.ListExistencesFunc = func(req models.ListExistencesRequest) ([]models.Existence, error) {
		return expectedExistences, nil
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/existences", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListExistences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ExistencesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, expectedExistences[0].ID, response.Data[0].ID)
	assert.Equal(t, 1, response.Total)
}

func TestHttpHandler_ListExistences_WithFilters(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	expectedExistences := []models.Existence{
		{
			ID:                     "existence-1",
			ExistenceReferenceCode: 1001,
			IngredientID:           "ingredient-id-123",
			InvoiceDetailID:        "invoice-detail-1",
			UnitsPurchased:         10.0,
			UnitsAvailable:         1.0, // Low stock
			UnitType:               "Liters",
			ItemsPerUnit:           31,
			CostPerItem:            387.10,
			CostPerUnit:            12000.00,
			TotalPurchaseCost:      120000.00,
			RemainingValue:         12000.00,
			ExpirationDate:         timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)), // Not expired
			IncomeMarginPercentage: 30.0,
			IncomeMarginAmount:     36000.00,
			IvaPercentage:          13.0,
			IvaAmount:              20280.00,
			ServiceTaxPercentage:   10.0,
			ServiceTaxAmount:       15600.00,
			CalculatedPrice:        191880.00,
			FinalPrice:             float64Ptr(15000.00),
			CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	// Mock setup
	mockDB.ListExistencesFunc = func(req models.ListExistencesRequest) ([]models.Existence, error) {
		// Verify filters are passed correctly
		assert.NotNil(t, req.IngredientID)
		assert.Equal(t, "ingredient-id-123", *req.IngredientID)
		assert.NotNil(t, req.UnitType)
		assert.Equal(t, "Liters", *req.UnitType)
		assert.NotNil(t, req.Expired)
		assert.False(t, *req.Expired)
		assert.NotNil(t, req.LowStock)
		assert.True(t, *req.LowStock)
		return expectedExistences, nil
	}

	// Prepare request with query parameters
	req := httptest.NewRequest(http.MethodGet, "/existences?ingredient_id=ingredient-id-123&unit_type=Liters&expired=false&low_stock=true", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListExistences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ExistencesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, expectedExistences[0].ID, response.Data[0].ID)
}

func TestHttpHandler_ListExistences_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	// Mock setup
	mockDB.ListExistencesFunc = func(req models.ListExistencesRequest) ([]models.Existence, error) {
		return nil, fmt.Errorf("database error")
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/existences", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListExistences(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHttpHandler_UpdateExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "existence-id-123"
	reqBody := models.UpdateExistenceRequest{
		UnitsAvailable: float64Ptr(5.0),
	}

	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         5.0, // Updated
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         60000.00, // Updated based on new units
		ExpirationDate:         timePtr(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		IvaPercentage:          13.0,
		IvaAmount:              20280.00,
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       15600.00,
		CalculatedPrice:        191880.00,
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
	}

	// Mock setup
	mockDB.UpdateExistenceFunc = func(id string, req models.UpdateExistenceRequest) (*models.Existence, error) {
		if id == existenceID {
			return &expectedExistence, nil
		}
		return nil, sql.ErrNoRows
	}

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/existences/"+existenceID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.UpdateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ExistenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, expectedExistence.ID, response.Data.ID)
	assert.Equal(t, expectedExistence.UnitsAvailable, response.Data.UnitsAvailable)
	assert.Contains(t, response.Message, "updated successfully")
}

func TestHttpHandler_UpdateExistence_NotFound(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "nonexistent-id"
	reqBody := models.UpdateExistenceRequest{
		UnitsAvailable: float64Ptr(5.0),
	}

	// Mock setup
	mockDB.UpdateExistenceFunc = func(id string, req models.UpdateExistenceRequest) (*models.Existence, error) {
		return nil, sql.ErrNoRows
	}

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/existences/"+existenceID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.UpdateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHttpHandler_UpdateExistence_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHttpHandler()

	existenceID := "existence-id-123"

	// Prepare invalid request
	req := httptest.NewRequest(http.MethodPut, "/existences/"+existenceID, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.UpdateExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHttpHandler_DeleteExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "existence-id-123"

	// Mock setup
	mockDB.DeleteExistenceFunc = func(id string) error {
		if id == existenceID {
			return nil
		}
		return sql.ErrNoRows
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodDelete, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.DeleteExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GenericResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "deleted successfully")
}

func TestHttpHandler_DeleteExistence_NotFound(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "nonexistent-id"

	// Mock setup
	mockDB.DeleteExistenceFunc = func(id string) error {
		return sql.ErrNoRows
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodDelete, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.DeleteExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHttpHandler_DeleteExistence_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "existence-id-123"

	// Mock setup
	mockDB.DeleteExistenceFunc = func(id string) error {
		return fmt.Errorf("database error")
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodDelete, "/existences/"+existenceID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": existenceID})
	w := httptest.NewRecorder()

	// Execute
	handler.DeleteExistence(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
