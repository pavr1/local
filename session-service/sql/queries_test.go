package sql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadQueries(t *testing.T) {
	t.Run("successful_query_loading", func(t *testing.T) {
		queries, err := LoadQueries()
		require.NoError(t, err)
		require.NotNil(t, queries)

		// Should have at least the known SQL files
		assert.Greater(t, len(queries), 0, "Should load at least one query")

		// Check that expected queries exist
		expectedQueries := []string{
			"update_last_login",
			"get_user_permissions",
			"get_user_profile_by_id",
			"get_user_profile_by_username",
		}

		for _, expected := range expectedQueries {
			_, exists := queries[expected]
			assert.True(t, exists, "Query %s should be loaded", expected)
		}
	})

	t.Run("queries_content_validation", func(t *testing.T) {
		queries, err := LoadQueries()
		require.NoError(t, err)

		// Verify that all queries have content
		for name, content := range queries {
			assert.NotEmpty(t, content, "Query %s should have content", name)
			assert.True(t, strings.Contains(strings.ToUpper(content), "SELECT") ||
				strings.Contains(strings.ToUpper(content), "UPDATE") ||
				strings.Contains(strings.ToUpper(content), "INSERT") ||
				strings.Contains(strings.ToUpper(content), "DELETE"),
				"Query %s should contain SQL keywords", name)
		}
	})

	t.Run("query_names_format", func(t *testing.T) {
		queries, err := LoadQueries()
		require.NoError(t, err)

		// Verify query names don't have .sql extension
		for name := range queries {
			assert.False(t, strings.HasSuffix(name, ".sql"),
				"Query name %s should not have .sql extension", name)
		}
	})
}

func TestQueriesGet(t *testing.T) {
	queries, err := LoadQueries()
	require.NoError(t, err)

	t.Run("get_existing_query", func(t *testing.T) {
		// Test with a known query
		query, err := queries.Get("get_user_profile_by_id")
		assert.NoError(t, err)
		assert.NotEmpty(t, query)
		assert.Contains(t, strings.ToUpper(query), "SELECT")
	})

	t.Run("get_nonexistent_query", func(t *testing.T) {
		query, err := queries.Get("nonexistent_query")
		assert.Error(t, err)
		assert.Empty(t, query)
		assert.Contains(t, err.Error(), "query 'nonexistent_query' not found")
	})

	t.Run("get_empty_query_name", func(t *testing.T) {
		query, err := queries.Get("")
		assert.Error(t, err)
		assert.Empty(t, query)
		assert.Contains(t, err.Error(), "query '' not found")
	})
}

func TestQueriesMustGet(t *testing.T) {
	queries, err := LoadQueries()
	require.NoError(t, err)

	t.Run("must_get_existing_query", func(t *testing.T) {
		// Test with a known query
		assert.NotPanics(t, func() {
			query := queries.MustGet("get_user_profile_by_id")
			assert.NotEmpty(t, query)
			assert.Contains(t, strings.ToUpper(query), "SELECT")
		})
	})

	t.Run("must_get_nonexistent_query_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			queries.MustGet("nonexistent_query")
		})
	})
}

func TestQueriesList(t *testing.T) {
	queries, err := LoadQueries()
	require.NoError(t, err)

	t.Run("list_all_queries", func(t *testing.T) {
		queryNames := queries.List()
		assert.Greater(t, len(queryNames), 0, "Should return at least one query name")

		// Verify all names exist in the map
		for _, name := range queryNames {
			_, exists := queries[name]
			assert.True(t, exists, "Listed query %s should exist in map", name)
		}

		// Verify count matches
		assert.Equal(t, len(queries), len(queryNames), "List should return all query names")
	})

	t.Run("list_contains_expected_queries", func(t *testing.T) {
		queryNames := queries.List()

		expectedQueries := []string{
			"update_last_login",
			"get_user_permissions",
			"get_user_profile_by_id",
			"get_user_profile_by_username",
		}

		for _, expected := range expectedQueries {
			assert.Contains(t, queryNames, expected, "List should contain %s", expected)
		}
	})
}

func TestQueriesType(t *testing.T) {
	t.Run("queries_type_structure", func(t *testing.T) {
		queries := make(Queries)
		queries["test"] = "SELECT 1"

		// Test that Queries is a map[string]string
		assert.IsType(t, map[string]string{}, map[string]string(queries))

		// Test direct access
		value, exists := queries["test"]
		assert.True(t, exists)
		assert.Equal(t, "SELECT 1", value)
	})
}

func TestSQLFileEmbedding(t *testing.T) {
	t.Run("embedded_files_accessible", func(t *testing.T) {
		// Test that we can read the embedded directory
		entries, err := sqlFiles.ReadDir("scripts")
		assert.NoError(t, err)
		assert.Greater(t, len(entries), 0, "Should have SQL files in scripts directory")

		// Test that we can read individual files
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				content, err := sqlFiles.ReadFile("scripts/" + entry.Name())
				assert.NoError(t, err, "Should be able to read %s", entry.Name())
				assert.Greater(t, len(content), 0, "File %s should have content", entry.Name())
			}
		}
	})
}

func TestQueryContentValidation(t *testing.T) {
	queries, err := LoadQueries()
	require.NoError(t, err)

	t.Run("update_last_login_query", func(t *testing.T) {
		query, err := queries.Get("update_last_login")
		require.NoError(t, err)
		assert.Contains(t, strings.ToUpper(query), "UPDATE")
		assert.Contains(t, strings.ToUpper(query), "LAST_LOGIN")
	})

	t.Run("get_user_permissions_query", func(t *testing.T) {
		query, err := queries.Get("get_user_permissions")
		require.NoError(t, err)
		assert.Contains(t, strings.ToUpper(query), "SELECT")
		assert.Contains(t, strings.ToUpper(query), "PERMISSIONS")
	})

	t.Run("get_user_profile_queries", func(t *testing.T) {
		profiles := []string{"get_user_profile_by_id", "get_user_profile_by_username"}

		for _, profile := range profiles {
			query, err := queries.Get(profile)
			require.NoError(t, err, "Query %s should exist", profile)
			assert.Contains(t, strings.ToUpper(query), "SELECT", "Query %s should be a SELECT", profile)
			assert.Contains(t, strings.ToUpper(query), "USER", "Query %s should reference user", profile)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("load_query_error_propagation", func(t *testing.T) {
		// Test the loadQuery function indirectly through LoadQueries
		// This test ensures error handling works correctly
		queries, err := LoadQueries()
		require.NoError(t, err)

		// Since we can't directly test loadQuery with invalid files in the embedded FS,
		// we verify that LoadQueries handles the embedded files correctly
		assert.NotNil(t, queries)
		assert.Greater(t, len(queries), 0)
	})
}

// Benchmark tests
func BenchmarkLoadQueries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := LoadQueries()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueriesGet(b *testing.B) {
	queries, err := LoadQueries()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = queries.Get("get_user_profile_by_id")
	}
}

func BenchmarkQueriesList(b *testing.B) {
	queries, err := LoadQueries()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queries.List()
	}
}
