package requestlogger

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// GetRequestLogger creates a logger with request ID from context
func GetRequestLogger(baseLogger *logrus.Logger, r *http.Request) *logrus.Entry {
	if requestID := r.Context().Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return baseLogger.WithField("request_id", id)
		}
	}

	// Fallback to base logger if no request ID found
	return baseLogger.WithFields(logrus.Fields{})
}
