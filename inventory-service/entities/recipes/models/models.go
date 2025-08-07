package models

import (
	"time"
)

// Recipe represents a recipe
type Recipe struct {
	ID                string    `json:"id" db:"id"`
	RecipeName        string    `json:"recipe_name" db:"recipe_name"`
	RecipeDescription *string   `json:"recipe_description" db:"recipe_description"`
	PictureURL        *string   `json:"picture_url" db:"picture_url"`
	RecipeCategoryID  string    `json:"recipe_category_id" db:"recipe_category_id"`
	TotalRecipeCost   float64   `json:"total_recipe_cost" db:"total_recipe_cost"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RecipeIngredient represents a recipe ingredient with quantity
type RecipeIngredient struct {
	IngredientID   string  `json:"ingredient_id" validate:"required,uuid"`
	NumberOfUnits  float64 `json:"number_of_units" validate:"required,min=0.001"`
	IngredientName string  `json:"ingredient_name,omitempty"` // For display purposes
	UnitType       string  `json:"unit_type,omitempty"`       // For display purposes
}

// CreateRecipeRequest represents the request to create a new recipe
type CreateRecipeRequest struct {
	RecipeName        string             `json:"recipe_name" validate:"required,min=1,max=255"`
	RecipeDescription *string            `json:"recipe_description,omitempty"`
	PictureURL        *string            `json:"picture_url,omitempty"`
	RecipeCategoryID  string             `json:"recipe_category_id" validate:"required,uuid"`
	TotalRecipeCost   float64            `json:"total_recipe_cost" validate:"required,min=0"`
	Ingredients       []RecipeIngredient `json:"ingredients" validate:"required,min=1"`
}

// UpdateRecipeRequest represents the request to update a recipe
type UpdateRecipeRequest struct {
	RecipeName        *string  `json:"recipe_name,omitempty" validate:"omitempty,min=1,max=255"`
	RecipeDescription *string  `json:"recipe_description,omitempty"`
	PictureURL        *string  `json:"picture_url,omitempty"`
	RecipeCategoryID  *string  `json:"recipe_category_id,omitempty" validate:"omitempty,uuid"`
	TotalRecipeCost   *float64 `json:"total_recipe_cost,omitempty" validate:"omitempty,min=0"`
}

// GetRecipeRequest represents the request to get a recipe by ID
type GetRecipeRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteRecipeRequest represents the request to delete a recipe
type DeleteRecipeRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListRecipesRequest represents the request to list recipes
type ListRecipesRequest struct {
	RecipeName       *string `json:"recipe_name,omitempty"`
	RecipeCategoryID *string `json:"recipe_category_id,omitempty" validate:"omitempty,uuid"`
	Limit            *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset           *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// RecipeResponse represents a single recipe response
type RecipeResponse struct {
	Success bool   `json:"success"`
	Data    Recipe `json:"data"`
	Message string `json:"message,omitempty"`
}

// RecipesResponse represents multiple recipes response
type RecipesResponse struct {
	Success bool     `json:"success"`
	Data    []Recipe `json:"data"`
	Total   int      `json:"total,omitempty"`
	Message string   `json:"message,omitempty"`
}

// GenericResponse represents a generic response (for delete operations)
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
