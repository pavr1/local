package handlers

import (
	"database/sql"
	"inventory-service/entities/runout_ingredients/models"
	runoutSQL "inventory-service/entities/runout_ingredients/sql"
	"time"

	"github.com/sirupsen/logrus"
)

// RunoutIngredientDBHandler handles database operations for runout ingredients
type RunoutIngredientDBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for runout ingredients
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *RunoutIngredientDBHandler {
	return &RunoutIngredientDBHandler{
		db:     db,
		logger: logger,
	}
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
		h.logger.WithError(err).WithFields(logrus.Fields{
			"existence_id": req.ExistenceID,
			"employee_id":  req.EmployeeID,
			"quantity":     req.Quantity,
		}).Error("Failed to create runout ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"runout_ingredient_id": runoutIngredient.ID,
		"existence_id":         runoutIngredient.ExistenceID,
		"employee_id":          runoutIngredient.EmployeeID,
		"quantity":             runoutIngredient.Quantity,
	}).Info("Runout ingredient created successfully")

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
			h.logger.WithFields(logrus.Fields{
				"runout_ingredient_id": req.ID,
			}).Warn("Runout ingredient not found")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"runout_ingredient_id": req.ID,
		}).Error("Failed to get runout ingredient from database")
		return nil, err
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
		h.logger.WithError(err).Error("Failed to execute runout ingredients list query")
		return nil, err
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
			h.logger.WithError(err).Warn("Failed to scan runout ingredient row, skipping")
			continue
		}
		runoutIngredients = append(runoutIngredients, runoutIngredient)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	// Ensure we always return an empty slice instead of nil
	if runoutIngredients == nil {
		runoutIngredients = []models.RunoutIngredient{}
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
			h.logger.WithFields(logrus.Fields{
				"runout_ingredient_id": id,
			}).Warn("Runout ingredient not found for update")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"runout_ingredient_id": id,
		}).Error("Failed to update runout ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"runout_ingredient_id": runoutIngredient.ID,
		"existence_id":         runoutIngredient.ExistenceID,
		"quantity":             runoutIngredient.Quantity,
	}).Info("Runout ingredient updated successfully")

	return &runoutIngredient, nil
}

func (h *RunoutIngredientDBHandler) Delete(req models.DeleteRunoutIngredientRequest) error {
	result, err := h.db.Exec(runoutSQL.DeleteRunoutIngredientQuery, req.ID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"runout_ingredient_id": req.ID,
		}).Error("Failed to execute runout ingredient delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"runout_ingredient_id": req.ID,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"runout_ingredient_id": req.ID,
		}).Warn("No runout ingredient found to delete")
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"runout_ingredient_id": req.ID,
	}).Info("Runout ingredient deleted successfully")

	return nil
}
