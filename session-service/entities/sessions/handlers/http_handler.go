package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"session-service/entities/sessions/models"
	httpresponse "shared/http-response"
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
	// Parse request
	var req models.SessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		h.logger.Error("Username and password are required")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Username and password are required",
		})
		return
	}

	// Create session
	response, err := h.dbHandler.CreateSession(&req)
	if err != nil {
		h.logger.WithError(err).Error("Invalid username or password")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid username or password",
		})
		return
	}

	// Return success response
	h.logger.WithFields(logrus.Fields{
		"session_id": response.SessionID,
		"username":   req.Username,
	}).Info("Session created successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    response,
		Message: "Session created successfully",
	})
}

// ValidateSession handles session validation requests
func (h *HTTPHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.SessionValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Validate session
	response, err := h.dbHandler.ValidateSession(req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to validate session")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to validate session",
		})
		return
	}

	// Write response
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Data:    response,
		Message: "Session validated successfully",
	})
}

// LogoutSession handles session logout requests
func (h *HTTPHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.SessionLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")

		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.SessionID == "" {
		h.logger.WithError(errors.New("Session ID is required")).Error("Session ID is required")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Session ID is required",
		})
		return
	}

	// Delete session
	response, err := h.dbHandler.DeleteSession(req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to logout session")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to logout session",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": req.SessionID,
	}).Info("Session logged out successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_SESSION_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Data:    response,
		Message: "Session logged out successfully",
	})
}
