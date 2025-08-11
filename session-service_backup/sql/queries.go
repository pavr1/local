package sql

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed scripts/*.sql
var sqlFiles embed.FS

// Queries holds all SQL queries loaded from files as a map
// Key: filename without .sql extension
// Value: SQL query content
type Queries map[string]string

// LoadQueries dynamically loads all SQL queries from embedded files
func LoadQueries() (Queries, error) {
	queries := make(Queries)

	// Read all entries in the scripts directory
	entries, err := sqlFiles.ReadDir("scripts")
	if err != nil {
		return nil, fmt.Errorf("failed to read scripts directory: %w", err)
	}

	// Process each .sql file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		// Read the file content
		content, err := loadQuery(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", filename, err)
		}

		// Use filename without extension as the key
		queryName := strings.TrimSuffix(filename, ".sql")
		queries[queryName] = content
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("no SQL queries found in scripts directory")
	}

	return queries, nil
}

// Get retrieves a query by name, returns error if not found
func (q Queries) Get(name string) (string, error) {
	query, exists := q[name]
	if !exists {
		return "", fmt.Errorf("query '%s' not found", name)
	}
	return query, nil
}

// MustGet retrieves a query by name, panics if not found (use for required queries)
func (q Queries) MustGet(name string) string {
	query, err := q.Get(name)
	if err != nil {
		panic(err)
	}
	return query
}

// List returns all available query names
func (q Queries) List() []string {
	names := make([]string, 0, len(q))
	for name := range q {
		names = append(names, name)
	}
	return names
}

// loadQuery loads a single SQL query from the embedded file system
func loadQuery(filename string) (string, error) {
	path := filepath.Join("scripts", filename)
	content, err := sqlFiles.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read SQL file %s: %w", filename, err)
	}

	return string(content), nil
}
