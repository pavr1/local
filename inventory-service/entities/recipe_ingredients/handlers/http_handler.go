package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"inventory-service/entities/recipe_ingredients/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RecipeIngredientHTTPHandler struct {
	dbHandler *RecipeIngredientDBHandler
	logger    *logrus.Logger
}

func NewRecipeIngredientHTTPHandler(db *sql.DB, logger *logrus.Logger) *RecipeIngredientHTTPHandler {
	return &RecipeIngredientHTTPHandler{
		dbHandler: NewRecipeIngredientDBHandler(db),
		logger:    logger,
	}
}

// CreateRecipeIngredient handles POST /recipe-ingredients
func (h *RecipeIngredientHTTPHandler) CreateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create recipe ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	recipeIngredient, err := h.dbHandler.Create(req)
	if err != nil {
		response := models.RecipeIngredientResponse{
			Success: false,
			Data:    models.RecipeIngredient{},
			Message: "Failed to create recipe ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeIngredientResponse{
		Success: true,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetRecipeIngredient handles GET /recipe-ingredients/{id}
func (h *RecipeIngredientHTTPHandler) GetRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ingredient ID in get request")
		h.writeErrorResponse(w, "Recipe ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.GetRecipeIngredientRequest{ID: id}
	recipeIngredient, err := h.dbHandler.GetByID(req)
	if err != nil {
		if err.Error() == "recipe ingredient not found" {
			response := models.RecipeIngredientResponse{
				Success: false,
				Data:    models.RecipeIngredient{},
				Message: "Recipe ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeIngredientResponse{
			Success: false,
			Data:    models.RecipeIngredient{},
			Message: "Failed to get recipe ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeIngredientResponse{
		Success: true,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListRecipeIngredients handles GET /recipe-ingredients
func (h *RecipeIngredientHTTPHandler) ListRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	req := models.ListRecipeIngredientsRequest{}

	// Parse query parameters
	if recipeID := r.URL.Query().Get("recipe_id"); recipeID != "" {
		req.RecipeID = &recipeID
	}

	if ingredientID := r.URL.Query().Get("ingredient_id"); ingredientID != "" {
		req.IngredientID = &ingredientID
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

	recipeIngredients, err := h.dbHandler.List(req)
	if err != nil {
		response := models.RecipeIngredientsResponse{
			Success: false,
			Data:    []models.RecipeIngredient{},
			Message: "Failed to list recipe ingredients: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeIngredientsResponse{
		Success: true,
		Data:    recipeIngredients,
		Total:   len(recipeIngredients),
		Message: "Recipe ingredients retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateRecipeIngredient handles PUT /recipe-ingredients/{id}
func (h *RecipeIngredientHTTPHandler) UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ingredient ID in update request")
		h.writeErrorResponse(w, "Recipe ingredient ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update recipe ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	recipeIngredient, err := h.dbHandler.Update(req, id)
	if err != nil {
		if err.Error() == "recipe ingredient not found" {
			response := models.RecipeIngredientResponse{
				Success: false,
				Data:    models.RecipeIngredient{},
				Message: "Recipe ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RecipeIngredientResponse{
			Success: false,
			Data:    models.RecipeIngredient{},
			Message: "Failed to update recipe ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RecipeIngredientResponse{
		Success: true,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteRecipeIngredient handles DELETE /recipe-ingredients/{id}
func (h *RecipeIngredientHTTPHandler) DeleteRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing recipe ingredient ID in delete request")
		h.writeErrorResponse(w, "Recipe ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.DeleteRecipeIngredientRequest{ID: id}
	err := h.dbHandler.Delete(req)
	if err != nil {
		if err.Error() == "recipe ingredient not found" {
			response := models.GenericResponse{
				Success: false,
				Message: "Recipe ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.GenericResponse{
			Success: false,
			Message: "Failed to delete recipe ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Recipe ingredient deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *RecipeIngredientHTTPHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (h *RecipeIngredientHTTPHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, errorResponse, statusCode)
}
