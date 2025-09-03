package sessionmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway-service/models"
	httpresponse "shared/http-response"

	"github.com/sirupsen/logrus"
)

// SessionManager handles communication with the session service
type SessionManager struct {
	baseURL string
	client  *http.Client
	logger  *logrus.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionServiceURL string, logger *logrus.Logger) *SessionManager {
	return &SessionManager{
		baseURL: sessionServiceURL + "/api/v1/sessions",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// makeRequest makes a request to the session service with gateway headers
func (sm *SessionManager) makeRequest(method, path string, body io.Reader, requestID string) (*http.Response, error) {
	sm.logger.WithFields(logrus.Fields{
		"method": method,
		"path":   path,
		"url":    sm.baseURL + path,
	}).Debug("Making request to session service")

	httpReq, err := http.NewRequest(method, sm.baseURL+path, body)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"method": method,
			"path":   path,
		}).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add gateway headers (same as createProxyHandler)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	httpReq.Header.Set("X-Gateway-Session-Managed", "true")

	// Add request ID if provided
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}

	resp, err := sm.client.Do(httpReq)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"method": method,
			"path":   path,
			"url":    sm.baseURL + path,
		}).Error("Failed to make HTTP request to session service")
		return nil, err
	}

	sm.logger.WithFields(logrus.Fields{
		"method":     method,
		"path":       path,
		"status":     resp.StatusCode,
		"statusText": resp.Status,
	}).Debug("Received response from session service")

	return resp, nil
}

// ValidateSession validates a session ID against the session service
func (sm *SessionManager) ValidateSession(sessionId string, requestID string) (*models.SessionValidationResponse, error) {
	sm.logger.WithFields(logrus.Fields{
		"session_id": sessionId,
	}).Debug("Starting session validation")

	if sessionId == "" {
		sm.logger.Warn("Session validation attempted with empty session ID")
		return &models.SessionValidationResponse{
			Valid:   false,
			Message: "Session ID is required",
		}, nil
	}

	validationReq := models.SessionValidationRequest{
		SessionID: sessionId,
	}

	reqBody, err := json.Marshal(validationReq)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionId,
		}).Error("Failed to marshal session validation request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"session_id": sessionId,
		"request":    string(reqBody),
	}).Debug("Sending session validation request")

	resp, err := sm.makeRequest("POST", "/p/validate", bytes.NewBuffer(reqBody), requestID)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionId,
		}).Error("Failed to make session validation request")
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionId,
			"status":     resp.StatusCode,
		}).Error("Failed to read session validation response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"session_id":  sessionId,
		"status":      resp.StatusCode,
		"body":        string(body),
		"body_length": len(body),
	}).Debug("Received session validation response")

	// First, unmarshal the new response structure using shared Response
	var responseWrapper httpresponse.Response
	if err := json.Unmarshal(body, &responseWrapper); err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":  sessionId,
			"status":      resp.StatusCode,
			"body":        string(body),
			"body_length": len(body),
		}).Error("Failed to unmarshal session validation response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract the data from the response
	var validationResp models.SessionValidationResponse
	if responseWrapper.Data != nil {
		if dataBytes, err := json.Marshal(responseWrapper.Data); err == nil {
			json.Unmarshal(dataBytes, &validationResp)
		}
	}

	sm.logger.WithFields(logrus.Fields{
		"session_id":       sessionId,
		"valid":            validationResp.Valid,
		"message":          validationResp.Message,
		"user_id":          validationResp.UserID,
		"username":         validationResp.Username,
		"response_code":    responseWrapper.Code,
		"response_message": responseWrapper.Message,
	}).Debug("Session validation completed")

	return &validationResp, nil
}

// LoginSession creates a new session after successful login
func (sm *SessionManager) LoginSession(req *models.SessionCreateRequest, requestID string) (*models.SessionCreateResponse, error) {
	sm.logger.WithFields(logrus.Fields{
		"username": req.Username,
	}).Debug("Starting session creation")

	reqBody, err := json.Marshal(req)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
		}).Error("Failed to marshal session creation request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := sm.makeRequest("POST", "/p/login", bytes.NewBuffer(reqBody), requestID)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
		}).Error("Failed to make session creation request")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
			"status":   resp.StatusCode,
		}).Error("Failed to read session creation response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// First, unmarshal the new response structure using shared Response
	var responseWrapper httpresponse.Response
	if err := json.Unmarshal(body, &responseWrapper); err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
			"status":   resp.StatusCode,
			"body":     string(body),
		}).Error("Failed to unmarshal session creation response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if the response indicates success (200-299 range)
	if responseWrapper.Code < 200 || responseWrapper.Code >= 300 {
		sm.logger.WithFields(logrus.Fields{
			"username":         req.Username,
			"status":           resp.StatusCode,
			"response_code":    responseWrapper.Code,
			"response_message": responseWrapper.Message,
			"body":             string(body),
		}).Error("Session creation failed")
		return nil, fmt.Errorf("session creation failed: %s", responseWrapper.Message)
	}

	// Extract the data from the response
	var createResp models.SessionCreateResponse
	if responseWrapper.Data != nil {
		if dataBytes, err := json.Marshal(responseWrapper.Data); err == nil {
			json.Unmarshal(dataBytes, &createResp)
		}
	}

	// Debug logging removed to reduce noise

	return &createResp, nil
}

// LogoutSession revokes a session
func (sm *SessionManager) LogoutSession(sessionId string, requestID string) error {
	// Debug logging removed to reduce noise

	req := models.SessionLogoutRequest{
		SessionID: sessionId,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionId,
		}).Error("Failed to marshal session logout request")
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := sm.makeRequest("POST", "/logout", bytes.NewBuffer(reqBody), requestID)
	if err != nil {
		sm.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionId,
		}).Error("Failed to make session logout request")
		return fmt.Errorf("failed to logout session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sm.logger.WithFields(logrus.Fields{
			"session_id": sessionId,
			"status":     resp.StatusCode,
			"body":       string(body),
		}).Error("Session logout failed with non-200 status")
		return fmt.Errorf("logout failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Debug logging removed to reduce noise

	return nil
}
