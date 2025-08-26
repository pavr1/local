package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"inventory-service/entities/recipe_categories/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RecipeCategoryHTTPHandler struct {
	dbHandler *RecipeCategoryDBHandler
	logger    *logrus.Logger
}

// NewRecipeCategoryDBHandler creates a new database handler for recipe categories
func NewRecipeCategoryDBHandler(db *sql.DB, logger *logrus.Logger) *RecipeCategoryDBHandler {
	return &RecipeCategoryDBHandler{
		db:     db,
		logger: logger,
	}
}

func NewRecipeCategoryHTTPHandler(db *sql.DB, logger *logrus.Logger) *RecipeCategoryHTTPHandler {
	return &RecipeCategoryHTTPHandler{
		dbHandler: NewRecipeCategoryDBHandler(db, logger),
		logger:    logger,
	}
}

// CreateRecipeCategory handles POST /recipe-categories
func (h *RecipeCategoryHTTPHandler) CreateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create recipe category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateRecipeCategoryRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Validation failed for create recipe category request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	recipeCategory, err := h.dbHandler.Create(req)
	if err != nil {
		response := models.RecipeCategoryResponse{
			Success: false,
			Data:    models.RecipeCategory{},
			Message: "Failed to create recipe category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeCategoryResponse{
		Success: true,
		Data:    *recipeCategory,
		Message: "Recipe category created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetRecipeCategory handles GET /recipe-categories/{id}
func (h *RecipeCategoryHTTPHandler) GetRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in get request")
		h.writeErrorResponse(w, "Recipe category ID is required", http.StatusBadRequest)
		return
	}

	// Validate recipe category ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid recipe category ID format in get request")
		h.writeErrorResponse(w, "Recipe category ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	req := models.GetRecipeCategoryRequest{ID: id}
	recipeCategory, err := h.dbHandler.GetByID(req)
	if err != nil {
		if err.Error() == "recipe category not found" {
			response := models.RecipeCategoryResponse{
				Success: false,
				Data:    models.RecipeCategory{},
				Message: "Recipe category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeCategoryResponse{
			Success: false,
			Data:    models.RecipeCategory{},
			Message: "Failed to get recipe category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeCategoryResponse{
		Success: true,
		Data:    *recipeCategory,
		Message: "Recipe category retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListRecipeCategories handles GET /recipe-categories
func (h *RecipeCategoryHTTPHandler) ListRecipeCategories(w http.ResponseWriter, r *http.Request) {
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

	recipeCategories, err := h.dbHandler.List(req)
	if err != nil {
		response := models.RecipeCategoriesResponse{
			Success: false,
			Data:    []models.RecipeCategory{},
			Message: "Failed to list recipe categories: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeCategoriesResponse{
		Success: true,
		Data:    recipeCategories,
		Total:   len(recipeCategories),
		Message: "Recipe categories retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateRecipeCategory handles PUT /recipe-categories/{id}
func (h *RecipeCategoryHTTPHandler) UpdateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in update request")
		h.writeErrorResponse(w, "Recipe category ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update recipe category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateRecipeCategoryRequest(req, id); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Validation failed for update recipe category request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	recipeCategory, err := h.dbHandler.Update(req, id)
	if err != nil {
		if err.Error() == "recipe category not found" {
			response := models.RecipeCategoryResponse{
				Success: false,
				Data:    models.RecipeCategory{},
				Message: "Recipe category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeCategoryResponse{
			Success: false,
			Data:    models.RecipeCategory{},
			Message: "Failed to update recipe category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeCategoryResponse{
		Success: true,
		Data:    *recipeCategory,
		Message: "Recipe category updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteRecipeCategory handles DELETE /recipe-categories/{id}
func (h *RecipeCategoryHTTPHandler) DeleteRecipeCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe category ID in delete request")
		h.writeErrorResponse(w, "Recipe category ID is required", http.StatusBadRequest)
		return
	}

	// Validate recipe category ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid recipe category ID format in delete request")
		h.writeErrorResponse(w, "Recipe category ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	req := models.DeleteRecipeCategoryRequest{ID: id}
	err := h.dbHandler.Delete(req)
	if err != nil {
		if err.Error() == "recipe category not found" {
			response := models.GenericResponse{
				Success: false,
				Message: "Recipe category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.GenericResponse{
			Success: false,
			Message: "Failed to delete recipe category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Recipe category deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *RecipeCategoryHTTPHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (h *RecipeCategoryHTTPHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, errorResponse, statusCode)
}

// validateCreateRecipeCategoryRequest validates all required fields for recipe category creation
func (h *RecipeCategoryHTTPHandler) validateCreateRecipeCategoryRequest(req models.CreateRecipeCategoryRequest) error {
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
func (h *RecipeCategoryHTTPHandler) validateUpdateRecipeCategoryRequest(req models.UpdateRecipeCategoryRequest, id string) error {
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
