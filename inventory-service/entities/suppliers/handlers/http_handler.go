package handlers

import (
	"encoding/json"
	"net/http"

	"inventory-service/entities/suppliers/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HttpHandler handles HTTP requests for supplier operations
type HttpHandler struct {
	dbHandler *DBHandler
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateSupplier handles POST /suppliers
func (h *HttpHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode create supplier request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response := h.dbHandler.CreateSupplier(req)

	if response.Success {
		h.writeJSONResponse(w, response, http.StatusCreated)
	} else {
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
	}
}

// GetSupplier handles GET /suppliers/{id}
func (h *HttpHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	req := models.GetSupplierRequest{ID: id}
	response := h.dbHandler.GetSupplier(req)

	if response.Success {
		h.writeJSONResponse(w, response, http.StatusOK)
	} else {
		if response.Message == "Supplier not found" {
			h.writeJSONResponse(w, response, http.StatusNotFound)
		} else {
			h.writeJSONResponse(w, response, http.StatusInternalServerError)
		}
	}
}

// ListSuppliers handles GET /suppliers
func (h *HttpHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	// For now, we'll use empty request (no pagination implemented yet)
	req := models.ListSuppliersRequest{}

	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	response := h.dbHandler.ListSuppliers(req)

	if response.Success {
		h.writeJSONResponse(w, response, http.StatusOK)
	} else {
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
	}
}

// UpdateSupplier handles PUT /suppliers/{id}
func (h *HttpHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode update supplier request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response := h.dbHandler.UpdateSupplier(id, req)

	if response.Success {
		h.writeJSONResponse(w, response, http.StatusOK)
	} else {
		if response.Message == "Supplier not found" {
			h.writeJSONResponse(w, response, http.StatusNotFound)
		} else {
			h.writeJSONResponse(w, response, http.StatusInternalServerError)
		}
	}
}

// DeleteSupplier handles DELETE /suppliers/{id}
func (h *HttpHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	req := models.DeleteSupplierRequest{ID: id}
	response := h.dbHandler.DeleteSupplier(req)

	if response.Success {
		h.writeJSONResponse(w, response, http.StatusOK)
	} else {
		if response.Message == "Supplier not found" {
			h.writeJSONResponse(w, response, http.StatusNotFound)
		} else {
			h.writeJSONResponse(w, response, http.StatusInternalServerError)
		}
	}
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
