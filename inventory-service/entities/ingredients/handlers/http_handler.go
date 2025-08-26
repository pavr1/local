package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/ingredients/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error)
	GetIngredientByID(id string) (*models.Ingredient, error)
	ListIngredients() ([]models.Ingredient, error)
	UpdateIngredient(id string, req models.UpdateIngredientRequest) (*models.Ingredient, error)
	DeleteIngredient(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for ingredient operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// NewHttpHandlerWithInterface creates a new HTTP handler with interface (for testing)
func NewHttpHandlerWithInterface(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateIngredient handles POST /ingredients
func (h *HttpHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateIngredientRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_name": req.Name,
		}).Error("Validation failed for create ingredient request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	ingredient, err := h.dbHandler.CreateIngredient(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.IngredientResponse{
			Success: false,
			Data:    models.Ingredient{},
			Message: "Failed to create ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientResponse{
		Success: true,
		Data:    *ingredient,
		Message: "Ingredient created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetIngredient handles GET /ingredients/{id}
func (h *HttpHandler) GetIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in get request")
		h.writeErrorResponse(w, "Ingredient ID is required", http.StatusBadRequest)
		return
	}

	// Validate ingredient ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Warn("Invalid ingredient ID format in get request")
		h.writeErrorResponse(w, "Ingredient ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	ingredient, err := h.dbHandler.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientResponse{
				Success: false,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientResponse{
			Success: false,
			Data:    models.Ingredient{},
			Message: "Failed to get ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientResponse{
		Success: true,
		Data:    *ingredient,
		Message: "Ingredient retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListIngredients handles GET /ingredients
func (h *HttpHandler) ListIngredients(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	ingredients, err := h.dbHandler.ListIngredients()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.IngredientsListResponse{
			Success: false,
			Data:    []models.Ingredient{},
			Count:   0,
			Message: "Failed to list ingredients: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientsListResponse{
		Success: true,
		Data:    ingredients,
		Count:   len(ingredients),
		Message: "Ingredients retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateIngredient handles PUT /ingredients/{id}
func (h *HttpHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in update request")
		h.writeErrorResponse(w, "Ingredient ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateIngredientRequest(req, id); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Validation failed for update ingredient request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	ingredient, err := h.dbHandler.UpdateIngredient(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientResponse{
				Success: false,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientResponse{
			Success: false,
			Data:    models.Ingredient{},
			Message: "Failed to update ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientResponse{
		Success: true,
		Data:    *ingredient,
		Message: "Ingredient updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteIngredient handles DELETE /ingredients/{id}
func (h *HttpHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in delete request")
		h.writeErrorResponse(w, "Ingredient ID is required", http.StatusBadRequest)
		return
	}

	// Validate ingredient ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Warn("Invalid ingredient ID format in delete request")
		h.writeErrorResponse(w, "Ingredient ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteIngredient(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientDeleteResponse{
				Success: false,
				Message: "Ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientDeleteResponse{
			Success: false,
			Message: "Failed to delete ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientDeleteResponse{
		Success: true,
		Message: "Ingredient deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response using the ErrorResponse model
func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := models.ErrorResponse{
		Success: false,
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	h.writeJSONResponse(w, errorResponse, statusCode)
}

// validateCreateIngredientRequest validates all required fields for ingredient creation
func (h *HttpHandler) validateCreateIngredientRequest(req models.CreateIngredientRequest) error {
	// Validate name (required, non-empty, max 255 chars)
	if req.Name == "" {
		return fmt.Errorf("name is required and cannot be empty")
	}
	//pvillalobos - hardcoded values
	if len(req.Name) > 255 {
		return fmt.Errorf("name cannot exceed 255 characters, got: %d", len(req.Name))
	}

	// Validate ingredient_category_id if provided (valid UUID)
	if req.IngredientCategoryID != nil {
		if *req.IngredientCategoryID == "" {
			return fmt.Errorf("ingredient_category_id cannot be empty if provided")
		}
		if !isValidUUID(*req.IngredientCategoryID) {
			return fmt.Errorf("ingredient_category_id must be a valid UUID, got: %s", *req.IngredientCategoryID)
		}
	}

	return nil
}

// validateUpdateIngredientRequest validates all required fields for ingredient update
func (h *HttpHandler) validateUpdateIngredientRequest(req models.UpdateIngredientRequest, id string) error {
	// Validate ingredient ID (required, valid UUID)
	if id == "" {
		return fmt.Errorf("ingredient ID is required and cannot be empty")
	}
	if !isValidUUID(id) {
		return fmt.Errorf("ingredient ID must be a valid UUID, got: %s", id)
	}

	// Validate name if provided (non-empty, max 255 chars)
	if req.Name != nil {
		if *req.Name == "" {
			return fmt.Errorf("name cannot be empty if provided")
		}
		if len(*req.Name) > 255 {
			return fmt.Errorf("name cannot exceed 255 characters, got: %d", len(*req.Name))
		}
	}

	// Validate ingredient_category_id if provided (valid UUID)
	if req.IngredientCategoryID != nil {
		if *req.IngredientCategoryID == "" {
			return fmt.Errorf("ingredient_category_id cannot be empty if provided")
		}
		if !isValidUUID(*req.IngredientCategoryID) {
			return fmt.Errorf("ingredient_category_id must be a valid UUID, got: %s", *req.IngredientCategoryID)
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
