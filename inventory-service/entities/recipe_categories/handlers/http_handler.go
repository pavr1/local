package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"inventory-service/entities/recipe_categories/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	Create(req models.CreateRecipeCategoryRequest, logger *logrus.Logger) (*models.RecipeCategory, error)
	GetByID(req models.GetRecipeCategoryRequest, logger *logrus.Logger) (*models.RecipeCategory, error)
	List(req models.ListRecipeCategoriesRequest, logger *logrus.Logger) ([]models.RecipeCategory, error)
	Update(req models.UpdateRecipeCategoryRequest, id string, logger *logrus.Logger) (*models.RecipeCategory, error)
	Delete(req models.DeleteRecipeCategoryRequest, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for ingredient operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateRecipeCategory handles POST /recipe-categories
func (h *HttpHandler) CreateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create recipe category request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateRecipeCategoryRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Validation failed for create recipe category request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	recipeCategory, err := h.dbHandler.Create(req, h.logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeCategory{},
			Message: "Failed to create recipe category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *recipeCategory,
		Message: "Recipe category created successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"category_id":   recipeCategory.ID,
		"category_name": recipeCategory.Name,
	}).Info("Recipe category created successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetRecipeCategory handles GET /recipe-categories/{id}
func (h *HttpHandler) GetRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe category ID is required",
		})
		return
	}

	// Validate recipe category ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid recipe category ID format in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe category ID must be a valid UUID",
		})
		return
	}

	req := models.GetRecipeCategoryRequest{ID: id}
	recipeCategory, err := h.dbHandler.GetByID(req, h.logger)
	if err != nil {
		//pvillalobos - use notfoundError instead
		if err.Error() == "recipe category not found" {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.RecipeCategory{},
				Message: "Recipe category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeCategory{},
			Message: "Failed to get recipe category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipeCategory,
		Message: "Recipe category retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListRecipeCategories handles GET /recipe-categories
func (h *HttpHandler) ListRecipeCategories(w http.ResponseWriter, r *http.Request) {
	req := models.ListRecipeCategoriesRequest{}

	// Parse query parameters
	if name := r.URL.Query().Get("name"); name != "" {
		req.Name = &name
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

	recipeCategories, err := h.dbHandler.List(req, h.logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.RecipeCategory{},
			Message: "Failed to list recipe categories: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    recipeCategories,
		Message: "Recipe categories retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateRecipeCategory handles PUT /recipe-categories/{id}
func (h *HttpHandler) UpdateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in update request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe category ID is required",
		})
		return
	}

	var req models.UpdateRecipeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update recipe category request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateRecipeCategoryRequest(req, id); err != nil {
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	recipeCategory, err := h.dbHandler.Update(req, id, h.logger)
	if err != nil {
		if err.Error() == "recipe category not found" {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.RecipeCategory{},
				Message: "Recipe category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeCategory{},
			Message: "Failed to update recipe category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipeCategory,
		Message: "Recipe category updated successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"category_id":   recipeCategory.ID,
		"category_name": recipeCategory.Name,
	}).Info("Recipe category updated successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteRecipeCategory handles DELETE /recipe-categories/{id}
func (h *HttpHandler) DeleteRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe category ID is required",
		})
		return
	}

	// Validate recipe category ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid recipe category ID format in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe category ID must be a valid UUID",
		})
		return
	}

	req := models.DeleteRecipeCategoryRequest{ID: id}
	err := h.dbHandler.Delete(req, h.logger)
	if err != nil {
		if err.Error() == "recipe category not found" {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Recipe category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete recipe category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Recipe category deleted successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"category_id": id,
	}).Info("Recipe category deleted successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// validateCreateRecipeCategoryRequest validates all required fields for recipe category creation
func (h *HttpHandler) validateCreateRecipeCategoryRequest(req models.CreateRecipeCategoryRequest) error {
	// Validate name (required, non-empty, max 100 chars)
	if req.Name == "" {
		return fmt.Errorf("name is required and cannot be empty")
	}
	//pvillalobos - hardcoded values
	if len(req.Name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters, got: %d", len(req.Name))
	}

	// Validate description (required, non-empty, max 1000 chars)
	if req.Description == nil {
		return fmt.Errorf("description is required and cannot be empty")
	}
	if *req.Description == "" {
		return fmt.Errorf("description is required and cannot be empty")
	}
	if len(*req.Description) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters, got: %d", len(*req.Description))
	}

	return nil
}

// validateUpdateRecipeCategoryRequest validates all required fields for recipe category update
func (h *HttpHandler) validateUpdateRecipeCategoryRequest(req models.UpdateRecipeCategoryRequest, id string) error {
	// Validate recipe category ID (required, valid UUID)
	if id == "" {
		return fmt.Errorf("recipe category ID is required and cannot be empty")
	}
	if !isValidUUID(id) {
		return fmt.Errorf("recipe category ID must be a valid UUID, got: %s", id)
	}

	// Validate name if provided (non-empty, max 100 chars)
	if req.Name != nil {
		if *req.Name == "" {
			return fmt.Errorf("name cannot be empty if provided")
		}
		if len(*req.Name) > 100 {
			return fmt.Errorf("name cannot exceed 100 characters, got: %d", len(*req.Name))
		}
	}

	// Validate description if provided (non-empty, max 1000 chars)
	if req.Description != nil {
		if *req.Description == "" {
			return fmt.Errorf("description cannot be empty if provided")
		}
		if len(*req.Description) > 1000 {
			return fmt.Errorf("description cannot exceed 1000 characters, got: %d", len(*req.Description))
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
