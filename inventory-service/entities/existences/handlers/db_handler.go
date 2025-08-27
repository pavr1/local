package handlers

import (
	"database/sql"

	"inventory-service/entities/existences/models"
	existenceSQL "inventory-service/entities/existences/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for existences
type DBHandler struct {
	db *sql.DB
}

// NewDBHandler creates a new database handler for existences
func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{
		db: db,
	}
}

// CreateExistence creates a new existence in the database
func (h *DBHandler) CreateExistence(req models.CreateExistenceRequest, logger *logrus.Entry) (*models.Existence, error) {
	var existence models.Existence

	err := h.db.QueryRow(existenceSQL.CreateExistenceQuery,
		req.IngredientID,
		req.InvoiceDetailID,
		req.UnitsPurchased,
		req.UnitsAvailable,
		req.UnitType,
		req.ItemsPerUnit,
		req.CostPerUnit,
		req.ExpirationDate,
		req.IncomeMarginPercentage,
		req.MinimumPrice,
		req.MaximumPrice,
		req.FinalPrice).
		Scan(&existence.ID, &existence.ExistenceReferenceCode, &existence.IngredientID,
			&existence.InvoiceDetailID, &existence.UnitsPurchased, &existence.UnitsAvailable,
			&existence.UnitType, &existence.ItemsPerUnit, &existence.CostPerItem,
			&existence.CostPerUnit, &existence.TotalPurchaseCost, &existence.RemainingValue,
			&existence.ExpirationDate, &existence.IncomeMarginPercentage, &existence.IncomeMarginAmount,
			&existence.MinimumPrice, &existence.MaximumPrice, &existence.FinalPrice,
			&existence.CreatedAt, &existence.UpdatedAt)

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id":     req.IngredientID,
			"invoice_detail_id": req.InvoiceDetailID,
		}).Error("Failed to create existence in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"existence_id":   existence.ID,
		"reference_code": existence.ExistenceReferenceCode,
		"ingredient_id":  existence.IngredientID,
	}).Info("Existence created successfully")

	return &existence, nil
}

// GetExistenceByID retrieves an existence by ID from the database
func (h *DBHandler) GetExistenceByID(id string, logger *logrus.Entry) (*models.Existence, error) {
	var existence models.Existence

	err := h.db.QueryRow(existenceSQL.GetExistenceByIDQuery, id).
		Scan(&existence.ID, &existence.ExistenceReferenceCode, &existence.IngredientID,
			&existence.InvoiceDetailID, &existence.UnitsPurchased, &existence.UnitsAvailable,
			&existence.UnitType, &existence.ItemsPerUnit, &existence.CostPerItem,
			&existence.CostPerUnit, &existence.TotalPurchaseCost, &existence.RemainingValue,
			&existence.ExpirationDate, &existence.IncomeMarginPercentage, &existence.IncomeMarginAmount,
			&existence.MinimumPrice, &existence.MaximumPrice, &existence.FinalPrice,
			&existence.CreatedAt, &existence.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"existence_id": id,
			}).Warn("Existence not found")
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"existence_id": id,
		}).Error("Failed to get existence from database")
		return nil, err
	}

	return &existence, nil
}

// GetMostRecentExistenceByIngredientAndUnitType retrieves the most recent existence for a specific ingredient and unit type
func (h *DBHandler) GetMostRecentExistenceByIngredientAndUnitType(ingredientID, unitType string, logger *logrus.Entry) (*models.Existence, error) {
	var existence models.Existence

	err := h.db.QueryRow(existenceSQL.GetMostRecentExistenceByIngredientAndUnitTypeQuery, ingredientID, unitType).
		Scan(&existence.ID, &existence.ExistenceReferenceCode, &existence.IngredientID,
			&existence.InvoiceDetailID, &existence.UnitsPurchased, &existence.UnitsAvailable,
			&existence.UnitType, &existence.ItemsPerUnit, &existence.CostPerItem,
			&existence.CostPerUnit, &existence.TotalPurchaseCost, &existence.RemainingValue,
			&existence.ExpirationDate, &existence.IncomeMarginPercentage, &existence.IncomeMarginAmount,
			&existence.MinimumPrice, &existence.MaximumPrice, &existence.FinalPrice,
			&existence.CreatedAt, &existence.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": ingredientID,
			"unit_type":     unitType,
		}).Error("Failed to get most recent existence from database")
		return nil, err
	}

	return &existence, nil
}

// ListExistences retrieves all existences from the database with optional filtering
func (h *DBHandler) ListExistences(req models.ListExistencesRequest, logger *logrus.Entry) ([]models.Existence, error) {
	rows, err := h.db.Query(existenceSQL.ListExistencesQuery,
		req.IngredientID, req.UnitType, req.Expired, req.LowStock, req.Limit, req.Offset)
	if err != nil {
		logger.WithError(err).Error("Failed to list existences from database")
		return nil, err
	}
	defer rows.Close()

	var existences []models.Existence
	for rows.Next() {
		var existence models.Existence
		err := rows.Scan(&existence.ID, &existence.ExistenceReferenceCode, &existence.IngredientID,
			&existence.IngredientName, &existence.IngredientCategory, &existence.InvoiceDetailID,
			&existence.UnitsPurchased, &existence.UnitsAvailable, &existence.UnitType,
			&existence.ItemsPerUnit, &existence.CostPerItem, &existence.CostPerUnit,
			&existence.TotalPurchaseCost, &existence.RemainingValue, &existence.ExpirationDate,
			&existence.IncomeMarginPercentage, &existence.IncomeMarginAmount,
			&existence.MinimumPrice, &existence.MaximumPrice, &existence.FinalPrice, &existence.CreatedAt, &existence.UpdatedAt)

		if err != nil {
			logger.WithError(err).Error("Failed to scan existence row")
			return nil, err
		}
		existences = append(existences, existence)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	// Return empty slice instead of nil if no existences found
	if existences == nil {
		existences = []models.Existence{}
	}

	return existences, nil
}

// UpdateExistence updates an existence in the database
func (h *DBHandler) UpdateExistence(id string, req models.UpdateExistenceRequest, logger *logrus.Entry) (*models.Existence, error) {
	var existence models.Existence

	err := h.db.QueryRow(existenceSQL.UpdateExistenceQuery, id,
		req.UnitsAvailable, req.UnitType, req.ItemsPerUnit, req.CostPerUnit,
		req.ExpirationDate, req.IncomeMarginPercentage, req.MinimumPrice, req.MaximumPrice, req.FinalPrice).
		Scan(&existence.ID, &existence.ExistenceReferenceCode, &existence.IngredientID,
			&existence.InvoiceDetailID, &existence.UnitsPurchased, &existence.UnitsAvailable,
			&existence.UnitType, &existence.ItemsPerUnit, &existence.CostPerItem,
			&existence.CostPerUnit, &existence.TotalPurchaseCost, &existence.RemainingValue,
			&existence.ExpirationDate, &existence.IncomeMarginPercentage, &existence.IncomeMarginAmount,
			&existence.MinimumPrice, &existence.MaximumPrice, &existence.FinalPrice,
			&existence.CreatedAt, &existence.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"existence_id": id,
			}).Warn("Existence not found for update")
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"existence_id": id,
		}).Error("Failed to update existence in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"existence_id":   existence.ID,
		"reference_code": existence.ExistenceReferenceCode,
	}).Info("Existence updated successfully")

	return &existence, nil
}

// DeleteExistence deletes an existence from the database
func (h *DBHandler) DeleteExistence(id string, logger *logrus.Entry) error {
	result, err := h.db.Exec(existenceSQL.DeleteExistenceQuery, id)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"existence_id": id,
		}).Error("Failed to delete existence from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"existence_id": id,
		}).Error("Failed to get rows affected after deletion")
		return err
	}

	if rowsAffected == 0 {
		logger.WithFields(logrus.Fields{
			"existence_id": id,
		}).Warn("No existence found to delete")
		return sql.ErrNoRows
	}

	logger.WithFields(logrus.Fields{
		"existence_id": id,
	}).Info("Existence deleted successfully")

	return nil
}
