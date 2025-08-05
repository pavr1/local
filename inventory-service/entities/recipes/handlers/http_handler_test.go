package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-service/entities/recipes/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeHTTPHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.dbHandler.db)
	assert.Equal(t, logger, handler.logger)
}

func TestRecipeHTTPHandler_CreateRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	req := models.CreateRecipeRequest{
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440000",
		TotalRecipeCost:   15.50,
	}

	now := time.Now()
	expectedRecipe := models.Recipe{
		ID:                "550e8400-e29b-41d4-a716-446655440001",
		RecipeName:        req.RecipeName,
		RecipeDescription: req.RecipeDescription,
		PictureURL:        req.PictureURL,
		RecipeCategoryID:  req.RecipeCategoryID,
		TotalRecipeCost:   req.TotalRecipeCost,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	}).AddRow(
		expectedRecipe.ID,
		expectedRecipe.RecipeName,
		expectedRecipe.RecipeDescription,
		expectedRecipe.PictureURL,
		expectedRecipe.RecipeCategoryID,
		expectedRecipe.TotalRecipeCost,
		expectedRecipe.CreatedAt,
		expectedRecipe.UpdatedAt,
	)

	mock.ExpectQuery("INSERT INTO recipes").
		WithArgs(req.RecipeName, req.RecipeDescription, req.PictureURL, req.RecipeCategoryID, req.TotalRecipeCost).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipe(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipe.ID, result.Data.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.Data.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.Data.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.Data.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.Data.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.Data.TotalRecipeCost)
	assert.Contains(t, result.Message, "created successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_CreateRecipe_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	request := httptest.NewRequest("POST", "/recipes", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipe(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeHTTPHandler_CreateRecipe_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	req := models.CreateRecipeRequest{
		RecipeName:        "Test Recipe",
		RecipeDescription: nil,
		PictureURL:        nil,
		RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440000",
		TotalRecipeCost:   15.50,
	}

	mock.ExpectQuery("INSERT INTO recipes").
		WithArgs(req.RecipeName, req.RecipeDescription, req.PictureURL, req.RecipeCategoryID, req.TotalRecipeCost).
		WillReturnError(sql.ErrConnDone)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRecipe(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to create recipe")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_GetRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	now := time.Now()
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	expectedRecipe := models.Recipe{
		ID:                "550e8400-e29b-41d4-a716-446655440000",
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440001",
		TotalRecipeCost:   15.50,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	}).AddRow(
		expectedRecipe.ID,
		expectedRecipe.RecipeName,
		expectedRecipe.RecipeDescription,
		expectedRecipe.PictureURL,
		expectedRecipe.RecipeCategoryID,
		expectedRecipe.TotalRecipeCost,
		expectedRecipe.CreatedAt,
		expectedRecipe.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs(expectedRecipe.ID).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipes/"+expectedRecipe.ID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.GetRecipe)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipe.ID, result.Data.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.Data.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.Data.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.Data.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.Data.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.Data.TotalRecipeCost)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_GetRecipe_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest("GET", "/recipes/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.GetRecipe)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_GetRecipe_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	request := httptest.NewRequest("GET", "/recipes/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.GetRecipe(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRecipeHTTPHandler_ListRecipes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	now := time.Now()
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	expectedRecipes := []models.Recipe{
		{
			ID:                "550e8400-e29b-41d4-a716-446655440000",
			RecipeName:        "Test Recipe",
			RecipeDescription: &description,
			PictureURL:        &pictureURL,
			RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440001",
			TotalRecipeCost:   15.50,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	})
	for _, r := range expectedRecipes {
		rows.AddRow(
			r.ID, r.RecipeName, r.RecipeDescription, r.PictureURL, r.RecipeCategoryID, r.TotalRecipeCost, r.CreatedAt, r.UpdatedAt,
		)
	}

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs(nil, nil, 50, 0).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipes", nil)
	response := httptest.NewRecorder()

	handler.ListRecipes(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipesResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, len(expectedRecipes), len(result.Data))
	for i, expected := range expectedRecipes {
		assert.Equal(t, expected.ID, result.Data[i].ID)
		assert.Equal(t, expected.RecipeName, result.Data[i].RecipeName)
		assert.Equal(t, expected.RecipeDescription, result.Data[i].RecipeDescription)
		assert.Equal(t, expected.PictureURL, result.Data[i].PictureURL)
		assert.Equal(t, expected.RecipeCategoryID, result.Data[i].RecipeCategoryID)
		assert.Equal(t, expected.TotalRecipeCost, result.Data[i].TotalRecipeCost)
	}
	assert.Equal(t, 1, result.Total)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_ListRecipes_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs("Test", "550e8400-e29b-41d4-a716-446655440000", 10, 5).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/recipes?recipe_name=Test&recipe_category_id=550e8400-e29b-41d4-a716-446655440000&limit=10&offset=5", nil)
	response := httptest.NewRecorder()

	handler.ListRecipes(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipesResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.Data)
	assert.Equal(t, 0, result.Total)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_UpdateRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	now := time.Now()
	recipeName := "Updated Recipe"
	description := "Updated description"
	pictureURL := "https://example.com/updated-image.jpg"
	recipeCategoryID := "550e8400-e29b-41d4-a716-446655440001"
	totalRecipeCost := 20.00
	req := models.UpdateRecipeRequest{
		RecipeName:        &recipeName,
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  &recipeCategoryID,
		TotalRecipeCost:   &totalRecipeCost,
	}

	expectedRecipe := models.Recipe{
		ID:                "550e8400-e29b-41d4-a716-446655440000",
		RecipeName:        recipeName,
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  recipeCategoryID,
		TotalRecipeCost:   totalRecipeCost,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	}).AddRow(
		expectedRecipe.ID,
		expectedRecipe.RecipeName,
		expectedRecipe.RecipeDescription,
		expectedRecipe.PictureURL,
		expectedRecipe.RecipeCategoryID,
		expectedRecipe.TotalRecipeCost,
		expectedRecipe.CreatedAt,
		expectedRecipe.UpdatedAt,
	)

	mock.ExpectQuery("UPDATE recipes").
		WithArgs(expectedRecipe.ID, &recipeName, &description, &pictureURL, &recipeCategoryID, &totalRecipeCost).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipes/"+expectedRecipe.ID, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.UpdateRecipe)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRecipe.ID, result.Data.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.Data.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.Data.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.Data.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.Data.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.Data.TotalRecipeCost)
	assert.Contains(t, result.Message, "updated successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_UpdateRecipe_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	recipeName := "Updated Recipe"
	req := models.UpdateRecipeRequest{
		RecipeName: &recipeName,
	}

	mock.ExpectQuery("UPDATE recipes").
		WithArgs("non-existent-id", &recipeName, nil, nil, nil, nil).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/recipes/non-existent-id", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.UpdateRecipe)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RecipeResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeHTTPHandler_DeleteRecipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	recipeID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(recipeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest("DELETE", "/recipes/"+recipeID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.DeleteRecipe)
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

func TestRecipeHTTPHandler_DeleteRecipe_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	mock.ExpectExec("DELETE FROM recipes").
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	request := httptest.NewRequest("DELETE", "/recipes/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/recipes/{id}", handler.DeleteRecipe)
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

func TestRecipeHTTPHandler_DeleteRecipe_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRecipeHTTPHandler(db, logger)

	request := httptest.NewRequest("DELETE", "/recipes/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.DeleteRecipe(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}
