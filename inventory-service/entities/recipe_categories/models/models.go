package models

import (
	"time"
)

// RecipeCategory represents a recipe category
type RecipeCategory struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateRecipeCategoryRequest represents the request to create a new recipe category
type CreateRecipeCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// UpdateRecipeCategoryRequest represents the request to update a recipe category
type UpdateRecipeCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// GetRecipeCategoryRequest represents the request to get a recipe category by ID
type GetRecipeCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteRecipeCategoryRequest represents the request to delete a recipe category
type DeleteRecipeCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListRecipeCategoriesRequest represents the request to list recipe categories
type ListRecipeCategoriesRequest struct {
	Name   *string `json:"name,omitempty"`
	Limit  *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// RecipeCategoryResponse represents a single recipe category response
type RecipeCategoryResponse struct {
	Success bool           `json:"success"`
	Data    RecipeCategory `json:"data"`
	Message string         `json:"message,omitempty"`
}

// RecipeCategoriesResponse represents multiple recipe categories response
type RecipeCategoriesResponse struct {
	Success bool             `json:"success"`
	Data    []RecipeCategory `json:"data"`
	Total   int              `json:"total,omitempty"`
	Message string           `json:"message,omitempty"`
}

// GenericResponse represents a generic response (for delete operations)
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
