package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"inventory-service/entities/suppliers/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateSupplier(req models.CreateSupplierRequest) (*models.Supplier, error)
	GetSupplierByID(id string) (*models.Supplier, error)
	ListSuppliers() ([]models.Supplier, error)
	UpdateSupplier(id string, req models.UpdateSupplierRequest) (*models.Supplier, error)
	DeleteSupplier(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for supplier operations
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

// CreateSupplier handles POST /suppliers
func (h *HttpHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create supplier request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	supplier, err := h.dbHandler.CreateSupplier(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to create supplier: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.SupplierResponse{
		Success: true,
		Data:    *supplier,
		Message: "Supplier created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetSupplier handles GET /suppliers/{id}
func (h *HttpHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing supplier ID in get request")
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	supplier, err := h.dbHandler.GetSupplierByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.SupplierResponse{
				Success: false,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to get supplier: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.SupplierResponse{
		Success: true,
		Data:    *supplier,
		Message: "Supplier retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListSuppliers handles GET /suppliers
func (h *HttpHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	suppliers, err := h.dbHandler.ListSuppliers()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.SuppliersListResponse{
			Success: false,
			Data:    []models.Supplier{},
			Count:   0,
			Message: "Failed to list suppliers: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.SuppliersListResponse{
		Success: true,
		Data:    suppliers,
		Count:   len(suppliers),
		Message: "Suppliers retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateSupplier handles PUT /suppliers/{id}
func (h *HttpHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing supplier ID in update request")
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update supplier request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	supplier, err := h.dbHandler.UpdateSupplier(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.SupplierResponse{
				Success: false,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to update supplier: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.SupplierResponse{
		Success: true,
		Data:    *supplier,
		Message: "Supplier updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteSupplier handles DELETE /suppliers/{id}
func (h *HttpHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing supplier ID in delete request")
		h.writeErrorResponse(w, "Supplier ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteSupplier(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.SupplierDeleteResponse{
				Success: false,
				Message: "Supplier not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.SupplierDeleteResponse{
			Success: false,
			Message: "Failed to delete supplier: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.SupplierDeleteResponse{
		Success: true,
		Message: "Supplier deleted successfully",
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
