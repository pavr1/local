package models

import (
	"time"
)

// Receipt represents a receipt in the database
type Receipt struct {
	ID                string    `json:"id" db:"id"`
	ReceiptNumber     string    `json:"receipt_number" db:"receipt_number"`
	PurchaseDate      time.Time `json:"purchase_date" db:"purchase_date"`
	SupplierID        *string   `json:"supplier_id" db:"supplier_id"`
	ExpenseCategoryID string    `json:"expense_category_id" db:"expense_category_id"`
	TotalAmount       *float64  `json:"total_amount" db:"total_amount"`
	ImageURL          string    `json:"image_url" db:"image_url"`
	Notes             *string   `json:"notes" db:"notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ReceiptItem represents a line item within a receipt
type ReceiptItem struct {
	ID             string     `json:"id" db:"id"`
	ReceiptID      string     `json:"receipt_id" db:"receipt_id"`
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

// CreateReceiptItemRequest represents the request to create a new receipt item
type CreateReceiptItemRequest struct {
	ReceiptID      string     `json:"receipt_id" validate:"required,uuid"`
	IngredientID   *string    `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Detail         string     `json:"detail" validate:"required"`
	Count          float64    `json:"count" validate:"required,gt=0"`
	UnitType       string     `json:"unit_type" validate:"required,oneof=Liters Gallons Units Bag"`
	Price          float64    `json:"price" validate:"required,gt=0"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// CreateReceiptRequest represents the request to create a new receipt with items
type CreateReceiptRequest struct {
	ReceiptNumber     string                     `json:"receipt_number" validate:"required"`
	PurchaseDate      time.Time                  `json:"purchase_date" validate:"required"`
	SupplierID        *string                    `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	ExpenseCategoryID string                     `json:"expense_category_id" validate:"required,uuid"`
	ImageURL          string                     `json:"image_url" validate:"required,url"`
	Notes             *string                    `json:"notes,omitempty"`
	Items             []CreateReceiptItemRequest `json:"items" validate:"required,dive"`
}

// UpdateReceiptRequest represents the request to update a receipt
type UpdateReceiptRequest struct {
	ReceiptNumber     *string    `json:"receipt_number,omitempty" validate:"omitempty,min=1"`
	PurchaseDate      *time.Time `json:"purchase_date,omitempty"`
	SupplierID        *string    `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
	ExpenseCategoryID *string    `json:"expense_category_id,omitempty" validate:"omitempty,uuid"`
	ImageURL          *string    `json:"image_url,omitempty" validate:"omitempty,url"`
	Notes             *string    `json:"notes,omitempty"`
}

// UpdateReceiptItemRequest represents the request to update a receipt item
type UpdateReceiptItemRequest struct {
	IngredientID   *string    `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
	Detail         *string    `json:"detail,omitempty" validate:"omitempty,min=1"`
	Count          *float64   `json:"count,omitempty" validate:"omitempty,gt=0"`
	UnitType       *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=Liters Gallons Units Bag"`
	Price          *float64   `json:"price,omitempty" validate:"omitempty,gt=0"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// GetReceiptRequest represents the request to get a receipt by ID
type GetReceiptRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// GetReceiptItemRequest represents the request to get a receipt item by ID
type GetReceiptItemRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteReceiptRequest represents the request to delete a receipt
type DeleteReceiptRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// DeleteReceiptItemRequest represents the request to delete a receipt item
type DeleteReceiptItemRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

// ListReceiptsRequest represents the request to list receipts
type ListReceiptsRequest struct {
	Limit             *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset            *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
	ExpenseCategoryID *string `json:"expense_category_id,omitempty" validate:"omitempty,uuid"`
	SupplierID        *string `json:"supplier_id,omitempty" validate:"omitempty,uuid"`
}

// ListReceiptItemsRequest represents the request to list receipt items
type ListReceiptItemsRequest struct {
	Limit        *int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset       *int    `json:"offset,omitempty" validate:"omitempty,min=0"`
	ReceiptID    *string `json:"receipt_id,omitempty" validate:"omitempty,uuid"`
	IngredientID *string `json:"ingredient_id,omitempty" validate:"omitempty,uuid"`
}

// Response Structs
// ReceiptResponse represents a single receipt response
type ReceiptResponse struct {
	Success bool    `json:"success"`
	Data    Receipt `json:"data"`
	Message string  `json:"message,omitempty"`
}

// ReceiptsListResponse represents a list of receipts response
type ReceiptsListResponse struct {
	Success bool      `json:"success"`
	Data    []Receipt `json:"data"`
	Count   int       `json:"count"`
	Message string    `json:"message,omitempty"`
}

// ReceiptDeleteResponse represents a delete operation response
type ReceiptDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReceiptItemResponse represents a single receipt item response
type ReceiptItemResponse struct {
	Success bool        `json:"success"`
	Data    ReceiptItem `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ReceiptItemsListResponse represents a list of receipt items response
type ReceiptItemsListResponse struct {
	Success bool          `json:"success"`
	Data    []ReceiptItem `json:"data"`
	Count   int           `json:"count"`
	Message string        `json:"message,omitempty"`
}

// ReceiptItemDeleteResponse represents a delete operation response
type ReceiptItemDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
