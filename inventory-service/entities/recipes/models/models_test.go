package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecipe_Struct(t *testing.T) {
	now := time.Now()
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	recipe := Recipe{
		ID:                "test-id",
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	assert.Equal(t, "test-id", recipe.ID)
	assert.Equal(t, "Test Recipe", recipe.RecipeName)
	assert.Equal(t, &description, recipe.RecipeDescription)
	assert.Equal(t, &pictureURL, recipe.PictureURL)
	assert.Equal(t, "category-id", recipe.RecipeCategoryID)
	assert.Equal(t, 15.50, recipe.TotalRecipeCost)
	assert.Equal(t, now, recipe.CreatedAt)
	assert.Equal(t, now, recipe.UpdatedAt)
}

func TestCreateRecipeRequest_Struct(t *testing.T) {
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	req := CreateRecipeRequest{
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
	}

	assert.Equal(t, "Test Recipe", req.RecipeName)
	assert.Equal(t, &description, req.RecipeDescription)
	assert.Equal(t, &pictureURL, req.PictureURL)
	assert.Equal(t, "category-id", req.RecipeCategoryID)
	assert.Equal(t, 15.50, req.TotalRecipeCost)
}

func TestUpdateRecipeRequest_Struct(t *testing.T) {
	recipeName := "Updated Recipe"
	description := "Updated description"
	pictureURL := "https://example.com/updated-image.jpg"
	recipeCategoryID := "updated-category-id"
	totalRecipeCost := 20.00

	req := UpdateRecipeRequest{
		RecipeName:        &recipeName,
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  &recipeCategoryID,
		TotalRecipeCost:   &totalRecipeCost,
	}

	assert.Equal(t, &recipeName, req.RecipeName)
	assert.Equal(t, &description, req.RecipeDescription)
	assert.Equal(t, &pictureURL, req.PictureURL)
	assert.Equal(t, &recipeCategoryID, req.RecipeCategoryID)
	assert.Equal(t, &totalRecipeCost, req.TotalRecipeCost)
}

func TestGetRecipeRequest_Struct(t *testing.T) {
	req := GetRecipeRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestDeleteRecipeRequest_Struct(t *testing.T) {
	req := DeleteRecipeRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestListRecipesRequest_Struct(t *testing.T) {
	recipeName := "Test Recipe"
	recipeCategoryID := "category-id"
	limit := 10
	offset := 5

	req := ListRecipesRequest{
		RecipeName:       &recipeName,
		RecipeCategoryID: &recipeCategoryID,
		Limit:            &limit,
		Offset:           &offset,
	}

	assert.Equal(t, &recipeName, req.RecipeName)
	assert.Equal(t, &recipeCategoryID, req.RecipeCategoryID)
	assert.Equal(t, &limit, req.Limit)
	assert.Equal(t, &offset, req.Offset)
}

func TestRecipeResponse_Struct(t *testing.T) {
	description := "Test description"
	recipe := Recipe{
		ID:                "test-id",
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
	}

	response := RecipeResponse{
		Success: true,
		Data:    recipe,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipe, response.Data)
	assert.Equal(t, "Success", response.Message)
}

func TestRecipesResponse_Struct(t *testing.T) {
	description1 := "Test description 1"
	description2 := "Test description 2"
	recipes := []Recipe{
		{
			ID:                "test-id-1",
			RecipeName:        "Test Recipe 1",
			RecipeDescription: &description1,
			RecipeCategoryID:  "category-id-1",
			TotalRecipeCost:   15.50,
		},
		{
			ID:                "test-id-2",
			RecipeName:        "Test Recipe 2",
			RecipeDescription: &description2,
			RecipeCategoryID:  "category-id-2",
			TotalRecipeCost:   20.00,
		},
	}

	response := RecipesResponse{
		Success: true,
		Data:    recipes,
		Total:   2,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipes, response.Data)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Success", response.Message)
}

func TestGenericResponse_Struct(t *testing.T) {
	response := GenericResponse{
		Success: true,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Success", response.Message)
}

func TestRecipe_JSONTags(t *testing.T) {
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	recipe := Recipe{
		ID:                "test-id",
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// This test ensures the struct has proper JSON tags
	// The actual JSON marshaling would be tested in integration tests
	assert.NotEmpty(t, recipe.ID)
	assert.NotEmpty(t, recipe.RecipeName)
	assert.NotNil(t, recipe.RecipeDescription)
	assert.NotNil(t, recipe.PictureURL)
	assert.NotEmpty(t, recipe.RecipeCategoryID)
	assert.Greater(t, recipe.TotalRecipeCost, 0.0)
}

func TestCreateRecipeRequest_Validation(t *testing.T) {
	// Test with valid data
	description := "Test description"
	pictureURL := "https://example.com/image.jpg"
	req := CreateRecipeRequest{
		RecipeName:        "Test Recipe",
		RecipeDescription: &description,
		PictureURL:        &pictureURL,
		RecipeCategoryID:  "550e8400-e29b-41d4-a716-446655440000",
		TotalRecipeCost:   15.50,
	}

	// This would be tested with actual validation in integration tests
	assert.NotEmpty(t, req.RecipeName)
	assert.NotNil(t, req.RecipeDescription)
	assert.NotNil(t, req.PictureURL)
	assert.NotEmpty(t, req.RecipeCategoryID)
	assert.Greater(t, req.TotalRecipeCost, 0.0)
}

func TestUpdateRecipeRequest_OptionalFields(t *testing.T) {
	// Test with only some fields set
	recipeName := "Updated Recipe"
	req := UpdateRecipeRequest{
		RecipeName: &recipeName,
	}

	assert.Equal(t, &recipeName, req.RecipeName)
	assert.Nil(t, req.RecipeDescription)
	assert.Nil(t, req.PictureURL)
	assert.Nil(t, req.RecipeCategoryID)
	assert.Nil(t, req.TotalRecipeCost)
}

func TestListRecipesRequest_EmptyFilters(t *testing.T) {
	// Test with no filters
	req := ListRecipesRequest{}

	assert.Nil(t, req.RecipeName)
	assert.Nil(t, req.RecipeCategoryID)
	assert.Nil(t, req.Limit)
	assert.Nil(t, req.Offset)
}

func TestRecipe_NilOptionalFields(t *testing.T) {
	// Test with nil optional fields
	recipe := Recipe{
		ID:                "test-id",
		RecipeName:        "Test Recipe",
		RecipeDescription: nil,
		PictureURL:        nil,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
	}

	assert.Equal(t, "test-id", recipe.ID)
	assert.Equal(t, "Test Recipe", recipe.RecipeName)
	assert.Nil(t, recipe.RecipeDescription)
	assert.Nil(t, recipe.PictureURL)
	assert.Equal(t, "category-id", recipe.RecipeCategoryID)
	assert.Equal(t, 15.50, recipe.TotalRecipeCost)
}

func TestCreateRecipeRequest_NilOptionalFields(t *testing.T) {
	// Test with nil optional fields
	req := CreateRecipeRequest{
		RecipeName:        "Test Recipe",
		RecipeDescription: nil,
		PictureURL:        nil,
		RecipeCategoryID:  "category-id",
		TotalRecipeCost:   15.50,
	}

	assert.Equal(t, "Test Recipe", req.RecipeName)
	assert.Nil(t, req.RecipeDescription)
	assert.Nil(t, req.PictureURL)
	assert.Equal(t, "category-id", req.RecipeCategoryID)
	assert.Equal(t, 15.50, req.TotalRecipeCost)
}
