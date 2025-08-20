package handlers

import (
	"database/sql"
	recipeIngredientsSQL "inventory-service/entities/recipe_ingredients/sql"
	"inventory-service/entities/recipes/models"
	recipeSQL "inventory-service/entities/recipes/sql"

	"github.com/sirupsen/logrus"
)

// RecipeDBHandler handles database operations for recipes
type RecipeDBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewRecipeDBHandler creates a new database handler for recipes
func NewRecipeDBHandler(db *sql.DB, logger *logrus.Logger) *RecipeDBHandler {
	return &RecipeDBHandler{
		db:     db,
		logger: logger,
	}
}

func (h *RecipeDBHandler) Create(req models.CreateRecipeRequest) (*models.Recipe, error) {
	// Start a transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback() // Rollback if not committed

	var recipe models.Recipe
	err = tx.QueryRow(
		recipeSQL.CreateRecipeQuery,
		req.RecipeName,
		req.RecipeDescription,
		req.PictureURL,
		req.RecipeCategoryID,
		req.TotalRecipeCost,
	).Scan(
		&recipe.ID,
		&recipe.RecipeName,
		&recipe.RecipeDescription,
		&recipe.PictureURL,
		&recipe.RecipeCategoryID,
		&recipe.TotalRecipeCost,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_name": req.RecipeName,
		}).Error("Failed to create recipe in database")
		return nil, err
	}

	// Create recipe ingredients
	for _, ingredient := range req.Ingredients {
		_, err = tx.Exec(
			recipeIngredientsSQL.CreateRecipeIngredientQuery,
			recipe.ID,
			ingredient.IngredientID,
			ingredient.NumberOfUnits,
		)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"recipe_id":     recipe.ID,
				"ingredient_id": ingredient.IngredientID,
			}).Error("Failed to create recipe ingredient in database")
			return nil, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_id":         recipe.ID,
		"recipe_name":       recipe.RecipeName,
		"ingredients_count": len(req.Ingredients),
	}).Info("Recipe and ingredients created successfully")

	return &recipe, nil
}

func (h *RecipeDBHandler) GetByID(req models.GetRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := h.db.QueryRow(recipeSQL.GetRecipeByIDQuery, req.ID).Scan(
		&recipe.ID,
		&recipe.RecipeName,
		&recipe.RecipeDescription,
		&recipe.PictureURL,
		&recipe.RecipeCategoryID,
		&recipe.RecipeCategoryName,
		&recipe.TotalRecipeCost,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"recipe_id": req.ID,
			}).Warn("Recipe not found")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": req.ID,
		}).Error("Failed to get recipe from database")
		return nil, err
	}

	return &recipe, nil
}

func (h *RecipeDBHandler) List(req models.ListRecipesRequest) ([]models.Recipe, error) {
	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}

	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	rows, err := h.db.Query(
		recipeSQL.ListRecipesQuery,
		req.RecipeName,
		req.RecipeCategoryID,
		limit,
		offset,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute recipes list query")
		return nil, err
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var recipe models.Recipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.RecipeName,
			&recipe.RecipeDescription,
			&recipe.PictureURL,
			&recipe.RecipeCategoryID,
			&recipe.RecipeCategoryName,
			&recipe.TotalRecipeCost,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan recipe row, skipping")
			continue
		}
		recipes = append(recipes, recipe)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	// Ensure we always return an empty slice instead of nil
	if recipes == nil {
		recipes = []models.Recipe{}
	}

	h.logger.WithFields(logrus.Fields{
		"recipes_count": len(recipes),
	}).Info("Listed recipes successfully")

	return recipes, nil
}

func (h *RecipeDBHandler) Update(req models.UpdateRecipeRequest, id string) (*models.Recipe, error) {
	var recipe models.Recipe
	err := h.db.QueryRow(
		recipeSQL.UpdateRecipeQuery,
		id,
		req.RecipeName,
		req.RecipeDescription,
		req.PictureURL,
		req.RecipeCategoryID,
		req.TotalRecipeCost,
	).Scan(
		&recipe.ID,
		&recipe.RecipeName,
		&recipe.RecipeDescription,
		&recipe.PictureURL,
		&recipe.RecipeCategoryID,
		&recipe.TotalRecipeCost,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"recipe_id": id,
			}).Warn("Recipe not found for update")
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": id,
		}).Error("Failed to update recipe in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_id":   recipe.ID,
		"recipe_name": recipe.RecipeName,
	}).Info("Recipe updated successfully")

	return &recipe, nil
}

func (h *RecipeDBHandler) Delete(req models.DeleteRecipeRequest) error {
	result, err := h.db.Exec(recipeSQL.DeleteRecipeQuery, req.ID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": req.ID,
		}).Error("Failed to execute recipe delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": req.ID,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"recipe_id": req.ID,
		}).Warn("No recipe found to delete")
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_id": req.ID,
	}).Info("Recipe deleted successfully")

	return nil
}
