package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/ingredients/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateIngredient(req models.CreateIngredientRequest, logger *logrus.Logger) (*models.Ingredient, error)
	GetIngredientByID(id string, logger *logrus.Logger) (*models.Ingredient, error)
	ListIngredients(logger *logrus.Logger) ([]models.Ingredient, error)
	UpdateIngredient(id string, req models.UpdateIngredientRequest, logger *logrus.Logger) (*models.Ingredient, error)
	DeleteIngredient(id string, logger *logrus.Logger) error
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

// CreateIngredient handles POST /ingredients
func (h *HttpHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create ingredient request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateIngredientRequest(req); err != nil {
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed for create ingredient request",
		})
		return
	}

	ingredient, err := h.dbHandler.CreateIngredient(req, h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Ingredient{},
			Message: "Failed to create ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *ingredient,
		Message: "Ingredient created successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetIngredient handles GET /ingredients/{id}
func (h *HttpHandler) GetIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient ID is required",
		})
		return
	}

	// Validate ingredient ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Warn("Invalid ingredient ID format in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient ID must be a valid UUID",
		})
		return
	}

	ingredient, err := h.dbHandler.GetIngredientByID(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Ingredient{},
			Message: "Failed to get ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *ingredient,
		Message: "Ingredient retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListIngredients handles GET /ingredients
func (h *HttpHandler) ListIngredients(w http.ResponseWriter, r *http.Request) {
	ingredients, err := h.dbHandler.ListIngredients(h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.Ingredient{},
			Message: "Failed to list ingredients: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    ingredients,
		Message: "Ingredients retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateIngredient handles PUT /ingredients/{id}
func (h *HttpHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in update request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient ID is required",
		})
		return
	}

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update ingredient request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateIngredientRequest(req, id); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Error("Validation failed for update ingredient request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	ingredient, err := h.dbHandler.UpdateIngredient(id, req, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Ingredient{},
			Message: "Failed to update ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *ingredient,
		Message: "Ingredient updated successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id":   ingredient.ID,
		"ingredient_name": ingredient.Name,
	}).Info("Ingredient updated successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteIngredient handles DELETE /ingredients/{id}
func (h *HttpHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing ingredient ID in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient ID is required",
		})
		return
	}

	// Validate ingredient ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": id,
		}).Warn("Invalid ingredient ID format in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Ingredient ID must be a valid UUID",
		})
		return
	}

	err := h.dbHandler.DeleteIngredient(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Ingredient{},
				Message: "Ingredient not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete ingredient: " + err.Error(),
		})
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    models.Ingredient{},
		Message: "Ingredient deleted successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id": id,
	}).Info("Ingredient deleted successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
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
