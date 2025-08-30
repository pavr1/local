package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"inventory-service/entities/recipes/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HttpHandler struct {
	dbHandler DBHandlerInterface
}

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	Create(req models.CreateRecipeRequest, logger *logrus.Logger) (*models.Recipe, error)
	GetByID(req models.GetRecipeRequest, logger *logrus.Logger) (*models.Recipe, error)
	List(req models.ListRecipesRequest, logger *logrus.Logger) ([]models.Recipe, error)
	Update(req models.UpdateRecipeRequest, id string, logger *logrus.Logger) (*models.Recipe, error)
	Delete(req models.DeleteRecipeRequest, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
	}
}

// CreateRecipe handles POST /recipes
func (h *HttpHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	var req models.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create recipe request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateRecipeRequest(req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"recipe_name": req.RecipeName,
		}).Error("Validation failed for create recipe request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	recipe, err := h.dbHandler.Create(req, logger.Logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Recipe{},
			Message: "Failed to create recipe: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *recipe,
		Message: "Recipe created successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_id":   recipe.ID,
		"recipe_name": recipe.RecipeName,
	}).Info("Recipe created successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetRecipe handles GET /recipes/{id}
func (h *HttpHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ID in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ID is required",
		})
		return
	}

	// Validate recipe ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"recipe_id": id,
		}).Warn("Invalid recipe ID format in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ID must be a valid UUID",
		})
		return
	}

	req := models.GetRecipeRequest{ID: id}
	recipe, err := h.dbHandler.GetByID(req, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Recipe{},
				Message: "Recipe not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Recipe{},
			Message: "Failed to get recipe: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipe,
		Message: "Recipe retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListRecipes handles GET /recipes
func (h *HttpHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	req := models.ListRecipesRequest{}

	// Parse query parameters
	if recipeName := r.URL.Query().Get("recipe_name"); recipeName != "" {
		req.RecipeName = &recipeName
	}

	if recipeCategoryID := r.URL.Query().Get("recipe_category_id"); recipeCategoryID != "" {
		req.RecipeCategoryID = &recipeCategoryID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	recipes, err := h.dbHandler.List(req, logger.Logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.Recipe{},
			Message: "Failed to list recipes: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    recipes,
		Message: "Recipes retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateRecipe handles PUT /recipes/{id}
func (h *HttpHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ID in update request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ID is required",
		})
		return
	}

	var req models.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update recipe request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateRecipeRequest(req, id); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": id,
		}).Error("Validation failed for update recipe request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	recipe, err := h.dbHandler.Update(req, id, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Recipe{},
				Message: "Recipe not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Recipe{},
			Message: "Failed to update recipe: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipe,
		Message: "Recipe updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_id":   recipe.ID,
		"recipe_name": recipe.RecipeName,
	}).Info("Recipe updated successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteRecipe handles DELETE /recipes/{id}
func (h *HttpHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ID in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ID is required",
		})
		return
	}

	// Validate recipe ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"recipe_id": id,
		}).Warn("Invalid recipe ID format in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ID must be a valid UUID",
		})
		return
	}

	req := models.DeleteRecipeRequest{ID: id}
	err := h.dbHandler.Delete(req, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Recipe not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete recipe: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Recipe deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_id": id,
	}).Info("Recipe deleted successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// validateCreateRecipeRequest validates all required fields for recipe creation
func (h *HttpHandler) validateCreateRecipeRequest(req models.CreateRecipeRequest) error {
	// Validate recipe_name (required, non-empty, max 255 chars)
	if req.RecipeName == "" {
		return fmt.Errorf("recipe_name is required and cannot be empty")
	}
	//pvillalobos - hardcoded values
	if len(req.RecipeName) > 255 {
		return fmt.Errorf("recipe_name cannot exceed 255 characters, got: %d", len(req.RecipeName))
	}

	// Validate recipe_category_id (required, valid UUID)
	if req.RecipeCategoryID == "" {
		return fmt.Errorf("recipe_category_id is required and cannot be empty")
	}
	if !isValidUUID(req.RecipeCategoryID) {
		return fmt.Errorf("recipe_category_id must be a valid UUID, got: %s", req.RecipeCategoryID)
	}

	// Validate total_recipe_cost (required, non-negative)
	if req.TotalRecipeCost < 0 {
		return fmt.Errorf("total_recipe_cost must be non-negative, got: %f", req.TotalRecipeCost)
	}

	// Validate ingredients (required, non-empty array)
	if len(req.Ingredients) == 0 {
		return fmt.Errorf("at least one ingredient is required")
	}

	// Validate each ingredient
	for i, ingredient := range req.Ingredients {
		if err := h.validateRecipeIngredient(ingredient, i); err != nil {
			return fmt.Errorf("ingredient %d validation failed: %w", i, err)
		}
	}

	// Validate image data (required for recipe creation)
	if req.ImageData == nil || req.ImageName == nil || len(req.ImageData) == 0 {
		return fmt.Errorf("image is required for recipe creation (image_data, image_name, and non-empty image_data)")
	}

	return nil
}

// validateRecipeIngredient validates a single recipe ingredient
func (h *HttpHandler) validateRecipeIngredient(ingredient models.RecipeIngredient, index int) error {
	// Validate ingredient_id (required, valid UUID)
	if ingredient.IngredientID == "" {
		return fmt.Errorf("ingredient_id is required and cannot be empty")
	}
	if !isValidUUID(ingredient.IngredientID) {
		return fmt.Errorf("ingredient_id must be a valid UUID, got: %s", ingredient.IngredientID)
	}

	// Validate quantity (required, greater than 0.001)
	if ingredient.Quantity <= 0.001 {
		return fmt.Errorf("quantity must be greater than 0.001, got: %f", ingredient.Quantity)
	}

	return nil
}

// validateUpdateRecipeRequest validates all required fields for recipe update
func (h *HttpHandler) validateUpdateRecipeRequest(req models.UpdateRecipeRequest, id string) error {
	// Validate recipe ID (required, valid UUID)
	if id == "" {
		return fmt.Errorf("recipe ID is required and cannot be empty")
	}
	if !isValidUUID(id) {
		return fmt.Errorf("recipe ID must be a valid UUID, got: %s", id)
	}

	// Validate recipe_name if provided (non-empty, max 255 chars)
	if req.RecipeName != nil {
		if *req.RecipeName == "" {
			return fmt.Errorf("recipe_name cannot be empty if provided")
		}
		if len(*req.RecipeName) > 255 {
			return fmt.Errorf("recipe_name cannot exceed 255 characters, got: %d", len(*req.RecipeName))
		}
	}

	// Validate recipe_category_id if provided (valid UUID)
	if req.RecipeCategoryID != nil {
		if *req.RecipeCategoryID == "" {
			return fmt.Errorf("recipe_category_id cannot be empty if provided")
		}
		if !isValidUUID(*req.RecipeCategoryID) {
			return fmt.Errorf("recipe_category_id must be a valid UUID, got: %s", *req.RecipeCategoryID)
		}
	}

	// Validate total_recipe_cost if provided (non-negative)
	if req.TotalRecipeCost != nil {
		if *req.TotalRecipeCost < 0 {
			return fmt.Errorf("total_recipe_cost must be non-negative, got: %f", *req.TotalRecipeCost)
		}
	}

	// Validate ingredients if provided (non-empty array)
	if req.Ingredients != nil {
		if len(req.Ingredients) == 0 {
			return fmt.Errorf("ingredients cannot be empty if provided")
		}

		// Validate each ingredient
		for i, ingredient := range req.Ingredients {
			if err := h.validateRecipeIngredient(ingredient, i); err != nil {
				return fmt.Errorf("ingredient %d validation failed: %w", i, err)
			}
		}
	}

	return nil
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(uuid string) bool {
	// Simple UUID validation - check length and format
	if len(uuid) != 36 {
		return false
	}

	// Check if it matches UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		return false
	}

	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}

	return true
}
