package handlers

import (
	"database/sql"
	"inventory-service/entities/recipe_categories/models"
	recipeSQL "inventory-service/entities/recipe_categories/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for recipe categories
type DBHandler struct {
	db *sql.DB
}

// NewDBHandler creates a new database handler for recipe categories
func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{
		db: db,
	}
}

func (h *DBHandler) Create(req models.CreateRecipeCategoryRequest, logger *logrus.Logger) (*models.RecipeCategory, error) {
	var recipeCategory models.RecipeCategory
	err := h.db.QueryRow(
		recipeSQL.CreateRecipeCategoryQuery,
		req.Name,
		req.Description,
	).Scan(
		&recipeCategory.ID,
		&recipeCategory.Name,
		&recipeCategory.Description,
		&recipeCategory.CreatedAt,
		&recipeCategory.UpdatedAt,
	)

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Failed to create recipe category in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"category_id":   recipeCategory.ID,
		"category_name": recipeCategory.Name,
	}).Info("Recipe category created successfully")

	return &recipeCategory, nil
}

func (h *DBHandler) GetByID(req models.GetRecipeCategoryRequest, logger *logrus.Logger) (*models.RecipeCategory, error) {
	var recipeCategory models.RecipeCategory
	err := h.db.QueryRow(recipeSQL.GetRecipeCategoryByIDQuery, req.ID).Scan(
		&recipeCategory.ID,
		&recipeCategory.Name,
		&recipeCategory.Description,
		&recipeCategory.CreatedAt,
		&recipeCategory.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"category_id": req.ID,
			}).Warn("Recipe category not found")
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": req.ID,
		}).Error("Failed to get recipe category from database")
		return nil, err
	}

	return &recipeCategory, nil
}

func (h *DBHandler) List(req models.ListRecipeCategoriesRequest, logger *logrus.Logger) ([]models.RecipeCategory, error) {
	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}

	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	rows, err := h.db.Query(
		recipeSQL.ListRecipeCategoriesQuery,
		req.Name,
		limit,
		offset,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to execute recipe categories list query")
		return nil, err
	}
	defer rows.Close()

	var recipeCategories []models.RecipeCategory
	for rows.Next() {
		var recipeCategory models.RecipeCategory
		err := rows.Scan(
			&recipeCategory.ID,
			&recipeCategory.Name,
			&recipeCategory.Description,
			&recipeCategory.CreatedAt,
			&recipeCategory.UpdatedAt,
		)
		if err != nil {
			logger.WithError(err).Warn("Failed to scan recipe category row, skipping")
			continue
		}
		recipeCategories = append(recipeCategories, recipeCategory)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	// Ensure we always return an empty slice instead of nil
	if recipeCategories == nil {
		recipeCategories = []models.RecipeCategory{}
	}

	return recipeCategories, nil
}

func (h *DBHandler) Update(req models.UpdateRecipeCategoryRequest, id string, logger *logrus.Logger) (*models.RecipeCategory, error) {
	var recipeCategory models.RecipeCategory
	err := h.db.QueryRow(
		recipeSQL.UpdateRecipeCategoryQuery,
		id,
		req.Name,
		req.Description,
	).Scan(
		&recipeCategory.ID,
		&recipeCategory.Name,
		&recipeCategory.Description,
		&recipeCategory.CreatedAt,
		&recipeCategory.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"category_id": id,
			}).Warn("Recipe category not found for update")
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Failed to update recipe category in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"category_id":   recipeCategory.ID,
		"category_name": recipeCategory.Name,
	}).Info("Recipe category updated successfully")

	return &recipeCategory, nil
}

func (h *DBHandler) Delete(req models.DeleteRecipeCategoryRequest, logger *logrus.Logger) error {
	result, err := h.db.Exec(recipeSQL.DeleteRecipeCategoryQuery, req.ID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": req.ID,
		}).Error("Failed to execute recipe category delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": req.ID,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		logger.WithFields(logrus.Fields{
			"category_id": req.ID,
		}).Warn("No recipe category found to delete")
		return sql.ErrNoRows
	}

	logger.WithFields(logrus.Fields{
		"category_id": req.ID,
	}).Info("Recipe category deleted successfully")

	return nil
}
