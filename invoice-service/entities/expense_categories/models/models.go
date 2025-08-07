package models

import (
	"time"
)

// ExpenseCategory represents an expense category in the database
type ExpenseCategory struct {
	ID           string    `json:"id" db:"id"`
	CategoryName string    `json:"category_name" db:"category_name"`
	Description  *string   `json:"description" db:"description"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateExpenseCategoryRequest represents the request to create a new expense category
type CreateExpenseCategoryRequest struct {
	CategoryName string  `json:"category_name" validate:"required,min=1,max=100"`
	Description  *string `json:"description,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// UpdateExpenseCategoryRequest represents the request to update an expense category
type UpdateExpenseCategoryRequest struct {
	CategoryName *string `json:"category_name,omitempty" validate:"omitempty,min=1,max=100"`
	Description  *string `json:"description,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// GetExpenseCategoryRequest represents the request to get an expense category by ID
type GetExpenseCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteExpenseCategoryRequest represents the request to delete an expense category
type DeleteExpenseCategoryRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListExpenseCategoriesRequest represents the request to list expense categories
type ListExpenseCategoriesRequest struct {
	Limit    *int  `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset   *int  `json:"offset,omitempty" validate:"omitempty,min=0"`
	IsActive *bool `json:"is_active,omitempty"`
}

// Response Structs
// ExpenseCategoryResponse represents a single expense category response
type ExpenseCategoryResponse struct {
	Success bool            `json:"success"`
	Data    ExpenseCategory `json:"data"`
	Message string          `json:"message,omitempty"`
}

// ExpenseCategoriesListResponse represents a list of expense categories response
type ExpenseCategoryListResponse struct {
	Success bool              `json:"success"`
	Data    []ExpenseCategory `json:"data"`
	Count   int               `json:"count"`
	Message string            `json:"message,omitempty"`
}

// ExpenseCategoryDeleteResponse represents a delete operation response
type ExpenseCategoryDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
} 