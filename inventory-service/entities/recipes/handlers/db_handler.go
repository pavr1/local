package handlers

import (
	"database/sql"
	"fmt"

	"inventory-service/entities/recipes/models"
	recipeSQL "inventory-service/entities/recipes/sql"
)

type RecipeDBHandler struct {
	db *sql.DB
}

func NewRecipeDBHandler(db *sql.DB) *RecipeDBHandler {
	return &RecipeDBHandler{db: db}
}

func (h *RecipeDBHandler) Create(req models.CreateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := h.db.QueryRow(
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
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

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
		&recipe.TotalRecipeCost,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to get recipe: %w", err)
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
		return nil, fmt.Errorf("failed to list recipes: %w", err)
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
			&recipe.TotalRecipeCost,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}
		recipes = append(recipes, recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipes: %w", err)
	}

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
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	return &recipe, nil
}

func (h *RecipeDBHandler) Delete(req models.DeleteRecipeRequest) error {
	result, err := h.db.Exec(recipeSQL.DeleteRecipeQuery, req.ID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recipe not found")
	}

	return nil
}
