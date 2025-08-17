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
}

// NewSettingsHandler creates a new settings HTTP handler
func NewSettingsHandler(service *SettingsService, logger *logrus.Logger) *SettingsHandler {
	return &SettingsHandler{
		service: service,
		logger:  logger,
	}
}

// LoadAllSettings handles the request to load all settings into memory
func (h *SettingsHandler) LoadAllSettings(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received request to load all settings")

	// Check if request is from gateway
	if !h.validateGatewayRequest(r) {
		h.logger.Warn("Request not from gateway, rejecting")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, err := h.service.LoadAllSettings()
	if err != nil {
		h.logger.WithError(err).Error("Failed to load all settings")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, response)
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

	response, err := h.service.GetSettingsByService(req.Service)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get settings by service")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	response, err := h.service.GetSettingsByName(req.Key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get settings by name")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	response, err := h.service.GetSettingsByServiceGrouped()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get settings grouped by service")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
