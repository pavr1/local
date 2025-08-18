package settings

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// SettingsHandler handles HTTP requests for settings
type SettingsHandler struct {
	service *SettingsService
	logger  *logrus.Logger
	config  map[string]Setting
}

// NewSettingsHandler creates a new settings HTTP handler
func NewSettingsHandler(service *SettingsService, logger *logrus.Logger) (*SettingsHandler, error) {
	handler := &SettingsHandler{
		service: service,
		logger:  logger,
		config:  make(map[string]Setting),
	}

	// Load all settings into memory during initialization
	if err := service.loadAllSettings(); err != nil {
		logger.WithError(err).Error("Failed to load settings during handler initialization")
		return nil, err
	}

	// Copy settings to handler config
	service.cacheMux.RLock()
	handler.config = make(map[string]Setting)
	for k, v := range service.cache {
		handler.config[k] = v
	}
	service.cacheMux.RUnlock()
	logger.Info("Settings loaded into handler config successfully")

	return handler, nil
}

// GetAllSettings returns all settings (requires gateway validation for UI access)
func (h *SettingsHandler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Received request to get all settings")

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

	h.writeJSONResponse(w, response)
}

// GetByService returns general settings + service-specific settings (for business services)
func (h *SettingsHandler) GetByService(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req GetSettingsByServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" {
		h.logger.Error("Service parameter is required")
		http.Error(w, "Service parameter is required", http.StatusBadRequest)
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

	response := &SettingsResponse{
		Success: true,
		Data:    settings,
		Total:   len(settings),
		Message: "Settings retrieved successfully",
	}

	h.writeJSONResponse(w, response)
}

// GetByKey returns a specific setting by service and key
func (h *SettingsHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req GetSettingByKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" {
		h.logger.Error("Service and Key parameters are required")
		http.Error(w, "Service and Key parameters are required", http.StatusBadRequest)
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
		http.Error(w, "Setting not found", http.StatusNotFound)
		return
	}

	response := &SettingResponse{
		Success: true,
		Data:    *foundSetting,
		Message: "Setting retrieved successfully",
	}

	h.writeJSONResponse(w, response)
}

// Reload reloads all settings into memory
func (h *SettingsHandler) Reload(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received request to reload settings")

	// Reload settings into memory
	if err := h.service.loadAllSettings(); err != nil {
		h.logger.WithError(err).Error("Failed to reload settings")
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

	h.writeJSONResponse(w, response)
}

// UpdateSetting updates a specific setting (requires gateway validation for UI access)
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway (UI access)
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" || req.Key == "" || req.Value == "" {
		h.logger.Error("Service, Key, and Value parameters are required")
		http.Error(w, "Service, Key, and Value parameters are required", http.StatusBadRequest)
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

	response := &GenericResponse{
		Success: true,
		Message: "Setting updated successfully",
	}

	h.writeJSONResponse(w, response)
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
	return gatewayService == "gateway" &&
		gatewaySessionManaged == "true" &&
		userID != "" &&
		username != "" &&
		userRole != ""
}

// writeJSONResponse writes a JSON response to the HTTP response writer
func (h *SettingsHandler) writeJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
