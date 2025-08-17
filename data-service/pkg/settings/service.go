package settings

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SettingsService handles settings business logic and caching
type SettingsService struct {
	dbHandler *SettingsDBHandler
	logger    *logrus.Logger
	cache     map[string]Setting
	cacheMux  sync.RWMutex
	lastLoad  time.Time
}

// NewSettingsService creates a new settings service
func NewSettingsService(dbHandler *SettingsDBHandler, logger *logrus.Logger) *SettingsService {
	return &SettingsService{
		dbHandler: dbHandler,
		logger:    logger,
		cache:     make(map[string]Setting),
	}
}

// loadAllSettings loads all settings into memory cache (private method)
func (s *SettingsService) loadAllSettings() error {
	s.logger.Info("Loading all settings into memory cache")

	settingsMap, err := s.dbHandler.LoadAllSettings()
	if err != nil {
		s.logger.WithError(err).Error("Failed to load settings from database")
		return err
	}

	// Update cache
	s.cacheMux.Lock()
	s.cache = settingsMap
	s.lastLoad = time.Now()
	s.cacheMux.Unlock()

	s.logger.WithFields(logrus.Fields{
		"settings_count": len(settingsMap),
		"cache_size":     len(s.cache),
	}).Info("Settings loaded into cache successfully")

	return nil
}

// GetSettingsByService retrieves settings for a specific service
func (s *SettingsService) GetSettingsByService(service string) (*SettingsResponse, error) {
	s.logger.WithField("service", service).Info("Getting settings by service")

	// Use cached settings
	s.cacheMux.RLock()
	settingsMap := s.cache
	s.cacheMux.RUnlock()

	// Filter settings by service
	var settings []Setting
	for _, setting := range settingsMap {
		if setting.Service == service {
			settings = append(settings, setting)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"service":        service,
		"settings_count": len(settings),
	}).Info("Settings retrieved by service successfully")

	return &SettingsResponse{
		Success: true,
		Data:    settings,
		Total:   len(settings),
		Message: "Settings retrieved successfully",
	}, nil
}

// GetSettingsByName retrieves settings by key name across all services
func (s *SettingsService) GetSettingsByName(key string) (*SettingsResponse, error) {
	s.logger.WithField("key", key).Info("Getting settings by name")

	// Use cached settings
	s.cacheMux.RLock()
	settingsMap := s.cache
	s.cacheMux.RUnlock()

	// Filter settings by key
	var settings []Setting
	for _, setting := range settingsMap {
		if setting.Key == key {
			settings = append(settings, setting)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"key":            key,
		"settings_count": len(settings),
	}).Info("Settings retrieved by name successfully")

	return &SettingsResponse{
		Success: true,
		Data:    settings,
		Total:   len(settings),
		Message: "Settings retrieved successfully",
	}, nil
}

// GetSettingsByServiceGrouped returns settings grouped by service
func (s *SettingsService) GetSettingsByServiceGrouped() (*SettingsByServiceResponse, error) {
	s.logger.Info("Getting all settings grouped by service")

	// Use cached settings
	s.cacheMux.RLock()
	settingsMap := s.cache
	s.cacheMux.RUnlock()

	// Convert map to slice for grouping
	var settings []Setting
	for _, setting := range settingsMap {
		settings = append(settings, setting)
	}

	groupedSettings := s.dbHandler.GroupSettingsByService(settings)

	s.logger.WithFields(logrus.Fields{
		"services_count": len(groupedSettings),
		"total_settings": len(settings),
	}).Info("Settings grouped by service successfully")

	return &SettingsByServiceResponse{
		Success: true,
		Data:    groupedSettings,
		Total:   len(settings),
		Message: "Settings grouped successfully",
	}, nil
}

// GetCachedSetting retrieves a setting from cache
func (s *SettingsService) GetCachedSetting(service, key string) (*Setting, bool) {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	cacheKey := s.getCacheKey(service, key)
	setting, exists := s.cache[cacheKey]
	if !exists {
		return nil, false
	}

	return &setting, true
}

// IsCacheValid checks if the cache is still valid (less than 5 minutes old)
func (s *SettingsService) IsCacheValid() bool {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	return time.Since(s.lastLoad) < 5*time.Minute
}

// getCacheKey generates a cache key for a service and key combination
func (s *SettingsService) getCacheKey(service, key string) string {
	return service + ":" + key
}
