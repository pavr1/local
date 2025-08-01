package models

import (
	"time"
)

// Ingredient represents an ingredient used in ice cream production
type Ingredient struct {
	ID                string    `json:"id" db:"id"`
	IngredientName    string    `json:"ingredient_name" db:"ingredient_name"`
	IngredientType    *string   `json:"ingredient_type" db:"ingredient_type"`
	UnitOfMeasure     *string   `json:"unit_of_measure" db:"unit_of_measure"`
	CostPerUnit       *float64  `json:"cost_per_unit" db:"cost_per_unit"`
	SupplierID        *string   `json:"supplier_id" db:"supplier_id"`
	MinimumStockLevel *int      `json:"minimum_stock_level" db:"minimum_stock_level"`
	Notes             *string   `json:"notes" db:"notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateIngredientRequest represents the request to create a new ingredient
type CreateIngredientRequest struct {
	IngredientName    string   `json:"ingredient_name" validate:"required,min=1,max=255"`
	IngredientType    *string  `json:"ingredient_type,omitempty" validate:"omitempty,max=100"`
	UnitOfMeasure     *string  `json:"unit_of_measure,omitempty" validate:"omitempty,max=50"`
	CostPerUnit       *float64 `json:"cost_per_unit,omitempty" validate:"omitempty,min=0"`
	SupplierID        *string  `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	MinimumStockLevel *int     `json:"minimum_stock_level,omitempty" validate:"omitempty,min=0"`
	Notes             *string  `json:"notes,omitempty"`
}

// UpdateIngredientRequest represents the request to update an ingredient
type UpdateIngredientRequest struct {
	IngredientName    *string  `json:"ingredient_name,omitempty" validate:"omitempty,min=1,max=255"`
	IngredientType    *string  `json:"ingredient_type,omitempty" validate:"omitempty,max=100"`
	UnitOfMeasure     *string  `json:"unit_of_measure,omitempty" validate:"omitempty,max=50"`
	CostPerUnit       *float64 `json:"cost_per_unit,omitempty" validate:"omitempty,min=0"`
	SupplierID        *string  `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	MinimumStockLevel *int     `json:"minimum_stock_level,omitempty" validate:"omitempty,min=0"`
	Notes             *string  `json:"notes,omitempty"`
}

// GetIngredientRequest represents the request to get an ingredient by ID
type GetIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteIngredientRequest represents the request to delete an ingredient
type DeleteIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListIngredientsRequest represents the request to list ingredients (for future pagination)
type ListIngredientsRequest struct {
	Limit  *int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset *int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// IngredientResponse represents a single ingredient response
type IngredientResponse struct {
	Success bool       `json:"success"`
	Data    Ingredient `json:"data"`
	Message string     `json:"message,omitempty"`
}

// IngredientsListResponse represents a list of ingredients response
type IngredientsListResponse struct {
	Success bool         `json:"success"`
	Data    []Ingredient `json:"data"`
	Count   int          `json:"count"`
	Message string       `json:"message,omitempty"`
}

// IngredientDeleteResponse represents a delete operation response
type IngredientDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
