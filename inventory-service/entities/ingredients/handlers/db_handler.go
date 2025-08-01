package handlers

import (
	"database/sql"

	"inventory-service/entities/ingredients/models"
	ingredientSQL "inventory-service/entities/ingredients/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for ingredients
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for ingredients
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// CreateIngredient creates a new ingredient in the database
func (h *DBHandler) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	var ingredient models.Ingredient

	err := h.db.QueryRow(ingredientSQL.CreateIngredientQuery,
		req.Name, req.SupplierID).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_name": req.Name,
		}).Error("Failed to create ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id":   ingredient.ID,
		"ingredient_name": ingredient.Name,
	}).Info("Ingredient created successfully")

	return &ingredient, nil
}

// GetIngredientByID retrieves an ingredient by ID from the database
func (h *DBHandler) GetIngredientByID(id string) (*models.Ingredient, error) {
	var ingredient models.Ingredient

	err := h.db.QueryRow(ingredientSQL.GetIngredientByIDQuery, id).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Failed to retrieve ingredient from database")
		return nil, err
	}

	return &ingredient, nil
}

// ListIngredients retrieves all ingredients from the database
func (h *DBHandler) ListIngredients() ([]models.Ingredient, error) {
	rows, err := h.db.Query(ingredientSQL.ListIngredientsQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute ingredients list query")
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ingredient models.Ingredient
		err := rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan ingredient row, skipping")
			continue
		}
		ingredients = append(ingredients, ingredient)
	}

	h.logger.WithFields(logrus.Fields{
		"ingredients_count": len(ingredients),
	}).Info("Listed ingredients successfully")

	return ingredients, nil
}

// UpdateIngredient updates an ingredient in the database
func (h *DBHandler) UpdateIngredient(id string, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	var ingredient models.Ingredient

	err := h.db.QueryRow(ingredientSQL.UpdateIngredientQuery,
		id, req.Name, req.SupplierID).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Failed to update ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id":   ingredient.ID,
		"ingredient_name": ingredient.Name,
	}).Info("Ingredient updated successfully")

	return &ingredient, nil
}

// DeleteIngredient deletes an ingredient from the database
func (h *DBHandler) DeleteIngredient(id string) error {
	result, err := h.db.Exec(ingredientSQL.DeleteIngredientQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Failed to execute ingredient delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		// Don't log as error since "not found" is a normal business case
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id": id,
	}).Info("Ingredient deleted successfully")

	return nil
}
