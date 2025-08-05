package models

import (
	"time"
)

// RunoutIngredient represents a runout ingredient report
type RunoutIngredient struct {
	ID          string    `json:"id" db:"id"`
	ExistenceID string    `json:"existence_id" db:"existence_id"`
	EmployeeID  string    `json:"employee_id" db:"employee_id"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitType    string    `json:"unit_type" db:"unit_type"`
	ReportDate  time.Time `json:"report_date" db:"report_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateRunoutIngredientRequest represents the request to create a new runout ingredient
type CreateRunoutIngredientRequest struct {
	ExistenceID string     `json:"existence_id" validate:"required,uuid"`
	EmployeeID  string     `json:"employee_id" validate:"required,uuid"`
	Quantity    float64    `json:"quantity" validate:"required,min=0.01"`
	UnitType    string     `json:"unit_type" validate:"required,oneof=Liters Gallons Units Bag"`
	ReportDate  *time.Time `json:"report_date,omitempty"`
}

// UpdateRunoutIngredientRequest represents the request to update a runout ingredient
type UpdateRunoutIngredientRequest struct {
	Quantity   *float64   `json:"quantity,omitempty" validate:"omitempty,min=0.01"`
	UnitType   *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	ReportDate *time.Time `json:"report_date,omitempty"`
}

// GetRunoutIngredientRequest represents the request to get a runout ingredient by ID
type GetRunoutIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteRunoutIngredientRequest represents the request to delete a runout ingredient
type DeleteRunoutIngredientRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListRunoutIngredientsRequest represents the request to list runout ingredients
type ListRunoutIngredientsRequest struct {
	ExistenceID *string    `json:"existence_id,omitempty" validate:"omitempty,uuid"`
	EmployeeID  *string    `json:"employee_id,omitempty" validate:"omitempty,uuid"`
	UnitType    *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	ReportDate  *time.Time `json:"report_date,omitempty"`
	Limit       *int       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset      *int       `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// RunoutIngredientResponse represents a single runout ingredient response
type RunoutIngredientResponse struct {
	Success bool             `json:"success"`
	Data    RunoutIngredient `json:"data"`
	Message string           `json:"message,omitempty"`
}

// RunoutIngredientsResponse represents multiple runout ingredients response
type RunoutIngredientsResponse struct {
	Success bool               `json:"success"`
	Data    []RunoutIngredient `json:"data"`
	Total   int                `json:"total,omitempty"`
	Message string             `json:"message,omitempty"`
}

// GenericResponse represents a generic response (for delete operations)
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
