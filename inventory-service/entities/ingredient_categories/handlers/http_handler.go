package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/ingredient_categories/models"
	"inventory-service/pkg/requestlogger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateIngredientCategory(req models.CreateIngredientCategoryRequest, logger *logrus.Entry) (*models.IngredientCategory, error)
	GetIngredientCategoryByID(id string, logger *logrus.Entry) (*models.IngredientCategory, error)
	ListIngredientCategories(logger *logrus.Entry) ([]models.IngredientCategory, error)
	UpdateIngredientCategory(id string, req models.UpdateIngredientCategoryRequest, logger *logrus.Entry) (*models.IngredientCategory, error)
	DeleteIngredientCategory(id string, logger *logrus.Entry) error
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
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	var req models.CreateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create ingredient category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateIngredientCategoryRequest(req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Validation failed for create ingredient category request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.CreateIngredientCategory(req, logger)
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
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in get request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	// Validate ingredient category ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid ingredient category ID format in get request")
		h.writeErrorResponse(w, "Ingredient category ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.GetIngredientCategoryByID(id, logger)
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
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	categories, err := h.dbHandler.ListIngredientCategories(logger)
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
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in update request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update ingredient category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateIngredientCategoryRequest(req, id); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Validation failed for update ingredient category request")
		h.writeErrorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.dbHandler.UpdateIngredientCategory(id, req, logger)
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
	// Get logger with request ID
	logger := requestlogger.GetRequestLogger(h.logger, r)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in delete request")
		h.writeErrorResponse(w, "Ingredient category ID is required", http.StatusBadRequest)
		return
	}

	// Validate ingredient category ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid ingredient category ID format in delete request")
		h.writeErrorResponse(w, "Ingredient category ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteIngredientCategory(id, logger)
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
		h.writeErrorResponse(w, "Failed to delete ingredient category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.IngredientCategoryResponse{
		Success: true,
		Data:    models.IngredientCategory{},
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

// validateCreateIngredientCategoryRequest validates all required fields for ingredient category creation
func (h *HttpHandler) validateCreateIngredientCategoryRequest(req models.CreateIngredientCategoryRequest) error {
	// Validate name (required, non-empty, max 100 chars)
	if req.Name == "" {
		return fmt.Errorf("name is required and cannot be empty")
	}
	if len(req.Name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters, got: %d", len(req.Name))
	}

	// Validate description (required, non-empty, max 1000 chars)
	if req.Description == "" {
		return fmt.Errorf("description is required and cannot be empty")
	}
	if len(req.Description) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters, got: %d", len(req.Description))
	}

	return nil
}

// validateUpdateIngredientCategoryRequest validates all required fields for ingredient category update
func (h *HttpHandler) validateUpdateIngredientCategoryRequest(req models.UpdateIngredientCategoryRequest, id string) error {
	// Validate ingredient category ID (required, valid UUID)
	if id == "" {
		return fmt.Errorf("ingredient category ID is required and cannot be empty")
	}
	if !isValidUUID(id) {
		return fmt.Errorf("ingredient category ID must be a valid UUID, got: %s", id)
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
