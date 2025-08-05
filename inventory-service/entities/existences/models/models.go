package models

import (
	"time"
)

// Existence represents a specific ingredient purchase/acquisition batch
type Existence struct {
	ID                     string     `json:"id" db:"id"`
	ExistenceReferenceCode int        `json:"existence_reference_code" db:"existence_reference_code"`
	IngredientID           string     `json:"ingredient_id" db:"ingredient_id"`
	InvoiceDetailID        string     `json:"invoice_detail_id" db:"invoice_detail_id"`
	UnitsPurchased         float64    `json:"units_purchased" db:"units_purchased"`
	UnitsAvailable         float64    `json:"units_available" db:"units_available"`
	UnitType               string     `json:"unit_type" db:"unit_type"`
	ItemsPerUnit           int        `json:"items_per_unit" db:"items_per_unit"`
	CostPerItem            float64    `json:"cost_per_item" db:"cost_per_item"`
	CostPerUnit            float64    `json:"cost_per_unit" db:"cost_per_unit"`
	TotalPurchaseCost      float64    `json:"total_purchase_cost" db:"total_purchase_cost"`
	RemainingValue         float64    `json:"remaining_value" db:"remaining_value"`
	ExpirationDate         *time.Time `json:"expiration_date" db:"expiration_date"`
	IncomeMarginPercentage float64    `json:"income_margin_percentage" db:"income_margin_percentage"`
	IncomeMarginAmount     float64    `json:"income_margin_amount" db:"income_margin_amount"`
	IvaPercentage          float64    `json:"iva_percentage" db:"iva_percentage"`
	IvaAmount              float64    `json:"iva_amount" db:"iva_amount"`
	ServiceTaxPercentage   float64    `json:"service_tax_percentage" db:"service_tax_percentage"`
	ServiceTaxAmount       float64    `json:"service_tax_amount" db:"service_tax_amount"`
	CalculatedPrice        float64    `json:"calculated_price" db:"calculated_price"`
	FinalPrice             *float64   `json:"final_price" db:"final_price"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateExistenceRequest represents the request to create a new existence
type CreateExistenceRequest struct {
	IngredientID           string     `json:"ingredient_id" validate:"required,uuid"`
	InvoiceDetailID        string     `json:"invoice_detail_id" validate:"required,uuid"`
	UnitsPurchased         float64    `json:"units_purchased" validate:"required,min=0.01"`
	UnitsAvailable         float64    `json:"units_available" validate:"required,min=0"`
	UnitType               string     `json:"unit_type" validate:"required,oneof=Liters Gallons Units Bag"`
	ItemsPerUnit           int        `json:"items_per_unit" validate:"required,min=1"`
	CostPerUnit            float64    `json:"cost_per_unit" validate:"required,min=0.01"`
	ExpirationDate         *time.Time `json:"expiration_date,omitempty"`
	IncomeMarginPercentage *float64   `json:"income_margin_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	IvaPercentage          *float64   `json:"iva_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	ServiceTaxPercentage   *float64   `json:"service_tax_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	FinalPrice             *float64   `json:"final_price,omitempty" validate:"omitempty,min=0"`
}

// UpdateExistenceRequest represents the request to update an existence
type UpdateExistenceRequest struct {
	UnitsAvailable         *float64   `json:"units_available,omitempty" validate:"omitempty,min=0"`
	UnitType               *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	ItemsPerUnit           *int       `json:"items_per_unit,omitempty" validate:"omitempty,min=1"`
	CostPerUnit            *float64   `json:"cost_per_unit,omitempty" validate:"omitempty,min=0.01"`
	ExpirationDate         *time.Time `json:"expiration_date,omitempty"`
	IncomeMarginPercentage *float64   `json:"income_margin_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	IvaPercentage          *float64   `json:"iva_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	ServiceTaxPercentage   *float64   `json:"service_tax_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	FinalPrice             *float64   `json:"final_price,omitempty" validate:"omitempty,min=0"`
}

// GetExistenceRequest represents the request to get an existence by ID
type GetExistenceRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteExistenceRequest represents the request to delete an existence
type DeleteExistenceRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListExistencesRequest represents the request to list existences (for pagination)
type ListExistencesRequest struct {
	IngredientID *string `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	UnitType     *string `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	Expired      *bool   `json:"expired,omitempty"`
	LowStock     *bool   `json:"low_stock,omitempty"`
	Limit        *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset       *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Response Structs
// ExistenceResponse represents a single existence response
type ExistenceResponse struct {
	Success bool      `json:"success"`
	Data    Existence `json:"data"`
	Message string    `json:"message,omitempty"`
}

// ExistencesResponse represents multiple existences response
type ExistencesResponse struct {
	Success bool        `json:"success"`
	Data    []Existence `json:"data"`
	Total   int         `json:"total,omitempty"`
	Message string      `json:"message,omitempty"`
}

// GenericResponse represents a generic response (for delete operations)
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
