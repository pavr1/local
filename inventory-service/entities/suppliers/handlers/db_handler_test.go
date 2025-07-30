package handlers

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"inventory-service/entities/suppliers/models"

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

func TestDBHandler_CreateSupplier_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	// Test data
	req := models.CreateSupplierRequest{
		SupplierName:  "Test Supplier",
		ContactNumber: stringPtr("+1234567890"),
		Email:         stringPtr("test@example.com"),
		Address:       stringPtr("123 Test St"),
		Notes:         stringPtr("Test notes"),
	}

	expectedSupplier := models.Supplier{
		ID:            "123e4567-e89b-12d3-a456-426614174000",
		SupplierName:  "Test Supplier",
		ContactNumber: stringPtr("+1234567890"),
		Email:         stringPtr("test@example.com"),
		Address:       stringPtr("123 Test St"),
		Notes:         stringPtr("Test notes"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Mock the database query
	rows := sqlmock.NewRows([]string{"id", "supplier_name", "contact_number", "email", "address", "notes", "created_at", "updated_at"}).
		AddRow(expectedSupplier.ID, expectedSupplier.SupplierName, expectedSupplier.ContactNumber,
			expectedSupplier.Email, expectedSupplier.Address, expectedSupplier.Notes,
			expectedSupplier.CreatedAt, expectedSupplier.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO suppliers")).
		WithArgs(req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		WillReturnRows(rows)

	// Execute
	result, err := handler.CreateSupplier(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSupplier.ID, result.ID)
	assert.Equal(t, expectedSupplier.SupplierName, result.SupplierName)
	assert.Equal(t, expectedSupplier.ContactNumber, result.ContactNumber)
	assert.Equal(t, expectedSupplier.Email, result.Email)
}

func TestDBHandler_CreateSupplier_DatabaseError(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	req := models.CreateSupplierRequest{
		SupplierName: "Test Supplier",
	}

	// Mock database error
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO suppliers")).
		WithArgs(req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		WillReturnError(fmt.Errorf("database connection failed"))

	// Execute
	result, err := handler.CreateSupplier(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestDBHandler_GetSupplierByID_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"
	expectedSupplier := models.Supplier{
		ID:           supplierID,
		SupplierName: "Test Supplier",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock the database query
	rows := sqlmock.NewRows([]string{"id", "supplier_name", "contact_number", "email", "address", "notes", "created_at", "updated_at"}).
		AddRow(expectedSupplier.ID, expectedSupplier.SupplierName, nil, nil, nil, nil,
			expectedSupplier.CreatedAt, expectedSupplier.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, supplier_name, contact_number, email, address, notes, created_at, updated_at FROM suppliers WHERE id = $1")).
		WithArgs(supplierID).
		WillReturnRows(rows)

	// Execute
	result, err := handler.GetSupplierByID(supplierID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSupplier.ID, result.ID)
	assert.Equal(t, expectedSupplier.SupplierName, result.SupplierName)
}

func TestDBHandler_GetSupplierByID_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "nonexistent-id"

	// Mock no rows returned
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, supplier_name, contact_number, email, address, notes, created_at, updated_at FROM suppliers WHERE id = $1")).
		WithArgs(supplierID).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := handler.GetSupplierByID(supplierID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDBHandler_ListSuppliers_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	expectedSuppliers := []models.Supplier{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174001",
			SupplierName: "Supplier 1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "123e4567-e89b-12d3-a456-426614174002",
			SupplierName: "Supplier 2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Mock the database query
	rows := sqlmock.NewRows([]string{"id", "supplier_name", "contact_number", "email", "address", "notes", "created_at", "updated_at"}).
		AddRow(expectedSuppliers[0].ID, expectedSuppliers[0].SupplierName, nil, nil, nil, nil,
			expectedSuppliers[0].CreatedAt, expectedSuppliers[0].UpdatedAt).
		AddRow(expectedSuppliers[1].ID, expectedSuppliers[1].SupplierName, nil, nil, nil, nil,
			expectedSuppliers[1].CreatedAt, expectedSuppliers[1].UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, supplier_name, contact_number, email, address, notes, created_at, updated_at FROM suppliers ORDER BY supplier_name ASC")).
		WillReturnRows(rows)

	// Execute
	result, err := handler.ListSuppliers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedSuppliers[0].ID, result[0].ID)
	assert.Equal(t, expectedSuppliers[1].ID, result[1].ID)
}

func TestDBHandler_ListSuppliers_Empty(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	// Mock empty result set
	rows := sqlmock.NewRows([]string{"id", "supplier_name", "contact_number", "email", "address", "notes", "created_at", "updated_at"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, supplier_name, contact_number, email, address, notes, created_at, updated_at FROM suppliers ORDER BY supplier_name ASC")).
		WillReturnRows(rows)

	// Execute
	result, err := handler.ListSuppliers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestDBHandler_UpdateSupplier_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"
	req := models.UpdateSupplierRequest{
		SupplierName: stringPtr("Updated Supplier"),
		Email:        stringPtr("updated@example.com"),
	}

	expectedSupplier := models.Supplier{
		ID:           supplierID,
		SupplierName: "Updated Supplier",
		Email:        stringPtr("updated@example.com"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock the database query
	rows := sqlmock.NewRows([]string{"id", "supplier_name", "contact_number", "email", "address", "notes", "created_at", "updated_at"}).
		AddRow(expectedSupplier.ID, expectedSupplier.SupplierName, nil, expectedSupplier.Email, nil, nil,
			expectedSupplier.CreatedAt, expectedSupplier.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE suppliers SET")).
		WithArgs(supplierID, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		WillReturnRows(rows)

	// Execute
	result, err := handler.UpdateSupplier(supplierID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSupplier.ID, result.ID)
	assert.Equal(t, expectedSupplier.SupplierName, result.SupplierName)
}

func TestDBHandler_UpdateSupplier_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "nonexistent-id"
	req := models.UpdateSupplierRequest{
		SupplierName: stringPtr("Updated Supplier"),
	}

	// Mock no rows affected
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE suppliers SET")).
		WithArgs(supplierID, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := handler.UpdateSupplier(supplierID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDBHandler_DeleteSupplier_Success(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"

	// Mock successful delete
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM suppliers WHERE id = $1")).
		WithArgs(supplierID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	// Execute
	err := handler.DeleteSupplier(supplierID)

	// Assert
	assert.NoError(t, err)
}

func TestDBHandler_DeleteSupplier_NotFound(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "nonexistent-id"

	// Mock no rows affected
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM suppliers WHERE id = $1")).
		WithArgs(supplierID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	// Execute
	err := handler.DeleteSupplier(supplierID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDBHandler_DeleteSupplier_DatabaseError(t *testing.T) {
	handler, mock, cleanup := setupTestDBHandler(t)
	defer cleanup()

	supplierID := "123e4567-e89b-12d3-a456-426614174000"

	// Mock database error
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM suppliers WHERE id = $1")).
		WithArgs(supplierID).
		WillReturnError(fmt.Errorf("database connection failed"))

	// Execute
	err := handler.DeleteSupplier(supplierID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
