package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"inventory-service/entities/existences/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateExistence(req models.CreateExistenceRequest) (*models.Existence, error)
	GetExistenceByID(id string) (*models.Existence, error)
	ListExistences(req models.ListExistencesRequest) ([]models.Existence, error)
	UpdateExistence(id string, req models.UpdateExistenceRequest) (*models.Existence, error)
	DeleteExistence(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for existence operations
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

// CreateExistence handles POST /existences
func (h *HttpHandler) CreateExistence(w http.ResponseWriter, r *http.Request) {
	var req models.CreateExistenceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create existence request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existence, err := h.dbHandler.CreateExistence(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create existence")
		http.Error(w, "Failed to create existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    *existence,
		Message: "Existence created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetExistence handles GET /existences/{id}
func (h *HttpHandler) GetExistence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	existence, err := h.dbHandler.GetExistenceByID(id)
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
		Data:    *existence,
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

	existences, err := h.dbHandler.ListExistences(req)
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

	existence, err := h.dbHandler.UpdateExistence(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Existence not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to update existence")
		http.Error(w, "Failed to update existence", http.StatusInternalServerError)
		return
	}

	response := models.ExistenceResponse{
		Success: true,
		Data:    *existence,
		Message: "Existence updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteExistence handles DELETE /existences/{id}
func (h *HttpHandler) DeleteExistence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.dbHandler.DeleteExistence(id)
	if err != nil {
		if err == sql.ErrNoRows {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
