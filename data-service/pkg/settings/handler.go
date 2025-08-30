package settings

import (
	"encoding/json"
	"net/http"

	sharedLogger "shared/logger"

	"github.com/sirupsen/logrus"
)

// SettingsHandler handles HTTP requests for settings
type SettingsHandler struct {
	service *SettingsService
	config  map[string]Setting
}

// NewSettingsHandler creates a new settings HTTP handler
func NewSettingsHandler(service *SettingsService) (*SettingsHandler, error) {
	handler := &SettingsHandler{
		service: service,
		config:  make(map[string]Setting),
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
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_DATA_SERVICE)
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Info("Received request to get all settings")

	// Get all settings from config map
	var settings []Setting
	for _, setting := range h.config {
		settings = append(settings, setting)
	}

	response := &SettingsResponse{
		Success: true,
		Data:    settings,
		Total:   len(settings),
		Message: "All settings retrieved successfully",
	}

	h.writeJSONResponse(w, response, logger.Logger)
}

// GetByService returns general settings + service-specific settings (for business services)
func (h *SettingsHandler) GetByService(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_DATA_SERVICE)
	// Parse request body
	var req GetSettingsByServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" {
		logger.Error("Service parameter is required")
		http.Error(w, "Service parameter is required", http.StatusBadRequest)
		return
	}

	logger.WithField("service", req.Service).Info("Received request to get settings by service")

	// Get general settings + service-specific settings
	var settings []Setting
	for _, setting := range h.config {
		if setting.Service == "General" || setting.Service == req.Service {
			settings = append(settings, setting)
		}
	}

	response := &SettingsResponse{
		Success: true,
		Data:    settings,
		Total:   len(settings),
		Message: "Settings retrieved successfully",
	}

	h.writeJSONResponse(w, response, logger.Logger)
}

// GetByKey returns a specific setting by service and key
func (h *SettingsHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_DATA_SERVICE)
	// Parse request body
	var req GetSettingByKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" {
		logger.Error("Service and Key parameters are required")
		http.Error(w, "Service and Key parameters are required", http.StatusBadRequest)
		return
	}

	// Validate service name format (basic validation)
	if len(req.Service) > 50 {
		logger.Error("Service name too long")
		http.Error(w, "Service name cannot exceed 50 characters", http.StatusBadRequest)
		return
	}

	// Validate key format (basic validation)
	if len(req.Key) > 100 {
		logger.Error("Key name too long")
		http.Error(w, "Key name cannot exceed 100 characters", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
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
		logger.WithFields(logrus.Fields{
			"service": req.Service,
			"key":     req.Key,
		}).Warn("Setting not found")
		http.Error(w, "Setting not found", http.StatusNotFound)
		return
	}

	response := &SettingResponse{
		Success: true,
		Data:    *foundSetting,
		Message: "Setting retrieved successfully",
	}

	h.writeJSONResponse(w, response, logger.Logger)
}

// Reload reloads all settings into memory
func (h *SettingsHandler) Reload(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_DATA_SERVICE)
	logger.Info("Received request to reload settings")

	// Reload settings into memory
	if err := h.service.loadAllSettings(); err != nil {
		logger.WithError(err).Error("Failed to reload settings")
		http.Error(w, "Failed to reload settings", http.StatusInternalServerError)
		return
	}

	// Update handler config
	h.service.cacheMux.RLock()
	h.config = make(map[string]Setting)
	for k, v := range h.service.cache {
		h.config[k] = v
	}
	h.service.cacheMux.RUnlock()

	response := &GenericResponse{
		Success: true,
		Message: "Settings reloaded successfully",
	}

	h.writeJSONResponse(w, response, logger.Logger)
}

// UpdateSetting updates a specific setting (requires gateway validation for UI access)
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_DATA_SERVICE)
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" || req.Value == "" {
		logger.Error("Service, Key, and Value parameters are required")
		http.Error(w, "Service, Key, and Value parameters are required", http.StatusBadRequest)
		return
	}

	// Validate service name format (basic validation)
	if len(req.Service) > 50 {
		logger.Error("Service name too long")
		http.Error(w, "Service name cannot exceed 50 characters", http.StatusBadRequest)
		return
	}

	// Validate key format (basic validation)
	if len(req.Key) > 100 {
		logger.Error("Key name too long")
		http.Error(w, "Key name cannot exceed 100 characters", http.StatusBadRequest)
		return
	}

	// Validate value format (basic validation)
	if len(req.Value) > 1000 {
		logger.Error("Value too long")
		http.Error(w, "Value cannot exceed 1000 characters", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"service": req.Service,
		"key":     req.Key,
		"value":   req.Value,
	}).Info("Received request to update setting")

	// Update setting in database
	if err := h.service.dbHandler.UpdateSetting(req.Service, req.Key, req.Value); err != nil {
		logger.WithError(err).Error("Failed to update setting in database")
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	// Reload settings into memory
	if err := h.service.loadAllSettings(); err != nil {
		logger.WithError(err).Error("Failed to reload settings after update")
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

	response := &GenericResponse{
		Success: true,
		Message: "Setting updated successfully",
	}

	h.writeJSONResponse(w, response, logger.Logger)
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
