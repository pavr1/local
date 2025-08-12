package fluentd

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// FluentdClient represents a client for sending logs to Fluentd
type FluentdClient struct {
	conn      net.Conn
	host      string
	port      int
	tag       string
	service   string
	reconnect bool
}

// LogEntry represents a log entry to be sent to Fluentd
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
}

// NewFluentdClient creates a new Fluentd client
func NewFluentdClient(host string, port int, service string) *FluentdClient {
	return &FluentdClient{
		host:      host,
		port:      port,
		tag:       fmt.Sprintf("service.%s", service),
		service:   service,
		reconnect: true,
	}
}

// Connect establishes connection to Fluentd
func (fc *FluentdClient) Connect() error {
	var err error
	fc.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", fc.host, fc.port))
	if err != nil {
		return fmt.Errorf("failed to connect to Fluentd: %v", err)
	}
	return nil
}

// Close closes the connection to Fluentd
func (fc *FluentdClient) Close() error {
	if fc.conn != nil {
		return fc.conn.Close()
	}
	return nil
}

// SendLog sends a log entry to Fluentd
func (fc *FluentdClient) SendLog(level, message string, fields map[string]interface{}) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Service:   fc.service,
		Message:   message,
		Fields:    fields,
	}

	// Try to send the log
	err := fc.sendEntry(entry)
	if err != nil && fc.reconnect {
		// Try to reconnect and send again
		fc.Close()
		if err := fc.Connect(); err == nil {
			err = fc.sendEntry(entry)
		}
	}

	return err
}

// sendEntry sends a single log entry
func (fc *FluentdClient) sendEntry(entry LogEntry) error {
	if fc.conn == nil {
		return fmt.Errorf("not connected to Fluentd")
	}

	// Create the message to send
	message := map[string]interface{}{
		"tag":    fc.tag,
		"time":   entry.Timestamp.Unix(),
		"record": entry,
	}

	// Encode as JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	// Send to Fluentd
	_, err = fc.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send log to Fluentd: %v", err)
	}

	return nil
}

// Convenience methods
func (fc *FluentdClient) Debug(message string, fields map[string]interface{}) error {
	return fc.SendLog("DEBUG", message, fields)
}

func (fc *FluentdClient) Info(message string, fields map[string]interface{}) error {
	return fc.SendLog("INFO", message, fields)
}

func (fc *FluentdClient) Warn(message string, fields map[string]interface{}) error {
	return fc.SendLog("WARN", message, fields)
}

func (fc *FluentdClient) Error(message string, fields map[string]interface{}) error {
	return fc.SendLog("ERROR", message, fields)
}

func (fc *FluentdClient) Fatal(message string, fields map[string]interface{}) error {
	return fc.SendLog("FATAL", message, fields)
}

// LogrusHook integrates with logrus
type LogrusHook struct {
	client *FluentdClient
}

// NewLogrusHook creates a new logrus hook for Fluentd
func NewLogrusHook(client *FluentdClient) *LogrusHook {
	return &LogrusHook{client: client}
}

// Fire implements the logrus.Hook interface
func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	fields := make(map[string]interface{})
	for k, v := range entry.Data {
		fields[k] = v
	}

	return h.client.SendLog(entry.Level.String(), entry.Message, fields)
}

// Levels implements the logrus.Hook interface
func (h *LogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
