package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-service/entities/recipe_ingredients/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeIngredientHTTPHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.dbHandler.db)
	assert.Equal(t, logger, handler.logger)
}

func TestRecipeIngredientHTTPHandler_CreateRecipeIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	req := models.CreateRecipeIngredientRequest{
		RecipeID:     "550e8400-e29b-41d4-a716-446655440000",
		IngredientID: "550e8400-e29b-41d4-a716-446655440001",
		Quantity:     2.5,
		UnitType:     "cups",
	}

	now := time.Now()
	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440002",
		RecipeID:     req.RecipeID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		UnitType:     req.UnitType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "unit_type", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.UnitType,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("INSERT INTO recipe_ingredients").
		WithArgs(req.RecipeID, req.IngredientID, req.Quantity, req.UnitType).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipe-ingredients", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeIngredient(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.Data.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.Data.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRecipeIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "created successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_CreateRecipeIngredient_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("POST", "/recipe-ingredients", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeIngredientHTTPHandler_CreateRecipeIngredient_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	req := models.CreateRecipeIngredientRequest{
		RecipeID:     "550e8400-e29b-41d4-a716-446655440000",
		IngredientID: "550e8400-e29b-41d4-a716-446655440001",
		Quantity:     2.5,
		UnitType:     "cups",
	}

	mock.ExpectQuery("INSERT INTO recipe_ingredients").
		WithArgs(req.RecipeID, req.IngredientID, req.Quantity, req.UnitType).
		WillReturnError(sql.ErrConnDone)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipe-ingredients", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipeIngredient(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to create recipe ingredient")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_GetRecipeIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	now := time.Now()
	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		RecipeID:     "550e8400-e29b-41d4-a716-446655440001",
		IngredientID: "550e8400-e29b-41d4-a716-446655440002",
		Quantity:     2.5,
		UnitType:     "cups",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "unit_type", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.UnitType,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at").
		WithArgs(expectedRecipeIngredient.ID).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-ingredients/"+expectedRecipeIngredient.ID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.GetRecipeIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.Data.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.Data.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRecipeIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_GetRecipeIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest("GET", "/recipe-ingredients/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.GetRecipeIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_GetRecipeIngredient_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("GET", "/recipe-ingredients/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.GetRecipeIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeIngredientHTTPHandler_ListRecipeIngredients(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	now := time.Now()
	expectedRecipeIngredients := []models.RecipeIngredient{
		{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			RecipeID:     "550e8400-e29b-41d4-a716-446655440001",
			IngredientID: "550e8400-e29b-41d4-a716-446655440002",
			Quantity:     2.5,
			UnitType:     "cups",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "unit_type", "created_at", "updated_at",
	})
	for _, ri := range expectedRecipeIngredients {
		rows.AddRow(
			ri.ID, ri.RecipeID, ri.IngredientID, ri.Quantity, ri.UnitType, ri.CreatedAt, ri.UpdatedAt,
		)
	}

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at").
		WithArgs(nil, nil, 50, 0).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-ingredients", nil)
	response := httptest.NewRecorder()

	handler.ListRecipeIngredients(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeIngredientsResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, len(expectedRecipeIngredients), len(result.Data))
	for i, expected := range expectedRecipeIngredients {
		assert.Equal(t, expected.ID, result.Data[i].ID)
		assert.Equal(t, expected.RecipeID, result.Data[i].RecipeID)
		assert.Equal(t, expected.IngredientID, result.Data[i].IngredientID)
		assert.Equal(t, expected.Quantity, result.Data[i].Quantity)
		assert.Equal(t, expected.UnitType, result.Data[i].UnitType)
	}
	assert.Equal(t, 1, result.Total)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_ListRecipeIngredients_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "unit_type", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at").
		WithArgs("550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001", 10, 5).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipe-ingredients?recipe_id=550e8400-e29b-41d4-a716-446655440000&ingredient_id=550e8400-e29b-41d4-a716-446655440001&limit=10&offset=5", nil)
	response := httptest.NewRecorder()

	handler.ListRecipeIngredients(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeIngredientsResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.Data)
	assert.Equal(t, 0, result.Total)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_UpdateRecipeIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	now := time.Now()
	recipeID := "550e8400-e29b-41d4-a716-446655440001"
	ingredientID := "550e8400-e29b-41d4-a716-446655440002"
	quantity := 3.0
	unitType := "tablespoons"
	req := models.UpdateRecipeIngredientRequest{
		RecipeID:     &recipeID,
		IngredientID: &ingredientID,
		Quantity:     &quantity,
		UnitType:     &unitType,
	}

	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		RecipeID:     recipeID,
		IngredientID: ingredientID,
		Quantity:     quantity,
		UnitType:     unitType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "unit_type", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.UnitType,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("UPDATE recipe_ingredients").
		WithArgs(expectedRecipeIngredient.ID, &recipeID, &ingredientID, &quantity, &unitType).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipe-ingredients/"+expectedRecipeIngredient.ID, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.UpdateRecipeIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipeIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.Data.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.Data.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRecipeIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "updated successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_UpdateRecipeIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	quantity := 3.0
	req := models.UpdateRecipeIngredientRequest{
		Quantity: &quantity,
	}

	mock.ExpectQuery("UPDATE recipe_ingredients").
		WithArgs("non-existent-id", nil, nil, &quantity, nil).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipe-ingredients/non-existent-id", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.UpdateRecipeIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientHTTPHandler_DeleteRecipeIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	recipeIngredientID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(recipeIngredientID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest("DELETE", "/recipe-ingredients/"+recipeIngredientID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.DeleteRecipeIngredient)
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

func TestRecipeIngredientHTTPHandler_DeleteRecipeIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	request := httptest.NewRequest("DELETE", "/recipe-ingredients/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipe-ingredients/{id}", handler.DeleteRecipeIngredient)
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

func TestRecipeIngredientHTTPHandler_DeleteRecipeIngredient_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("DELETE", "/recipe-ingredients/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.DeleteRecipeIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}
