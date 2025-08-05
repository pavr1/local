package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"inventory-service/entities/ingredient_categories/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateIngredientCategory(req models.CreateIngredientCategoryRequest) (*models.IngredientCategory, error)
	GetIngredientCategoryByID(id string) (*models.IngredientCategory, error)
	ListIngredientCategories() ([]models.IngredientCategory, error)
	UpdateIngredientCategory(id string, req models.UpdateIngredientCategoryRequest) (*models.IngredientCategory, error)
	DeleteIngredientCategory(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for ingredient category operations
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

// CreateIngredientCategory handles POST /ingredient-categories
func (h *HttpHandler) CreateIngredientCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create ingredient category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.CreateIngredientCategory(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.IngredientCategoryResponse{
			Success: false,
			Data:    models.IngredientCategory{},
			Message: "Failed to create ingredient category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoryResponse{
		Success: true,
		Data:    *category,
		Message: "Ingredient category created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetIngredientCategory handles GET /ingredient-categories/{id}
func (h *HttpHandler) GetIngredientCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient category ID in get request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.GetIngredientCategoryByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientCategoryResponse{
				Success: false,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientCategoryResponse{
			Success: false,
			Data:    models.IngredientCategory{},
			Message: "Failed to get ingredient category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoryResponse{
		Success: true,
		Data:    *category,
		Message: "Ingredient category retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListIngredientCategories handles GET /ingredient-categories
func (h *HttpHandler) ListIngredientCategories(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	categories, err := h.dbHandler.ListIngredientCategories()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.IngredientCategoriesListResponse{
			Success: false,
			Data:    []models.IngredientCategory{},
			Count:   0,
			Message: "Failed to list ingredient categories: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoriesListResponse{
		Success: true,
		Data:    categories,
		Count:   len(categories),
		Message: "Ingredient categories retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateIngredientCategory handles PUT /ingredient-categories/{id}
func (h *HttpHandler) UpdateIngredientCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient category ID in update request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update ingredient category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.UpdateIngredientCategory(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientCategoryResponse{
				Success: false,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientCategoryResponse{
			Success: false,
			Data:    models.IngredientCategory{},
			Message: "Failed to update ingredient category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoryResponse{
		Success: true,
		Data:    *category,
		Message: "Ingredient category updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteIngredientCategory handles DELETE /ingredient-categories/{id}
func (h *HttpHandler) DeleteIngredientCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient category ID in delete request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteIngredientCategory(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.IngredientCategoryDeleteResponse{
				Success: false,
				Message: "Ingredient category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.IngredientCategoryDeleteResponse{
			Success: false,
			Message: "Failed to delete ingredient category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoryDeleteResponse{
		Success: true,
		Message: "Ingredient category deleted successfully",
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
