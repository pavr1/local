package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"orders-service/utils"

	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	logger     *logrus.Logger
}

type contextKey string

const ClaimsContextKey contextKey = "claims"

func NewAuthMiddleware(jwtManager *utils.JWTManager, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Authenticate validates JWT token and adds claims to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Check if token starts with "Bearer "
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.logger.WithError(err).Warn("Token validation failed")
			m.respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if user has specific permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*utils.Claims)
			if !ok {
				m.respondWithError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !claims.HasPermission(permission) && !claims.IsAdmin() {
				m.logger.WithFields(logrus.Fields{
					"user_id":    claims.UserID,
					"permission": permission,
					"user_perms": claims.Permissions,
				}).Warn("Permission denied")
				m.respondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrdersPermission checks if user has orders-specific permission
func (m *AuthMiddleware) RequireOrdersPermission(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*utils.Claims)
			if !ok {
				m.respondWithError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !claims.HasOrdersPermission(action) {
				m.logger.WithFields(logrus.Fields{
					"user_id":         claims.UserID,
					"required_action": action,
					"user_perms":      claims.Permissions,
				}).Warn("Orders permission denied")
				m.respondWithError(w, http.StatusForbidden, "Insufficient orders permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly restricts access to admin users only
func (m *AuthMiddleware) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*utils.Claims)
		if !ok {
			m.respondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		if !claims.IsAdmin() {
			m.logger.WithFields(logrus.Fields{
				"user_id":   claims.UserID,
				"role_name": claims.RoleName,
			}).Warn("Admin access denied")
			m.respondWithError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORS middleware
func (m *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func (m *AuthMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		// Log request details
		duration := time.Since(start)
		m.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": rw.statusCode,
			"duration":    duration,
			"user_agent":  r.UserAgent(),
			"remote_addr": r.RemoteAddr,
		}).Info("HTTP request")
	})
}

// Helper functions

func (m *AuthMiddleware) respondWithError(w http.ResponseWriter, status int, message string) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(r *http.Request) (*utils.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsContextKey).(*utils.Claims)
	return claims, ok
}

// responseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
