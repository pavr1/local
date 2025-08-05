package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory-service/entities/ingredient_categories/models"

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

func (m *MockDBHandler) CreateIngredientCategory(req models.CreateIngredientCategoryRequest) (*models.IngredientCategory, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IngredientCategory), args.Error(1)
}

func (m *MockDBHandler) GetIngredientCategoryByID(id string) (*models.IngredientCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IngredientCategory), args.Error(1)
}

func (m *MockDBHandler) ListIngredientCategories() ([]models.IngredientCategory, error) {
	args := m.Called()
	return args.Get(0).([]models.IngredientCategory), args.Error(1)
}

func (m *MockDBHandler) UpdateIngredientCategory(id string, req models.UpdateIngredientCategoryRequest) (*models.IngredientCategory, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IngredientCategory), args.Error(1)
}

func (m *MockDBHandler) DeleteIngredientCategory(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateIngredientCategoryHTTP(t *testing.T) {
	testCases := map[string]struct {
		requestBody        interface{}
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_creation": {
			requestBody: models.CreateIngredientCategoryRequest{
				Name:        "dairy_products",
				Description: "Milk, cream, butter, eggs, cheese, yogurt",
				IsActive:    boolPtr(true),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("CreateIngredientCategory", mock.AnythingOfType("models.CreateIngredientCategoryRequest")).Return(
					&models.IngredientCategory{
						ID:          "category-123",
						Name:        "dairy_products",
						Description: "Milk, cream, butter, eggs, cheese, yogurt",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: models.IngredientCategoryResponse{
				Success: true,
				Data: models.IngredientCategory{
					ID:          "category-123",
					Name:        "dairy_products",
					Description: "Milk, cream, butter, eggs, cheese, yogurt",
					IsActive:    true,
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
				Message: "Ingredient category created successfully",
			},
		},
		"invalid_json": {
			requestBody:        "invalid json",
			mockSetup:          func(mockDB *MockDBHandler) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		"database_error": {
			requestBody: models.CreateIngredientCategoryRequest{
				Name:        "test_category",
				Description: "Test description",
				IsActive:    boolPtr(true),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("CreateIngredientCategory", mock.AnythingOfType("models.CreateIngredientCategoryRequest")).Return(
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
			req := httptest.NewRequest(http.MethodPost, "/ingredient-categories", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// Execute
			handler.CreateIngredientCategory(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientCategoryResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetIngredientCategoryHTTP(t *testing.T) {
	testCases := map[string]struct {
		categoryID         string
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_retrieval": {
			categoryID: "category-123",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("GetIngredientCategoryByID", "category-123").Return(
					&models.IngredientCategory{
						ID:          "category-123",
						Name:        "dairy_products",
						Description: "Milk, cream, butter, eggs, cheese, yogurt",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientCategoryResponse{
				Success: true,
				Data: models.IngredientCategory{
					ID:          "category-123",
					Name:        "dairy_products",
					Description: "Milk, cream, butter, eggs, cheese, yogurt",
					IsActive:    true,
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
				Message: "Ingredient category retrieved successfully",
			},
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("GetIngredientCategoryByID", "nonexistent-id").Return(nil, sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientCategoryResponse{
				Success: false,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
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
			req := httptest.NewRequest(http.MethodGet, "/ingredient-categories/"+tc.categoryID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.categoryID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.GetIngredientCategory(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientCategoryResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestListIngredientCategoriesHTTP(t *testing.T) {
	testCases := map[string]struct {
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_list": {
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("ListIngredientCategories").Return([]models.IngredientCategory{
					{
						ID:          "category-1",
						Name:        "dairy_products",
						Description: "Milk, cream, butter",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
					{
						ID:          "category-2",
						Name:        "sweeteners",
						Description: "Sugar, honey, syrups",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientCategoriesListResponse{
				Success: true,
				Data: []models.IngredientCategory{
					{
						ID:          "category-1",
						Name:        "dairy_products",
						Description: "Milk, cream, butter",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
					{
						ID:          "category-2",
						Name:        "sweeteners",
						Description: "Sugar, honey, syrups",
						IsActive:    true,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T00:00:00Z",
					},
				},
				Count:   2,
				Message: "Ingredient categories retrieved successfully",
			},
		},
		"empty_list": {
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("ListIngredientCategories").Return([]models.IngredientCategory{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientCategoriesListResponse{
				Success: true,
				Data:    []models.IngredientCategory{},
				Count:   0,
				Message: "Ingredient categories retrieved successfully",
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
			req := httptest.NewRequest(http.MethodGet, "/ingredient-categories", nil)
			recorder := httptest.NewRecorder()

			// Execute
			handler.ListIngredientCategories(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientCategoriesListResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateIngredientCategoryHTTP(t *testing.T) {
	testCases := map[string]struct {
		categoryID         string
		requestBody        interface{}
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_update": {
			categoryID: "category-123",
			requestBody: models.UpdateIngredientCategoryRequest{
				Name:        stringPtr("updated_dairy"),
				Description: stringPtr("Updated dairy products description"),
				IsActive:    boolPtr(false),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("UpdateIngredientCategory", "category-123", mock.AnythingOfType("models.UpdateIngredientCategoryRequest")).Return(
					&models.IngredientCategory{
						ID:          "category-123",
						Name:        "updated_dairy",
						Description: "Updated dairy products description",
						IsActive:    false,
						CreatedAt:   "2024-01-01T00:00:00Z",
						UpdatedAt:   "2024-01-01T12:00:00Z",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientCategoryResponse{
				Success: true,
				Data: models.IngredientCategory{
					ID:          "category-123",
					Name:        "updated_dairy",
					Description: "Updated dairy products description",
					IsActive:    false,
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T12:00:00Z",
				},
				Message: "Ingredient category updated successfully",
			},
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			requestBody: models.UpdateIngredientCategoryRequest{
				Name: stringPtr("Test Name"),
			},
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("UpdateIngredientCategory", "nonexistent-id", mock.AnythingOfType("models.UpdateIngredientCategoryRequest")).Return(
					nil, sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientCategoryResponse{
				Success: false,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
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
			req := httptest.NewRequest(http.MethodPut, "/ingredient-categories/"+tc.categoryID, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tc.categoryID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.UpdateIngredientCategory(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientCategoryResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteIngredientCategoryHTTP(t *testing.T) {
	testCases := map[string]struct {
		categoryID         string
		mockSetup          func(*MockDBHandler)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		"successful_delete": {
			categoryID: "category-123",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("DeleteIngredientCategory", "category-123").Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: models.IngredientCategoryDeleteResponse{
				Success: true,
				Message: "Ingredient category deleted successfully",
			},
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			mockSetup: func(mockDB *MockDBHandler) {
				mockDB.On("DeleteIngredientCategory", "nonexistent-id").Return(sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse: models.IngredientCategoryDeleteResponse{
				Success: false,
				Message: "Ingredient category not found",
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
			req := httptest.NewRequest(http.MethodDelete, "/ingredient-categories/"+tc.categoryID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.categoryID})
			recorder := httptest.NewRecorder()

			// Execute
			handler.DeleteIngredientCategory(recorder, req)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedResponse != nil {
				var response models.IngredientCategoryDeleteResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
