package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"inventory-service/entities/recipe_categories/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RecipeCategoryHTTPHandler struct {
	dbHandler *RecipeCategoryDBHandler
	logger    *logrus.Logger
}

func NewRecipeCategoryHTTPHandler(db *sql.DB, logger *logrus.Logger) *RecipeCategoryHTTPHandler {
	return &RecipeCategoryHTTPHandler{
		dbHandler: NewRecipeCategoryDBHandler(db),
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
