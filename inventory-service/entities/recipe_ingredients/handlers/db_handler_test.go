package handlers

import (
	"database/sql"
	"testing"
	"time"

	"inventory-service/entities/recipe_ingredients/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

func TestNewRecipeIngredientDBHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

func TestRecipeIngredientDBHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	req := models.CreateRecipeIngredientRequest{
		RecipeID:     "550e8400-e29b-41d4-a716-446655440000",
		IngredientID: "550e8400-e29b-41d4-a716-446655440001",
		Quantity:     2.5,
	}

	now := time.Now()
	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440002",
		RecipeID:     req.RecipeID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("INSERT INTO recipe_ingredients").
		WithArgs(req.RecipeID, req.IngredientID, req.Quantity).
		WillReturnRows(rows)

	result, err := handler.Create(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeIngredient.ID, result.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Quantity)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	req := models.CreateRecipeIngredientRequest{
		RecipeID:     "550e8400-e29b-41d4-a716-446655440000",
		IngredientID: "550e8400-e29b-41d4-a716-446655440001",
		Quantity:     2.5,
	}

	mock.ExpectQuery("INSERT INTO recipe_ingredients").
		WithArgs(req.RecipeID, req.IngredientID, req.Quantity).
		WillReturnError(sql.ErrConnDone)

	result, err := handler.Create(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "sql: connection is already closed")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	now := time.Now()
	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		RecipeID:     "550e8400-e29b-41d4-a716-446655440001",
		IngredientID: "550e8400-e29b-41d4-a716-446655440002",
		Quantity:     2.5,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, created_at, updated_at").
		WithArgs(expectedRecipeIngredient.ID).
		WillReturnRows(rows)

	req := models.GetRecipeIngredientRequest{ID: expectedRecipeIngredient.ID}
	result, err := handler.GetByID(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeIngredient.ID, result.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Quantity)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	req := models.GetRecipeIngredientRequest{ID: "non-existent-id"}
	result, err := handler.GetByID(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	now := time.Now()
	expectedRecipeIngredients := []models.RecipeIngredient{
		{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			RecipeID:     "550e8400-e29b-41d4-a716-446655440001",
			IngredientID: "550e8400-e29b-41d4-a716-446655440002",
			Quantity:     2.5,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "550e8400-e29b-41d4-a716-446655440003",
			RecipeID:     "550e8400-e29b-41d4-a716-446655440001",
			IngredientID: "550e8400-e29b-41d4-a716-446655440004",
			Quantity:     1.0,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "created_at", "updated_at",
	})
	for _, ri := range expectedRecipeIngredients {
		rows.AddRow(
			ri.ID, ri.RecipeID, ri.IngredientID, ri.Quantity, ri.CreatedAt, ri.UpdatedAt,
		)
	}

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, created_at, updated_at").
		WithArgs(nil, nil, 50, 0).
		WillReturnRows(rows)

	req := models.ListRecipeIngredientsRequest{}
	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Equal(t, len(expectedRecipeIngredients), len(result))
	for i, expected := range expectedRecipeIngredients {
		assert.Equal(t, expected.ID, result[i].ID)
		assert.Equal(t, expected.RecipeID, result[i].RecipeID)
		assert.Equal(t, expected.IngredientID, result[i].IngredientID)
		assert.Equal(t, expected.Quantity, result[i].Quantity)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	recipeID := "550e8400-e29b-41d4-a716-446655440000"
	ingredientID := "550e8400-e29b-41d4-a716-446655440001"
	limit := 10
	offset := 5

	req := models.ListRecipeIngredientsRequest{
		RecipeID:     &recipeID,
		IngredientID: &ingredientID,
		Limit:        &limit,
		Offset:       &offset,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, recipe_id, ingredient_id, quantity, created_at, updated_at").
		WithArgs(&recipeID, &ingredientID, limit, offset).
		WillReturnRows(rows)

	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Empty(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	now := time.Now()
	recipeID := "550e8400-e29b-41d4-a716-446655440001"
	ingredientID := "550e8400-e29b-41d4-a716-446655440002"
	quantity := 3.0
	req := models.UpdateRecipeIngredientRequest{
		RecipeID:     &recipeID,
		IngredientID: &ingredientID,
		Quantity:     &quantity,
	}

	expectedRecipeIngredient := models.RecipeIngredient{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		RecipeID:     recipeID,
		IngredientID: ingredientID,
		Quantity:     quantity,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_id", "ingredient_id", "quantity", "created_at", "updated_at",
	}).AddRow(
		expectedRecipeIngredient.ID,
		expectedRecipeIngredient.RecipeID,
		expectedRecipeIngredient.IngredientID,
		expectedRecipeIngredient.Quantity,
		expectedRecipeIngredient.CreatedAt,
		expectedRecipeIngredient.UpdatedAt,
	)

	mock.ExpectQuery("UPDATE recipe_ingredients").
		WithArgs(expectedRecipeIngredient.ID, &recipeID, &ingredientID, &quantity).
		WillReturnRows(rows)

	result, err := handler.Update(req, expectedRecipeIngredient.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeIngredient.ID, result.ID)
	assert.Equal(t, expectedRecipeIngredient.RecipeID, result.RecipeID)
	assert.Equal(t, expectedRecipeIngredient.IngredientID, result.IngredientID)
	assert.Equal(t, expectedRecipeIngredient.Quantity, result.Quantity)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	quantity := 3.0
	req := models.UpdateRecipeIngredientRequest{
		Quantity: &quantity,
	}

	mock.ExpectQuery("UPDATE recipe_ingredients").
		WithArgs("non-existent-id", nil, nil, &quantity).
		WillReturnError(sql.ErrNoRows)

	result, err := handler.Update(req, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	req := models.DeleteRecipeIngredientRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = handler.Delete(req)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	req := models.DeleteRecipeIngredientRequest{ID: "non-existent-id"}

	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeIngredientDBHandler_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeIngredientDBHandler(db, testLogger())

	req := models.DeleteRecipeIngredientRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipe_ingredients").
		WithArgs(req.ID).
		WillReturnError(sql.ErrConnDone)

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql: connection is already closed")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
