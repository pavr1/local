package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"inventory-service/entities/runout_ingredients/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type RunoutIngredientHTTPHandler struct {
	dbHandler *RunoutIngredientDBHandler
	logger    *logrus.Logger
}

func NewRunoutIngredientHTTPHandler(db *sql.DB, logger *logrus.Logger) *RunoutIngredientHTTPHandler {
	return &RunoutIngredientHTTPHandler{
		dbHandler: NewRunoutIngredientDBHandler(db),
		logger:    logger,
	}
}

// CreateRunoutIngredient handles POST /runout-ingredients
func (h *RunoutIngredientHTTPHandler) CreateRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRunoutIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create runout ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	runoutIngredient, err := h.dbHandler.Create(req)
	if err != nil {
		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to create runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetRunoutIngredient handles GET /runout-ingredients/{id}
func (h *RunoutIngredientHTTPHandler) GetRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing runout ingredient ID in get request")
		h.writeErrorResponse(w, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.GetRunoutIngredientRequest{ID: id}
	runoutIngredient, err := h.dbHandler.GetByID(req)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.RunoutIngredientResponse{
				Success: false,
				Data:    models.RunoutIngredient{},
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to get runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListRunoutIngredients handles GET /runout-ingredients
func (h *RunoutIngredientHTTPHandler) ListRunoutIngredients(w http.ResponseWriter, r *http.Request) {
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

	runoutIngredients, err := h.dbHandler.List(req)
	if err != nil {
		response := models.RunoutIngredientsResponse{
			Success: false,
			Data:    []models.RunoutIngredient{},
			Message: "Failed to list runout ingredients: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientsResponse{
		Success: true,
		Data:    runoutIngredients,
		Total:   len(runoutIngredients),
		Message: "Runout ingredients retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateRunoutIngredient handles PUT /runout-ingredients/{id}
func (h *RunoutIngredientHTTPHandler) UpdateRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing runout ingredient ID in update request")
		h.writeErrorResponse(w, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRunoutIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update runout ingredient request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	runoutIngredient, err := h.dbHandler.Update(req, id)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.RunoutIngredientResponse{
				Success: false,
				Data:    models.RunoutIngredient{},
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.RunoutIngredientResponse{
			Success: false,
			Data:    models.RunoutIngredient{},
			Message: "Failed to update runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.RunoutIngredientResponse{
		Success: true,
		Data:    *runoutIngredient,
		Message: "Runout ingredient updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteRunoutIngredient handles DELETE /runout-ingredients/{id}
func (h *RunoutIngredientHTTPHandler) DeleteRunoutIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing runout ingredient ID in delete request")
		h.writeErrorResponse(w, "Runout ingredient ID is required", http.StatusBadRequest)
		return
	}

	req := models.DeleteRunoutIngredientRequest{ID: id}
	err := h.dbHandler.Delete(req)
	if err != nil {
		if err.Error() == "runout ingredient not found" {
			response := models.GenericResponse{
				Success: false,
				Message: "Runout ingredient not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		response := models.GenericResponse{
			Success: false,
			Message: "Failed to delete runout ingredient: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.GenericResponse{
		Success: true,
		Message: "Runout ingredient deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for HTTP responses

// writeJSONResponse writes a JSON response with the specified status code
func (h *RunoutIngredientHTTPHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		// If we can't encode the response, send a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (h *RunoutIngredientHTTPHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, errorResponse, statusCode)
}
