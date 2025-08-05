package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-service/entities/runout_ingredients/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunoutIngredientHTTPHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.dbHandler.db)
	assert.Equal(t, logger, handler.logger)
}

func TestRunoutIngredientHTTPHandler_CreateRunoutIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	req := models.CreateRunoutIngredientRequest{
		ExistenceID: "550e8400-e29b-41d4-a716-446655440000",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440001",
		Quantity:    10.5,
		UnitType:    "Liters",
	}

	now := time.Now()
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
		WithArgs(req.ExistenceID, req.EmployeeID, req.Quantity, req.UnitType, sqlmock.AnyArg()).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/runout-ingredients", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRunoutIngredient(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRunoutIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRunoutIngredient.ExistenceID, result.Data.ExistenceID)
	assert.Equal(t, expectedRunoutIngredient.EmployeeID, result.Data.EmployeeID)
	assert.Equal(t, expectedRunoutIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRunoutIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "created successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_CreateRunoutIngredient_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("POST", "/runout-ingredients", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRunoutIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRunoutIngredientHTTPHandler_CreateRunoutIngredient_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	req := models.CreateRunoutIngredientRequest{
		ExistenceID: "550e8400-e29b-41d4-a716-446655440000",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440001",
		Quantity:    10.5,
		UnitType:    "Liters",
	}

	mock.ExpectQuery("INSERT INTO runout_ingredient_report").
		WithArgs(req.ExistenceID, req.EmployeeID, req.Quantity, req.UnitType, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/runout-ingredients", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.CreateRunoutIngredient(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to create runout ingredient")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_GetRunoutIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

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

	request := httptest.NewRequest("GET", "/runout-ingredients/"+expectedRunoutIngredient.ID, nil)
	response := httptest.NewRecorder()

	// Set up the router to extract the ID parameter
	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.GetRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRunoutIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRunoutIngredient.ExistenceID, result.Data.ExistenceID)
	assert.Equal(t, expectedRunoutIngredient.EmployeeID, result.Data.EmployeeID)
	assert.Equal(t, expectedRunoutIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRunoutIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_GetRunoutIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	request := httptest.NewRequest("GET", "/runout-ingredients/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.GetRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_GetRunoutIngredient_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("GET", "/runout-ingredients/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.GetRunoutIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRunoutIngredientHTTPHandler_ListRunoutIngredients(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

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

	request := httptest.NewRequest("GET", "/runout-ingredients", nil)
	response := httptest.NewRecorder()

	handler.ListRunoutIngredients(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RunoutIngredientsResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, len(expectedRunoutIngredients), len(result.Data))
	for i, expected := range expectedRunoutIngredients {
		assert.Equal(t, expected.ID, result.Data[i].ID)
		assert.Equal(t, expected.ExistenceID, result.Data[i].ExistenceID)
		assert.Equal(t, expected.EmployeeID, result.Data[i].EmployeeID)
		assert.Equal(t, expected.Quantity, result.Data[i].Quantity)
		assert.Equal(t, expected.UnitType, result.Data[i].UnitType)
	}
	assert.Equal(t, 1, result.Total)
	assert.Contains(t, result.Message, "retrieved successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_ListRunoutIngredients_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	rows := sqlmock.NewRows([]string{
		"id", "existence_id", "employee_id", "quantity", "unit_type",
		"report_date", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at").
		WithArgs("existence-id", "employee-id", "Liters", sqlmock.AnyArg(), 10, 5).
		WillReturnRows(rows)

	request := httptest.NewRequest("GET", "/runout-ingredients?existence_id=existence-id&employee_id=employee-id&unit_type=Liters&limit=10&offset=5", nil)
	response := httptest.NewRecorder()

	handler.ListRunoutIngredients(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RunoutIngredientsResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.Data)
	assert.Equal(t, 0, result.Total)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_UpdateRunoutIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

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
		WithArgs(expectedRunoutIngredient.ID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/runout-ingredients/"+expectedRunoutIngredient.ID, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.UpdateRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Compare fields individually to avoid time precision issues
	assert.Equal(t, expectedRunoutIngredient.ID, result.Data.ID)
	assert.Equal(t, expectedRunoutIngredient.ExistenceID, result.Data.ExistenceID)
	assert.Equal(t, expectedRunoutIngredient.EmployeeID, result.Data.EmployeeID)
	assert.Equal(t, expectedRunoutIngredient.Quantity, result.Data.Quantity)
	assert.Equal(t, expectedRunoutIngredient.UnitType, result.Data.UnitType)
	assert.Contains(t, result.Message, "updated successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_UpdateRunoutIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	quantity := 15.0
	req := models.UpdateRunoutIngredientRequest{
		Quantity: &quantity,
	}

	mock.ExpectQuery("UPDATE runout_ingredient_report").
		WithArgs("non-existent-id", &quantity, nil, nil).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("PUT", "/runout-ingredients/non-existent-id", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.UpdateRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.RunoutIngredientResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_DeleteRunoutIngredient(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	runoutIngredientID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectExec("DELETE FROM runout_ingredient_report").
		WithArgs(runoutIngredientID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest("DELETE", "/runout-ingredients/"+runoutIngredientID, nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.DeleteRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result models.GenericResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "deleted successfully")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_DeleteRunoutIngredient_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	mock.ExpectExec("DELETE FROM runout_ingredient_report").
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	request := httptest.NewRequest("DELETE", "/runout-ingredients/non-existent-id", nil)
	response := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/runout-ingredients/{id}", handler.DeleteRunoutIngredient)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)

	var result models.GenericResponse
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRunoutIngredientHTTPHandler_DeleteRunoutIngredient_MissingID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	handler := NewRunoutIngredientHTTPHandler(db, logger)

	request := httptest.NewRequest("DELETE", "/runout-ingredients/", nil)
	response := httptest.NewRecorder()

	// Call the handler directly since router won't match empty ID
	handler.DeleteRunoutIngredient(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}
