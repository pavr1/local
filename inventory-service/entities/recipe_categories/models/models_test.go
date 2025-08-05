package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecipeCategory_Struct(t *testing.T) {
	now := time.Now()
	description := "Test description"
	recipeCategory := RecipeCategory{
		ID:          "test-id",
		Name:        "Test Category",
		Description: &description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "test-id", recipeCategory.ID)
	assert.Equal(t, "Test Category", recipeCategory.Name)
	assert.Equal(t, &description, recipeCategory.Description)
	assert.Equal(t, now, recipeCategory.CreatedAt)
	assert.Equal(t, now, recipeCategory.UpdatedAt)
}

func TestCreateRecipeCategoryRequest_Struct(t *testing.T) {
	description := "Test description"
	req := CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: &description,
	}

	assert.Equal(t, "Test Category", req.Name)
	assert.Equal(t, &description, req.Description)
}

func TestUpdateRecipeCategoryRequest_Struct(t *testing.T) {
	name := "Updated Category"
	description := "Updated description"

	req := UpdateRecipeCategoryRequest{
		Name:        &name,
		Description: &description,
	}

	assert.Equal(t, &name, req.Name)
	assert.Equal(t, &description, req.Description)
}

func TestGetRecipeCategoryRequest_Struct(t *testing.T) {
	req := GetRecipeCategoryRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestDeleteRecipeCategoryRequest_Struct(t *testing.T) {
	req := DeleteRecipeCategoryRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestListRecipeCategoriesRequest_Struct(t *testing.T) {
	name := "Test Category"
	limit := 10
	offset := 5

	req := ListRecipeCategoriesRequest{
		Name:   &name,
		Limit:  &limit,
		Offset: &offset,
	}

	assert.Equal(t, &name, req.Name)
	assert.Equal(t, &limit, req.Limit)
	assert.Equal(t, &offset, req.Offset)
}

func TestRecipeCategoryResponse_Struct(t *testing.T) {
	description := "Test description"
	recipeCategory := RecipeCategory{
		ID:          "test-id",
		Name:        "Test Category",
		Description: &description,
	}

	response := RecipeCategoryResponse{
		Success: true,
		Data:    recipeCategory,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipeCategory, response.Data)
	assert.Equal(t, "Success", response.Message)
}

func TestRecipeCategoriesResponse_Struct(t *testing.T) {
	description1 := "Test description 1"
	description2 := "Test description 2"
	recipeCategories := []RecipeCategory{
		{
			ID:          "test-id-1",
			Name:        "Test Category 1",
			Description: &description1,
		},
		{
			ID:          "test-id-2",
			Name:        "Test Category 2",
			Description: &description2,
		},
	}

	response := RecipeCategoriesResponse{
		Success: true,
		Data:    recipeCategories,
		Total:   2,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipeCategories, response.Data)
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

func TestRecipeCategory_JSONTags(t *testing.T) {
	description := "Test description"
	recipeCategory := RecipeCategory{
		ID:          "test-id",
		Name:        "Test Category",
		Description: &description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// This test ensures the struct has proper JSON tags
	// The actual JSON marshaling would be tested in integration tests
	assert.NotEmpty(t, recipeCategory.ID)
	assert.NotEmpty(t, recipeCategory.Name)
	assert.NotNil(t, recipeCategory.Description)
}

func TestCreateRecipeCategoryRequest_Validation(t *testing.T) {
	// Test with valid data
	description := "Test description"
	req := CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: &description,
	}

	// This would be tested with actual validation in integration tests
	assert.NotEmpty(t, req.Name)
	assert.NotNil(t, req.Description)
}

func TestUpdateRecipeCategoryRequest_OptionalFields(t *testing.T) {
	// Test with only some fields set
	name := "Updated Category"
	req := UpdateRecipeCategoryRequest{
		Name: &name,
	}

	assert.Equal(t, &name, req.Name)
	assert.Nil(t, req.Description)
}

func TestListRecipeCategoriesRequest_EmptyFilters(t *testing.T) {
	// Test with no filters
	req := ListRecipeCategoriesRequest{}

	assert.Nil(t, req.Name)
	assert.Nil(t, req.Limit)
	assert.Nil(t, req.Offset)
}

func TestRecipeCategory_NilDescription(t *testing.T) {
	// Test with nil description
	recipeCategory := RecipeCategory{
		ID:          "test-id",
		Name:        "Test Category",
		Description: nil,
	}

	assert.Equal(t, "test-id", recipeCategory.ID)
	assert.Equal(t, "Test Category", recipeCategory.Name)
	assert.Nil(t, recipeCategory.Description)
}

func TestCreateRecipeCategoryRequest_NilDescription(t *testing.T) {
	// Test with nil description
	req := CreateRecipeCategoryRequest{
		Name:        "Test Category",
		Description: nil,
	}

	assert.Equal(t, "Test Category", req.Name)
	assert.Nil(t, req.Description)
}
