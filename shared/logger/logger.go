package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel represents log levels
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

const (
	SERVICE_GATEWAY_SERVICE   = "gateway-service"
	SERVICE_DATA_SERVICE      = "data-service"
	SERVICE_INVENTORY_SERVICE = "inventory-service"
	SERVICE_INVOICE_SERVICE   = "invoice-service"
	SERVICE_ORDERS_SERVICE    = "orders-service"
	SERVICE_SESSION_SERVICE   = "session-service"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
}

// CentralizedLogger manages centralized logging
type CentralizedLogger struct {
	conn      net.Conn
	host      string
	port      int
	service   string
	connected bool
	reconnect bool
}

// Global instance
var (
	centralLogger *CentralizedLogger
)

// InitLogger initializes the centralized logger
func InitLogger(serviceName, fluentdHost string, fluentdPort int) *CentralizedLogger {
	if centralLogger == nil {
		centralLogger = &CentralizedLogger{
			host:      fluentdHost,
			port:      fluentdPort,
			service:   serviceName,
			reconnect: true,
		}

		// Try to connect
		if err := centralLogger.Connect(); err != nil {
			log.Printf("Warning: Could not connect to Fluentd: %v (falling back to local logging)", err)
		} else {
			log.Printf("Connected to centralized logging at %s:%d", fluentdHost, fluentdPort)
		}
	}

	return centralLogger
}

// GetLogger returns the global logger instance
func GetLogger() *CentralizedLogger {
	if centralLogger == nil {
		// Initialize with defaults
		InitLogger("unknown", "localhost", 24224)
	}
	return centralLogger
}

// Connect establishes connection to Fluentd
func (cl *CentralizedLogger) Connect() error {
	var err error
	cl.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", cl.host, cl.port))
	if err != nil {
		cl.connected = false
		return fmt.Errorf("failed to connect to Fluentd: %v", err)
	}
	cl.connected = true
	return nil
}

// Close closes the connection to Fluentd
func (cl *CentralizedLogger) Close() error {
	if cl.conn != nil {
		cl.connected = false
		return cl.conn.Close()
	}
	return nil
}

// Log sends a log entry to Fluentd
func (cl *CentralizedLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     string(level),
		Service:   cl.service,
		Message:   message,
		Fields:    fields,
	}

	// Try to send to Fluentd if connected
	if cl.connected {
		if err := cl.sendToFluentd(entry); err != nil {
			// Try to reconnect and send again
			if cl.reconnect {
				cl.Close()
				if err := cl.Connect(); err == nil {
					cl.sendToFluentd(entry)
				}
			}
		}
	}

	// Always log locally as fallback
	cl.logLocally(entry)
}

// sendToFluentd sends a log entry to Fluentd
func (cl *CentralizedLogger) sendToFluentd(entry LogEntry) error {
	if cl.conn == nil {
		return fmt.Errorf("not connected to Fluentd")
	}

	// Create the message to send (Fluentd forward protocol format)
	message := map[string]interface{}{
		"tag":    fmt.Sprintf("service.%s", cl.service),
		"time":   entry.Timestamp.Unix(),
		"record": entry,
	}

	// Encode as JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	// Send to Fluentd
	_, err = cl.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send log to Fluentd: %v", err)
	}

	return nil
}

// logLocally logs to local file as fallback
func (cl *CentralizedLogger) logLocally(entry LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s [%s] %s", timestamp, entry.Level, entry.Service, entry.Message)

	// Also log to file
	if logFile, err := os.OpenFile(fmt.Sprintf("%s.log", cl.service), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		defer logFile.Close()
		fmt.Fprintf(logFile, "[%s] %s [%s] %s\n", timestamp, entry.Level, entry.Service, entry.Message)
	}
}

// Convenience methods
func (cl *CentralizedLogger) Debug(message string, fields map[string]interface{}) {
	cl.Log(DEBUG, message, fields)
}

func (cl *CentralizedLogger) Info(message string, fields map[string]interface{}) {
	cl.Log(INFO, message, fields)
}

func (cl *CentralizedLogger) Warn(message string, fields map[string]interface{}) {
	cl.Log(WARN, message, fields)
}

func (cl *CentralizedLogger) Error(message string, fields map[string]interface{}) {
	cl.Log(ERROR, message, fields)
}

func (cl *CentralizedLogger) Fatal(message string, fields map[string]interface{}) {
	cl.Log(FATAL, message, fields)
	os.Exit(1)
}

// IsConnected returns whether the logger is connected to Fluentd
func (cl *CentralizedLogger) IsConnected() bool {
	return cl.connected
}

// GetRequestLogger creates a logger with request ID from context
func GetRequestLogger(r *http.Request, service string) *logrus.Entry {
	//pvillalobos - hardcoded value
	logger := setupLogger("info") // Default log level

	if logger != nil {
		if requestID := r.Context().Value("request_id"); requestID != nil {
			if id, ok := requestID.(string); ok {
				return logger.WithField("request_id", id)
			}
		}
	}
	// Fallback to base logger if no request ID found
	return logger.WithFields(logrus.Fields{"service": service})
}

// SetupLogger configures the logrus logger with consistent formatting
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format with line numbers and better formatting
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	// Enable caller reporting for line numbers
	logger.SetReportCaller(true)

	return logger
}
