package sql

import (
	"embed"
	"io/fs"
)

//go:embed scripts/*.sql
var sqlFiles embed.FS

// SQL query constants
var (
	LoadAllSettingsQuery     string
	GetSettingsByServiceQuery string
	GetSettingsByNameQuery   string
)

// init loads SQL queries from embedded files
func init() {
	LoadAllSettingsQuery = mustReadSQLFile("scripts/load_all_settings.sql")
	GetSettingsByServiceQuery = mustReadSQLFile("scripts/get_settings_by_service.sql")
	GetSettingsByNameQuery = mustReadSQLFile("scripts/get_settings_by_name.sql")
}

// mustReadSQLFile reads a SQL file from the embedded filesystem
func mustReadSQLFile(filename string) string {
	content, err := fs.ReadFile(sqlFiles, filename)
	if err != nil {
		panic(err)
	}
	return string(content)
}
