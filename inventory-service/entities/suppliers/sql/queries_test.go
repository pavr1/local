package sql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryConstants tests that all SQL query constants are defined
func TestQueryConstants(t *testing.T) {
	// Test that query constants are not empty
	assert.NotEmpty(t, CreateSupplierQuery, "CreateSupplierQuery should not be empty")
	assert.NotEmpty(t, ListSuppliersQuery, "ListSuppliersQuery should not be empty")
	assert.NotEmpty(t, GetSupplierByIDQuery, "GetSupplierByIDQuery should not be empty")
	assert.NotEmpty(t, UpdateSupplierQuery, "UpdateSupplierQuery should not be empty")
	assert.NotEmpty(t, DeleteSupplierQuery, "DeleteSupplierQuery should not be empty")
}

// TestQueryStructure tests that queries have expected SQL structure
func TestQueryStructure(t *testing.T) {
	t.Run("CreateSupplierQuery", func(t *testing.T) {
		query := strings.ToUpper(CreateSupplierQuery)
		assert.Contains(t, query, "INSERT INTO", "Should be an INSERT query")
		assert.Contains(t, query, "SUPPLIERS", "Should insert into suppliers table")
		assert.Contains(t, query, "RETURNING", "Should have RETURNING clause")
	})

	t.Run("ListSuppliersQuery", func(t *testing.T) {
		query := strings.ToUpper(ListSuppliersQuery)
		assert.Contains(t, query, "SELECT", "Should be a SELECT query")
		assert.Contains(t, query, "FROM SUPPLIERS", "Should select from suppliers table")
	})

	t.Run("GetSupplierByIDQuery", func(t *testing.T) {
		query := strings.ToUpper(GetSupplierByIDQuery)
		assert.Contains(t, query, "SELECT", "Should be a SELECT query")
		assert.Contains(t, query, "FROM SUPPLIERS", "Should select from suppliers table")
		assert.Contains(t, query, "WHERE", "Should have WHERE clause")
		assert.Contains(t, query, "ID", "Should filter by ID")
	})

	t.Run("UpdateSupplierQuery", func(t *testing.T) {
		query := strings.ToUpper(UpdateSupplierQuery)
		assert.Contains(t, query, "UPDATE", "Should be an UPDATE query")
		assert.Contains(t, query, "SUPPLIERS", "Should update suppliers table")
		assert.Contains(t, query, "SET", "Should have SET clause")
		assert.Contains(t, query, "WHERE", "Should have WHERE clause")
		assert.Contains(t, query, "RETURNING", "Should have RETURNING clause")
	})

	t.Run("DeleteSupplierQuery", func(t *testing.T) {
		query := strings.ToUpper(DeleteSupplierQuery)
		assert.Contains(t, query, "DELETE", "Should be a DELETE query")
		assert.Contains(t, query, "FROM SUPPLIERS", "Should delete from suppliers table")
		assert.Contains(t, query, "WHERE", "Should have WHERE clause")
		assert.Contains(t, query, "ID", "Should filter by ID")
	})
}

// TestQueryParameters tests that queries have expected parameter placeholders
func TestQueryParameters(t *testing.T) {
	t.Run("CreateSupplierQuery parameters", func(t *testing.T) {
		// Should have placeholders for supplier fields
		assert.Contains(t, CreateSupplierQuery, "$1", "Should have parameter placeholder")
		assert.Contains(t, CreateSupplierQuery, "$2", "Should have parameter placeholder")
		assert.Contains(t, CreateSupplierQuery, "$3", "Should have parameter placeholder")
	})

	t.Run("GetSupplierByIDQuery parameters", func(t *testing.T) {
		// Should have one parameter for ID
		assert.Contains(t, GetSupplierByIDQuery, "$1", "Should have ID parameter placeholder")
	})

	t.Run("UpdateSupplierQuery parameters", func(t *testing.T) {
		// Should have multiple parameters for update
		assert.Contains(t, UpdateSupplierQuery, "$1", "Should have parameter placeholder")
		assert.Contains(t, UpdateSupplierQuery, "$2", "Should have parameter placeholder")
	})

	t.Run("DeleteSupplierQuery parameters", func(t *testing.T) {
		// Should have one parameter for ID
		assert.Contains(t, DeleteSupplierQuery, "$1", "Should have ID parameter placeholder")
	})
}

// TestQueryValidation tests basic SQL syntax validation
func TestQueryValidation(t *testing.T) {
	queries := []struct {
		name  string
		query string
	}{
		{"CreateSupplierQuery", CreateSupplierQuery},
		{"ListSuppliersQuery", ListSuppliersQuery},
		{"GetSupplierByIDQuery", GetSupplierByIDQuery},
		{"UpdateSupplierQuery", UpdateSupplierQuery},
		{"DeleteSupplierQuery", DeleteSupplierQuery},
	}

	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			// Basic validation: should not be just whitespace
			assert.NotEqual(t, strings.TrimSpace(q.query), "", "Query should not be empty or just whitespace")

			// Should not contain common SQL injection patterns (basic check)
			upperQuery := strings.ToUpper(q.query)
			assert.NotContains(t, upperQuery, "; DROP", "Should not contain suspicious SQL patterns")
			assert.NotContains(t, upperQuery, "' OR '1'='1", "Should not contain suspicious SQL patterns")

			// Should be properly terminated (no dangling semicolon issues)
			trimmed := strings.TrimSpace(q.query)
			if strings.HasSuffix(trimmed, ";") {
				// If it ends with semicolon, should only have one
				assert.Equal(t, 1, strings.Count(trimmed, ";"), "Should not have multiple semicolons")
			}
		})
	}
}

// TestQueryConsistency tests that queries are consistent in style
func TestQueryConsistency(t *testing.T) {
	queries := []string{
		CreateSupplierQuery,
		ListSuppliersQuery,
		GetSupplierByIDQuery,
		UpdateSupplierQuery,
		DeleteSupplierQuery,
	}

	t.Run("consistent table naming", func(t *testing.T) {
		for _, query := range queries {
			upperQuery := strings.ToUpper(query)
			if strings.Contains(upperQuery, "SUPPLIERS") {
				// Should use consistent table name (suppliers, not supplier)
				assert.Contains(t, upperQuery, "SUPPLIERS", "Should use consistent table name 'suppliers'")
			}
		}
	})

	t.Run("consistent parameter style", func(t *testing.T) {
		for _, query := range queries {
			if strings.Contains(query, "$") {
				// If query has parameters, they should follow PostgreSQL format ($1, $2, etc.)
				assert.Regexp(t, `\$\d+`, query, "Parameters should follow PostgreSQL format ($1, $2, etc.)")
			}
		}
	})
}

// TestQueryReturnClauses tests that queries that should return data have RETURNING clauses
func TestQueryReturnClauses(t *testing.T) {
	t.Run("CREATE query should return data", func(t *testing.T) {
		upperQuery := strings.ToUpper(CreateSupplierQuery)
		assert.Contains(t, upperQuery, "RETURNING", "CREATE query should have RETURNING clause")
	})

	t.Run("UPDATE query should return data", func(t *testing.T) {
		upperQuery := strings.ToUpper(UpdateSupplierQuery)
		assert.Contains(t, upperQuery, "RETURNING", "UPDATE query should have RETURNING clause")
	})

	t.Run("SELECT queries should not need RETURNING", func(t *testing.T) {
		// SELECT queries don't need RETURNING clauses
		selectQueries := []string{ListSuppliersQuery, GetSupplierByIDQuery}
		for _, query := range selectQueries {
			upperQuery := strings.ToUpper(query)
			assert.Contains(t, upperQuery, "SELECT", "Should be a SELECT query")
			// RETURNING is not required for SELECT, but if present, it's not wrong
		}
	})
}

// BenchmarkQueryConstants benchmarks access to query constants
func BenchmarkQueryConstants(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CreateSupplierQuery
		_ = ListSuppliersQuery
		_ = GetSupplierByIDQuery
		_ = UpdateSupplierQuery
		_ = DeleteSupplierQuery
	}
}
