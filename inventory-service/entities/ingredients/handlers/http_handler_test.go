package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory-service/entities/ingredients/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDBHandler is a mock implementation of DBHandlerInterface
type MockDBHandler struct {
	mock.Mock
}

func (m *MockDBHandler) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockDBHandler) GetIngredientByID(id string) (*models.Ingredient, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockDBHandler) ListIngredients() ([]models.Ingredient, error) {
	args := m.Called()
	return args.Get(0).([]models.Ingredient), args.Error(1)
}

func (m *MockDBHandler) UpdateIngredient(id string, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockDBHandler) DeleteIngredient(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateIngredientHTTP(t *testing.T) {
	testCases := map[string]struct {
		requestBody        interface{}
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_creation": {
			requestBody: models.CreateIngredientRequest{
				Name:        "Vanilla Extract",
				Description: stringPtr("Pure vanilla extract"),
				SupplierID:  stringPtr("supplier-123"),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).Return(
					&models.Ingredient{
						ID:          "ingredient-123",
						Name:        "Vanilla Extract",
						Description: stringPtr("Pure vanilla extract"),
						SupplierID:  stringPtr("supplier-123"),
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: models.IngredientResponse{
				Success: true,
				Data: models.Ingredient{
					ID:          "ingredient-123",
					Name:        "Vanilla Extract",
					Description: stringPtr("Pure vanilla extract"),
					SupplierID:  stringPtr("supplier-123"),
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
				Message: "Ingredient created successfully",
			},
		},
		"invalid_json": {
			requestBody:        "invalid json",
			mockSetup:          func(mockDB *MockDBHandler) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		"database_error": {
			requestBody: models.CreateIngredientRequest{
				Name:        "Test Ingredient",
				Description: nil,
				SupplierID:  nil,
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).Return(
					nil, assert.AnError)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDBHandler)
			tc.mockSetup(mockDB)

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewHttpHandlerWithInterface(mockDB, logger)

			// Create request body
			var requestBody []byte
			var err error
			if tc.requestBody == "invalid json" {
				requestBody = []byte("invalid json")
			} else {
				requestBody, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/ingredients", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// Execute
			handler.CreateIngredient(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetIngredientHTTP(t *testing.T) {
	testCases := map[string]struct {
		ingredientID       string
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_retrieval": {
			ingredientID: "ingredient-123",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("GetIngredientByID", "ingredient-123").Return(
					&models.Ingredient{
						ID:          "ingredient-123",
						Name:        "Vanilla Extract",
						Description: stringPtr("Pure vanilla extract"),
						SupplierID:  stringPtr("supplier-123"),
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientResponse{
				Success: true,
				Data: models.Ingredient{
					ID:          "ingredient-123",
					Name:        "Vanilla Extract",
					Description: stringPtr("Pure vanilla extract"),
					SupplierID:  stringPtr("supplier-123"),
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
				Message: "Ingredient retrieved successfully",
			},
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("GetIngredientByID", "nonexistent-id").Return(nil, sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientResponse{
				Success: false,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDBHandler)
			tc.mockSetup(mockDB)

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewHttpHandlerWithInterface(mockDB, logger)

			// Create HTTP request with mux vars
			req := httptest.NewRequest(http.MethodGet, "/ingredients/"+tc.ingredientID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.ingredientID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.GetIngredient(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestListIngredientsHTTP(t *testing.T) {
	testCases := map[string]struct {
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_list": {
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("ListIngredients").Return([]models.Ingredient{
					{
						ID:          "ingredient-1",
						Name:        "Sugar",
						Description: nil,
						SupplierID:  nil,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
					{
						ID:          "ingredient-2",
						Name:        "Vanilla",
						Description: stringPtr("Pure vanilla extract"),
						SupplierID:  stringPtr("supplier-123"),
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientsListResponse{
				Success: true,
				Data: []models.Ingredient{
					{
						ID:          "ingredient-1",
						Name:        "Sugar",
						Description: nil,
						SupplierID:  nil,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
					{
						ID:          "ingredient-2",
						Name:        "Vanilla",
						Description: stringPtr("Pure vanilla extract"),
						SupplierID:  stringPtr("supplier-123"),
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
				},
				Count:   2,
				Message: "Ingredients retrieved successfully",
			},
		},
		"empty_list": {
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("ListIngredients").Return([]models.Ingredient{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientsListResponse{
				Success: true,
				Data:    []models.Ingredient{},
				Count:   0,
				Message: "Ingredients retrieved successfully",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDBHandler)
			tc.mockSetup(mockDB)

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewHttpHandlerWithInterface(mockDB, logger)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/ingredients", nil)
			recorder := httptest.NewRecorder()

			// Execute
			handler.ListIngredients(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientsListResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateIngredientHTTP(t *testing.T) {
	testCases := map[string]struct {
		ingredientID       string
		requestBody        interface{}
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_update": {
			ingredientID: "ingredient-123",
			requestBody: models.UpdateIngredientRequest{
				Name:        stringPtr("Updated Vanilla"),
				Description: stringPtr("Updated description"),
				SupplierID:  stringPtr("new-supplier-456"),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("UpdateIngredient", "ingredient-123", mock.AnythingOfType("models.UpdateIngredientRequest")).Return(
					&models.Ingredient{
						ID:          "ingredient-123",
						Name:        "Updated Vanilla",
						Description: stringPtr("Updated description"),
						SupplierID:  stringPtr("new-supplier-456"),
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T12:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientResponse{
				Success: true,
				Data: models.Ingredient{
					ID:          "ingredient-123",
					Name:        "Updated Vanilla",
					Description: stringPtr("Updated description"),
					SupplierID:  stringPtr("new-supplier-456"),
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T12:00:00Z",
				},
				Message: "Ingredient updated successfully",
			},
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			requestBody: models.UpdateIngredientRequest{
				Name: stringPtr("Test Name"),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("UpdateIngredient", "nonexistent-id", mock.AnythingOfType("models.UpdateIngredientRequest")).Return(
					nil, sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientResponse{
				Success: false,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDBHandler)
			tc.mockSetup(mockDB)

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewHttpHandlerWithInterface(mockDB, logger)

			// Create request body
			requestBody, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			// Create HTTP request with mux vars
			req := httptest.NewRequest(http.MethodPut, "/ingredients/"+tc.ingredientID, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tc.ingredientID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.UpdateIngredient(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteIngredientHTTP(t *testing.T) {
	testCases := map[string]struct {
		ingredientID       string
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_delete": {
			ingredientID: "ingredient-123",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("DeleteIngredient", "ingredient-123").Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientDeleteResponse{
				Success: true,
				Message: "Ingredient deleted successfully",
			},
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("DeleteIngredient", "nonexistent-id").Return(sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientDeleteResponse{
				Success: false,
				Message: "Ingredient not found",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDBHandler)
			tc.mockSetup(mockDB)

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewHttpHandlerWithInterface(mockDB, logger)

			// Create HTTP request with mux vars
			req := httptest.NewRequest(http.MethodDelete, "/ingredients/"+tc.ingredientID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.ingredientID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.DeleteIngredient(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientDeleteResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
