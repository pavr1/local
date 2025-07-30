package models

import (
	"database/sql/driver"
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

// Ingredient represents raw materials/ingredients
type Ingredient struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	SupplierID *string   `json:"supplier_id" db:"supplier_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// UnitType represents the different unit types for measurements
type UnitType string

const (
	UnitTypeLiters  UnitType = "Liters"
	UnitTypeGallons UnitType = "Gallons"
	UnitTypeUnits   UnitType = "Units"
	UnitTypeBag     UnitType = "Bag"
)

// Existence represents individual ingredient purchases/acquisitions
type Existence struct {
	ID                     string     `json:"id" db:"id"`
	ExistenceReferenceCode int        `json:"existence_reference_code" db:"existence_reference_code"`
	IngredientID           string     `json:"ingredient_id" db:"ingredient_id"`
	ExpenseReceiptID       string     `json:"expense_receipt_id" db:"expense_receipt_id"`
	UnitsPurchased         float64    `json:"units_purchased" db:"units_purchased"`
	UnitsAvailable         float64    `json:"units_available" db:"units_available"`
	UnitType               UnitType   `json:"unit_type" db:"unit_type"`
	ItemsPerUnit           int        `json:"items_per_unit" db:"items_per_unit"`
	CostPerItem            float64    `json:"cost_per_item" db:"cost_per_item"`
	CostPerUnit            float64    `json:"cost_per_unit" db:"cost_per_unit"`
	TotalPurchaseCost      float64    `json:"total_purchase_cost" db:"total_purchase_cost"`
	RemainingValue         float64    `json:"remaining_value" db:"remaining_value"`
	ExpirationDate         *time.Time `json:"expiration_date" db:"expiration_date"`
	IncomeMarginPercentage float64    `json:"income_margin_percentage" db:"income_margin_percentage"`
	IncomeMarginAmount     float64    `json:"income_margin_amount" db:"income_margin_amount"`
	IVAPercentage          float64    `json:"iva_percentage" db:"iva_percentage"`
	IVAAmount              float64    `json:"iva_amount" db:"iva_amount"`
	ServiceTaxPercentage   float64    `json:"service_tax_percentage" db:"service_tax_percentage"`
	ServiceTaxAmount       float64    `json:"service_tax_amount" db:"service_tax_amount"`
	CalculatedPrice        float64    `json:"calculated_price" db:"calculated_price"`
	FinalPrice             *float64   `json:"final_price" db:"final_price"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// RunoutIngredientReport represents ingredient usage reported by employees
type RunoutIngredientReport struct {
	ID          string    `json:"id" db:"id"`
	ExistenceID string    `json:"existence_id" db:"existence_id"`
	EmployeeID  string    `json:"employee_id" db:"employee_id"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitType    UnitType  `json:"unit_type" db:"unit_type"`
	ReportDate  time.Time `json:"report_date" db:"report_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RecipeCategory represents recipe categorization
type RecipeCategory struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Recipe represents product recipes with pricing information
type Recipe struct {
	ID                string    `json:"id" db:"id"`
	RecipeName        string    `json:"recipe_name" db:"recipe_name"`
	RecipeDescription *string   `json:"recipe_description" db:"recipe_description"`
	PictureURL        *string   `json:"picture_url" db:"picture_url"`
	RecipeCategoryID  string    `json:"recipe_category_id" db:"recipe_category_id"`
	TotalRecipeCost   float64   `json:"total_recipe_cost" db:"total_recipe_cost"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RecipeIngredient represents junction table linking recipes to ingredients
type RecipeIngredient struct {
	ID            string    `json:"id" db:"id"`
	RecipeID      string    `json:"recipe_id" db:"recipe_id"`
	IngredientID  string    `json:"ingredient_id" db:"ingredient_id"`
	NumberOfUnits float64   `json:"number_of_units" db:"number_of_units"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Request and Response DTOs
type CreateSupplierRequest struct {
	SupplierName  string  `json:"supplier_name" validate:"required,min=1,max=255"`
	ContactNumber *string `json:"contact_number,omitempty" validate:"omitempty,max=50"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address       *string `json:"address,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type UpdateSupplierRequest struct {
	SupplierName  *string `json:"supplier_name,omitempty" validate:"omitempty,min=1,max=255"`
	ContactNumber *string `json:"contact_number,omitempty" validate:"omitempty,max=50"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address       *string `json:"address,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type CreateIngredientRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=255"`
	SupplierID *string `json:"supplier_id,omitempty"`
}

type UpdateIngredientRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	SupplierID *string `json:"supplier_id,omitempty"`
}

type CreateExistenceRequest struct {
	IngredientID           string     `json:"ingredient_id" validate:"required"`
	ExpenseReceiptID       string     `json:"expense_receipt_id" validate:"required"`
	UnitsPurchased         float64    `json:"units_purchased" validate:"required,gt=0"`
	UnitType               UnitType   `json:"unit_type" validate:"required"`
	ItemsPerUnit           int        `json:"items_per_unit" validate:"required,gt=0"`
	CostPerUnit            float64    `json:"cost_per_unit" validate:"required,gt=0"`
	ExpirationDate         *time.Time `json:"expiration_date,omitempty"`
	IncomeMarginPercentage *float64   `json:"income_margin_percentage,omitempty"`
	IVAPercentage          *float64   `json:"iva_percentage,omitempty"`
	ServiceTaxPercentage   *float64   `json:"service_tax_percentage,omitempty"`
	FinalPrice             *float64   `json:"final_price,omitempty"`
}

type CreateRunoutReportRequest struct {
	ExistenceID string   `json:"existence_id" validate:"required"`
	EmployeeID  string   `json:"employee_id" validate:"required"`
	Quantity    float64  `json:"quantity" validate:"required,gt=0"`
	UnitType    UnitType `json:"unit_type" validate:"required"`
}

type CreateRecipeCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

type CreateRecipeRequest struct {
	RecipeName        string  `json:"recipe_name" validate:"required,min=1,max=255"`
	RecipeDescription *string `json:"recipe_description,omitempty"`
	PictureURL        *string `json:"picture_url,omitempty" validate:"omitempty,url,max=500"`
	RecipeCategoryID  string  `json:"recipe_category_id" validate:"required"`
}

type CreateRecipeIngredientRequest struct {
	RecipeID      string  `json:"recipe_id" validate:"required"`
	IngredientID  string  `json:"ingredient_id" validate:"required"`
	NumberOfUnits float64 `json:"number_of_units" validate:"required,gt=0"`
}

// Generic response structures
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Code      int    `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Implement database/sql/driver.Valuer interface for UnitType
func (ut UnitType) Value() (driver.Value, error) {
	return string(ut), nil
}

// Implement database/sql.Scanner interface for UnitType
func (ut *UnitType) Scan(value interface{}) error {
	if value == nil {
		*ut = ""
		return nil
	}
	switch s := value.(type) {
	case string:
		*ut = UnitType(s)
	case []byte:
		*ut = UnitType(s)
	}
	return nil
}
