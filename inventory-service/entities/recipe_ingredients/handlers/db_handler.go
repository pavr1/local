package handlers

import (
	"database/sql"
	"fmt"

	"inventory-service/entities/recipe_ingredients/models"
	recipeIngredientSQL "inventory-service/entities/recipe_ingredients/sql"
)

type RecipeIngredientDBHandler struct {
	db *sql.DB
}

func NewRecipeIngredientDBHandler(db *sql.DB) *RecipeIngredientDBHandler {
	return &RecipeIngredientDBHandler{db: db}
}

func (h *RecipeIngredientDBHandler) Create(req models.CreateRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	var recipeIngredient models.RecipeIngredient
	err := h.db.QueryRow(
		recipeIngredientSQL.CreateRecipeIngredientQuery,
		req.RecipeID,
		req.IngredientID,
		req.Quantity,
		req.UnitType,
	).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.UnitType,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create recipe ingredient: %w", err)
	}

	return &recipeIngredient, nil
}

func (h *RecipeIngredientDBHandler) GetByID(req models.GetRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	var recipeIngredient models.RecipeIngredient
	err := h.db.QueryRow(recipeIngredientSQL.GetRecipeIngredientByIDQuery, req.ID).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.UnitType,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe ingredient not found")
		}
		return nil, fmt.Errorf("failed to get recipe ingredient: %w", err)
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
		return nil, fmt.Errorf("failed to list recipe ingredients: %w", err)
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
			&recipeIngredient.UnitType,
			&recipeIngredient.CreatedAt,
			&recipeIngredient.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe ingredient: %w", err)
		}
		recipeIngredients = append(recipeIngredients, recipeIngredient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipe ingredients: %w", err)
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
		req.UnitType,
	).Scan(
		&recipeIngredient.ID,
		&recipeIngredient.RecipeID,
		&recipeIngredient.IngredientID,
		&recipeIngredient.Quantity,
		&recipeIngredient.UnitType,
		&recipeIngredient.CreatedAt,
		&recipeIngredient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe ingredient not found")
		}
		return nil, fmt.Errorf("failed to update recipe ingredient: %w", err)
	}

	return &recipeIngredient, nil
}

func (h *RecipeIngredientDBHandler) Delete(req models.DeleteRecipeIngredientRequest) error {
	result, err := h.db.Exec(recipeIngredientSQL.DeleteRecipeIngredientQuery, req.ID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe ingredient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recipe ingredient not found")
	}

	return nil
}
