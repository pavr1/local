package logger

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/sirupsen/logrus"
)

const (
	SERVICE_GATEWAY_SERVICE   = "gateway-service"
	SERVICE_DATA_SERVICE      = "data-service"
	SERVICE_INVENTORY_SERVICE = "inventory-service"
	SERVICE_INVOICE_SERVICE   = "invoice-service"
	SERVICE_ORDERS_SERVICE    = "orders-service"
	SERVICE_SESSION_SERVICE   = "session-service"
)

// CustomizeLogger creates an instance of the main logger with the request id if presented
func CustomizeLogger(logger *logrus.Logger, r *http.Request, service string) *logrus.Logger {
	var customizedLogger *logrus.Logger
	if r != nil {
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			customizedLogger = logger.WithFields(logrus.Fields{
				"service":    service,
				"request_id": requestID,
			}).Logger
		} else {
			customizedLogger = logger.WithFields(logrus.Fields{
				"service":    service,
				"request_id": "not-found",
			}).Logger
		}
	}

	return customizedLogger
}

// SetupLogger configures the logrus logger with consistent formatting
func SetupLogger(serviceName, logLevel string) *logrus.Logger {
	logger := logrus.NewEntry(logrus.New())

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.Logger.SetLevel(level)

	// Set log format with line numbers and better formatting
	logger.Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// Get the relative path from the project root
			// This will show the full relative path instead of just the filename
			// Extract the relative path by finding the project root
			filePath := f.File
			if idx := findProjectRoot(filePath); idx != -1 {
				filePath = filePath[idx:]
			}
			return "", fmt.Sprintf("%s:%d", filePath, f.Line)
		},
	})

	// Enable caller reporting for line numbers
	logger.Logger.SetReportCaller(true)
	logger = logger.WithFields(logrus.Fields{
		"service": serviceName,
	})

	return logger.Logger
}

// findProjectRoot finds the index of the project root in the file path
// It looks for common project directory names to identify the root
func findProjectRoot(filePath string) int {
	// Common project root indicators
	indicators := []string{
		"/gateway-service/",
		"/data-service/",
		"/inventory-service/",
		"/invoice-service/",
		"/orders-service/",
		"/session-service/",
		"/ui/",
		"/shared/",
	}

	for _, indicator := range indicators {
		if idx := findLastIndex(filePath, indicator); idx != -1 {
			return idx
		}
	}

	return -1
}

// findLastIndex finds the last occurrence of a substring in a string
func findLastIndex(s, substr string) int {
	lastIdx := -1
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			lastIdx = i
		}
	}
	return lastIdx
}
