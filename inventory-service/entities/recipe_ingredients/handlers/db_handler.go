package handlers

import (
	"database/sql"
	"inventory-service/entities/recipe_ingredients/models"
	recipeIngredientSQL "inventory-service/entities/recipe_ingredients/sql"

	"github.com/sirupsen/logrus"
)

// RecipeIngredientDBHandler handles database operations for recipe ingredients
type RecipeIngredientDBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for recipe ingredients
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *RecipeIngredientDBHandler {
	return &RecipeIngredientDBHandler{
		db:     db,
		logger: logger,
	}
}

func (h *RecipeIngredientDBHandler) Create(req models.CreateRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	var recipeIngredient models.RecipeIngredient
	err := h.db.QueryRow(
		recipeIngredientSQL.CreateRecipeIngredientQuery,
		req.RecipeID,
		req.IngredientID,
		req.Quantity,
	).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id":     req.RecipeID,
			"ingredient_id": req.IngredientID,
		}).Error("Failed to create recipe ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": recipeIngredient.ID,
		"recipe_id":            recipeIngredient.RecipeID,
		"ingredient_id":        recipeIngredient.IngredientID,
	}).Info("Recipe ingredient created successfully")

	return &recipeIngredient, nil
}

func (h *RecipeIngredientDBHandler) GetByID(req models.GetRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	var recipeIngredient models.RecipeIngredient
	err := h.db.QueryRow(recipeIngredientSQL.GetRecipeIngredientByIDQuery, req.ID).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"recipe_ingredient_id": req.ID,
			}).Warn("Recipe ingredient not found")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_ingredient_id": req.ID,
		}).Error("Failed to get recipe ingredient from database")
		return nil, err
	}

	return &recipeIngredient, nil
}

func (h *RecipeIngredientDBHandler) List(req models.ListRecipeIngredientsRequest) ([]models.RecipeIngredient, error) {
	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}

	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	rows, err := h.db.Query(
		recipeIngredientSQL.ListRecipeIngredientsQuery,
		req.RecipeID,
		req.IngredientID,
		limit,
		offset,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute recipe ingredients list query")
		return nil, err
	}
	defer rows.Close()

	var recipeIngredients []models.RecipeIngredient
	for rows.Next() {
		var recipeIngredient models.RecipeIngredient
		err := rows.Scan(
			&recipeIngredient.ID,
			&recipeIngredient.RecipeID,
			&recipeIngredient.IngredientID,
			&recipeIngredient.Quantity,
			&recipeIngredient.IngredientName,
			&recipeIngredient.FinalPrice,
			&recipeIngredient.CreatedAt,
			&recipeIngredient.UpdatedAt,
		)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan recipe ingredient row, skipping")
			continue
		}
		recipeIngredients = append(recipeIngredients, recipeIngredient)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	// Ensure we always return an empty slice instead of nil
	if recipeIngredients == nil {
		recipeIngredients = []models.RecipeIngredient{}
	}

	return recipeIngredients, nil
}

func (h *RecipeIngredientDBHandler) Update(req models.UpdateRecipeIngredientRequest, id string) (*models.RecipeIngredient, error) {
	var recipeIngredient models.RecipeIngredient
	err := h.db.QueryRow(
		recipeIngredientSQL.UpdateRecipeIngredientQuery,
		id,
		req.RecipeID,
		req.IngredientID,
		req.Quantity,
	).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"recipe_ingredient_id": id,
			}).Warn("Recipe ingredient not found for update")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_ingredient_id": id,
		}).Error("Failed to update recipe ingredient in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": recipeIngredient.ID,
		"recipe_id":            recipeIngredient.RecipeID,
		"ingredient_id":        recipeIngredient.IngredientID,
	}).Info("Recipe ingredient updated successfully")

	return &recipeIngredient, nil
}

func (h *RecipeIngredientDBHandler) Delete(req models.DeleteRecipeIngredientRequest) error {
	result, err := h.db.Exec(recipeIngredientSQL.DeleteRecipeIngredientQuery, req.ID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_ingredient_id": req.ID,
		}).Error("Failed to execute recipe ingredient delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_ingredient_id": req.ID,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"recipe_ingredient_id": req.ID,
		}).Warn("No recipe ingredient found to delete")
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": req.ID,
	}).Info("Recipe ingredient deleted successfully")

	return nil
}
