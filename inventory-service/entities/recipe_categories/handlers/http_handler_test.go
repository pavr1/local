package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-service/entities/recipe_categories/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeCategoryHTTPHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.dbHandler.db)
	assert.Equal(t, logger, handler.logger)
}

func TestRecipeCategoryHTTPHandler_CreateRecipeCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	description := "Test description"
	req := models.CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: &description,
	}

	now := time.Now()
	expectedRecipeCategory := models.RecipeCategory{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeCategory.ID,
		expectedRecipeCategory.Name,
		expectedRecipeCategory.Description,
		expectedRecipeCategory.CreatedAt,
		expectedRecipeCategory.UpdatedAt,
	)

	mock.ExpectQuery("INSERT INTO recipe_categories").
		WithArgs(req.Name, req.Description).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipe-categories", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeCategory(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeCategory.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Data.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Data.Description)
	assert.Contains(t, result.Message, "created successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_CreateRecipeCategory_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	request := httptest.NewRequest("POST", "/recipe-categories", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeCategory(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeCategoryHTTPHandler_CreateRecipeCategory_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	req := models.CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: nil,
	}

	mock.ExpectQuery("INSERT INTO recipe_categories").
		WithArgs(req.Name, req.Description).
		WillReturnError(sql.ErrConnDone)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipe-categories", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeCategory(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to create recipe category")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_GetRecipeCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	now := time.Now()
	description := "Test description"
	expectedRecipeCategory := models.RecipeCategory{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Test Category",
		Description: &description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeCategory.ID,
		expectedRecipeCategory.Name,
		expectedRecipeCategory.Description,
		expectedRecipeCategory.CreatedAt,
		expectedRecipeCategory.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs(expectedRecipeCategory.ID).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-categories/"+expectedRecipeCategory.ID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.GetRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeCategory.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Data.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Data.Description)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_GetRecipeCategory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest("GET", "/recipe-categories/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.GetRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_GetRecipeCategory_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	request := httptest.NewRequest("GET", "/recipe-categories/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.GetRecipeCategory(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeCategoryHTTPHandler_ListRecipeCategories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	now := time.Now()
	description := "Test description"
	expectedRecipeCategories := []models.RecipeCategory{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "Test Category",
			Description: &description,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	})
	for _, rc := range expectedRecipeCategories {
		rows.AddRow(
			rc.ID, rc.Name, rc.Description, rc.CreatedAt, rc.UpdatedAt,
		)
	}

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs(nil, 50, 0).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-categories", nil)
	response := httptest.NewRecorder()

	handler.ListRecipeCategories(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeCategoriesResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, len(expectedRecipeCategories), len(result.Data))
	for i, expected := range expectedRecipeCategories {
		assert.Equal(t, expected.ID, result.Data[i].ID)
		assert.Equal(t, expected.Name, result.Data[i].Name)
		assert.Equal(t, expected.Description, result.Data[i].Description)
	}
	assert.Equal(t, 1, result.Total)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_ListRecipeCategories_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs("Test", 10, 5).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-categories?name=Test&limit=10&offset=5", nil)
	response := httptest.NewRecorder()

	handler.ListRecipeCategories(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeCategoriesResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.Data)
	assert.Equal(t, 0, result.Total)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_UpdateRecipeCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	now := time.Now()
	name := "Updated Category"
	description := "Updated description"
	req := models.UpdateRecipeCategoryRequest{
		Name:        &name,
		Description: &description,
	}

	expectedRecipeCategory := models.RecipeCategory{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        name,
		Description: &description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeCategory.ID,
		expectedRecipeCategory.Name,
		expectedRecipeCategory.Description,
		expectedRecipeCategory.CreatedAt,
		expectedRecipeCategory.UpdatedAt,
	)

	mock.ExpectQuery("UPDATE recipe_categories").
		WithArgs(expectedRecipeCategory.ID, &name, &description).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipe-categories/"+expectedRecipeCategory.ID, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.UpdateRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeCategory.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Data.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Data.Description)
	assert.Contains(t, result.Message, "updated successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_UpdateRecipeCategory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	name := "Updated Category"
	req := models.UpdateRecipeCategoryRequest{
		Name: &name,
	}

	mock.ExpectQuery("UPDATE recipe_categories").
		WithArgs("non-existent-id", &name, nil).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipe-categories/non-existent-id", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.UpdateRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeCategoryResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_DeleteRecipeCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	recipeCategoryID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectExec("DELETE FROM recipe_categories").
		WithArgs(recipeCategoryID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest("DELETE", "/recipe-categories/"+recipeCategoryID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.DeleteRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.GenericResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "deleted successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_DeleteRecipeCategory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	mock.ExpectExec("DELETE FROM recipe_categories").
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	request := httptest.NewRequest("DELETE", "/recipe-categories/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-categories/{id}", handler.DeleteRecipeCategory)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.GenericResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryHTTPHandler_DeleteRecipeCategory_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeCategoryHTTPHandler(db, logger)

	request := httptest.NewRequest("DELETE", "/recipe-categories/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.DeleteRecipeCategory(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}
