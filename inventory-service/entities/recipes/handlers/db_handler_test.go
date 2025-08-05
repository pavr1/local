package handlers

import (
	"database/sql"
	"testing"
	"time"

	"inventory-service/entities/recipes/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeDBHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

func TestRecipeDBHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

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

	result, err := handler.Create(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipe.ID, result.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.TotalRecipeCost)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

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

	result, err := handler.Create(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create recipe")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

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

	req := models.GetRecipeRequest{ID: expectedRecipe.ID}
	result, err := handler.GetByID(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipe.ID, result.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.TotalRecipeCost)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	req := models.GetRecipeRequest{ID: "non-existent-id"}
	result, err := handler.GetByID(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipe not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	now := time.Now()
	description1 := "Test description 1"
	description2 := "Test description 2"
	pictureURL1 := "https://example.com/image1.jpg"
	pictureURL2 := "https://example.com/image2.jpg"
	expectedRecipes := []models.Recipe{
		{
			ID:                "550e8400-e29b-41d4-a716-446655440000",
			RecipeName:        "Test Recipe 1",
			RecipeDescription: &description1,
			PictureURL:        &pictureURL1,
			RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440001",
			TotalRecipeCost:   15.50,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "550e8400-e29b-41d4-a716-446655440002",
			RecipeName:        "Test Recipe 2",
			RecipeDescription: &description2,
			PictureURL:        &pictureURL2,
			RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440003",
			TotalRecipeCost:   20.00,
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

	req := models.ListRecipesRequest{}
	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Equal(t, len(expectedRecipes), len(result))
	for i, expected := range expectedRecipes {
		assert.Equal(t, expected.ID, result[i].ID)
		assert.Equal(t, expected.RecipeName, result[i].RecipeName)
		assert.Equal(t, expected.RecipeDescription, result[i].RecipeDescription)
		assert.Equal(t, expected.PictureURL, result[i].PictureURL)
		assert.Equal(t, expected.RecipeCategoryID, result[i].RecipeCategoryID)
		assert.Equal(t, expected.TotalRecipeCost, result[i].TotalRecipeCost)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	recipeName := "Test"
	recipeCategoryID := "550e8400-e29b-41d4-a716-446655440000"
	limit := 10
	offset := 5

	req := models.ListRecipesRequest{
		RecipeName:       &recipeName,
		RecipeCategoryID: &recipeCategoryID,
		Limit:            &limit,
		Offset:           &offset,
	}

	rows := sqlmock.NewRows([]string{
		"id", "recipe_name", "recipe_description", "picture_url", "recipe_category_id", "total_recipe_cost", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at").
		WithArgs(&recipeName, &recipeCategoryID, limit, offset).
		WillReturnRows(rows)

	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Empty(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

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

	result, err := handler.Update(req, expectedRecipe.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipe.ID, result.ID)
	assert.Equal(t, expectedRecipe.RecipeName, result.RecipeName)
	assert.Equal(t, expectedRecipe.RecipeDescription, result.RecipeDescription)
	assert.Equal(t, expectedRecipe.PictureURL, result.PictureURL)
	assert.Equal(t, expectedRecipe.RecipeCategoryID, result.RecipeCategoryID)
	assert.Equal(t, expectedRecipe.TotalRecipeCost, result.TotalRecipeCost)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	recipeName := "Updated Recipe"
	req := models.UpdateRecipeRequest{
		RecipeName: &recipeName,
	}

	mock.ExpectQuery("UPDATE recipes").
		WithArgs("non-existent-id", &recipeName, nil, nil, nil, nil).
		WillReturnError(sql.ErrNoRows)

	result, err := handler.Update(req, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipe not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	req := models.DeleteRecipeRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = handler.Delete(req)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	req := models.DeleteRecipeRequest{ID: "non-existent-id"}

	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipe not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeDBHandler_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeDBHandler(db)

	req := models.DeleteRecipeRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(req.ID).
		WillReturnError(sql.ErrConnDone)

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete recipe")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
