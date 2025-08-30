package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"inventory-service/entities/suppliers/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateSupplier(req models.CreateSupplierRequest, logger *logrus.Entry) (*models.Supplier, error)
	GetSupplierByID(id string, logger *logrus.Entry) (*models.Supplier, error)
	ListSuppliers(logger *logrus.Entry) ([]models.Supplier, error)
	UpdateSupplier(id string, req models.UpdateSupplierRequest, logger *logrus.Entry) (*models.Supplier, error)
	DeleteSupplier(id string, logger *logrus.Entry) error
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
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	var req models.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create supplier request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateCreateSupplierRequest(req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_name": req.SupplierName,
		}).Error("Validation failed for create supplier request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	supplier, err := h.dbHandler.CreateSupplier(req, logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Supplier{},
			Message: "Failed to create supplier: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	logger.WithFields(logrus.Fields{
		"supplier_id":   supplier.ID,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier created successfully")

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *supplier,
		Message: "Supplier created successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetSupplier handles GET /suppliers/{id}
func (h *HttpHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing supplier ID in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Supplier ID is required",
		})
		return
	}

	// Validate supplier ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"supplier_id": id,
		}).Warn("Invalid supplier ID format in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Supplier ID must be a valid UUID",
		})
		return
	}

	supplier, err := h.dbHandler.GetSupplierByID(id, logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Supplier{},
			Message: "Failed to get supplier: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *supplier,
		Message: "Supplier retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListSuppliers handles GET /suppliers
func (h *HttpHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	// TODO: Parse query parameters for pagination when needed
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")

	suppliers, err := h.dbHandler.ListSuppliers(logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.Supplier{},
			Message: "Failed to list suppliers: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    suppliers,
		Message: "Suppliers retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateSupplier handles PUT /suppliers/{id}
func (h *HttpHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing supplier ID in update request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Supplier ID is required",
		})
		return
	}

	var req models.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update supplier request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate required fields based on database schema
	if err := h.validateUpdateSupplierRequest(req, id); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_id": id,
		}).Error("Validation failed for update supplier request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	supplier, err := h.dbHandler.UpdateSupplier(id, req, logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.Supplier{},
			Message: "Failed to update supplier: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *supplier,
		Message: "Supplier updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"supplier_id":   supplier.ID,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier updated successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteSupplier handles DELETE /suppliers/{id}
func (h *HttpHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing supplier ID in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Supplier ID is required",
		})
		return
	}

	// Validate supplier ID format
	if !isValidUUID(id) {
		logger.WithFields(logrus.Fields{
			"supplier_id": id,
		}).Warn("Invalid supplier ID format in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Supplier ID must be a valid UUID",
		})
		return
	}

	err := h.dbHandler.DeleteSupplier(id, logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Supplier not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete supplier: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Supplier deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"supplier_id": id,
	}).Info("Supplier deleted successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// validateCreateSupplierRequest validates all required fields for supplier creation
func (h *HttpHandler) validateCreateSupplierRequest(req models.CreateSupplierRequest) error {
	// Validate supplier_name (required, non-empty, max 255 chars)
	if req.SupplierName == "" {
		return fmt.Errorf("supplier_name is required and cannot be empty")
	}
	if len(req.SupplierName) > 255 {
		return fmt.Errorf("supplier_name cannot exceed 255 characters, got: %d", len(req.SupplierName))
	}

	return nil
}

// validateUpdateSupplierRequest validates all required fields for supplier update
func (h *HttpHandler) validateUpdateSupplierRequest(req models.UpdateSupplierRequest, id string) error {
	// Validate supplier ID (required, valid UUID)
	if id == "" {
		return fmt.Errorf("supplier ID is required and cannot be empty")
	}
	if !isValidUUID(id) {
		return fmt.Errorf("supplier ID must be a valid UUID, got: %s", id)
	}

	// Validate supplier_name if provided (non-empty, max 255 chars)
	if req.SupplierName != nil {
		if *req.SupplierName == "" {
			return fmt.Errorf("supplier_name cannot be empty if provided")
		}
		if len(*req.SupplierName) > 255 {
			return fmt.Errorf("supplier_name cannot exceed 255 characters, got: %d", len(*req.SupplierName))
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
