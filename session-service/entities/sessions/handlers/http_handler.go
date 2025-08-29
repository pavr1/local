package handlers

import (
	"encoding/json"
	"net/http"
	"session-service/entities/sessions/models"
	sharedLogger "shared/logger"

	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the interface for database operations
type DBHandlerInterface interface {
	CreateSession(req *models.SessionCreateRequest) (*models.SessionCreateResponse, error)
	ValidateSession(sessionID string) (*models.SessionValidationResponse, error)
	DeleteSession(sessionID string) (*models.SessionLogoutResponse, error)
	Close() error
}

// HTTPHandler handles HTTP requests for sessions
type HTTPHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(dbHandler DBHandlerInterface, logger *logrus.Logger) *HTTPHandler {
	return &HTTPHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateSession handles session creation requests
func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_SESSION_SERVICE)

	// Parse request
	var req models.SessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_credentials", "Username and password are required")
		return
	}

	// Create session
	response, err := h.dbHandler.CreateSession(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create session")
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication_failed", "Invalid username or password")
		return
	}

	// Return success response
	logger.WithFields(logrus.Fields{
		"session_id": response.SessionID,
		"username":   req.Username,
	}).Info("Session created successfully")

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// ValidateSession handles session validation requests
func (h *HTTPHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.SessionValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate session
	response, err := h.dbHandler.ValidateSession(req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to validate session")
		h.writeErrorResponse(w, http.StatusInternalServerError, "validation_failed", "Failed to validate session")
		return
	}

	// Write response
	h.writeJSONResponse(w, http.StatusOK, response)
}

// LogoutSession handles session logout requests
func (h *HTTPHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_SESSION_SERVICE)

	// Parse request body
	var req models.SessionLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.SessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_session_id", "Session ID is required")
		return
	}

	// Delete session
	response, err := h.dbHandler.DeleteSession(req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to logout session")
		h.writeErrorResponse(w, http.StatusInternalServerError, "logout_failed", "Failed to logout session")
		return
	}

	// Write response
	if response.Success {
		logger.WithFields(logrus.Fields{
			"session_id": req.SessionID,
		}).Info("Session logged out successfully")

		h.writeJSONResponse(w, http.StatusOK, response)
	} else {
		h.writeJSONResponse(w, http.StatusNotFound, response)
	}
}

// writeJSONResponse writes a JSON response
func (h *HTTPHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes an error response
func (h *HTTPHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	response := models.ErrorResponse{
		Error:   errorCode,
		Message: message,
		Code:    errorCode,
	}

	h.writeJSONResponse(w, statusCode, response)
}
