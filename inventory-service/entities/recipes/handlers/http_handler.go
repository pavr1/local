package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"inventory-service/entities/recipes/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RecipeHTTPHandler struct {
	dbHandler *RecipeDBHandler
	logger    *logrus.Logger
}

func NewRecipeHTTPHandler(db *sql.DB, logger *logrus.Logger) *RecipeHTTPHandler {
	return &RecipeHTTPHandler{
		dbHandler: NewRecipeDBHandler(db),
		logger:    logger,
	}
}

// CreateRecipe handles POST /recipes
func (h *RecipeHTTPHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create recipe request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	recipe, err := h.dbHandler.Create(req)
	if err != nil {
		response := models.RecipeResponse{
			Success: false,
			Data:    models.Recipe{},
			Message: "Failed to create recipe: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeResponse{
		Success: true,
		Data:    *recipe,
		Message: "Recipe created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetRecipe handles GET /recipes/{id}
func (h *RecipeHTTPHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ID in get request")
		h.writeErrorResponse(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	req := models.GetRecipeRequest{ID: id}
	recipe, err := h.dbHandler.GetByID(req)
	if err != nil {
		if err.Error() == "recipe not found" {
			response := models.RecipeResponse{
				Success: false,
				Data:    models.Recipe{},
				Message: "Recipe not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeResponse{
			Success: false,
			Data:    models.Recipe{},
			Message: "Failed to get recipe: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeResponse{
		Success: true,
		Data:    *recipe,
		Message: "Recipe retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListRecipes handles GET /recipes
func (h *RecipeHTTPHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
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

	recipes, err := h.dbHandler.List(req)
	if err != nil {
		response := models.RecipesResponse{
			Success: false,
			Data:    []models.Recipe{},
			Message: "Failed to list recipes: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipesResponse{
		Success: true,
		Data:    recipes,
		Total:   len(recipes),
		Message: "Recipes retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateRecipe handles PUT /recipes/{id}
func (h *RecipeHTTPHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ID in update request")
		h.writeErrorResponse(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update recipe request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	recipe, err := h.dbHandler.Update(req, id)
	if err != nil {
		if err.Error() == "recipe not found" {
			response := models.RecipeResponse{
				Success: false,
				Data:    models.Recipe{},
				Message: "Recipe not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeResponse{
			Success: false,
			Data:    models.Recipe{},
			Message: "Failed to update recipe: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeResponse{
		Success: true,
		Data:    *recipe,
		Message: "Recipe updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteRecipe handles DELETE /recipes/{id}
func (h *RecipeHTTPHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ID in delete request")
		h.writeErrorResponse(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	req := models.DeleteRecipeRequest{ID: id}
	err := h.dbHandler.Delete(req)
	if err != nil {
		if err.Error() == "recipe not found" {
			response := models.GenericResponse{
				Success: false,
				Message: "Recipe not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.GenericResponse{
			Success: false,
			Message: "Failed to delete recipe: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Recipe deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *RecipeHTTPHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (h *RecipeHTTPHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, errorResponse, statusCode)
}
