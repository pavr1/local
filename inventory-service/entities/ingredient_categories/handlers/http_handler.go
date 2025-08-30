package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/ingredient_categories/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateIngredientCategory(req models.CreateIngredientCategoryRequest, logger *logrus.Logger) (*models.IngredientCategory, error)
	GetIngredientCategoryByID(id string, logger *logrus.Logger) (*models.IngredientCategory, error)
	ListIngredientCategories(logger *logrus.Logger) ([]models.IngredientCategory, error)
	UpdateIngredientCategory(id string, req models.UpdateIngredientCategoryRequest, logger *logrus.Logger) (*models.IngredientCategory, error)
	DeleteIngredientCategory(id string, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for ingredient category operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
	}
}

// CreateIngredientCategory handles POST /ingredient-categories
func (h *HttpHandler) CreateIngredientCategory(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	var req models.CreateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create ingredient category request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateIngredientCategoryRequest(req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.Name,
		}).Error("Validation failed for create ingredient category request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	category, err := h.dbHandler.CreateIngredientCategory(req, logger.Logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.IngredientCategory{},
			Message: "Failed to create ingredient category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)

		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *category,
		Message: "Ingredient category created successfully",
	}

	logger.WithFields(logrus.Fields{
		"category_id":   category.ID,
		"category_name": category.Name,
	}).Info("Ingredient category created successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetIngredientCategory handles GET /ingredient-categories/{id}
func (h *HttpHandler) GetIngredientCategory(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient category ID is required",
		})
		return
	}

	// Validate ingredient category ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid ingredient category ID format in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient category ID must be a valid UUID",
		})
		return
	}

	category, err := h.dbHandler.GetIngredientCategoryByID(id, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.IngredientCategory{},
			Message: "Failed to get ingredient category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *category,
		Message: "Ingredient category retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListIngredientCategories handles GET /ingredient-categories
func (h *HttpHandler) ListIngredientCategories(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	categories, err := h.dbHandler.ListIngredientCategories(logger.Logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.IngredientCategory{},
			Message: "Failed to list ingredient categories: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    categories,
		Message: "Ingredient categories retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateIngredientCategory handles PUT /ingredient-categories/{id}
func (h *HttpHandler) UpdateIngredientCategory(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in update request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient category ID is required",
		})
		return
	}

	var req models.UpdateIngredientCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update ingredient category request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateIngredientCategoryRequest(req, id); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"category_id": id,
		}).Error("Validation failed for update ingredient category request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	category, err := h.dbHandler.UpdateIngredientCategory(id, req, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.IngredientCategory{},
			Message: "Failed to update ingredient category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *category,
		Message: "Ingredient category updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"category_id":   category.ID,
		"category_name": category.Name,
	}).Info("Ingredient category updated successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteIngredientCategory handles DELETE /ingredient-categories/{id}
func (h *HttpHandler) DeleteIngredientCategory(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing ingredient category ID in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient category ID is required",
		})
		return
	}

	// Validate ingredient category ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"category_id": id,
		}).Warn("Invalid ingredient category ID format in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient category ID must be a valid UUID",
		})
		return
	}

	err := h.dbHandler.DeleteIngredientCategory(id, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.IngredientCategory{},
				Message: "Ingredient category not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete ingredient category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    models.IngredientCategory{},
		Message: "Ingredient category deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"category_id": id,
	}).Info("Ingredient category deleted successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
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
