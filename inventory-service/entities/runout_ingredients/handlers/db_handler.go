package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"inventory-service/entities/runout_ingredients/models"
	runoutSQL "inventory-service/entities/runout_ingredients/sql"
)

type RunoutIngredientDBHandler struct {
	db *sql.DB
}

func NewRunoutIngredientDBHandler(db *sql.DB) *RunoutIngredientDBHandler {
	return &RunoutIngredientDBHandler{db: db}
}

func (h *RunoutIngredientDBHandler) Create(req models.CreateRunoutIngredientRequest) (*models.RunoutIngredient, error) {
	reportDate := time.Now()
	if req.ReportDate != nil {
		reportDate = *req.ReportDate
	}

	var runoutIngredient models.RunoutIngredient
	err := h.db.QueryRow(
		runoutSQL.CreateRunoutIngredientQuery,
		req.ExistenceID,
		req.EmployeeID,
		req.Quantity,
		req.UnitType,
		reportDate,
	).Scan(
		&runoutIngredient.ID,
		&runoutIngredient.ExistenceID,
		&runoutIngredient.EmployeeID,
		&runoutIngredient.Quantity,
		&runoutIngredient.UnitType,
		&runoutIngredient.ReportDate,
		&runoutIngredient.CreatedAt,
		&runoutIngredient.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create runout ingredient: %w", err)
	}

	return &runoutIngredient, nil
}

func (h *RunoutIngredientDBHandler) GetByID(req models.GetRunoutIngredientRequest) (*models.RunoutIngredient, error) {
	var runoutIngredient models.RunoutIngredient
	err := h.db.QueryRow(runoutSQL.GetRunoutIngredientByIDQuery, req.ID).Scan(
		&runoutIngredient.ID,
		&runoutIngredient.ExistenceID,
		&runoutIngredient.EmployeeID,
		&runoutIngredient.Quantity,
		&runoutIngredient.UnitType,
		&runoutIngredient.ReportDate,
		&runoutIngredient.CreatedAt,
		&runoutIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("runout ingredient not found")
		}
		return nil, fmt.Errorf("failed to get runout ingredient: %w", err)
	}

	return &runoutIngredient, nil
}

func (h *RunoutIngredientDBHandler) List(req models.ListRunoutIngredientsRequest) ([]models.RunoutIngredient, error) {
	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}

	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	rows, err := h.db.Query(
		runoutSQL.ListRunoutIngredientsQuery,
		req.ExistenceID,
		req.EmployeeID,
		req.UnitType,
		req.ReportDate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list runout ingredients: %w", err)
	}
	defer rows.Close()

	var runoutIngredients []models.RunoutIngredient
	for rows.Next() {
		var runoutIngredient models.RunoutIngredient
		err := rows.Scan(
			&runoutIngredient.ID,
			&runoutIngredient.ExistenceID,
			&runoutIngredient.EmployeeID,
			&runoutIngredient.Quantity,
			&runoutIngredient.UnitType,
			&runoutIngredient.ReportDate,
			&runoutIngredient.CreatedAt,
			&runoutIngredient.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan runout ingredient: %w", err)
		}
		runoutIngredients = append(runoutIngredients, runoutIngredient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating runout ingredients: %w", err)
	}

	return runoutIngredients, nil
}

func (h *RunoutIngredientDBHandler) Update(req models.UpdateRunoutIngredientRequest, id string) (*models.RunoutIngredient, error) {
	var runoutIngredient models.RunoutIngredient
	err := h.db.QueryRow(
		runoutSQL.UpdateRunoutIngredientQuery,
		id,
		req.Quantity,
		req.UnitType,
		req.ReportDate,
	).Scan(
		&runoutIngredient.ID,
		&runoutIngredient.ExistenceID,
		&runoutIngredient.EmployeeID,
		&runoutIngredient.Quantity,
		&runoutIngredient.UnitType,
		&runoutIngredient.ReportDate,
		&runoutIngredient.CreatedAt,
		&runoutIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("runout ingredient not found")
		}
		return nil, fmt.Errorf("failed to update runout ingredient: %w", err)
	}

	return &runoutIngredient, nil
}

func (h *RunoutIngredientDBHandler) Delete(req models.DeleteRunoutIngredientRequest) error {
	result, err := h.db.Exec(runoutSQL.DeleteRunoutIngredientQuery, req.ID)
	if err != nil {
		return fmt.Errorf("failed to delete runout ingredient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("runout ingredient not found")
	}

	return nil
}
