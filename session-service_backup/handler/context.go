package handler

import (
	"context"
	"session-service/models"
)

// Context keys for session data
type contextKey string

const (
	sessionContextKey contextKey = "session"
)

// addSessionToContext adds session data to the request context
func addSessionToContext(ctx context.Context, session *models.SessionData) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

// getSessionFromContext retrieves session data from the request context
func getSessionFromContext(ctx context.Context) *models.SessionData {
	session, ok := ctx.Value(sessionContextKey).(*models.SessionData)
	if !ok {
		return nil
	}
	return session
}
