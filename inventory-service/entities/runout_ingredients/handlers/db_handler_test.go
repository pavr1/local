package handlers

import (
	"database/sql"
	"testing"
	"time"

	"inventory-service/entities/runout_ingredients/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunoutIngredientDBHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

func TestRunoutIngredientDBHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	now := time.Now()
	req := models.CreateRunoutIngredientRequest{
		ExistenceID: "550e8400-e29b-41d4-a716-446655440000",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440001",
		Quantity:    10.5,
		UnitType:    "Liters",
		ReportDate:  &now,
	}

	expectedRunoutIngredient := models.RunoutIngredient{
		ID:          "550e8400-e29b-41d4-a716-446655440002",
		ExistenceID: req.ExistenceID,
		EmployeeID:  req.EmployeeID,
		Quantity:    req.Quantity,
		UnitType:    req.UnitType,
		ReportDate:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	}).AddRow(
		expectedRunoutIngredient.ID,
		expectedRunoutIngredient.ExistenceID,
		expectedRunoutIngredient.EmployeeID,
		expectedRunoutIngredient.Quantity,
		expectedRunoutIngredient.UnitType,
		expectedRunoutIngredient.ReportDate,
		expectedRunoutIngredient.CreatedAt,
		expectedRunoutIngredient.UpdatedAt,
	)

	mock.ExpectQuery("INSERT INTO runout_ingredient_report").
		WithArgs(req.ExistenceID, req.EmployeeID, req.Quantity, req.UnitType, now).
		WillReturnRows(rows)

	result, err := handler.Create(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRunoutIngredient, *result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	req := models.CreateRunoutIngredientRequest{
		ExistenceID: "550e8400-e29b-41d4-a716-446655440000",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440001",
		Quantity:    10.5,
		UnitType:    "Liters",
	}

	mock.ExpectQuery("INSERT INTO runout_ingredient_report").
		WithArgs(req.ExistenceID, req.EmployeeID, req.Quantity, req.UnitType, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	result, err := handler.Create(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create runout ingredient")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	now := time.Now()
	expectedRunoutIngredient := models.RunoutIngredient{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		ExistenceID: "550e8400-e29b-41d4-a716-446655440001",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440002",
		Quantity:    10.5,
		UnitType:    "Liters",
		ReportDate:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	}).AddRow(
		expectedRunoutIngredient.ID,
		expectedRunoutIngredient.ExistenceID,
		expectedRunoutIngredient.EmployeeID,
		expectedRunoutIngredient.Quantity,
		expectedRunoutIngredient.UnitType,
		expectedRunoutIngredient.ReportDate,
		expectedRunoutIngredient.CreatedAt,
		expectedRunoutIngredient.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs(expectedRunoutIngredient.ID).
		WillReturnRows(rows)

	req := models.GetRunoutIngredientRequest{ID: expectedRunoutIngredient.ID}
	result, err := handler.GetByID(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRunoutIngredient, *result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	req := models.GetRunoutIngredientRequest{ID: "non-existent-id"}
	result, err := handler.GetByID(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "runout ingredient not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	now := time.Now()
	expectedRunoutIngredients := []models.RunoutIngredient{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			ExistenceID: "550e8400-e29b-41d4-a716-446655440001",
			EmployeeID:  "550e8400-e29b-41d4-a716-446655440002",
			Quantity:    10.5,
			UnitType:    "Liters",
			ReportDate:  now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440003",
			ExistenceID: "550e8400-e29b-41d4-a716-446655440004",
			EmployeeID:  "550e8400-e29b-41d4-a716-446655440005",
			Quantity:    20.0,
			UnitType:    "Gallons",
			ReportDate:  now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	})
	for _, ri := range expectedRunoutIngredients {
		rows.AddRow(
			ri.ID, ri.ExistenceID, ri.EmployeeID, ri.Quantity,
			ri.UnitType, ri.ReportDate, ri.CreatedAt, ri.UpdatedAt,
		)
	}

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs(nil, nil, nil, nil, 50, 0).
		WillReturnRows(rows)

	req := models.ListRunoutIngredientsRequest{}
	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Equal(t, expectedRunoutIngredients, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	existenceID := "550e8400-e29b-41d4-a716-446655440000"
	employeeID := "550e8400-e29b-41d4-a716-446655440001"
	unitType := "Liters"
	now := time.Now()
	limit := 10
	offset := 5

	req := models.ListRunoutIngredientsRequest{
		ExistenceID: &existenceID,
		EmployeeID:  &employeeID,
		UnitType:    &unitType,
		ReportDate:  &now,
		Limit:       &limit,
		Offset:      &offset,
	}

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs(&existenceID, &employeeID, &unitType, &now, limit, offset).
		WillReturnRows(rows)

	result, err := handler.List(req)
	require.NoError(t, err)
	assert.Empty(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	now := time.Now()
	quantity := 15.0
	unitType := "Gallons"
	req := models.UpdateRunoutIngredientRequest{
		Quantity:   &quantity,
		UnitType:   &unitType,
		ReportDate: &now,
	}

	expectedRunoutIngredient := models.RunoutIngredient{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		ExistenceID: "550e8400-e29b-41d4-a716-446655440001",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440002",
		Quantity:    quantity,
		UnitType:    unitType,
		ReportDate:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	}).AddRow(
		expectedRunoutIngredient.ID,
		expectedRunoutIngredient.ExistenceID,
		expectedRunoutIngredient.EmployeeID,
		expectedRunoutIngredient.Quantity,
		expectedRunoutIngredient.UnitType,
		expectedRunoutIngredient.ReportDate,
		expectedRunoutIngredient.CreatedAt,
		expectedRunoutIngredient.UpdatedAt,
	)

	mock.ExpectQuery("UPDATE runout_ingredient_report").
		WithArgs(expectedRunoutIngredient.ID, &quantity, &unitType, &now).
		WillReturnRows(rows)

	result, err := handler.Update(req, expectedRunoutIngredient.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedRunoutIngredient, *result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	quantity := 15.0
	req := models.UpdateRunoutIngredientRequest{
		Quantity: &quantity,
	}

	mock.ExpectQuery("UPDATE runout_ingredient_report").
		WithArgs("non-existent-id", &quantity, nil, nil).
		WillReturnError(sql.ErrNoRows)

	result, err := handler.Update(req, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "runout ingredient not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	req := models.DeleteRunoutIngredientRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM runout_ingredient_report").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = handler.Delete(req)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	req := models.DeleteRunoutIngredientRequest{ID: "non-existent-id"}

	mock.ExpectExec("DELETE FROM runout_ingredient_report").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runout ingredient not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientDBHandler_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := NewRunoutIngredientDBHandler(db)

	req := models.DeleteRunoutIngredientRequest{ID: "550e8400-e29b-41d4-a716-446655440000"}

	mock.ExpectExec("DELETE FROM runout_ingredient_report").
		WithArgs(req.ID).
		WillReturnError(sql.ErrConnDone)

	err = handler.Delete(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete runout ingredient")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
