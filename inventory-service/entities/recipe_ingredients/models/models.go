package models

import (
	"time"
)

// RecipeIngredient represents a recipe ingredient
type RecipeIngredient struct {
	ID           string    `json:"id" db:"id"`
	RecipeID     string    `json:"recipe_id" db:"recipe_id"`
	IngredientID string    `json:"ingredient_id" db:"ingredient_id"`
	Quantity     float64   `json:"quantity" db:"quantity"`
	UnitType     string    `json:"unit_type" db:"unit_type"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateRecipeIngredientRequest represents the request to create a new recipe ingredient
type CreateRecipeIngredientRequest struct {
	RecipeID     string  `json:"recipe_id" validate:"required,uuid"`
	IngredientID string  `json:"ingredient_id" validate:"required,uuid"`
	Quantity     float64 `json:"quantity" validate:"required,min=0"`
	UnitType     string  `json:"unit_type" validate:"required,min=1,max=50"`
}

// UpdateRecipeIngredientRequest represents the request to update a recipe ingredient
type UpdateRecipeIngredientRequest struct {
	RecipeID     *string  `json:"recipe_id,omitempty" validate:"omitempty,uuid"`
	IngredientID *string  `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Quantity     *float64 `json:"quantity,omitempty" validate:"omitempty,min=0"`
	UnitType     *string  `json:"unit_type,omitempty" validate:"omitempty,min=1,max=50"`
}

// GetRecipeIngredientRequest represents the request to get a recipe ingredient by ID
type GetRecipeIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteRecipeIngredientRequest represents the request to delete a recipe ingredient
type DeleteRecipeIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListRecipeIngredientsRequest represents the request to list recipe ingredients
type ListRecipeIngredientsRequest struct {
	RecipeID     *string `json:"recipe_id,omitempty" validate:"omitempty,uuid"`
	IngredientID *string `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Limit        *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset       *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// RecipeIngredientResponse represents a single recipe ingredient response
type RecipeIngredientResponse struct {
	Success bool             `json:"success"`
	Data    RecipeIngredient `json:"data"`
	Message string           `json:"message,omitempty"`
}

// RecipeIngredientsResponse represents multiple recipe ingredients response
type RecipeIngredientsResponse struct {
	Success bool               `json:"success"`
	Data    []RecipeIngredient `json:"data"`
	Total   int                `json:"total,omitempty"`
	Message string             `json:"message,omitempty"`
}

// GenericResponse represents a generic response (for delete operations)
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
