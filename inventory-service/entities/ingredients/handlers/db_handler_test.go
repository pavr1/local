package handlers

import (
	"database/sql"
	"testing"

	"inventory-service/entities/ingredients/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIngredient(t *testing.T) {
	testCases := map[string]struct {
		request        models.CreateIngredientRequest
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.Ingredient
	}{
		"successful_creation_with_description": {
			request: models.CreateIngredientRequest{
				Name:        "Vanilla Extract",
				Description: stringPtr("Pure vanilla extract for flavoring"),
				SupplierID:  stringPtr("supplier-123"),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"}).
					AddRow("ingredient-123", "Vanilla Extract", "Pure vanilla extract for flavoring", "supplier-123", "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("INSERT INTO ingredients").
					WithArgs("Vanilla Extract", "Pure vanilla extract for flavoring", "supplier-123").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.Ingredient{
				ID:          "ingredient-123",
				Name:        "Vanilla Extract",
				Description: stringPtr("Pure vanilla extract for flavoring"),
				SupplierID:  stringPtr("supplier-123"),
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"successful_creation_without_description": {
			request: models.CreateIngredientRequest{
				Name:        "Sugar",
				Description: nil,
				SupplierID:  nil,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"}).
					AddRow("ingredient-456", "Sugar", nil, nil, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("INSERT INTO ingredients").
					WithArgs("Sugar", nil, nil).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.Ingredient{
				ID:          "ingredient-456",
				Name:        "Sugar",
				Description: nil,
				SupplierID:  nil,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"database_error": {
			request: models.CreateIngredientRequest{
				Name:        "Test Ingredient",
				Description: stringPtr("Test description"),
				SupplierID:  nil,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO ingredients").
					WithArgs("Test Ingredient", "Test description", nil).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel) // Suppress logs during testing

			handler := NewDBHandler(db, logger)
			tc.setupMock(mock)

			// Execute
			result, err := handler.CreateIngredient(tc.request)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetIngredientByID(t *testing.T) {
	testCases := map[string]struct {
		ingredientID   string
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.Ingredient
	}{
		"successful_retrieval": {
			ingredientID: "ingredient-123",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"}).
					AddRow("ingredient-123", "Vanilla Extract", "Pure vanilla extract", "supplier-123", "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("SELECT id, name, description, supplier_id, created_at, updated_at FROM ingredients WHERE id").
					WithArgs("ingredient-123").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.Ingredient{
				ID:          "ingredient-123",
				Name:        "Vanilla Extract",
				Description: stringPtr("Pure vanilla extract"),
				SupplierID:  stringPtr("supplier-123"),
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name, description, supplier_id, created_at, updated_at FROM ingredients WHERE id").
					WithArgs("nonexistent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewDBHandler(db, logger)
			tc.setupMock(mock)

			// Execute
			result, err := handler.GetIngredientByID(tc.ingredientID)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestListIngredients(t *testing.T) {
	testCases := map[string]struct {
		setupMock       func(sqlmock.Sqlmock)
		expectedError   bool
		expectedResults []models.Ingredient
	}{
		"successful_list": {
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"}).
					AddRow("ingredient-1", "Sugar", nil, nil, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z").
					AddRow("ingredient-2", "Vanilla", "Pure vanilla extract", "supplier-123", "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("SELECT id, name, description, supplier_id, created_at, updated_at FROM ingredients ORDER BY name").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResults: []models.Ingredient{
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
		},
		"empty_result": {
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT id, name, description, supplier_id, created_at, updated_at FROM ingredients ORDER BY name").
					WillReturnRows(rows)
			},
			expectedError:   false,
			expectedResults: []models.Ingredient{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewDBHandler(db, logger)
			tc.setupMock(mock)

			// Execute
			results, err := handler.ListIngredients()

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResults, results)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateIngredient(t *testing.T) {
	testCases := map[string]struct {
		ingredientID   string
		request        models.UpdateIngredientRequest
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.Ingredient
	}{
		"successful_update_all_fields": {
			ingredientID: "ingredient-123",
			request: models.UpdateIngredientRequest{
				Name:        stringPtr("Updated Vanilla"),
				Description: stringPtr("Updated description"),
				SupplierID:  stringPtr("new-supplier-456"),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "supplier_id", "created_at", "updated_at"}).
					AddRow("ingredient-123", "Updated Vanilla", "Updated description", "new-supplier-456", "2024-01-01T00:00:00Z", "2024-01-01T12:00:00Z")
				mock.ExpectQuery("UPDATE ingredients SET").
					WithArgs("ingredient-123", "Updated Vanilla", "Updated description", "new-supplier-456").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.Ingredient{
				ID:          "ingredient-123",
				Name:        "Updated Vanilla",
				Description: stringPtr("Updated description"),
				SupplierID:  stringPtr("new-supplier-456"),
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T12:00:00Z",
			},
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			request: models.UpdateIngredientRequest{
				Name: stringPtr("Test Name"),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE ingredients SET").
					WithArgs("nonexistent-id", "Test Name", nil, nil).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewDBHandler(db, logger)
			tc.setupMock(mock)

			// Execute
			result, err := handler.UpdateIngredient(tc.ingredientID, tc.request)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteIngredient(t *testing.T) {
	testCases := map[string]struct {
		ingredientID  string
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		"successful_delete": {
			ingredientID: "ingredient-123",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM ingredients WHERE id").
					WithArgs("ingredient-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		"ingredient_not_found": {
			ingredientID: "nonexistent-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM ingredients WHERE id").
					WithArgs("nonexistent-id").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			handler := NewDBHandler(db, logger)
			tc.setupMock(mock)

			// Execute
			err = handler.DeleteIngredient(tc.ingredientID)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
