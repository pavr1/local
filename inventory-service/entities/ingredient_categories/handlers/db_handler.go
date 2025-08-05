package handlers

import (
	"database/sql"

	"inventory-service/entities/ingredient_categories/models"
	ingredientCategorySQL "inventory-service/entities/ingredient_categories/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for ingredient categories
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for ingredient categories
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// CreateIngredientCategory creates a new ingredient category in the database
func (h *DBHandler) CreateIngredientCategory(req models.CreateIngredientCategoryRequest) (*models.IngredientCategory, error) {
	var category models.IngredientCategory

	err := h.db.QueryRow(ingredientCategorySQL.CreateIngredientCategoryQuery,
		req.Name, req.Description, req.IsActive).
		Scan(&category.ID, &category.Name, &category.Description, &category.IsActive, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Failed to create ingredient category in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"category_id":   category.ID,
		"category_name": category.Name,
	}).Info("Ingredient category created successfully")

	return &category, nil
}

// GetIngredientCategoryByID retrieves an ingredient category by ID from the database
func (h *DBHandler) GetIngredientCategoryByID(id string) (*models.IngredientCategory, error) {
	var category models.IngredientCategory

	err := h.db.QueryRow(ingredientCategorySQL.GetIngredientCategoryByIDQuery, id).
		Scan(&category.ID, &category.Name, &category.Description, &category.IsActive, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Failed to retrieve ingredient category from database")
		return nil, err
	}

	return &category, nil
}

// ListIngredientCategories retrieves all ingredient categories from the database
func (h *DBHandler) ListIngredientCategories() ([]models.IngredientCategory, error) {
	rows, err := h.db.Query(ingredientCategorySQL.ListIngredientCategoriesQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute ingredient categories list query")
		return nil, err
	}
	defer rows.Close()

	var categories []models.IngredientCategory
	for rows.Next() {
		var category models.IngredientCategory
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.IsActive, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan ingredient category row, skipping")
			continue
		}
		categories = append(categories, category)
	}

	// Ensure we return an empty slice instead of nil for consistency
	if categories == nil {
		categories = []models.IngredientCategory{}
	}

	h.logger.WithFields(logrus.Fields{
		"categories_count": len(categories),
	}).Info("Listed ingredient categories successfully")

	return categories, nil
}

// UpdateIngredientCategory updates an ingredient category in the database
func (h *DBHandler) UpdateIngredientCategory(id string, req models.UpdateIngredientCategoryRequest) (*models.IngredientCategory, error) {
	var category models.IngredientCategory

	err := h.db.QueryRow(ingredientCategorySQL.UpdateIngredientCategoryQuery,
		id, req.Name, req.Description, req.IsActive).
		Scan(&category.ID, &category.Name, &category.Description, &category.IsActive, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Failed to update ingredient category in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"category_id":   category.ID,
		"category_name": category.Name,
	}).Info("Ingredient category updated successfully")

	return &category, nil
}

// DeleteIngredientCategory deletes an ingredient category from the database
func (h *DBHandler) DeleteIngredientCategory(id string) error {
	result, err := h.db.Exec(ingredientCategorySQL.DeleteIngredientCategoryQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Failed to execute ingredient category delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		// Don't log as error since "not found" is a normal business case
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"category_id": id,
	}).Info("Ingredient category deleted successfully")

	return nil
}
