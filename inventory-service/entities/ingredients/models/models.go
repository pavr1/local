package models

// Ingredient represents an ingredient used in ice cream production
type Ingredient struct {
	ID         string  `json:"id" db:"id"`
	Name       string  `json:"name" db:"name"`
	SupplierID *string `json:"supplier_id" db:"supplier_id"`
}

// CreateIngredientRequest represents the request to create a new ingredient
type CreateIngredientRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=255"`
	SupplierID *string `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateIngredientRequest represents the request to update an ingredient
type UpdateIngredientRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	SupplierID *string `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
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
