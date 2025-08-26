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

// MockDBHandler for testing
type MockDBHandler struct {
	CreateExistenceFunc                               func(req models.CreateExistenceRequest) (*models.Existence, error)
	GetExistenceByIDFunc                              func(id string) (*models.Existence, error)
	ListExistencesFunc                                func(req models.ListExistencesRequest) ([]models.Existence, error)
	UpdateExistenceFunc                               func(id string, req models.UpdateExistenceRequest) (*models.Existence, error)
	DeleteExistenceFunc                               func(id string) error
	GetMostRecentExistenceByIngredientAndUnitTypeFunc func(ingredientID, unitType string) (*models.Existence, error)
}

func (m *MockDBHandler) CreateExistence(req models.CreateExistenceRequest) (*models.Existence, error) {
	return m.CreateExistenceFunc(req)
}

func (m *MockDBHandler) GetExistenceByID(id string) (*models.Existence, error) {
	return m.GetExistenceByIDFunc(id)
}

func (m *MockDBHandler) ListExistences(req models.ListExistencesRequest) ([]models.Existence, error) {
	return m.ListExistencesFunc(req)
}

func (m *MockDBHandler) UpdateExistence(id string, req models.UpdateExistenceRequest) (*models.Existence, error) {
	return m.UpdateExistenceFunc(id, req)
}

func (m *MockDBHandler) DeleteExistence(id string) error {
	return m.DeleteExistenceFunc(id)
}

func (m *MockDBHandler) GetMostRecentExistenceByIngredientAndUnitType(ingredientID, unitType string) (*models.Existence, error) {
	return m.GetMostRecentExistenceByIngredientAndUnitTypeFunc(ingredientID, unitType)
}

func setupTestHttpHandler() (*HttpHandler, *MockDBHandler) {
	mockDB := &MockDBHandler{}
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing
	handler := NewHttpHandlerWithInterface(mockDB, logger)
	return handler, mockDB
}

func TestHttpHandler_CreateExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	reqBody := models.CreateExistenceRequest{
		IngredientID:           "550e8400-e29b-41d4-a716-446655440000",
		InvoiceDetailID:        "550e8400-e29b-41d4-a716-446655440001",
		UnitsPurchased:         10.0,
		UnitsAvailable:         10.0,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerUnit:            12000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: float64Ptr(30.0),
		FinalPrice:             float64Ptr(15000.00),
	}

	expectedExistence := models.Existence{
		ID:                     "550e8400-e29b-41d4-a716-446655440002",
		ExistenceReferenceCode: 1001,
		IngredientID:           "550e8400-e29b-41d4-a716-446655440000",
		InvoiceDetailID:        "550e8400-e29b-41d4-a716-446655440001",
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
		MinimumPrice:           156000.00,             // cost + income margin
		MaximumPrice:           float64Ptr(156100.00), // rounded up to nearest 100
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
		IngredientID:    "550e8400-e29b-41d4-a716-446655440000",
		InvoiceDetailID: "550e8400-e29b-41d4-a716-446655440001",
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

	existenceID := "550e8400-e29b-41d4-a716-446655440002"
	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "550e8400-e29b-41d4-a716-446655440000",
		InvoiceDetailID:        "550e8400-e29b-41d4-a716-446655440001",
		UnitsPurchased:         10.0,
		UnitsAvailable:         8.5,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         102000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		MinimumPrice:           156000.00,             // cost + income margin
		MaximumPrice:           float64Ptr(156100.00), // rounded up to nearest 100
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Mock setup
	mockDB.GetExistenceByIDFunc = func(id string) (*models.Existence, error) {
		return &expectedExistence, nil
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
	assert.Equal(t, expectedExistence.UnitType, response.Data.UnitType)
}

func TestHttpHandler_GetExistence_NotFound(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "550e8400-e29b-41d4-a716-446655440004"

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
	assert.Contains(t, w.Body.String(), "Existence not found")
}

func TestHttpHandler_ListExistences_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	expectedExistences := []models.Existence{
		{
			ID:                     "550e8400-e29b-41d4-a716-446655440003",
			ExistenceReferenceCode: 1001,
			IngredientID:           "550e8400-e29b-41d4-a716-446655440000",
			InvoiceDetailID:        "550e8400-e29b-41d4-a716-446655440001",
			UnitsPurchased:         10.0,
			UnitsAvailable:         8.5,
			UnitType:               "Liters",
			ItemsPerUnit:           31,
			CostPerItem:            387.10,
			CostPerUnit:            12000.00,
			TotalPurchaseCost:      120000.00,
			RemainingValue:         102000.00,
			ExpirationDate:         &expirationDate,
			IncomeMarginPercentage: 30.0,
			IncomeMarginAmount:     36000.00,
			MinimumPrice:           156000.00,             // cost + income margin
			MaximumPrice:           float64Ptr(156100.00), // rounded up to nearest 100
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
}

func TestHttpHandler_UpdateExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "550e8400-e29b-41d4-a716-446655440002"
	newUnitsAvailable := 5.0
	reqBody := models.UpdateExistenceRequest{
		UnitsAvailable: &newUnitsAvailable,
	}

	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "550e8400-e29b-41d4-a716-446655440000",
		InvoiceDetailID:        "550e8400-e29b-41d4-a716-446655440001",
		UnitsPurchased:         10.0,
		UnitsAvailable:         newUnitsAvailable,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         60000.00, // Updated based on new units available
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		MinimumPrice:           156000.00,             // cost + income margin
		MaximumPrice:           float64Ptr(156100.00), // rounded up to nearest 100
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
	}

	// Mock setup
	mockDB.UpdateExistenceFunc = func(id string, req models.UpdateExistenceRequest) (*models.Existence, error) {
		return &expectedExistence, nil
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

	existenceID := "550e8400-e29b-41d4-a716-446655440004"
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
	assert.Contains(t, w.Body.String(), "Existence not found")
}

func TestHttpHandler_DeleteExistence_Success(t *testing.T) {
	handler, mockDB := setupTestHttpHandler()

	existenceID := "550e8400-e29b-41d4-a716-446655440002"

	// Mock setup
	mockDB.DeleteExistenceFunc = func(id string) error {
		return nil
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

	existenceID := "550e8400-e29b-41d4-a716-446655440004"

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
	assert.Contains(t, w.Body.String(), "Existence not found")
}
