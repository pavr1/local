package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"inventory-service/config"
	recipeIngredientsHandler "inventory-service/entities/recipe_ingredients/handlers"
	recipeIngredientsModels "inventory-service/entities/recipe_ingredients/models"
	recipeIngredientsSQL "inventory-service/entities/recipe_ingredients/sql"
	"inventory-service/entities/recipes/models"
	recipeSQL "inventory-service/entities/recipes/sql"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RecipeDBHandler handles database operations for recipes
type RecipeDBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
	config *config.Config
}

// NewRecipeDBHandler creates a new database handler for recipes
func NewRecipeDBHandler(db *sql.DB, logger *logrus.Logger, cfg *config.Config) *RecipeDBHandler {
	return &RecipeDBHandler{
		db:     db,
		logger: logger,
		config: cfg,
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

	// Validate that image data is provided
	if req.ImageData == nil || req.ImageName == nil || len(req.ImageData) == 0 {
		h.logger.WithFields(logrus.Fields{
			"recipe_name": req.RecipeName,
		}).Error("Recipe creation failed: image is required")
		return nil, fmt.Errorf("image is required for recipe creation")
	}

	req.RecipeName = strings.ReplaceAll(req.RecipeName, " ", "_")
	req.RecipeName += ".jpg"

	// Store the image in data service (image is required)
	imageURL, err := h.storeImageInDataService("recipes", req.RecipeName, req.ImageData)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_name": req.RecipeName,
			"image_name":  *req.ImageName,
		}).Error("Failed to store recipe image in data service")
		return nil, err
	}
	pictureURL := &imageURL

	var recipe models.Recipe
	err = tx.QueryRow(
		recipeSQL.CreateRecipeQuery,
		req.RecipeName,
		req.RecipeDescription,
		pictureURL,
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
		"has_image":         true,
	}).Info("Recipe and ingredients created successfully")

	return &recipe, nil
}

// storeImageInDataService stores an image in the data service and returns the image URL
func (h *RecipeDBHandler) storeImageInDataService(service, imageName string, imageData []byte) (string, error) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create form file
	part, err := writer.CreateFormFile("image", imageName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create form file")
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Write image data
	_, err = part.Write(imageData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to write image data")
		return "", fmt.Errorf("failed to write image data: %w", err)
	}

	// Close writer
	err = writer.Close()
	if err != nil {
		h.logger.WithError(err).Error("Failed to close writer")
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request
	// TODO: get gateway URL from config. We need to leave it localhost for images
	gatewayURL := "http://localhost:8082"

	url := fmt.Sprintf("%s/api/v1/data/images/%s", gatewayURL, service)

	h.logger.WithFields(logrus.Fields{
		"service":    service,
		"image_name": imageName,
		"image_data": len(imageData),
		"url":        url,
	}).Info("Image data written successfully, calling data service")

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create HTTP request")
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add gateway headers for internal service communication
	req.Header.Set("X-Gateway-Service", "gateway")
	req.Header.Set("X-Gateway-Session-Managed", "true")
	req.Header.Set("X-User-ID", "system")
	req.Header.Set("X-User-Role", "admin")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to make HTTP request")
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
		}).Error("Data service returned non-OK status")
		return "", fmt.Errorf("data service returned status %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		Success  bool   `json:"success"`
		Message  string `json:"message"`
		ImageURL string `json:"image_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		h.logger.WithError(err).Error("Failed to decode response")
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		h.logger.WithFields(logrus.Fields{
			"message": response.Message,
		}).Error("Data service returned non-success response")
		return "", fmt.Errorf("data service error: %s", response.Message)
	}

	return response.ImageURL, nil
}

// deleteImageFromDataService deletes an image from the data service
func (h *RecipeDBHandler) deleteImageFromDataService(service, filename string) error {
	// Construct the delete URL
	deleteURL := fmt.Sprintf("%s/api/v1/data/images/%s/%s", h.config.GatewayURL, service, filename)

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create DELETE request
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("data service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
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

	// Load ingredients for the recipe using the recipe ingredients handler
	riHandler := recipeIngredientsHandler.NewDBHandler(h.db, h.logger)
	recipeID := recipe.ID
	riReq := recipeIngredientsModels.ListRecipeIngredientsRequest{
		RecipeID: &recipeID,
	}

	recipeIngredientsList, err := riHandler.List(riReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": recipe.ID,
		}).Warn("Failed to load recipe ingredients, continuing without ingredients")
		// Don't fail the entire request if ingredients fail to load
		recipe.Ingredients = []models.RecipeIngredient{}
	} else {
		// Convert recipe ingredients to the format expected by the recipe model
		ingredients := make([]models.RecipeIngredient, len(recipeIngredientsList))
		for i, ri := range recipeIngredientsList {
			ingredients[i] = models.RecipeIngredient{
				IngredientID:   ri.IngredientID,
				NumberOfUnits:  ri.Quantity,
				IngredientName: ri.IngredientName,
				FinalPrice:     ri.FinalPrice,
			}
		}
		recipe.Ingredients = ingredients
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

	// Load ingredients for each recipe
	riHandler := recipeIngredientsHandler.NewDBHandler(h.db, h.logger)
	for i := range recipes {
		recipeID := recipes[i].ID
		riReq := recipeIngredientsModels.ListRecipeIngredientsRequest{
			RecipeID: &recipeID,
		}

		recipeIngredientsList, err := riHandler.List(riReq)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"recipe_id": recipes[i].ID,
			}).Warn("Failed to load recipe ingredients, continuing without ingredients")
			// Don't fail the entire request if ingredients fail to load
			recipes[i].Ingredients = []models.RecipeIngredient{}
		} else {
			// Convert recipe ingredients to the format expected by the recipe model
			ingredients := make([]models.RecipeIngredient, len(recipeIngredientsList))
			for j, ri := range recipeIngredientsList {
				ingredients[j] = models.RecipeIngredient{
					IngredientID:   ri.IngredientID,
					NumberOfUnits:  ri.Quantity,
					IngredientName: ri.IngredientName,
					FinalPrice:     ri.FinalPrice,
				}
			}
			recipes[i].Ingredients = ingredients
		}
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
	// Start a transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback() // Rollback if not committed

	// Handle image storage if image data is provided
	var pictureURL *string
	if req.ImageData != nil && req.ImageName != nil && len(req.ImageData) > 0 {
		h.logger.WithFields(logrus.Fields{
			"recipe_id":  id,
			"image_name": *req.ImageName,
			"image_data": len(req.ImageData),
		}).Info("Storing recipe image in data service")

		// Store the image in data service
		imageURL, err := h.storeImageInDataService("recipes", *req.ImageName, req.ImageData)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"recipe_id":  id,
				"image_name": *req.ImageName,
			}).Error("Failed to store recipe image in data service")
			return nil, err
		}
		pictureURL = &imageURL
	} else {
		// Use the provided PictureURL if no image data
		pictureURL = req.PictureURL
	}

	var recipe models.Recipe
	err = tx.QueryRow(
		recipeSQL.UpdateRecipeQuery,
		id,
		req.RecipeName,
		req.RecipeDescription,
		pictureURL,
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

	// Update ingredients if provided
	if req.Ingredients != nil {
		// Delete existing ingredients
		_, err = tx.Exec(recipeIngredientsSQL.DeleteRecipeIngredientsByRecipeIDQuery, id)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"recipe_id": id,
			}).Error("Failed to delete existing recipe ingredients")
			return nil, err
		}

		// Create new ingredients
		for _, ingredient := range req.Ingredients {
			_, err = tx.Exec(
				recipeIngredientsSQL.CreateRecipeIngredientQuery,
				id,
				ingredient.IngredientID,
				ingredient.NumberOfUnits,
			)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"recipe_id":     id,
					"ingredient_id": ingredient.IngredientID,
				}).Error("Failed to create recipe ingredient in database")
				return nil, err
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_id":   recipe.ID,
		"recipe_name": recipe.RecipeName,
	}).Info("Recipe updated successfully")

	return &recipe, nil
}

func (h *RecipeDBHandler) Delete(req models.DeleteRecipeRequest) error {
	// First, get the recipe details to extract the image filename
	recipe, err := h.GetByID(models.GetRecipeRequest{ID: req.ID})
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"recipe_id": req.ID,
			}).Warn("No recipe found to delete")
			return sql.ErrNoRows
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": req.ID,
		}).Error("Failed to get recipe details before deletion")
		return err
	}

	// Delete the recipe from the database
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

	// Delete the associated image file if it exists
	if recipe.PictureURL != nil && *recipe.PictureURL != "" {
		// Extract filename from the URL
		// URL format: http://localhost:8082/api/v1/data/images/recipes/filename.jpg
		imageURL := *recipe.PictureURL
		parts := strings.Split(imageURL, "/")
		if len(parts) >= 2 {
			filename := parts[len(parts)-1] // Get the last part as filename

			// Delete the image from data service
			err = h.deleteImageFromDataService("recipes", filename)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"recipe_id": req.ID,
					"filename":  filename,
				}).Warn("Failed to delete recipe image from data service, but recipe was deleted from database")
				// Don't fail the entire operation if image deletion fails
			} else {
				h.logger.WithFields(logrus.Fields{
					"recipe_id": req.ID,
					"filename":  filename,
				}).Info("Recipe image deleted from data service")
			}
		}
	}

	h.logger.WithFields(logrus.Fields{
		"recipe_id": req.ID,
	}).Info("Recipe deleted successfully")

	return nil
}
