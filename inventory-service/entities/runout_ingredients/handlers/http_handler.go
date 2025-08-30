package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"inventory-service/entities/runout_ingredients/models"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HttpHandler struct {
	dbHandler *DBHandler
}

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	Create(req models.CreateRunoutIngredientRequest, logger *logrus.Logger) (*models.RunoutIngredient, error)
	GetByID(req models.GetRunoutIngredientRequest, logger *logrus.Logger) (*models.RunoutIngredient, error)
	List(req models.ListRunoutIngredientsRequest, logger *logrus.Logger) ([]models.RunoutIngredient, error)
	Update(req models.UpdateRunoutIngredientRequest, id string, logger *logrus.Logger) (*models.RunoutIngredient, error)
	Delete(req models.DeleteRunoutIngredientRequest, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// NewHttpHandler creates a new HTTP handler for runout ingredients
func NewHttpHandler(db *sql.DB) *HttpHandler {
	return &HttpHandler{
		dbHandler: NewDBHandler(db),
	}
}

// CreateRunoutIngredient handles POST /runout-ingredients
func (h *HttpHandler) CreateRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	var req models.CreateRunoutIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create runout ingredient request")
		h.writeErrorResponse(w, r, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	runoutIngredient, err := h.dbHandler.Create(req, logger.Logger)
	if err != nil {
		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to create runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, r, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient created successfully",
	}

	logger.WithFields(logrus.Fields{
		"runout_ingredient_id": runoutIngredient.ID,
		"existence_id":         runoutIngredient.ExistenceID,
		"employee_id":          runoutIngredient.EmployeeID,
	}).Info("Runout ingredient created successfully")

	h.writeJSONResponse(w, r, response, http.StatusCreated)
}

// GetRunoutIngredient handles GET /runout-ingredients/{id}
func (h *HttpHandler) GetRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing runout ingredient ID in get request")
		h.writeErrorResponse(w, r, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.GetRunoutIngredientRequest{ID: id}
	runoutIngredient, err := h.dbHandler.GetByID(req, logger.Logger)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.RunoutIngredientResponse{
				Success: false,
				Data:    models.RunoutIngredient{},
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, r, response, http.StatusNotFound)
			return
		}

		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to get runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, r, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient retrieved successfully",
	}
	h.writeJSONResponse(w, r, response, http.StatusOK)
}

// ListRunoutIngredients handles GET /runout-ingredients
func (h *HttpHandler) ListRunoutIngredients(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	req := models.ListRunoutIngredientsRequest{}

	// Parse query parameters
	if existenceID := r.URL.Query().Get("existence_id"); existenceID != "" {
		req.ExistenceID = &existenceID
	}

	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		req.EmployeeID = &employeeID
	}

	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		req.UnitType = &unitType
	}

	if reportDateStr := r.URL.Query().Get("report_date"); reportDateStr != "" {
		if reportDate, err := time.Parse("2006-01-02", reportDateStr); err == nil {
			req.ReportDate = &reportDate
		}
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

	runoutIngredients, err := h.dbHandler.List(req, logger.Logger)
	if err != nil {
		response := models.RunoutIngredientsResponse{
			Success: false,
			Data:    []models.RunoutIngredient{},
			Message: "Failed to list runout ingredients: " + err.Error(),
		}
		h.writeJSONResponse(w, r, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientsResponse{
		Success: true,
		Data:    runoutIngredients,
		Total:   len(runoutIngredients),
		Message: "Runout ingredients retrieved successfully",
	}
	h.writeJSONResponse(w, r, response, http.StatusOK)
}

// UpdateRunoutIngredient handles PUT /runout-ingredients/{id}
func (h *HttpHandler) UpdateRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing runout ingredient ID in update request")
		h.writeErrorResponse(w, r, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRunoutIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update runout ingredient request")
		h.writeErrorResponse(w, r, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	runoutIngredient, err := h.dbHandler.Update(req, id, logger.Logger)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.RunoutIngredientResponse{
				Success: false,
				Data:    models.RunoutIngredient{},
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, r, response, http.StatusNotFound)
			return
		}

		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to update runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, r, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"runout_ingredient_id": runoutIngredient.ID,
		"existence_id":         runoutIngredient.ExistenceID,
		"employee_id":          runoutIngredient.EmployeeID,
	}).Info("Runout ingredient updated successfully")

	h.writeJSONResponse(w, r, response, http.StatusOK)
}

// DeleteRunoutIngredient handles DELETE /runout-ingredients/{id}
func (h *HttpHandler) DeleteRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing runout ingredient ID in delete request")
		h.writeErrorResponse(w, r, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.DeleteRunoutIngredientRequest{ID: id}
	err := h.dbHandler.Delete(req, logger.Logger)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.GenericResponse{
				Success: false,
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, r, response, http.StatusNotFound)
			return
		}

		response := models.GenericResponse{
			Success: false,
			Message: "Failed to delete runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, r, response, http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Runout ingredient deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"runout_ingredient_id": id,
	}).Info("Runout ingredient deleted successfully")

	h.writeJSONResponse(w, r, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, r, errorResponse, statusCode)
}
