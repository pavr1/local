package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser tests the User struct and its JSON serialization
func TestUser(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-1 * time.Hour)

	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed-password-here",
		FullName:     "Test User",
		RoleID:       "role-456",
		IsActive:     true,
		LastLogin:    &lastLogin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Test JSON marshaling - password should not be included
	jsonData, err := json.Marshal(user)
	require.NoError(t, err)

	// Verify password is not in JSON
	assert.NotContains(t, string(jsonData), "hashed-password-here")
	assert.NotContains(t, string(jsonData), "password_hash")

	// Verify other fields are present
	assert.Contains(t, string(jsonData), "user-123")
	assert.Contains(t, string(jsonData), "testuser")
	assert.Contains(t, string(jsonData), "Test User")

	// Test JSON unmarshaling
	var unmarshaled User
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, user.ID, unmarshaled.ID)
	assert.Equal(t, user.Username, unmarshaled.Username)
	assert.Equal(t, user.FullName, unmarshaled.FullName)
	assert.Equal(t, user.RoleID, unmarshaled.RoleID)
	assert.Equal(t, user.IsActive, unmarshaled.IsActive)

	// Password should not be unmarshaled
	assert.Empty(t, unmarshaled.PasswordHash)

	// Time fields
	assert.WithinDuration(t, user.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, user.UpdatedAt, unmarshaled.UpdatedAt, time.Second)
	require.NotNil(t, unmarshaled.LastLogin)
	assert.WithinDuration(t, *user.LastLogin, *unmarshaled.LastLogin, time.Second)
}

// TestUserWithNilLastLogin tests User with nil LastLogin
func TestUserWithNilLastLogin(t *testing.T) {
	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed-password",
		FullName:     "Test User",
		RoleID:       "role-456",
		IsActive:     true,
		LastLogin:    nil, // nil last login
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	jsonData, err := json.Marshal(user)
	require.NoError(t, err)

	var unmarshaled User
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.LastLogin)
}

// TestRole tests the Role struct
func TestRole(t *testing.T) {
	now := time.Now()

	role := &Role{
		ID:          "role-123",
		RoleName:    "admin",
		Description: "Administrator role with full access",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	jsonData, err := json.Marshal(role)
	require.NoError(t, err)

	var unmarshaled Role
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, role.ID, unmarshaled.ID)
	assert.Equal(t, role.RoleName, unmarshaled.RoleName)
	assert.Equal(t, role.Description, unmarshaled.Description)
	assert.WithinDuration(t, role.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, role.UpdatedAt, unmarshaled.UpdatedAt, time.Second)
}

// TestPermission tests the Permission struct
func TestPermission(t *testing.T) {
	now := time.Now()

	permission := &Permission{
		ID:             "perm-123",
		PermissionName: "user.create",
		Description:    "Create new users",
		RoleID:         "role-456",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	jsonData, err := json.Marshal(permission)
	require.NoError(t, err)

	var unmarshaled Permission
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, permission.ID, unmarshaled.ID)
	assert.Equal(t, permission.PermissionName, unmarshaled.PermissionName)
	assert.Equal(t, permission.Description, unmarshaled.Description)
	assert.Equal(t, permission.RoleID, unmarshaled.RoleID)
	assert.WithinDuration(t, permission.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, permission.UpdatedAt, unmarshaled.UpdatedAt, time.Second)
}

// TestUserProfile tests the UserProfile struct
func TestUserProfile(t *testing.T) {
	now := time.Now()

	user := User{
		ID:        "user-123",
		Username:  "testuser",
		FullName:  "Test User",
		RoleID:    "role-456",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	role := Role{
		ID:          "role-456",
		RoleName:    "admin",
		Description: "Administrator",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	permissions := []Permission{
		{
			ID:             "perm-1",
			PermissionName: "user.create",
			Description:    "Create users",
			RoleID:         "role-456",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "perm-2",
			PermissionName: "user.delete",
			Description:    "Delete users",
			RoleID:         "role-456",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	profile := &UserProfile{
		User:        user,
		Role:        role,
		Permissions: permissions,
	}

	jsonData, err := json.Marshal(profile)
	require.NoError(t, err)

	var unmarshaled UserProfile
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, profile.User.ID, unmarshaled.User.ID)
	assert.Equal(t, profile.Role.RoleName, unmarshaled.Role.RoleName)
	assert.Len(t, unmarshaled.Permissions, 2)
	assert.Equal(t, profile.Permissions[0].PermissionName, unmarshaled.Permissions[0].PermissionName)
	assert.Equal(t, profile.Permissions[1].PermissionName, unmarshaled.Permissions[1].PermissionName)
}

// TestLoginRequest tests the LoginRequest struct
func TestLoginRequest(t *testing.T) {
	request := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled LoginRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Username, unmarshaled.Username)
	assert.Equal(t, request.Password, unmarshaled.Password)
}

// TestLoginResponse tests the LoginResponse struct
func TestLoginResponse(t *testing.T) {
	now := time.Now()

	user := User{
		ID:        "user-123",
		Username:  "testuser",
		FullName:  "Test User",
		RoleID:    "role-456",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	role := Role{
		ID:          "role-456",
		RoleName:    "user",
		Description: "Regular user",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	response := &LoginResponse{
		User:  user,
		Role:  role,
		Token: "jwt-token-here",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled LoginResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.User.ID, unmarshaled.User.ID)
	assert.Equal(t, response.Role.RoleName, unmarshaled.Role.RoleName)
	assert.Equal(t, response.Token, unmarshaled.Token)
}

// TestRefreshTokenRequest tests the RefreshTokenRequest struct
func TestRefreshTokenRequest(t *testing.T) {
	request := &RefreshTokenRequest{
		Token: "existing-jwt-token",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled RefreshTokenRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Token, unmarshaled.Token)
}

// TestRefreshTokenResponse tests the RefreshTokenResponse struct
func TestRefreshTokenResponse(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	refreshAt := now.Add(25 * time.Minute)

	response := &RefreshTokenResponse{
		Token:     "new-jwt-token",
		ExpiresAt: expiresAt,
		RefreshAt: refreshAt,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled RefreshTokenResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.Token, unmarshaled.Token)
	assert.WithinDuration(t, response.ExpiresAt, unmarshaled.ExpiresAt, time.Second)
	assert.WithinDuration(t, response.RefreshAt, unmarshaled.RefreshAt, time.Second)
}

// TestLogoutRequest tests the LogoutRequest struct
func TestLogoutRequest(t *testing.T) {
	request := &LogoutRequest{
		Token: "jwt-token-to-logout",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled LogoutRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Token, unmarshaled.Token)
}

// TestJWTClaims tests the JWTClaims struct
func TestJWTClaims(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	claims := &JWTClaims{
		UserID:      "user-123",
		Username:    "testuser",
		FullName:    "Test User",
		RoleID:      "role-456",
		RoleName:    "admin",
		Permissions: []string{"read", "write", "delete"},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   "user-123",
			Issuer:    "test-service",
			Audience:  []string{"test-audience"},
		},
	}

	jsonData, err := json.Marshal(claims)
	require.NoError(t, err)

	var unmarshaled JWTClaims
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, claims.UserID, unmarshaled.UserID)
	assert.Equal(t, claims.Username, unmarshaled.Username)
	assert.Equal(t, claims.FullName, unmarshaled.FullName)
	assert.Equal(t, claims.RoleID, unmarshaled.RoleID)
	assert.Equal(t, claims.RoleName, unmarshaled.RoleName)
	assert.Equal(t, claims.Permissions, unmarshaled.Permissions)
	assert.Equal(t, claims.Subject, unmarshaled.Subject)
	assert.Equal(t, claims.Issuer, unmarshaled.Issuer)
	assert.Equal(t, claims.Audience, unmarshaled.Audience)

	// Check time fields
	assert.Equal(t, claims.IssuedAt.Unix(), unmarshaled.IssuedAt.Unix())
	assert.Equal(t, claims.ExpiresAt.Unix(), unmarshaled.ExpiresAt.Unix())
}

// TestErrorResponse tests the ErrorResponse struct
func TestErrorResponse(t *testing.T) {
	response := &ErrorResponse{
		Error:   "INVALID_CREDENTIALS",
		Message: "Username or password is incorrect",
		Code:    "AUTH_001",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled ErrorResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.Error, unmarshaled.Error)
	assert.Equal(t, response.Message, unmarshaled.Message)
	assert.Equal(t, response.Code, unmarshaled.Code)
}

// TestSuccessResponse tests the SuccessResponse struct
func TestSuccessResponse(t *testing.T) {
	data := map[string]interface{}{
		"user_id": "123",
		"count":   42,
	}

	response := &SuccessResponse{
		Success: true,
		Message: "Operation completed successfully",
		Data:    data,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled SuccessResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.Success, unmarshaled.Success)
	assert.Equal(t, response.Message, unmarshaled.Message)

	// Data is interface{}, so we need to type assert for comparison
	dataMap, ok := unmarshaled.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123", dataMap["user_id"])
	assert.Equal(t, float64(42), dataMap["count"]) // JSON numbers become float64
}

// TestValidationError tests the ValidationError struct
func TestValidationError(t *testing.T) {
	validationError := &ValidationError{
		Field:   "username",
		Message: "Username must be at least 3 characters",
	}

	jsonData, err := json.Marshal(validationError)
	require.NoError(t, err)

	var unmarshaled ValidationError
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, validationError.Field, unmarshaled.Field)
	assert.Equal(t, validationError.Message, unmarshaled.Message)
}

// TestValidationErrorResponse tests the ValidationErrorResponse struct
func TestValidationErrorResponse(t *testing.T) {
	errors := []ValidationError{
		{Field: "username", Message: "Username is required"},
		{Field: "password", Message: "Password must be at least 6 characters"},
	}

	response := &ValidationErrorResponse{
		Error:  "VALIDATION_FAILED",
		Errors: errors,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled ValidationErrorResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.Error, unmarshaled.Error)
	assert.Len(t, unmarshaled.Errors, 2)
	assert.Equal(t, response.Errors[0].Field, unmarshaled.Errors[0].Field)
	assert.Equal(t, response.Errors[0].Message, unmarshaled.Errors[0].Message)
	assert.Equal(t, response.Errors[1].Field, unmarshaled.Errors[1].Field)
	assert.Equal(t, response.Errors[1].Message, unmarshaled.Errors[1].Message)
}

// TestAuthStatus tests the AuthStatus struct
func TestAuthStatus(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	refreshAt := now.Add(25 * time.Minute)

	user := &User{
		ID:       "user-123",
		Username: "testuser",
		FullName: "Test User",
		IsActive: true,
	}

	role := &Role{
		ID:       "role-456",
		RoleName: "admin",
	}

	authStatus := &AuthStatus{
		IsAuthenticated: true,
		User:            user,
		Role:            role,
		ExpiresAt:       expiresAt,
		RefreshAt:       refreshAt,
	}

	jsonData, err := json.Marshal(authStatus)
	require.NoError(t, err)

	var unmarshaled AuthStatus
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, authStatus.IsAuthenticated, unmarshaled.IsAuthenticated)
	require.NotNil(t, unmarshaled.User)
	assert.Equal(t, authStatus.User.ID, unmarshaled.User.ID)
	require.NotNil(t, unmarshaled.Role)
	assert.Equal(t, authStatus.Role.RoleName, unmarshaled.Role.RoleName)
	assert.WithinDuration(t, authStatus.ExpiresAt, unmarshaled.ExpiresAt, time.Second)
	assert.WithinDuration(t, authStatus.RefreshAt, unmarshaled.RefreshAt, time.Second)
}

// TestAuthStatusUnauthenticated tests AuthStatus for unauthenticated user
func TestAuthStatusUnauthenticated(t *testing.T) {
	authStatus := &AuthStatus{
		IsAuthenticated: false,
		User:            nil,
		Role:            nil,
	}

	jsonData, err := json.Marshal(authStatus)
	require.NoError(t, err)

	var unmarshaled AuthStatus
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.False(t, unmarshaled.IsAuthenticated)
	assert.Nil(t, unmarshaled.User)
	assert.Nil(t, unmarshaled.Role)
}

// TestTokenInfo tests the TokenInfo struct
func TestTokenInfo(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	tokenInfo := &TokenInfo{
		Valid:       true,
		UserID:      "user-123",
		Username:    "testuser",
		RoleName:    "admin",
		Permissions: []string{"read", "write", "delete"},
		IssuedAt:    now,
		ExpiresAt:   expiresAt,
		Error:       "",
	}

	jsonData, err := json.Marshal(tokenInfo)
	require.NoError(t, err)

	var unmarshaled TokenInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, tokenInfo.Valid, unmarshaled.Valid)
	assert.Equal(t, tokenInfo.UserID, unmarshaled.UserID)
	assert.Equal(t, tokenInfo.Username, unmarshaled.Username)
	assert.Equal(t, tokenInfo.RoleName, unmarshaled.RoleName)
	assert.Equal(t, tokenInfo.Permissions, unmarshaled.Permissions)
	assert.WithinDuration(t, tokenInfo.IssuedAt, unmarshaled.IssuedAt, time.Second)
	assert.WithinDuration(t, tokenInfo.ExpiresAt, unmarshaled.ExpiresAt, time.Second)
	assert.Equal(t, tokenInfo.Error, unmarshaled.Error)
}

// TestTokenInfoInvalid tests TokenInfo for invalid token
func TestTokenInfoInvalid(t *testing.T) {
	tokenInfo := &TokenInfo{
		Valid: false,
		Error: "Token is expired",
	}

	jsonData, err := json.Marshal(tokenInfo)
	require.NoError(t, err)

	var unmarshaled TokenInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.False(t, unmarshaled.Valid)
	assert.Equal(t, tokenInfo.Error, unmarshaled.Error)
	assert.Empty(t, unmarshaled.UserID)
	assert.Empty(t, unmarshaled.Username)
}

// TestJWTClaimsValidation tests JWT claims validation
func TestJWTClaimsValidation(t *testing.T) {
	now := time.Now()

	// Test valid claims
	validClaims := &JWTClaims{
		UserID:   "user-123",
		Username: "testuser",
		RoleName: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			Subject:   "user-123",
		},
	}

	// Test that required fields are present
	assert.NotEmpty(t, validClaims.UserID)
	assert.NotEmpty(t, validClaims.Username)
	assert.NotEmpty(t, validClaims.Subject)
	assert.NotNil(t, validClaims.IssuedAt)
	assert.NotNil(t, validClaims.ExpiresAt)

	// Test that expiration is in the future
	assert.True(t, validClaims.ExpiresAt.Time.After(now))

	// Test that issued at is not in the future
	assert.True(t, validClaims.IssuedAt.Time.Before(now.Add(time.Minute)))
}

// TestEmptyPermissions tests handling of empty permissions
func TestEmptyPermissions(t *testing.T) {
	claims := &JWTClaims{
		UserID:      "user-123",
		Username:    "testuser",
		Permissions: []string{}, // empty permissions
	}

	jsonData, err := json.Marshal(claims)
	require.NoError(t, err)

	var unmarshaled JWTClaims
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.NotNil(t, unmarshaled.Permissions)
	assert.Len(t, unmarshaled.Permissions, 0)
}

// TestNilPermissions tests handling of nil permissions
func TestNilPermissions(t *testing.T) {
	claims := &JWTClaims{
		UserID:      "user-123",
		Username:    "testuser",
		Permissions: nil, // nil permissions
	}

	jsonData, err := json.Marshal(claims)
	require.NoError(t, err)

	var unmarshaled JWTClaims
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.Permissions)
}

// BenchmarkUserMarshal benchmarks User JSON marshaling
func BenchmarkUserMarshal(b *testing.B) {
	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed-password",
		FullName:     "Test User",
		RoleID:       "role-456",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJWTClaimsMarshal benchmarks JWTClaims JSON marshaling
func BenchmarkJWTClaimsMarshal(b *testing.B) {
	now := time.Now()

	claims := &JWTClaims{
		UserID:      "user-123",
		Username:    "testuser",
		FullName:    "Test User",
		RoleID:      "role-456",
		RoleName:    "admin",
		Permissions: []string{"read", "write", "delete", "admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			Subject:   "user-123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(claims)
		if err != nil {
			b.Fatal(err)
		}
	}
}
