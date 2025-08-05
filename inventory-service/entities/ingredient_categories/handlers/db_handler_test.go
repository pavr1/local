package handlers

import (
	"database/sql"
	"testing"

	"inventory-service/entities/ingredient_categories/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIngredientCategory(t *testing.T) {
	testCases := map[string]struct {
		request        models.CreateIngredientCategoryRequest
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.IngredientCategory
	}{
		"successful_creation_with_active_true": {
			request: models.CreateIngredientCategoryRequest{
				Name:        "dairy_products",
				Description: "Milk, cream, butter, eggs, cheese, yogurt",
				IsActive:    boolPtr(true),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
					AddRow("category-123", "dairy_products", "Milk, cream, butter, eggs, cheese, yogurt", true, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("INSERT INTO ingredient_categories").
					WithArgs("dairy_products", "Milk, cream, butter, eggs, cheese, yogurt", true).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.IngredientCategory{
				ID:          "category-123",
				Name:        "dairy_products",
				Description: "Milk, cream, butter, eggs, cheese, yogurt",
				IsActive:    true,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"successful_creation_with_nil_active": {
			request: models.CreateIngredientCategoryRequest{
				Name:        "sweeteners",
				Description: "Sugar, honey, artificial sweeteners",
				IsActive:    nil,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
					AddRow("category-456", "sweeteners", "Sugar, honey, artificial sweeteners", true, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("INSERT INTO ingredient_categories").
					WithArgs("sweeteners", "Sugar, honey, artificial sweeteners", nil).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.IngredientCategory{
				ID:          "category-456",
				Name:        "sweeteners",
				Description: "Sugar, honey, artificial sweeteners",
				IsActive:    true,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"database_error": {
			request: models.CreateIngredientCategoryRequest{
				Name:        "test_category",
				Description: "Test description",
				IsActive:    boolPtr(true),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO ingredient_categories").
					WithArgs("test_category", "Test description", true).
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
			result, err := handler.CreateIngredientCategory(tc.request)

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

func TestGetIngredientCategoryByID(t *testing.T) {
	testCases := map[string]struct {
		categoryID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.IngredientCategory
	}{
		"successful_retrieval": {
			categoryID: "category-123",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
					AddRow("category-123", "dairy_products", "Milk, cream, butter, eggs", true, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("SELECT id, name, description, is_active, created_at, updated_at FROM ingredient_categories WHERE id").
					WithArgs("category-123").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.IngredientCategory{
				ID:          "category-123",
				Name:        "dairy_products",
				Description: "Milk, cream, butter, eggs",
				IsActive:    true,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name, description, is_active, created_at, updated_at FROM ingredient_categories WHERE id").
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
			result, err := handler.GetIngredientCategoryByID(tc.categoryID)

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

func TestListIngredientCategories(t *testing.T) {
	testCases := map[string]struct {
		setupMock       func(sqlmock.Sqlmock)
		expectedError   bool
		expectedResults []models.IngredientCategory
	}{
		"successful_list": {
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
					AddRow("category-1", "dairy_products", "Milk, cream, butter", true, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z").
					AddRow("category-2", "sweeteners", "Sugar, honey, syrups", true, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
				mock.ExpectQuery("SELECT id, name, description, is_active, created_at, updated_at FROM ingredient_categories ORDER BY name").
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResults: []models.IngredientCategory{
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
		},
		"empty_result": {
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT id, name, description, is_active, created_at, updated_at FROM ingredient_categories ORDER BY name").
					WillReturnRows(rows)
			},
			expectedError:   false,
			expectedResults: []models.IngredientCategory{},
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
			results, err := handler.ListIngredientCategories()

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

func TestUpdateIngredientCategory(t *testing.T) {
	testCases := map[string]struct {
		categoryID     string
		request        models.UpdateIngredientCategoryRequest
		setupMock      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedResult *models.IngredientCategory
	}{
		"successful_update_all_fields": {
			categoryID: "category-123",
			request: models.UpdateIngredientCategoryRequest{
				Name:        stringPtr("updated_dairy"),
				Description: stringPtr("Updated dairy products description"),
				IsActive:    boolPtr(false),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
					AddRow("category-123", "updated_dairy", "Updated dairy products description", false, "2024-01-01T00:00:00Z", "2024-01-01T12:00:00Z")
				mock.ExpectQuery("UPDATE ingredient_categories SET").
					WithArgs("category-123", "updated_dairy", "Updated dairy products description", false).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedResult: &models.IngredientCategory{
				ID:          "category-123",
				Name:        "updated_dairy",
				Description: "Updated dairy products description",
				IsActive:    false,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T12:00:00Z",
			},
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			request: models.UpdateIngredientCategoryRequest{
				Name: stringPtr("Test Name"),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE ingredient_categories SET").
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
			result, err := handler.UpdateIngredientCategory(tc.categoryID, tc.request)

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

func TestDeleteIngredientCategory(t *testing.T) {
	testCases := map[string]struct {
		categoryID    string
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		"successful_delete": {
			categoryID: "category-123",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM ingredient_categories WHERE id").
					WithArgs("category-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		"category_not_found": {
			categoryID: "nonexistent-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM ingredient_categories WHERE id").
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
			err = handler.DeleteIngredientCategory(tc.categoryID)

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

// Helper functions to create pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
