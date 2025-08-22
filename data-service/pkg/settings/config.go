package settings

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"

	sqlqueries "data-service/pkg/settings/sql"
)

// ServiceConfig holds configuration for a specific service
type ServiceConfig struct {
	settings map[string]string
	mutex    sync.RWMutex
	logger   *logrus.Logger
	db       *sql.DB
}

// NewServiceConfig creates a new service configuration instance
func NewServiceConfig(serviceName string, logger *logrus.Logger) (*ServiceConfig, error) {
	config := &ServiceConfig{
		settings: make(map[string]string),
		logger:   logger,
	}

	// Connect to database with default credentials
	if err := config.connectToDatabase(); err != nil {
		logger.WithError(err).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer config.db.Close()

	// Load all settings (general + service specific)
	if err := config.loadAllSettings(serviceName); err != nil {
		logger.WithError(err).Error("Failed to load all settings")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"service": serviceName,
		"count":   len(config.settings),
	}).Info("Service configuration loaded successfully")

	return config, nil
}

// connectToDatabase establishes connection to PostgreSQL database with default credentials
func (sc *ServiceConfig) connectToDatabase() error {
	// Default database connection parameters
	//pvillalobos - change this
	dsn := "host=localhost port=5432 user=postgres password=postgres123 dbname=icecream_store sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to open database connection")
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		sc.logger.WithError(err).Error("Failed to ping database")
		return fmt.Errorf("failed to ping database: %w", err)
	}

	sc.db = db
	sc.logger.Info("Database connection established for configuration loading")
	return nil
}

// loadAllSettings loads both general and service-specific settings
func (sc *ServiceConfig) loadAllSettings(serviceName string) error {
	// Load general settings first
	generalSettings, err := sc.loadServiceSettings("General")
	if err != nil {
		sc.logger.WithError(err).Error("Failed to load general settings")
		return fmt.Errorf("failed to load general settings: %w", err)
	}

	// Load service-specific settings
	serviceSettings, err := sc.loadServiceSettings(serviceName)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to load service settings")
		return fmt.Errorf("failed to load service settings: %w", err)
	}

	// Merge settings (service-specific settings override general settings)
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	for k, v := range generalSettings {
		sc.settings[k] = v
	}
	for k, v := range serviceSettings {
		sc.settings[k] = v
	}

	sc.logger.WithFields(logrus.Fields{
		"service":          serviceName,
		"total":            len(sc.settings),
		"general":          len(generalSettings),
		"service_specific": len(serviceSettings),
	}).Info("All settings loaded and merged")

	return nil
}

// loadServiceSettings loads settings for a specific service from database
func (sc *ServiceConfig) loadServiceSettings(serviceName string) (map[string]string, error) {
	sc.logger.WithField("service", serviceName).Info("Loading settings from database")

	rows, err := sc.db.Query(sqlqueries.GetSettingsByServiceQuery, serviceName)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to query settings")
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	settingsMap := make(map[string]string)
	for rows.Next() {
		var setting Setting
		err := rows.Scan(
			&setting.SettingID,
			&setting.Service,
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			sc.logger.WithError(err).Error("Failed to scan setting row")
			return nil, fmt.Errorf("failed to scan setting row: %w", err)
		}
		settingsMap[setting.Key] = setting.Value
	}

	if err = rows.Err(); err != nil {
		sc.logger.WithError(err).Error("Error iterating settings rows")
		return nil, fmt.Errorf("error iterating settings rows: %w", err)
	}

	sc.logger.WithFields(logrus.Fields{
		"service": serviceName,
		"count":   len(settingsMap),
	}).Info("Settings loaded from database")

	return settingsMap, nil
}

// Get retrieves a setting value by key
func (sc *ServiceConfig) Get(key string) (string, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	value, exists := sc.settings[key]
	return value, exists
}

// GetOrDefault retrieves a setting value by key with a default fallback
func (sc *ServiceConfig) GetOrDefault(key, defaultValue string) string {
	if value, exists := sc.Get(key); exists {
		return value
	}
	return defaultValue
}

// GetInt retrieves a setting value as int
func (sc *ServiceConfig) GetInt(key string) (int, error) {
	value, exists := sc.Get(key)
	if !exists {
		sc.logger.WithField("key", key).Error("Setting not found")
		return 0, fmt.Errorf("setting not found: %s", key)
	}

	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		sc.logger.WithError(err).WithField("key", key).Error("Failed to parse setting as int")
		return 0, fmt.Errorf("failed to parse setting %s as int: %w", key, err)
	}

	return intValue, nil
}

// GetIntOrDefault retrieves a setting value as int with a default fallback
func (sc *ServiceConfig) GetIntOrDefault(key string, defaultValue int) int {
	if value, err := sc.GetInt(key); err == nil {
		return value
	}
	return defaultValue
}

// GetFloat retrieves a setting value as float64
func (sc *ServiceConfig) GetFloat(key string) (float64, error) {
	value, exists := sc.Get(key)
	if !exists {
		sc.logger.WithField("key", key).Error("Setting not found")
		return 0, fmt.Errorf("setting not found: %s", key)
	}

	var floatValue float64
	_, err := fmt.Sscanf(value, "%f", &floatValue)
	if err != nil {
		sc.logger.WithError(err).WithField("key", key).Error("Failed to parse setting as float")
		return 0, fmt.Errorf("failed to parse setting %s as float: %w", key, err)
	}

	return floatValue, nil
}

// GetFloatOrDefault retrieves a setting value as float64 with a default fallback
func (sc *ServiceConfig) GetFloatOrDefault(key string, defaultValue float64) float64 {
	if value, err := sc.GetFloat(key); err == nil {
		return value
	}
	return defaultValue
}

// GetAll returns all settings as a map (read-only copy)
func (sc *ServiceConfig) GetAll() map[string]string {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range sc.settings {
		result[k] = v
	}
	return result
}

// Refresh reloads all settings from the database
func (sc *ServiceConfig) Refresh(serviceName string) error {
	sc.logger.WithField("service", serviceName).Info("Refreshing configuration from database")

	// Reconnect to database
	if err := sc.connectToDatabase(); err != nil {
		sc.logger.WithError(err).Error("Failed to connect to database")
		return err
	}
	defer sc.db.Close()

	return sc.loadAllSettings(serviceName)
}
