package handlers

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"inventory-service/entities/existences/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDBHandler(t *testing.T) (*DBHandler, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing

	handler := NewDBHandler(db, logger)

	cleanup := func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}

	return handler, mock, cleanup
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestDBHandler_CreateExistence_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	// Test data
	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	req := models.CreateExistenceRequest{
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.5,
		UnitsAvailable:         10.5,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerUnit:            12000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: float64Ptr(35.0),
		IvaPercentage:          float64Ptr(13.0),
		ServiceTaxPercentage:   float64Ptr(10.0),
		FinalPrice:             float64Ptr(15000.00),
	}

	expectedExistence := models.Existence{
		ID:                     "existence-id-123",
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.5,
		UnitsAvailable:         10.5,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10, // 12000/31
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      126000.00, // 10.5 * 12000
		RemainingValue:         126000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: 35.0,
		IncomeMarginAmount:     44100.00, // 126000 * 35/100
		IvaPercentage:          13.0,
		IvaAmount:              22113.00, // (126000 + 44100) * 13/100
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       17010.00,  // (126000 + 44100) * 10/100
		CalculatedPrice:        209223.00, // sum of all
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Expected SQL query
	expectedSQL := `INSERT INTO existences`
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(
			req.IngredientID,
			req.InvoiceDetailID,
			req.UnitsPurchased,
			req.UnitsAvailable,
			req.UnitType,
			req.ItemsPerUnit,
			req.CostPerUnit,
			req.ExpirationDate,
			req.IncomeMarginPercentage,
			req.IvaPercentage,
			req.ServiceTaxPercentage,
			req.FinalPrice,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "existence_reference_code", "ingredient_id", "invoice_detail_id",
			"units_purchased", "units_available", "unit_type", "items_per_unit",
			"cost_per_item", "cost_per_unit", "total_purchase_cost", "remaining_value",
			"expiration_date", "income_margin_percentage", "income_margin_amount",
			"iva_percentage", "iva_amount", "service_tax_percentage", "service_tax_amount",
			"calculated_price", "final_price", "created_at", "updated_at",
		}).AddRow(
			expectedExistence.ID, expectedExistence.ExistenceReferenceCode,
			expectedExistence.IngredientID, expectedExistence.InvoiceDetailID,
			expectedExistence.UnitsPurchased, expectedExistence.UnitsAvailable,
			expectedExistence.UnitType, expectedExistence.ItemsPerUnit,
			expectedExistence.CostPerItem, expectedExistence.CostPerUnit,
			expectedExistence.TotalPurchaseCost, expectedExistence.RemainingValue,
			expectedExistence.ExpirationDate, expectedExistence.IncomeMarginPercentage,
			expectedExistence.IncomeMarginAmount, expectedExistence.IvaPercentage,
			expectedExistence.IvaAmount, expectedExistence.ServiceTaxPercentage,
			expectedExistence.ServiceTaxAmount, expectedExistence.CalculatedPrice,
			expectedExistence.FinalPrice, expectedExistence.CreatedAt,
			expectedExistence.UpdatedAt,
		))

	// Execute
	result, err := handler.CreateExistence(req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedExistence.ID, result.ID)
	assert.Equal(t, expectedExistence.ExistenceReferenceCode, result.ExistenceReferenceCode)
	assert.Equal(t, expectedExistence.IngredientID, result.IngredientID)
	assert.Equal(t, expectedExistence.UnitsPurchased, result.UnitsPurchased)
	assert.Equal(t, expectedExistence.UnitType, result.UnitType)
}

func TestDBHandler_CreateExistence_DatabaseError(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	req := models.CreateExistenceRequest{
		IngredientID:    "ingredient-id-123",
		InvoiceDetailID: "invoice-detail-id-123",
		UnitsPurchased:  10.0,
		UnitsAvailable:  10.0,
		UnitType:        "Units",
		ItemsPerUnit:    1,
		CostPerUnit:     100.0,
	}

	expectedSQL := `INSERT INTO existences`
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(
			req.IngredientID,
			req.InvoiceDetailID,
			req.UnitsPurchased,
			req.UnitsAvailable,
			req.UnitType,
			req.ItemsPerUnit,
			req.CostPerUnit,
			req.ExpirationDate,
			req.IncomeMarginPercentage,
			req.IvaPercentage,
			req.ServiceTaxPercentage,
			req.FinalPrice,
		).
		WillReturnError(fmt.Errorf("database connection failed"))

	// Execute
	result, err := handler.CreateExistence(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestDBHandler_GetExistenceByID_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "existence-id-123"
	expirationDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         8.5,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         102000.00,
		ExpirationDate:         &expirationDate,
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		IvaPercentage:          13.0,
		IvaAmount:              20280.00,
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       15600.00,
		CalculatedPrice:        191880.00,
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	expectedSQL := `SELECT.*FROM existences WHERE id = ?`
	mock.ExpectQuery(expectedSQL).
		WithArgs(existenceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "existence_reference_code", "ingredient_id", "invoice_detail_id",
			"units_purchased", "units_available", "unit_type", "items_per_unit",
			"cost_per_item", "cost_per_unit", "total_purchase_cost", "remaining_value",
			"expiration_date", "income_margin_percentage", "income_margin_amount",
			"iva_percentage", "iva_amount", "service_tax_percentage", "service_tax_amount",
			"calculated_price", "final_price", "created_at", "updated_at",
		}).AddRow(
			expectedExistence.ID, expectedExistence.ExistenceReferenceCode,
			expectedExistence.IngredientID, expectedExistence.InvoiceDetailID,
			expectedExistence.UnitsPurchased, expectedExistence.UnitsAvailable,
			expectedExistence.UnitType, expectedExistence.ItemsPerUnit,
			expectedExistence.CostPerItem, expectedExistence.CostPerUnit,
			expectedExistence.TotalPurchaseCost, expectedExistence.RemainingValue,
			expectedExistence.ExpirationDate, expectedExistence.IncomeMarginPercentage,
			expectedExistence.IncomeMarginAmount, expectedExistence.IvaPercentage,
			expectedExistence.IvaAmount, expectedExistence.ServiceTaxPercentage,
			expectedExistence.ServiceTaxAmount, expectedExistence.CalculatedPrice,
			expectedExistence.FinalPrice, expectedExistence.CreatedAt,
			expectedExistence.UpdatedAt,
		))

	// Execute
	result, err := handler.GetExistenceByID(existenceID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedExistence.ID, result.ID)
	assert.Equal(t, expectedExistence.ExistenceReferenceCode, result.ExistenceReferenceCode)
	assert.Equal(t, expectedExistence.UnitsAvailable, result.UnitsAvailable)
	assert.Equal(t, expectedExistence.UnitType, result.UnitType)
}

func TestDBHandler_GetExistenceByID_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "nonexistent-id"

	expectedSQL := `SELECT.*FROM existences WHERE id = ?`
	mock.ExpectQuery(expectedSQL).
		WithArgs(existenceID).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := handler.GetExistenceByID(existenceID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, result)
}

func TestDBHandler_ListExistences_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	ingredientID := "ingredient-id-123"
	unitType := "Liters"
	expired := false
	lowStock := true

	req := models.ListExistencesRequest{
		IngredientID: &ingredientID,
		UnitType:     &unitType,
		Expired:      &expired,
		LowStock:     &lowStock,
		Limit:        nil,
		Offset:       nil,
	}

	expectedExistences := []models.Existence{
		{
			ID:                     "existence-1",
			ExistenceReferenceCode: 1001,
			IngredientID:           ingredientID,
			InvoiceDetailID:        "invoice-detail-1",
			UnitsPurchased:         10.0,
			UnitsAvailable:         1.0, // Low stock
			UnitType:               unitType,
			ItemsPerUnit:           31,
			CostPerItem:            387.10,
			CostPerUnit:            12000.00,
			TotalPurchaseCost:      120000.00,
			RemainingValue:         12000.00,
			ExpirationDate:         timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
			IncomeMarginPercentage: 30.0,
			IncomeMarginAmount:     36000.00,
			IvaPercentage:          13.0,
			IvaAmount:              20280.00,
			ServiceTaxPercentage:   10.0,
			ServiceTaxAmount:       15600.00,
			CalculatedPrice:        191880.00,
			FinalPrice:             float64Ptr(15000.00),
			CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	expectedSQL := `SELECT.*FROM existences WHERE 1=1`
	mock.ExpectQuery(expectedSQL).
		WithArgs(&ingredientID, &unitType, &expired, &lowStock, nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "existence_reference_code", "ingredient_id", "invoice_detail_id",
			"units_purchased", "units_available", "unit_type", "items_per_unit",
			"cost_per_item", "cost_per_unit", "total_purchase_cost", "remaining_value",
			"expiration_date", "income_margin_percentage", "income_margin_amount",
			"iva_percentage", "iva_amount", "service_tax_percentage", "service_tax_amount",
			"calculated_price", "final_price", "created_at", "updated_at",
		}).AddRow(
			expectedExistences[0].ID, expectedExistences[0].ExistenceReferenceCode,
			expectedExistences[0].IngredientID, expectedExistences[0].InvoiceDetailID,
			expectedExistences[0].UnitsPurchased, expectedExistences[0].UnitsAvailable,
			expectedExistences[0].UnitType, expectedExistences[0].ItemsPerUnit,
			expectedExistences[0].CostPerItem, expectedExistences[0].CostPerUnit,
			expectedExistences[0].TotalPurchaseCost, expectedExistences[0].RemainingValue,
			expectedExistences[0].ExpirationDate, expectedExistences[0].IncomeMarginPercentage,
			expectedExistences[0].IncomeMarginAmount, expectedExistences[0].IvaPercentage,
			expectedExistences[0].IvaAmount, expectedExistences[0].ServiceTaxPercentage,
			expectedExistences[0].ServiceTaxAmount, expectedExistences[0].CalculatedPrice,
			expectedExistences[0].FinalPrice, expectedExistences[0].CreatedAt,
			expectedExistences[0].UpdatedAt,
		))

	// Execute
	result, err := handler.ListExistences(req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedExistences[0].ID, result[0].ID)
	assert.Equal(t, expectedExistences[0].UnitsAvailable, result[0].UnitsAvailable)
}

func TestDBHandler_ListExistences_EmptyResult(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	req := models.ListExistencesRequest{}

	expectedSQL := `SELECT.*FROM existences WHERE 1=1`
	mock.ExpectQuery(expectedSQL).
		WithArgs(req.IngredientID, req.UnitType, req.Expired, req.LowStock, req.Limit, req.Offset).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "existence_reference_code", "ingredient_id", "invoice_detail_id",
			"units_purchased", "units_available", "unit_type", "items_per_unit",
			"cost_per_item", "cost_per_unit", "total_purchase_cost", "remaining_value",
			"expiration_date", "income_margin_percentage", "income_margin_amount",
			"iva_percentage", "iva_amount", "service_tax_percentage", "service_tax_amount",
			"calculated_price", "final_price", "created_at", "updated_at",
		}))

	// Execute
	result, err := handler.ListExistences(req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	assert.Equal(t, []models.Existence{}, result)
}

func TestDBHandler_UpdateExistence_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "existence-id-123"
	newUnitsAvailable := 5.0
	req := models.UpdateExistenceRequest{
		UnitsAvailable: &newUnitsAvailable,
	}

	expectedExistence := models.Existence{
		ID:                     existenceID,
		ExistenceReferenceCode: 1001,
		IngredientID:           "ingredient-id-123",
		InvoiceDetailID:        "invoice-detail-id-123",
		UnitsPurchased:         10.0,
		UnitsAvailable:         newUnitsAvailable,
		UnitType:               "Liters",
		ItemsPerUnit:           31,
		CostPerItem:            387.10,
		CostPerUnit:            12000.00,
		TotalPurchaseCost:      120000.00,
		RemainingValue:         60000.00, // Updated based on new units available
		ExpirationDate:         timePtr(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
		IncomeMarginPercentage: 30.0,
		IncomeMarginAmount:     36000.00,
		IvaPercentage:          13.0,
		IvaAmount:              20280.00,
		ServiceTaxPercentage:   10.0,
		ServiceTaxAmount:       15600.00,
		CalculatedPrice:        191880.00,
		FinalPrice:             float64Ptr(15000.00),
		CreatedAt:              time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
	}

	expectedSQL := `UPDATE existences SET`
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(
			existenceID,
			req.UnitsAvailable,
			req.UnitType,
			req.ItemsPerUnit,
			req.CostPerUnit,
			req.ExpirationDate,
			req.IncomeMarginPercentage,
			req.IvaPercentage,
			req.ServiceTaxPercentage,
			req.FinalPrice,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "existence_reference_code", "ingredient_id", "invoice_detail_id",
			"units_purchased", "units_available", "unit_type", "items_per_unit",
			"cost_per_item", "cost_per_unit", "total_purchase_cost", "remaining_value",
			"expiration_date", "income_margin_percentage", "income_margin_amount",
			"iva_percentage", "iva_amount", "service_tax_percentage", "service_tax_amount",
			"calculated_price", "final_price", "created_at", "updated_at",
		}).AddRow(
			expectedExistence.ID, expectedExistence.ExistenceReferenceCode,
			expectedExistence.IngredientID, expectedExistence.InvoiceDetailID,
			expectedExistence.UnitsPurchased, expectedExistence.UnitsAvailable,
			expectedExistence.UnitType, expectedExistence.ItemsPerUnit,
			expectedExistence.CostPerItem, expectedExistence.CostPerUnit,
			expectedExistence.TotalPurchaseCost, expectedExistence.RemainingValue,
			expectedExistence.ExpirationDate, expectedExistence.IncomeMarginPercentage,
			expectedExistence.IncomeMarginAmount, expectedExistence.IvaPercentage,
			expectedExistence.IvaAmount, expectedExistence.ServiceTaxPercentage,
			expectedExistence.ServiceTaxAmount, expectedExistence.CalculatedPrice,
			expectedExistence.FinalPrice, expectedExistence.CreatedAt,
			expectedExistence.UpdatedAt,
		))

	// Execute
	result, err := handler.UpdateExistence(existenceID, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedExistence.ID, result.ID)
	assert.Equal(t, expectedExistence.UnitsAvailable, result.UnitsAvailable)
	assert.Equal(t, expectedExistence.UpdatedAt, result.UpdatedAt)
}

func TestDBHandler_UpdateExistence_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "nonexistent-id"
	req := models.UpdateExistenceRequest{
		UnitsAvailable: float64Ptr(5.0),
	}

	expectedSQL := `UPDATE existences SET`
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(
			existenceID,
			req.UnitsAvailable,
			req.UnitType,
			req.ItemsPerUnit,
			req.CostPerUnit,
			req.ExpirationDate,
			req.IncomeMarginPercentage,
			req.IvaPercentage,
			req.ServiceTaxPercentage,
			req.FinalPrice,
		).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := handler.UpdateExistence(existenceID, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, result)
}

func TestDBHandler_DeleteExistence_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "existence-id-123"

	expectedSQL := `DELETE FROM existences WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
		WithArgs(existenceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := handler.DeleteExistence(existenceID)

	// Assert
	assert.NoError(t, err)
}

func TestDBHandler_DeleteExistence_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "nonexistent-id"

	expectedSQL := `DELETE FROM existences WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
		WithArgs(existenceID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	// Execute
	err := handler.DeleteExistence(existenceID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDBHandler_DeleteExistence_DatabaseError(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	existenceID := "existence-id-123"

	expectedSQL := `DELETE FROM existences WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
		WithArgs(existenceID).
		WillReturnError(fmt.Errorf("database connection failed"))

	// Execute
	err := handler.DeleteExistence(existenceID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
}
