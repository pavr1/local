package models

import (
	"time"
)

// Invoice represents an invoice in the database
type Invoice struct {
	ID                string    `json:"id" db:"id"`
	InvoiceNumber     string    `json:"invoice_number" db:"invoice_number"`
	TransactionDate   time.Time `json:"transaction_date" db:"transaction_date"`
	TransactionType   string    `json:"transaction_type" db:"transaction_type"`
	SupplierID        *string   `json:"supplier_id" db:"supplier_id"`
	ExpenseCategoryID string    `json:"expense_category_id" db:"expense_category_id"`
	TotalAmount       *float64  `json:"total_amount" db:"total_amount"`
	ImageURL          string    `json:"image_url" db:"image_url"`
	Notes             *string   `json:"notes" db:"notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// InvoiceDetail represents a line item within an invoice
type InvoiceDetail struct {
	ID             string     `json:"id" db:"id"`
	InvoiceID      string     `json:"invoice_id" db:"invoice_id"`
	IngredientID   *string    `json:"ingredient_id" db:"ingredient_id"`
	Detail         string     `json:"detail" db:"detail"`
	Count          float64    `json:"count" db:"count"`
	UnitType       string     `json:"unit_type" db:"unit_type"`
	Price          float64    `json:"price" db:"price"`
	Total          float64    `json:"total" db:"total"`
	ExpirationDate *time.Time `json:"expiration_date" db:"expiration_date"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateInvoiceDetailRequest represents the request to create a new invoice detail
type CreateInvoiceDetailRequest struct {
	InvoiceID      string     `json:"invoice_id" validate:"required,uuid"`
	IngredientID   *string    `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Detail         string     `json:"detail" validate:"required"`
	Count          float64    `json:"count" validate:"required,gt=0"`
	UnitType       string     `json:"unit_type" validate:"required,oneof=Liters Gallons Units Bag"`
	Price          float64    `json:"price" validate:"required,gt=0"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// CreateInvoiceRequest represents the request to create a new invoice with details
type CreateInvoiceRequest struct {
	InvoiceNumber     string                       `json:"invoice_number" validate:"required"`
	TransactionDate   *time.Time                   `json:"transaction_date,omitempty"`
	TransactionType   string                       `json:"transaction_type" validate:"required,oneof=income outcome"`
	SupplierID        *string                      `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	ExpenseCategoryID string                       `json:"expense_category_id" validate:"required,uuid"`
	ImageURL          string                       `json:"image_url" validate:"required,url"`
	Notes             *string                      `json:"notes,omitempty"`
	Items             []CreateInvoiceDetailRequest `json:"items" validate:"required,dive"`
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	InvoiceNumber     *string    `json:"invoice_number,omitempty" validate:"omitempty,min=1"`
	TransactionDate   *time.Time `json:"transaction_date,omitempty"`
	TransactionType   *string    `json:"transaction_type,omitempty" validate:"omitempty,oneof=income outcome"`
	SupplierID        *string    `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	ExpenseCategoryID *string    `json:"expense_category_id,omitempty" validate:"omitempty,uuid"`
	ImageURL          *string    `json:"image_url,omitempty" validate:"omitempty,url"`
	Notes             *string    `json:"notes,omitempty"`
}

// UpdateInvoiceDetailRequest represents the request to update an invoice detail
type UpdateInvoiceDetailRequest struct {
	IngredientID   *string    `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Detail         *string    `json:"detail,omitempty" validate:"omitempty,min=1"`
	Count          *float64   `json:"count,omitempty" validate:"omitempty,gt=0"`
	UnitType       *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	Price          *float64   `json:"price,omitempty" validate:"omitempty,gt=0"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// GetInvoiceRequest represents the request to get an invoice by ID
type GetInvoiceRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// GetInvoiceDetailRequest represents the request to get an invoice detail by ID
type GetInvoiceDetailRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteInvoiceRequest represents the request to delete an invoice
type DeleteInvoiceRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteInvoiceDetailRequest represents the request to delete an invoice detail
type DeleteInvoiceDetailRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListInvoicesRequest represents the request to list invoices
type ListInvoicesRequest struct {
	Limit             *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset            *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
	TransactionType   *string `json:"transaction_type,omitempty" validate:"omitempty,oneof=income outcome"`
	ExpenseCategoryID *string `json:"expense_category_id,omitempty" validate:"omitempty,uuid"`
	SupplierID        *string `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
}

// ListInvoiceDetailsRequest represents the request to list invoice details
type ListInvoiceDetailsRequest struct {
	Limit        *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset       *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
	InvoiceID    *string `json:"invoice_id,omitempty" validate:"omitempty,uuid"`
	IngredientID *string `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
}

// Response Structs
// InvoiceResponse represents a single invoice response
type InvoiceResponse struct {
	Success bool    `json:"success"`
	Data    Invoice `json:"data"`
	Message string  `json:"message,omitempty"`
}

// InvoicesListResponse represents a list of invoices response
type InvoicesListResponse struct {
	Success bool      `json:"success"`
	Data    []Invoice `json:"data"`
	Count   int       `json:"count"`
	Message string    `json:"message,omitempty"`
}

// InvoiceDeleteResponse represents a delete operation response
type InvoiceDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// InvoiceDetailResponse represents a single invoice detail response
type InvoiceDetailResponse struct {
	Success bool          `json:"success"`
	Data    InvoiceDetail `json:"data"`
	Message string        `json:"message,omitempty"`
}

// InvoiceDetailsListResponse represents a list of invoice details response
type InvoiceDetailsListResponse struct {
	Success bool            `json:"success"`
	Data    []InvoiceDetail `json:"data"`
	Count   int             `json:"count"`
	Message string          `json:"message,omitempty"`
}

// InvoiceDetailDeleteResponse represents a delete operation response
type InvoiceDetailDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
