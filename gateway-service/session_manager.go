package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway-service/models"
)

// SessionManager handles communication with the session service
type SessionManager struct {
	baseURL string
	client  *http.Client
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionServiceURL string) *SessionManager {
	return &SessionManager{
		baseURL: sessionServiceURL + "/api/v1/sessions",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateSession validates a session ID against the session service
func (sm *SessionManager) ValidateSession(sessionId string) (*models.SessionValidationResponse, error) {
	if sessionId == "" {
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", sm.baseURL+"/validate", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	httpReq.Header.Set("X-Gateway-Session-Managed", "true")

	resp, err := sm.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var validationResp models.SessionValidationResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &validationResp, nil
}

// CreateSession creates a new session after successful login
func (sm *SessionManager) CreateSession(req *models.SessionCreateRequest) (*models.SessionCreateResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", sm.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	httpReq.Header.Set("X-Gateway-Session-Managed", "true")

	resp, err := sm.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("session creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp models.SessionCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &createResp, nil
}

// LogoutSession revokes a session
func (sm *SessionManager) LogoutSession(sessionId string) error {
	req := models.SessionLogoutRequest{
		SessionID: sessionId,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", sm.baseURL+"/logout", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	httpReq.Header.Set("X-Gateway-Session-Managed", "true")

	resp, err := sm.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to logout session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
