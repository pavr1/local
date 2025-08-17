package settings

import (
	sqlQueries "data-service/pkg/settings/sql"
	"database/sql"

	"github.com/sirupsen/logrus"
)

// SettingsDBHandler handles database operations for settings
type SettingsDBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewSettingsDBHandler creates a new database handler for settings
func NewSettingsDBHandler(db *sql.DB, logger *logrus.Logger) *SettingsDBHandler {
	return &SettingsDBHandler{
		db:     db,
		logger: logger,
	}
}

// LoadAllSettings loads all settings from the database into memory
func (h *SettingsDBHandler) LoadAllSettings() (map[string]Setting, error) {
	rows, err := h.db.Query(sqlQueries.LoadAllSettingsQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load all settings from database")
		return nil, err
	}
	defer rows.Close()

	settingsMap := make(map[string]Setting)
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
			h.logger.WithError(err).Error("Failed to scan setting row")
			return nil, err
		}
		// Use Service:Key as the composite key
		key := setting.Service + ":" + setting.Key
		settingsMap[key] = setting
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"settings_count": len(settingsMap),
	}).Info("Loaded all settings successfully")

	return settingsMap, nil
}

// GetSettingsByService retrieves settings for a specific service
func (h *SettingsDBHandler) GetSettingsByService(service string) ([]Setting, error) {
	rows, err := h.db.Query(sqlQueries.GetSettingsByServiceQuery, service)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"service": service,
		}).Error("Failed to get settings by service from database")
		return nil, err
	}
	defer rows.Close()

	var settings []Setting
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
			h.logger.WithError(err).Error("Failed to scan setting row")
			return nil, err
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"service":        service,
		"settings_count": len(settings),
	}).Info("Retrieved settings by service successfully")

	return settings, nil
}

// GetSettingsByName retrieves settings by key name across all services
func (h *SettingsDBHandler) GetSettingsByName(key string) ([]Setting, error) {
	rows, err := h.db.Query(sqlQueries.GetSettingsByNameQuery, key)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"key": key,
		}).Error("Failed to get settings by name from database")
		return nil, err
	}
	defer rows.Close()

	var settings []Setting
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
			h.logger.WithError(err).Error("Failed to scan setting row")
			return nil, err
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error occurred during rows iteration")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"key":            key,
		"settings_count": len(settings),
	}).Info("Retrieved settings by name successfully")

	return settings, nil
}

// GroupSettingsByService groups settings by service
func (h *SettingsDBHandler) GroupSettingsByService(settings []Setting) []SettingsByService {
	serviceMap := make(map[string][]Setting)

	for _, setting := range settings {
		serviceMap[setting.Service] = append(serviceMap[setting.Service], setting)
	}

	var result []SettingsByService
	for service, serviceSettings := range serviceMap {
		result = append(result, SettingsByService{
			Service:  service,
			Settings: serviceSettings,
		})
	}

	return result
}
