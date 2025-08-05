package handlers

import (
	"database/sql"
	"testing"
	"time"

	"inventory-service/entities/recipe_categories/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeCategoryDBHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

func TestRecipeCategoryDBHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

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

	result, err := handler.Create(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeCategory.ID, result.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Description)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	req := models.CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: nil,
	}

	mock.ExpectQuery("INSERT INTO recipe_categories").
		WithArgs(req.Name, req.Description).
		WillReturnError(sql.ErrConnDone)

	result, err := handler.Create(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create recipe category")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

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

	req := models.GetRecipeCategoryRequest{ID: expectedRecipeCategory.ID}
	result, err := handler.GetByID(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeCategory.ID, result.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Description)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	req := models.GetRecipeCategoryRequest{ID: "non-existent-id"}
	result, err := handler.GetByID(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipe category not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	now := time.Now()
	description1 := "Test description 1"
	description2 := "Test description 2"
	expectedRecipeCategories := []models.RecipeCategory{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "Test Category 1",
			Description: &description1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440001",
			Name:        "Test Category 2",
			Description: &description2,
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

	req := models.ListRecipeCategoriesRequest{}
	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Equal(t, len(expectedRecipeCategories), len(result))
	for i, expected := range expectedRecipeCategories {
		assert.Equal(t, expected.ID, result[i].ID)
		assert.Equal(t, expected.Name, result[i].Name)
		assert.Equal(t, expected.Description, result[i].Description)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	name := "Test"
	limit := 10
	offset := 5

	req := models.ListRecipeCategoriesRequest{
		Name:   &name,
		Limit:  &limit,
		Offset: &offset,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, name, description, created_at, updated_at").
		WithArgs(&name, limit, offset).
		WillReturnRows(rows)

	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Empty(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

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

	result, err := handler.Update(req, expectedRecipeCategory.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedRecipeCategory.ID, result.ID)
	assert.Equal(t, expectedRecipeCategory.Name, result.Name)
	assert.Equal(t, expectedRecipeCategory.Description, result.Description)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	name := "Updated Category"
	req := models.UpdateRecipeCategoryRequest{
		Name: &name,
	}

	mock.ExpectQuery("UPDATE recipe_categories").
		WithArgs("non-existent-id", &name, nil).
		WillReturnError(sql.ErrNoRows)

	result, err := handler.Update(req, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipe category not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	req := models.DeleteRecipeCategoryRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipe_categories").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = handler.Delete(req)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	req := models.DeleteRecipeCategoryRequest{ID: "non-existent-id"}

	mock.ExpectExec("DELETE FROM recipe_categories").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipe category not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRecipeCategoryDBHandler_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRecipeCategoryDBHandler(db)

	req := models.DeleteRecipeCategoryRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM recipe_categories").
		WithArgs(req.ID).
		WillReturnError(sql.ErrConnDone)

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete recipe category")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
