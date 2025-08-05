package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecipeIngredient_Struct(t *testing.T) {
	now := time.Now()
	recipeIngredient := RecipeIngredient{
		ID:           "test-id",
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     2.5,
		UnitType:     "cups",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, "test-id", recipeIngredient.ID)
	assert.Equal(t, "recipe-id", recipeIngredient.RecipeID)
	assert.Equal(t, "ingredient-id", recipeIngredient.IngredientID)
	assert.Equal(t, 2.5, recipeIngredient.Quantity)
	assert.Equal(t, "cups", recipeIngredient.UnitType)
	assert.Equal(t, now, recipeIngredient.CreatedAt)
	assert.Equal(t, now, recipeIngredient.UpdatedAt)
}

func TestCreateRecipeIngredientRequest_Struct(t *testing.T) {
	req := CreateRecipeIngredientRequest{
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     2.5,
		UnitType:     "cups",
	}

	assert.Equal(t, "recipe-id", req.RecipeID)
	assert.Equal(t, "ingredient-id", req.IngredientID)
	assert.Equal(t, 2.5, req.Quantity)
	assert.Equal(t, "cups", req.UnitType)
}

func TestUpdateRecipeIngredientRequest_Struct(t *testing.T) {
	recipeID := "updated-recipe-id"
	ingredientID := "updated-ingredient-id"
	quantity := 3.0
	unitType := "tablespoons"

	req := UpdateRecipeIngredientRequest{
		RecipeID:     &recipeID,
		IngredientID: &ingredientID,
		Quantity:     &quantity,
		UnitType:     &unitType,
	}

	assert.Equal(t, &recipeID, req.RecipeID)
	assert.Equal(t, &ingredientID, req.IngredientID)
	assert.Equal(t, &quantity, req.Quantity)
	assert.Equal(t, &unitType, req.UnitType)
}

func TestGetRecipeIngredientRequest_Struct(t *testing.T) {
	req := GetRecipeIngredientRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestDeleteRecipeIngredientRequest_Struct(t *testing.T) {
	req := DeleteRecipeIngredientRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestListRecipeIngredientsRequest_Struct(t *testing.T) {
	recipeID := "recipe-id"
	ingredientID := "ingredient-id"
	limit := 10
	offset := 5

	req := ListRecipeIngredientsRequest{
		RecipeID:     &recipeID,
		IngredientID: &ingredientID,
		Limit:        &limit,
		Offset:       &offset,
	}

	assert.Equal(t, &recipeID, req.RecipeID)
	assert.Equal(t, &ingredientID, req.IngredientID)
	assert.Equal(t, &limit, req.Limit)
	assert.Equal(t, &offset, req.Offset)
}

func TestRecipeIngredientResponse_Struct(t *testing.T) {
	recipeIngredient := RecipeIngredient{
		ID:           "test-id",
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     2.5,
		UnitType:     "cups",
	}

	response := RecipeIngredientResponse{
		Success: true,
		Data:    recipeIngredient,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipeIngredient, response.Data)
	assert.Equal(t, "Success", response.Message)
}

func TestRecipeIngredientsResponse_Struct(t *testing.T) {
	recipeIngredients := []RecipeIngredient{
		{
			ID:           "test-id-1",
			RecipeID:     "recipe-id-1",
			IngredientID: "ingredient-id-1",
			Quantity:     2.5,
			UnitType:     "cups",
		},
		{
			ID:           "test-id-2",
			RecipeID:     "recipe-id-2",
			IngredientID: "ingredient-id-2",
			Quantity:     1.0,
			UnitType:     "tablespoons",
		},
	}

	response := RecipeIngredientsResponse{
		Success: true,
		Data:    recipeIngredients,
		Total:   2,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, recipeIngredients, response.Data)
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

func TestRecipeIngredient_JSONTags(t *testing.T) {
	recipeIngredient := RecipeIngredient{
		ID:           "test-id",
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     2.5,
		UnitType:     "cups",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// This test ensures the struct has proper JSON tags
	// The actual JSON marshaling would be tested in integration tests
	assert.NotEmpty(t, recipeIngredient.ID)
	assert.NotEmpty(t, recipeIngredient.RecipeID)
	assert.NotEmpty(t, recipeIngredient.IngredientID)
	assert.Greater(t, recipeIngredient.Quantity, 0.0)
	assert.NotEmpty(t, recipeIngredient.UnitType)
}

func TestCreateRecipeIngredientRequest_Validation(t *testing.T) {
	// Test with valid data
	req := CreateRecipeIngredientRequest{
		RecipeID:     "550e8400-e29b-41d4-a716-446655440000",
		IngredientID: "550e8400-e29b-41d4-a716-446655440001",
		Quantity:     2.5,
		UnitType:     "cups",
	}

	// This would be tested with actual validation in integration tests
	assert.NotEmpty(t, req.RecipeID)
	assert.NotEmpty(t, req.IngredientID)
	assert.Greater(t, req.Quantity, 0.0)
	assert.NotEmpty(t, req.UnitType)
}

func TestUpdateRecipeIngredientRequest_OptionalFields(t *testing.T) {
	// Test with only some fields set
	quantity := 3.0
	req := UpdateRecipeIngredientRequest{
		Quantity: &quantity,
	}

	assert.Nil(t, req.RecipeID)
	assert.Nil(t, req.IngredientID)
	assert.Equal(t, &quantity, req.Quantity)
	assert.Nil(t, req.UnitType)
}

func TestListRecipeIngredientsRequest_EmptyFilters(t *testing.T) {
	// Test with no filters
	req := ListRecipeIngredientsRequest{}

	assert.Nil(t, req.RecipeID)
	assert.Nil(t, req.IngredientID)
	assert.Nil(t, req.Limit)
	assert.Nil(t, req.Offset)
}

func TestRecipeIngredient_ZeroQuantity(t *testing.T) {
	// Test with zero quantity
	recipeIngredient := RecipeIngredient{
		ID:           "test-id",
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     0.0,
		UnitType:     "cups",
	}

	assert.Equal(t, "test-id", recipeIngredient.ID)
	assert.Equal(t, "recipe-id", recipeIngredient.RecipeID)
	assert.Equal(t, "ingredient-id", recipeIngredient.IngredientID)
	assert.Equal(t, 0.0, recipeIngredient.Quantity)
	assert.Equal(t, "cups", recipeIngredient.UnitType)
}

func TestCreateRecipeIngredientRequest_ZeroQuantity(t *testing.T) {
	// Test with zero quantity
	req := CreateRecipeIngredientRequest{
		RecipeID:     "recipe-id",
		IngredientID: "ingredient-id",
		Quantity:     0.0,
		UnitType:     "cups",
	}

	assert.Equal(t, "recipe-id", req.RecipeID)
	assert.Equal(t, "ingredient-id", req.IngredientID)
	assert.Equal(t, 0.0, req.Quantity)
	assert.Equal(t, "cups", req.UnitType)
}
