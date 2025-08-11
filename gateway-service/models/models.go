package models

// Session validation request/response structures
type SessionValidationRequest struct {
	SessionID string `json:"session_id"`
}

type SessionValidationResponse struct {
	Valid       bool     `json:"valid"`
	SessionID   string   `json:"session_id,omitempty"`
	Message     string   `json:"message,omitempty"`
	UserID      string   `json:"user_id,omitempty"`
	Username    string   `json:"username,omitempty"`
	RoleName    string   `json:"role_name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Session creation request/response structures
type SessionCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionCreateResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// Session logout request
type SessionLogoutRequest struct {
	SessionID string `json:"session_id"`
}
