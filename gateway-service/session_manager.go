package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

// Session validation request/response structures
type SessionValidationRequest struct {
	Token string `json:"token"`
}

type SessionValidationResponse struct {
	IsValid       bool         `json:"is_valid"`
	Session       *SessionData `json:"session,omitempty"`
	ShouldRefresh bool         `json:"should_refresh"`
	NewToken      string       `json:"new_token,omitempty"`
	ErrorCode     string       `json:"error_code,omitempty"`
	ErrorMessage  string       `json:"error_message,omitempty"`
}

// Session creation request/response structures
type SessionCreateRequest struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	RoleName    string    `json:"role_name"`
	Permissions []string  `json:"permissions"`
	RememberMe  bool      `json:"remember_me"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

type SessionCreateResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	SessionID string      `json:"session_id"`
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      UserContext `json:"user"`
}

// Session data structure
type SessionData struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	RoleName     string    `json:"role_name"`
	Permissions  []string  `json:"permissions"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
}

// User context for adding to requests
type UserContext struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Session logout request
type SessionLogoutRequest struct {
	Token string `json:"token"`
}

// ValidateSession validates a token against the session service
func (sm *SessionManager) ValidateSession(token string) (*SessionValidationResponse, error) {
	if token == "" {
		return &SessionValidationResponse{
			IsValid:      false,
			ErrorCode:    "missing_token",
			ErrorMessage: "Token is required",
		}, nil
	}

	// For now, we'll continue using token validation
	// In the future, we can enhance this to extract session ID from JWT or store it separately
	validationReq := SessionValidationRequest{
		Token: token,
	}

	reqBody, err := json.Marshal(validationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: log the URL being called
	fmt.Printf("Gateway calling session service at: %s\n", sm.baseURL+"/validate")

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

	// Debug: log the response body
	fmt.Printf("Session service response: %s\n", string(body))

	var validationResp SessionValidationResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &validationResp, nil
}

// CreateSession creates a new session after successful login
func (sm *SessionManager) CreateSession(req *SessionCreateRequest) (*SessionCreateResponse, error) {
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

	var createResp SessionCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &createResp, nil
}

// LogoutSession revokes a session
func (sm *SessionManager) LogoutSession(token string) error {
	req := SessionLogoutRequest{
		Token: token,
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
