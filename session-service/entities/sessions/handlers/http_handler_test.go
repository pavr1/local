package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"session-service/entities/sessions/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDBHandler is a mock implementation of the DBHandler for testing
type MockDBHandler struct {
	createSessionResponse   *models.SessionCreateResponse
	createSessionError      error
	validateSessionResponse *models.SessionValidationResponse
	validateSessionError    error
	logoutSessionResponse   *models.SessionLogoutResponse
	logoutSessionError      error
}

func (m *MockDBHandler) CreateSession(req *models.SessionCreateRequest) (*models.SessionCreateResponse, error) {
	return m.createSessionResponse, m.createSessionError
}

func (m *MockDBHandler) ValidateSession(sessionID string) (*models.SessionValidationResponse, error) {
	return m.validateSessionResponse, m.validateSessionError
}

func (m *MockDBHandler) DeleteSession(sessionID string) (*models.SessionLogoutResponse, error) {
	return m.logoutSessionResponse, m.logoutSessionError
}

func (m *MockDBHandler) Close() error {
	return nil
}

func TestNewHTTPHandler(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}

	handler := NewHTTPHandler(mockDBHandler, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDBHandler, handler.dbHandler)
	assert.Equal(t, logger, handler.logger)
}

func TestCreateSession_Success(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		createSessionResponse: &models.SessionCreateResponse{
			SessionID: "session123",
			Message:   "Session created successfully",
		},
	}

	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create request body
	reqBody := models.SessionCreateRequest{
		Username: "testuser",
		Password: "testpass",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/sessions/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.SessionCreateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "session123", response.SessionID)
	assert.Equal(t, "Session created successfully", response.Message)
}

func TestCreateSession_InvalidRequest(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create invalid request body
	req := httptest.NewRequest("POST", "/api/v1/sessions/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestCreateSession_MissingCredentials(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create request with missing credentials
	reqBody := models.SessionCreateRequest{
		Username: "",
		Password: "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "missing_credentials", response.Error)
}

func TestCreateSession_DBError(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		createSessionError: assert.AnError,
	}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create valid request body
	reqBody := models.SessionCreateRequest{
		Username: "testuser",
		Password: "testpass",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "authentication_failed", response.Error)
}

func TestValidateSession_Success(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		validateSessionResponse: &models.SessionValidationResponse{
			Valid:       true,
			SessionID:   "session123",
			Message:     "Session validated",
			UserID:      "user123",
			Username:    "testuser",
			RoleName:    "admin",
			Permissions: []string{"read", "write"},
		},
	}

	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create request body
	reqBody := models.SessionValidationRequest{
		SessionID: "session123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/sessions/p/validate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.ValidateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SessionValidationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Valid)
	assert.Equal(t, "session123", response.SessionID)
	assert.Equal(t, "Session validated", response.Message)
	assert.Equal(t, "user123", response.UserID)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "admin", response.RoleName)
	assert.Equal(t, []string{"read", "write"}, response.Permissions)
}

func TestValidateSession_InvalidRequest(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create invalid request body
	req := httptest.NewRequest("POST", "/api/v1/sessions/p/validate", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.ValidateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestValidateSession_DBError(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		validateSessionError: assert.AnError,
	}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create valid request body
	reqBody := models.SessionValidationRequest{
		SessionID: "session123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/p/validate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.ValidateSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
}

func TestLogoutSession_Success(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		logoutSessionResponse: &models.SessionLogoutResponse{
			Success:   true,
			SessionID: "session123",
			Message:   "Session successfully logged out",
		},
	}

	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create request body
	reqBody := models.SessionLogoutRequest{
		SessionID: "session123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/sessions/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.LogoutSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SessionLogoutResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "session123", response.SessionID)
	assert.Equal(t, "Session successfully logged out", response.Message)
}

func TestLogoutSession_InvalidRequest(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create invalid request body
	req := httptest.NewRequest("POST", "/api/v1/sessions/logout", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.LogoutSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestLogoutSession_MissingSessionID(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create request with missing session ID
	reqBody := models.SessionLogoutRequest{
		SessionID: "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.LogoutSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "missing_session_id", response.Error)
}

func TestLogoutSession_DBError(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		logoutSessionError: assert.AnError,
	}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create valid request body
	reqBody := models.SessionLogoutRequest{
		SessionID: "session123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.LogoutSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "logout_failed", response.Error)
}

func TestLogoutSession_SessionNotFound(t *testing.T) {
	logger := logrus.New()
	mockDBHandler := &MockDBHandler{
		logoutSessionResponse: &models.SessionLogoutResponse{
			Success:   false,
			SessionID: "session123",
			Message:   "Session not found",
		},
	}
	handler := NewHTTPHandler(mockDBHandler, logger)

	// Create valid request body
	reqBody := models.SessionLogoutRequest{
		SessionID: "session123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/sessions/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.LogoutSession(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.SessionLogoutResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Session not found", response.Message)
}
