package settings

import (
	"encoding/json"
	"net/http"

	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/sirupsen/logrus"
)

// SettingsHandler handles HTTP requests for settings
type SettingsHandler struct {
	service *SettingsService
	config  map[string]Setting
	logger  *logrus.Logger
}

// NewSettingsHandler creates a new settings HTTP handler
func NewSettingsHandler(service *SettingsService, logger *logrus.Logger) (*SettingsHandler, error) {
	handler := &SettingsHandler{
		service: service,
		config:  make(map[string]Setting),
		logger:  logger,
	}

	// Load all settings into memory during initialization
	if err := service.loadAllSettings(); err != nil {
		return nil, err
	}

	// Copy settings to handler config
	service.cacheMux.RLock()
	handler.config = make(map[string]Setting)
	for k, v := range service.cache {
		handler.config[k] = v
	}
	service.cacheMux.RUnlock()

	return handler, nil
}

// GetAllSettings returns all settings (requires gateway validation for UI access)
func (h *SettingsHandler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Received request to get all settings")

	// Get all settings from config map
	var settings []Setting
	for _, setting := range h.config {
		settings = append(settings, setting)
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    settings,
		Message: "All settings retrieved successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, response)
}

// GetByService returns general settings + service-specific settings (for business services)
func (h *SettingsHandler) GetByService(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req GetSettingsByServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if req.Service == "" {
		h.logger.Error("Service parameter is required")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Service parameter is required",
		})
		return
	}

	h.logger.WithField("service", req.Service).Info("Received request to get settings by service")

	// Get general settings + service-specific settings
	var settings []Setting
	for _, setting := range h.config {
		if setting.Service == "General" || setting.Service == req.Service {
			settings = append(settings, setting)
		}
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    settings,
		Message: "Settings retrieved successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, response)
}

// GetByKey returns a specific setting by service and key
func (h *SettingsHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req GetSettingByKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" {
		h.logger.Error("Service and Key parameters are required")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Service and Key parameters are required",
		})
		return
	}

	// Validate service name format (basic validation)
	if len(req.Service) > 50 {
		h.logger.Error("Service name too long")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Service name cannot exceed 50 characters",
		})
		return
	}

	//pvillalobos - hardcoded values
	// Validate key format (basic validation)
	if len(req.Key) > 100 {
		h.logger.Error("Key name too long")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Key name cannot exceed 100 characters",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"service": req.Service,
		"key":     req.Key,
	}).Info("Received request to get setting by key")

	// Find the specific setting
	var foundSetting *Setting
	for _, setting := range h.config {
		if setting.Service == req.Service && setting.Key == req.Key {
			foundSetting = &setting
			break
		}
	}

	if foundSetting == nil {
		h.logger.WithFields(logrus.Fields{
			"service": req.Service,
			"key":     req.Key,
		}).Warn("Setting not found")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusNotFound,
			Message: "Setting not found",
		})
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *foundSetting,
		Message: "Setting retrieved successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, response)
}

// Reload reloads all settings into memory
func (h *SettingsHandler) Reload(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received request to reload settings")

	// Reload settings into memory
	if err := h.service.loadAllSettings(); err != nil {
		h.logger.WithError(err).Error("Failed to reload settings")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reload settings",
		})
		return
	}

	// Update handler config
	h.service.cacheMux.RLock()
	h.config = make(map[string]Setting)
	for k, v := range h.service.cache {
		h.config[k] = v
	}
	h.service.cacheMux.RUnlock()

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Settings reloaded successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, response)
}

// UpdateSetting updates a specific setting (requires gateway validation for UI access)
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	// Parse request body
	var req UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" || req.Value == "" {
		h.logger.Error("Service, Key, and Value parameters are required")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Service, Key, and Value parameters are required",
		})
		return
	}

	// Validate service name format (basic validation)
	if len(req.Service) > 50 {
		h.logger.Error("Service name too long")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Service name cannot exceed 50 characters",
		})
		return
	}

	// Validate key format (basic validation)
	if len(req.Key) > 100 {
		h.logger.Error("Key name too long")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Key name cannot exceed 100 characters",
		})
		return
	}

	// Validate value format (basic validation)
	if len(req.Value) > 1000 {
		h.logger.Error("Value too long")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Value cannot exceed 1000 characters",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"service": req.Service,
		"key":     req.Key,
		"value":   req.Value,
	}).Info("Received request to update setting")

	// Update setting in database
	if err := h.service.dbHandler.UpdateSetting(req.Service, req.Key, req.Value); err != nil {
		h.logger.WithError(err).Error("Failed to update setting in database")
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	// Reload settings into memory
	if err := h.service.loadAllSettings(); err != nil {
		h.logger.WithError(err).Error("Failed to reload settings after update")
		http.Error(w, "Setting updated but failed to reload cache", http.StatusInternalServerError)
		return
	}

	// Update handler config
	h.service.cacheMux.RLock()
	h.config = make(map[string]Setting)
	for k, v := range h.service.cache {
		h.config[k] = v
	}
	h.service.cacheMux.RUnlock()

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Setting updated successfully",
	}

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_DATA_SERVICE, response)
}

// validateGatewayRequest checks if the request is coming from the gateway
func (h *SettingsHandler) validateGatewayRequest(r *http.Request) bool {
	// Check for gateway headers
	gatewayService := r.Header.Get("X-Gateway-Service")
	gatewaySessionManaged := r.Header.Get("X-Gateway-Session-Managed")

	// Check for user context headers (indicating gateway authentication)
	userID := r.Header.Get("X-User-ID")
	username := r.Header.Get("X-Username")
	userRole := r.Header.Get("X-User-Role")

	// Request must come from gateway and have user context
	return gatewayService == "ice-cream-gateway" &&
		gatewaySessionManaged == "true" &&
		userID != "" &&
		username != "" &&
		userRole != ""
}
