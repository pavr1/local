package models

import (
	"time"
)

// Setting represents a configuration setting from the data service
type Setting struct {
	SettingID   string    `json:"setting_id"`
	Service     string    `json:"service"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SettingsResponse represents the response from the data service
type SettingsResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// GetSettingsByServiceRequest represents the request to get settings by service
type GetSettingsByServiceRequest struct {
	Service string `json:"service"`
}
