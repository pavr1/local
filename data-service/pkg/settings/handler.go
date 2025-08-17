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

// GetSettingsByService handles the request to get settings by service
func (h *SettingsHandler) GetSettingsByService(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Filter settings by service from config map
	var settings []Setting
	for _, setting := range h.config {
		if setting.Service == req.Service {
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

// GetSettingsByName handles the request to get settings by name
func (h *SettingsHandler) GetSettingsByName(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req GetSettingsByNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Key == "" {
		h.logger.Error("Key parameter is required")
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	h.logger.WithField("key", req.Key).Info("Received request to get settings by name")

	// Filter settings by key from config map
	var settings []Setting
	for _, setting := range h.config {
		if setting.Key == req.Key {
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

// GetSettingsByServiceGrouped handles the request to get all settings grouped by service
func (h *SettingsHandler) GetSettingsByServiceGrouped(w http.ResponseWriter, r *http.Request) {
	// Check if request is from gateway
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Received request to get settings grouped by service")

	// Group settings by service from config map
	serviceMap := make(map[string][]Setting)
	for _, setting := range h.config {
		serviceMap[setting.Service] = append(serviceMap[setting.Service], setting)
	}

	var groupedSettings []SettingsByService
	for service, serviceSettings := range serviceMap {
		groupedSettings = append(groupedSettings, SettingsByService{
			Service:  service,
			Settings: serviceSettings,
		})
	}

	response := &SettingsByServiceResponse{
		Success: true,
		Data:    groupedSettings,
		Total:   len(h.config),
		Message: "Settings grouped successfully",
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
