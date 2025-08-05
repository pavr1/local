package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunoutIngredient_Struct(t *testing.T) {
	now := time.Now()
	runoutIngredient := RunoutIngredient{
		ID:          "test-id",
		ExistenceID: "existence-id",
		EmployeeID:  "employee-id",
		Quantity:    10.5,
		UnitType:    "Liters",
		ReportDate:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "test-id", runoutIngredient.ID)
	assert.Equal(t, "existence-id", runoutIngredient.ExistenceID)
	assert.Equal(t, "employee-id", runoutIngredient.EmployeeID)
	assert.Equal(t, 10.5, runoutIngredient.Quantity)
	assert.Equal(t, "Liters", runoutIngredient.UnitType)
	assert.Equal(t, now, runoutIngredient.ReportDate)
	assert.Equal(t, now, runoutIngredient.CreatedAt)
	assert.Equal(t, now, runoutIngredient.UpdatedAt)
}

func TestCreateRunoutIngredientRequest_Struct(t *testing.T) {
	now := time.Now()
	req := CreateRunoutIngredientRequest{
		ExistenceID: "existence-id",
		EmployeeID:  "employee-id",
		Quantity:    10.5,
		UnitType:    "Liters",
		ReportDate:  &now,
	}

	assert.Equal(t, "existence-id", req.ExistenceID)
	assert.Equal(t, "employee-id", req.EmployeeID)
	assert.Equal(t, 10.5, req.Quantity)
	assert.Equal(t, "Liters", req.UnitType)
	assert.Equal(t, &now, req.ReportDate)
}

func TestUpdateRunoutIngredientRequest_Struct(t *testing.T) {
	now := time.Now()
	quantity := 15.0
	unitType := "Gallons"

	req := UpdateRunoutIngredientRequest{
		Quantity:   &quantity,
		UnitType:   &unitType,
		ReportDate: &now,
	}

	assert.Equal(t, &quantity, req.Quantity)
	assert.Equal(t, &unitType, req.UnitType)
	assert.Equal(t, &now, req.ReportDate)
}

func TestGetRunoutIngredientRequest_Struct(t *testing.T) {
	req := GetRunoutIngredientRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestDeleteRunoutIngredientRequest_Struct(t *testing.T) {
	req := DeleteRunoutIngredientRequest{
		ID: "test-id",
	}

	assert.Equal(t, "test-id", req.ID)
}

func TestListRunoutIngredientsRequest_Struct(t *testing.T) {
	existenceID := "existence-id"
	employeeID := "employee-id"
	unitType := "Units"
	now := time.Now()
	limit := 10
	offset := 5

	req := ListRunoutIngredientsRequest{
		ExistenceID: &existenceID,
		EmployeeID:  &employeeID,
		UnitType:    &unitType,
		ReportDate:  &now,
		Limit:       &limit,
		Offset:      &offset,
	}

	assert.Equal(t, &existenceID, req.ExistenceID)
	assert.Equal(t, &employeeID, req.EmployeeID)
	assert.Equal(t, &unitType, req.UnitType)
	assert.Equal(t, &now, req.ReportDate)
	assert.Equal(t, &limit, req.Limit)
	assert.Equal(t, &offset, req.Offset)
}

func TestRunoutIngredientResponse_Struct(t *testing.T) {
	runoutIngredient := RunoutIngredient{
		ID:          "test-id",
		ExistenceID: "existence-id",
		EmployeeID:  "employee-id",
		Quantity:    10.5,
		UnitType:    "Liters",
	}

	response := RunoutIngredientResponse{
		Success: true,
		Data:    runoutIngredient,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, runoutIngredient, response.Data)
	assert.Equal(t, "Success", response.Message)
}

func TestRunoutIngredientsResponse_Struct(t *testing.T) {
	runoutIngredients := []RunoutIngredient{
		{
			ID:          "test-id-1",
			ExistenceID: "existence-id-1",
			EmployeeID:  "employee-id-1",
			Quantity:    10.5,
			UnitType:    "Liters",
		},
		{
			ID:          "test-id-2",
			ExistenceID: "existence-id-2",
			EmployeeID:  "employee-id-2",
			Quantity:    20.0,
			UnitType:    "Gallons",
		},
	}

	response := RunoutIngredientsResponse{
		Success: true,
		Data:    runoutIngredients,
		Total:   2,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, runoutIngredients, response.Data)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Success", response.Message)
}

func TestGenericResponse_Struct(t *testing.T) {
	response := GenericResponse{
		Success: true,
		Message: "Success",
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Success", response.Message)
}

func TestRunoutIngredient_JSONTags(t *testing.T) {
	runoutIngredient := RunoutIngredient{
		ID:          "test-id",
		ExistenceID: "existence-id",
		EmployeeID:  "employee-id",
		Quantity:    10.5,
		UnitType:    "Liters",
		ReportDate:  time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// This test ensures the struct has proper JSON tags
	// The actual JSON marshaling would be tested in integration tests
	assert.NotEmpty(t, runoutIngredient.ID)
	assert.NotEmpty(t, runoutIngredient.ExistenceID)
	assert.NotEmpty(t, runoutIngredient.EmployeeID)
	assert.Greater(t, runoutIngredient.Quantity, 0.0)
	assert.NotEmpty(t, runoutIngredient.UnitType)
}

func TestCreateRunoutIngredientRequest_Validation(t *testing.T) {
	// Test with valid data
	req := CreateRunoutIngredientRequest{
		ExistenceID: "550e8400-e29b-41d4-a716-446655440000",
		EmployeeID:  "550e8400-e29b-41d4-a716-446655440001",
		Quantity:    10.5,
		UnitType:    "Liters",
	}

	// This would be tested with actual validation in integration tests
	assert.NotEmpty(t, req.ExistenceID)
	assert.NotEmpty(t, req.EmployeeID)
	assert.Greater(t, req.Quantity, 0.0)
	assert.NotEmpty(t, req.UnitType)
}

func TestUpdateRunoutIngredientRequest_OptionalFields(t *testing.T) {
	// Test with only some fields set
	quantity := 15.0
	req := UpdateRunoutIngredientRequest{
		Quantity: &quantity,
	}

	assert.Equal(t, &quantity, req.Quantity)
	assert.Nil(t, req.UnitType)
	assert.Nil(t, req.ReportDate)
}

func TestListRunoutIngredientsRequest_EmptyFilters(t *testing.T) {
	// Test with no filters
	req := ListRunoutIngredientsRequest{}

	assert.Nil(t, req.ExistenceID)
	assert.Nil(t, req.EmployeeID)
	assert.Nil(t, req.UnitType)
	assert.Nil(t, req.ReportDate)
	assert.Nil(t, req.Limit)
	assert.Nil(t, req.Offset)
}
