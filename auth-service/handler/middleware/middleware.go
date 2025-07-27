package middleware

import (
	"context"
	"fmt"
	"net/http"

	"auth-service/models"
	"auth-service/utils"

	"github.com/sirupsen/logrus"
)

// AuthMiddleware provides authentication middleware functionality
type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	logger     *logrus.Logger
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtManager *utils.JWTManager, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Authenticate validates JWT tokens and adds user context to requests
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := m.extractTokenFromHeader(r)
		if token == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "missing_token", "Authorization token is required")
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "invalid_token", err.Error())
			return
		}

		// Add user claims to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.RoleName)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission creates middleware that checks for a specific permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context (should be set by Authenticate middleware)
			claims, ok := r.Context().Value("user").(*models.JWTClaims)
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "missing_auth_context", "Authentication context is missing")
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, userPerm := range claims.Permissions {
				if userPerm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.logger.WithFields(logrus.Fields{
					"user_id":             claims.UserID,
					"username":            claims.Username,
					"required_permission": permission,
					"user_permissions":    claims.Permissions,
				}).Warn("Access denied: insufficient permissions")

				m.writeErrorResponse(w, http.StatusForbidden, "insufficient_permissions",
					fmt.Sprintf("Required permission '%s' not found", permission))
				return
			}

			// User has permission, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates middleware that checks for a specific role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context (should be set by Authenticate middleware)
			claims, ok := r.Context().Value("user").(*models.JWTClaims)
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "missing_auth_context", "Authentication context is missing")
				return
			}

			// Check if user has the required role
			if claims.RoleName != role {
				m.logger.WithFields(logrus.Fields{
					"user_id":       claims.UserID,
					"username":      claims.Username,
					"user_role":     claims.RoleName,
					"required_role": role,
				}).Warn("Access denied: insufficient role")

				m.writeErrorResponse(w, http.StatusForbidden, "insufficient_role",
					fmt.Sprintf("Required role '%s' not found", role))
				return
			}

			// User has required role, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates middleware that checks for any of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context (should be set by Authenticate middleware)
			claims, ok := r.Context().Value("user").(*models.JWTClaims)
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "missing_auth_context", "Authentication context is missing")
				return
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range claims.Permissions {
					if userPerm == requiredPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				m.logger.WithFields(logrus.Fields{
					"user_id":              claims.UserID,
					"username":             claims.Username,
					"required_permissions": permissions,
					"user_permissions":     claims.Permissions,
				}).Warn("Access denied: insufficient permissions")

				m.writeErrorResponse(w, http.StatusForbidden, "insufficient_permissions",
					fmt.Sprintf("One of the required permissions %v not found", permissions))
				return
			}

			// User has required permission, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

// extractTokenFromHeader extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return ""
	}

	return authHeader[7:]
}

// writeErrorResponse writes an error response
func (m *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	response := models.ErrorResponse{
		Error:   code,
		Message: message,
		Code:    code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Simple JSON encoding to avoid import issues
	jsonResponse := fmt.Sprintf(`{"error":"%s","message":"%s","code":"%s"}`,
		response.Error, response.Message, response.Code)
	w.Write([]byte(jsonResponse))
}
