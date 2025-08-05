package models

// IngredientCategory represents an ingredient category for classification and reporting
type IngredientCategory struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// CreateIngredientCategoryRequest represents the request to create a new ingredient category
type CreateIngredientCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// UpdateIngredientCategoryRequest represents the request to update an ingredient category
type UpdateIngredientCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=1,max=1000"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// GetIngredientCategoryRequest represents the request to get an ingredient category by ID
type GetIngredientCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteIngredientCategoryRequest represents the request to delete an ingredient category
type DeleteIngredientCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListIngredientCategoriesRequest represents the request to list ingredient categories (for future pagination)
type ListIngredientCategoriesRequest struct {
	Limit  *int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset *int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// IngredientCategoryResponse represents a single ingredient category response
type IngredientCategoryResponse struct {
	Success bool               `json:"success"`
	Data    IngredientCategory `json:"data"`
	Message string             `json:"message,omitempty"`
}

// IngredientCategoriesListResponse represents a list of ingredient categories response
type IngredientCategoriesListResponse struct {
	Success bool                 `json:"success"`
	Data    []IngredientCategory `json:"data"`
	Count   int                  `json:"count"`
	Message string               `json:"message,omitempty"`
}

// IngredientCategoryDeleteResponse represents a delete operation response
type IngredientCategoryDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
