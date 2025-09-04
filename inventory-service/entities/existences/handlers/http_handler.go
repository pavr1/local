package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/existences/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateExistence(req models.CreateExistenceRequest, logger *logrus.Logger) (*models.Existence, error)
	GetExistenceByID(id string, logger *logrus.Logger) (*models.Existence, error)
	GetMostRecentExistenceByIngredientAndUnitType(ingredientID, unitType string, logger *logrus.Logger) (*models.Existence, error)
	ListExistences(req models.ListExistencesRequest, logger *logrus.Logger) ([]models.Existence, error)
	UpdateExistence(id string, req models.UpdateExistenceRequest, logger *logrus.Logger) (*models.Existence, error)
	DeleteExistence(id string, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for existence operations
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

// CreateExistence handles POST /existences
func (h *HttpHandler) CreateExistence(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	var req models.CreateExistenceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create existence request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateExistenceRequest(req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": req.IngredientID,
		}).Error("Validation failed for create existence request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	existence, err := h.dbHandler.CreateExistence(req, h.logger)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create existence")
		http.Error(w, "Failed to create existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    existence,
		Message: "Existence created successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"existence_id":  existence.ID,
		"ingredient_id": existence.IngredientID,
	}).Info("Existence created successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetExistence handles GET /existences/{id}
func (h *HttpHandler) GetExistence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Validate existence ID format
	if !isValidUUID(id) {
		h.logger.WithFields(logrus.Fields{
			"existence_id": id,
		}).Warn("Invalid existence ID format in get request")
		http.Error(w, "Existence ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	existence, err := h.dbHandler.GetExistenceByID(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Existence not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get existence")
		http.Error(w, "Failed to get existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    existence,
		Message: "Existence retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMostRecentExistenceByIngredientAndUnitType handles GET /existences/ingredient/{ingredientId}/unit-type/{unitType}
func (h *HttpHandler) GetMostRecentExistenceByIngredientAndUnitType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ingredientID := vars["ingredientId"]
	unitType := vars["unitType"]

	if ingredientID == "" || unitType == "" {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": ingredientID,
			"unit_type":     unitType,
		}).Error("Missing ingredient ID or unit type")
		http.Error(w, "Missing ingredient ID or unit type", http.StatusBadRequest)
		return
	}

	// Validate ingredient ID format
	if !isValidUUID(ingredientID) {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": ingredientID,
		}).Warn("Invalid ingredient ID format")
		http.Error(w, "Ingredient ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	// Validate unit_type format
	//pvillalobos - hardcoded values
	allowedUnitTypes := []string{"Liters", "Gallons", "Units", "Bag"}
	unitTypeValid := false
	for _, allowed := range allowedUnitTypes {
		if unitType == allowed {
			unitTypeValid = true
			break
		}
	}
	if !unitTypeValid {
		h.logger.WithFields(logrus.Fields{
			"unit_type": unitType,
		}).Warn("Invalid unit type format")
		http.Error(w, "Unit type must be one of: Liters, Gallons, Units, Bag", http.StatusBadRequest)
		return
	}

	existence, err := h.dbHandler.GetMostRecentExistenceByIngredientAndUnitType(ingredientID, unitType, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"ingredient_id": ingredientID,
				"unit_type":     unitType,
			}).Warn("No existence found for this ingredient and unit type")
			// Return 404 when no existence is found
			http.Error(w, "No existence found for this ingredient and unit type", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get most recent existence")
		http.Error(w, "Failed to get most recent existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    existence,
		Message: "Most recent existence retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListExistences handles GET /existences
func (h *HttpHandler) ListExistences(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := models.ListExistencesRequest{}

	// Parse ingredient_id filter
	if ingredientID := r.URL.Query().Get("ingredient_id"); ingredientID != "" {
		req.IngredientID = &ingredientID
	}

	// Parse unit_type filter
	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		req.UnitType = &unitType
	}

	// Parse expired filter
	if expiredStr := r.URL.Query().Get("expired"); expiredStr != "" {
		expired := expiredStr == "true"
		req.Expired = &expired
	}

	// Parse low_stock filter
	if lowStockStr := r.URL.Query().Get("low_stock"); lowStockStr != "" {
		lowStock := lowStockStr == "true"
		req.LowStock = &lowStock
	}

	existences, err := h.dbHandler.ListExistences(req, h.logger)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list existences")
		http.Error(w, "Failed to list existences", http.StatusInternalServerError)
		return
	}

	response := models.ExistencesResponse{
		Success: true,
		Data:    existences,
		Total:   len(existences),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateExistence handles PUT /existences/{id}
func (h *HttpHandler) UpdateExistence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateExistenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode update existence request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existence, err := h.dbHandler.UpdateExistence(id, req, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"existence_id": id,
			}).Error("Existence not found")
			http.Error(w, "Existence not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to update existence")
		http.Error(w, "Failed to update existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    existence,
		Message: "Existence updated successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"existence_id":  existence.ID,
		"ingredient_id": existence.IngredientID,
	}).Info("Existence updated successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteExistence handles DELETE /existences/{id}
func (h *HttpHandler) DeleteExistence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.dbHandler.DeleteExistence(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"existence_id": id,
			}).Error("Existence not found")
			http.Error(w, "Existence not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to delete existence")
		http.Error(w, "Failed to delete existence", http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Existence deleted successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"existence_id": id,
	}).Info("Existence deleted successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validateCreateExistenceRequest validates all required fields for existence creation
func (h *HttpHandler) validateCreateExistenceRequest(req models.CreateExistenceRequest) error {
	// Validate ingredient_id (required, valid UUID)
	if req.IngredientID == "" {
		return fmt.Errorf("ingredient_id is required and cannot be empty")
	}
	if !isValidUUID(req.IngredientID) {
		return fmt.Errorf("ingredient_id must be a valid UUID, got: %s", req.IngredientID)
	}

	// Validate invoice_detail_id (required, valid UUID)
	if req.InvoiceDetailID == "" {
		return fmt.Errorf("invoice_detail_id is required and cannot be empty")
	}
	if !isValidUUID(req.InvoiceDetailID) {
		return fmt.Errorf("invoice_detail_id must be a valid UUID, got: %s", req.InvoiceDetailID)
	}

	// Validate units_purchased (required, greater than 0)
	if req.UnitsPurchased <= 0 {
		return fmt.Errorf("units_purchased must be greater than 0, got: %f", req.UnitsPurchased)
	}

	// Validate units_available (required, non-negative)
	if req.UnitsAvailable < 0 {
		return fmt.Errorf("units_available must be non-negative, got: %f", req.UnitsAvailable)
	}

	// Validate unit_type (required, must be one of the allowed values)
	if req.UnitType == "" {
		return fmt.Errorf("unit_type is required and cannot be empty")
	}
	allowedUnitTypes := []string{"Liters", "Gallons", "Units", "Bag"}
	unitTypeValid := false
	for _, allowed := range allowedUnitTypes {
		if req.UnitType == allowed {
			unitTypeValid = true
			break
		}
	}
	if !unitTypeValid {
		return fmt.Errorf("unit_type must be one of %v, got: %s", allowedUnitTypes, req.UnitType)
	}

	// Validate items_per_unit (required, greater than 0)
	if req.ItemsPerUnit <= 0 {
		return fmt.Errorf("items_per_unit must be greater than 0, got: %d", req.ItemsPerUnit)
	}

	// Validate cost_per_unit (required, greater than 0)
	if req.CostPerUnit <= 0 {
		return fmt.Errorf("cost_per_unit must be greater than 0, got: %f", req.CostPerUnit)
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
