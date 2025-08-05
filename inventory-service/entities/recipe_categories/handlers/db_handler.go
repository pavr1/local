package handlers

import (
	"database/sql"
	"fmt"

	"inventory-service/entities/recipe_categories/models"
	recipeSQL "inventory-service/entities/recipe_categories/sql"
)

type RecipeCategoryDBHandler struct {
	db *sql.DB
}

func NewRecipeCategoryDBHandler(db *sql.DB) *RecipeCategoryDBHandler {
	return &RecipeCategoryDBHandler{db: db}
}

func (h *RecipeCategoryDBHandler) Create(req models.CreateRecipeCategoryRequest) (*models.RecipeCategory, error) {
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
		return nil, fmt.Errorf("failed to create recipe category: %w", err)
	}

	return &recipeCategory, nil
}

func (h *RecipeCategoryDBHandler) GetByID(req models.GetRecipeCategoryRequest) (*models.RecipeCategory, error) {
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
			return nil, fmt.Errorf("recipe category not found")
		}
		return nil, fmt.Errorf("failed to get recipe category: %w", err)
	}

	return &recipeCategory, nil
}

func (h *RecipeCategoryDBHandler) List(req models.ListRecipeCategoriesRequest) ([]models.RecipeCategory, error) {
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
		return nil, fmt.Errorf("failed to list recipe categories: %w", err)
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
			return nil, fmt.Errorf("failed to scan recipe category: %w", err)
		}
		recipeCategories = append(recipeCategories, recipeCategory)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipe categories: %w", err)
	}

	return recipeCategories, nil
}

func (h *RecipeCategoryDBHandler) Update(req models.UpdateRecipeCategoryRequest, id string) (*models.RecipeCategory, error) {
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
			return nil, fmt.Errorf("recipe category not found")
		}
		return nil, fmt.Errorf("failed to update recipe category: %w", err)
	}

	return &recipeCategory, nil
}

func (h *RecipeCategoryDBHandler) Delete(req models.DeleteRecipeCategoryRequest) error {
	result, err := h.db.Exec(recipeSQL.DeleteRecipeCategoryQuery, req.ID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recipe category not found")
	}

	return nil
}
