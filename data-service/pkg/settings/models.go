package settings

import (
	"time"
)

// Setting represents a configuration setting
type Setting struct {
	SettingID  string    `json:"setting_id" db:"setting_id"`
	Service    string    `json:"service" db:"service"`
	Key        string    `json:"key" db:"key"`
	Value      string    `json:"value" db:"value"`
	Description *string   `json:"description" db:"description"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// SettingsByService represents settings grouped by service
type SettingsByService struct {
	Service  string     `json:"service"`
	Settings []Setting  `json:"settings"`
}

// LoadAllSettingsRequest represents the request to load all settings
type LoadAllSettingsRequest struct {
	// No specific parameters needed for loading all settings
}

// GetSettingsByServiceRequest represents the request to get settings by service
type GetSettingsByServiceRequest struct {
	Service string `json:"service" validate:"required"`
}

// GetSettingsByNameRequest represents the request to get settings by name
type GetSettingsByNameRequest struct {
	Key string `json:"key" validate:"required"`
}

// Response structs
type SettingsResponse struct {
	Success bool      `json:"success"`
	Data    []Setting `json:"data"`
	Total   int       `json:"total"`
	Message string    `json:"message,omitempty"`
}

type SettingsByServiceResponse struct {
	Success bool                `json:"success"`
	Data    []SettingsByService `json:"data"`
	Total   int                 `json:"total"`
	Message string              `json:"message,omitempty"`
}

type SettingResponse struct {
	Success bool     `json:"success"`
	Data    Setting  `json:"data"`
	Message string   `json:"message,omitempty"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
