package models

import (
	"time"
)

// Supplier represents a supplier/vendor for ingredient procurement
type Supplier struct {
	ID            string    `json:"id" db:"id"`
	SupplierName  string    `json:"supplier_name" db:"supplier_name"`
	ContactNumber *string   `json:"contact_number" db:"contact_number"`
	Email         *string   `json:"email" db:"email"`
	Address       *string   `json:"address" db:"address"`
	Notes         *string   `json:"notes" db:"notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreateSupplierRequest represents the request to create a new supplier
type CreateSupplierRequest struct {
	SupplierName  string  `json:"supplier_name" validate:"required,min=1,max=255"`
	ContactNumber *string `json:"contact_number,omitempty" validate:"omitempty,max=50"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address       *string `json:"address,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

// UpdateSupplierRequest represents the request to update a supplier
type UpdateSupplierRequest struct {
	SupplierName  *string `json:"supplier_name,omitempty" validate:"omitempty,min=1,max=255"`
	ContactNumber *string `json:"contact_number,omitempty" validate:"omitempty,max=50"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address       *string `json:"address,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}
