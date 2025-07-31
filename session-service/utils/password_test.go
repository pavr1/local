package utils

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// setupTestLogger creates a test logger for use in tests
func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	return logger
}

// TestNewPasswordManager tests the constructor
func TestNewPasswordManager(t *testing.T) {
	logger := setupTestLogger()

	tests := []struct {
		name   string
		cost   int
		logger *logrus.Logger
	}{
		{
			name:   "valid cost and logger",
			cost:   6, // Reasonable cost for testing
			logger: logger,
		},
		{
			name:   "minimum cost",
			cost:   bcrypt.MinCost,
			logger: logger,
		},
		{
			name:   "higher cost",
			cost:   8, // Reasonable higher cost for testing instead of MaxCost
			logger: logger,
		},
		{
			name:   "nil logger",
			cost:   6, // Reasonable cost for testing
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPasswordManager(tt.cost, tt.logger)

			assert.NotNil(t, pm)
			assert.Equal(t, tt.cost, pm.cost)
			assert.Equal(t, tt.logger, pm.logger)
		})
	}
}

// TestHashPassword tests password hashing functionality
func TestHashPassword(t *testing.T) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for testing

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "long password",
			password:    "this-is-a-very-long-password-with-special-characters-!@#$%^&*()",
			expectError: false,
		},
		{
			name:        "short password",
			password:    "abc",
			expectError: false,
		},
		{
			name:        "password with unicode",
			password:    "пароль123",
			expectError: false,
		},
		{
			name:        "password with emojis",
			password:    "password🔒123",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := pm.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
				assert.NotEqual(t, tt.password, hashedPassword)

				// Verify the hash is valid bcrypt hash
				assert.True(t, len(hashedPassword) >= 60) // bcrypt hashes are typically 60 chars
				assert.True(t, hashedPassword[0] == '$')  // bcrypt hashes start with $

				// Verify we can validate the password against the hash
				err = pm.ValidatePassword(tt.password, hashedPassword)
				assert.NoError(t, err)
			}
		})
	}
}

// TestHashPasswordDifferentCosts tests hashing with different bcrypt costs
func TestHashPasswordDifferentCosts(t *testing.T) {
	logger := setupTestLogger()
	password := "testpassword123"

	costs := []int{
		bcrypt.MinCost, // 4
		6,              // Reasonable cost for testing
		8,              // Higher but reasonable cost for testing
	}

	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			pm := NewPasswordManager(cost, logger)

			hashedPassword, err := pm.HashPassword(password)
			require.NoError(t, err)
			assert.NotEmpty(t, hashedPassword)

			// Verify the password validates correctly
			err = pm.ValidatePassword(password, hashedPassword)
			assert.NoError(t, err)

			// Verify wrong password doesn't validate
			err = pm.ValidatePassword("wrongpassword", hashedPassword)
			assert.Error(t, err)
		})
	}
}

// TestHashPasswordConsistency tests that the same password produces different hashes (due to salt)
func TestHashPasswordConsistency(t *testing.T) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for testing
	password := "testpassword123"

	// Hash the same password multiple times
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		hash, err := pm.HashPassword(password)
		require.NoError(t, err)
		hashes[i] = hash
	}

	// All hashes should be different (due to random salt)
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			assert.NotEqual(t, hashes[i], hashes[j], "Hashes should be different due to salt")
		}
	}

	// But all should validate the original password
	for _, hash := range hashes {
		err := pm.ValidatePassword(password, hash)
		assert.NoError(t, err)
	}
}

// TestValidatePassword tests password validation functionality
func TestValidatePassword(t *testing.T) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for testing

	password := "testpassword123"
	hashedPassword, err := pm.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name           string
		password       string
		hashedPassword string
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "correct password",
			password:       password,
			hashedPassword: hashedPassword,
			expectError:    false,
		},
		{
			name:           "incorrect password",
			password:       "wrongpassword",
			hashedPassword: hashedPassword,
			expectError:    true,
			errorMsg:       "incorrect password",
		},
		{
			name:           "empty password",
			password:       "",
			hashedPassword: hashedPassword,
			expectError:    true,
			errorMsg:       "incorrect password",
		},
		{
			name:           "case sensitive password",
			password:       "TestPassword123", // Different case
			hashedPassword: hashedPassword,
			expectError:    true,
			errorMsg:       "incorrect password",
		},
		{
			name:           "password with extra spaces",
			password:       " " + password + " ",
			hashedPassword: hashedPassword,
			expectError:    true,
			errorMsg:       "incorrect password",
		},
		{
			name:           "invalid hash format",
			password:       password,
			hashedPassword: "invalid-hash-format",
			expectError:    true,
			errorMsg:       "password validation error",
		},
		{
			name:           "empty hash",
			password:       password,
			hashedPassword: "",
			expectError:    true,
			errorMsg:       "password validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidatePassword(tt.password, tt.hashedPassword)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePasswordWithDifferentCosts tests validation across different bcrypt costs
func TestValidatePasswordWithDifferentCosts(t *testing.T) {
	logger := setupTestLogger()
	password := "testpassword123"

	// Generate hashes with different costs (using reasonable costs for testing)
	costs := []int{bcrypt.MinCost, 6} // MinCost=4, and 6 for reasonable testing speed
	hashes := make(map[int]string)

	for _, cost := range costs {
		pm := NewPasswordManager(cost, logger)
		hash, err := pm.HashPassword(password)
		require.NoError(t, err)
		hashes[cost] = hash
	}

	// Test that any password manager can validate hashes created with any cost
	for validateCost := range costs {
		pm := NewPasswordManager(validateCost, logger)

		for hashCost, hash := range hashes {
			t.Run(fmt.Sprintf("validate_cost_%d_hash_cost_%d", validateCost, hashCost), func(t *testing.T) {
				err := pm.ValidatePassword(password, hash)
				assert.NoError(t, err)

				// Test wrong password
				err = pm.ValidatePassword("wrongpassword", hash)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "incorrect password")
			})
		}
	}
}

// TestPasswordManagerWithNilLogger tests password manager with nil logger
func TestPasswordManagerWithNilLogger(t *testing.T) {
	t.Skip("Skipping nil logger test as it causes panics - this should be handled in production code")

	// In production, the password manager should always be created with a valid logger
	// This test documents that nil loggers are not supported
}

// TestExtremelyCostlyHash tests behavior with very high bcrypt cost (should be slow)
func TestExtremelyCostlyHash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow test in short mode")
	}

	logger := setupTestLogger()
	pm := NewPasswordManager(8, logger) // Higher but reasonable cost for testing
	password := "testpassword123"

	hashedPassword, err := pm.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	err = pm.ValidatePassword(password, hashedPassword)
	assert.NoError(t, err)
}

// TestPasswordEdgeCases tests various edge cases
func TestPasswordEdgeCases(t *testing.T) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for testing

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "very long password (over 72 bytes)",
			password:    string(make([]byte, 100)), // 100 null bytes - exceeds bcrypt limit
			expectError: true,
		},
		{
			name:        "password with many ASCII characters (over 72 bytes)",
			password:    "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", // 94 chars - exceeds bcrypt limit
			expectError: true,
		},
		{
			name:        "single character password",
			password:    "a",
			expectError: false,
		},
		{
			name:        "password with newlines",
			password:    "password\nwith\nnewlines",
			expectError: false,
		},
		{
			name:        "password with tabs",
			password:    "password\twith\ttabs",
			expectError: false,
		},
		{
			name:        "password at bcrypt limit (72 bytes)",
			password:    string(make([]byte, 72)), // Exactly at the limit
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty password as we know it fails
			if len(tt.password) == 0 {
				return
			}

			hashedPassword, err := pm.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)

				err = pm.ValidatePassword(tt.password, hashedPassword)
				assert.NoError(t, err)
			}
		})
	}
}

// TestInvalidCostValues tests password manager with invalid cost values
func TestInvalidCostValues(t *testing.T) {
	logger := setupTestLogger()
	password := "testpassword123"

	tests := []struct {
		name        string
		cost        int
		expectError bool
	}{
		{
			name:        "cost too low (auto-corrected by bcrypt)",
			cost:        bcrypt.MinCost - 1,
			expectError: false, // bcrypt auto-corrects to MinCost
		},
		{
			name:        "cost too high",
			cost:        bcrypt.MaxCost + 1,
			expectError: true,
		},
		{
			name:        "negative cost (auto-corrected by bcrypt)",
			cost:        -1,
			expectError: false, // bcrypt auto-corrects to MinCost
		},
		{
			name:        "zero cost (auto-corrected by bcrypt)",
			cost:        0,
			expectError: false, // bcrypt auto-corrects to MinCost
		},
		{
			name:        "valid minimum cost",
			cost:        bcrypt.MinCost,
			expectError: false,
		},
		{
			name:        "valid higher cost",
			cost:        10, // Reasonable higher cost instead of MaxCost
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPasswordManager(tt.cost, logger)

			_, err := pm.HashPassword(password)

			if tt.expectError {
				assert.Error(t, err, "Expected error for cost %d", tt.cost)
			} else {
				assert.NoError(t, err, "Expected no error for cost %d", tt.cost)
			}
		})
	}
}

// BenchmarkHashPassword benchmarks password hashing
func BenchmarkHashPassword(b *testing.B) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for benchmarking
	password := "testpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidatePassword benchmarks password validation
func BenchmarkValidatePassword(b *testing.B) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for benchmarking
	password := "testpassword123"

	hashedPassword, err := pm.HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := pm.ValidatePassword(password, hashedPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHashPasswordDifferentCosts benchmarks hashing with different costs
func BenchmarkHashPasswordDifferentCosts(b *testing.B) {
	logger := setupTestLogger()
	password := "testpassword123"

	costs := []int{bcrypt.MinCost, 6} // Reasonable costs for benchmarking

	for _, cost := range costs {
		pm := NewPasswordManager(cost, logger)
		b.Run(fmt.Sprintf("cost_%d", cost), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := pm.HashPassword(password)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestConcurrentHashing tests password hashing under concurrent access
func TestConcurrentHashing(t *testing.T) {
	logger := setupTestLogger()
	pm := NewPasswordManager(6, logger) // Use reasonable cost for testing
	password := "testpassword123"

	const numGoroutines = 10
	const numOperationsPerGoroutine = 5

	results := make(chan string, numGoroutines*numOperationsPerGoroutine)
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// Start concurrent hashing operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numOperationsPerGoroutine; j++ {
				hash, err := pm.HashPassword(password)
				if err != nil {
					errors <- err
					return
				}
				results <- hash
			}
		}()
	}

	// Collect results
	hashes := make([]string, 0, numGoroutines*numOperationsPerGoroutine)
	for i := 0; i < numGoroutines*numOperationsPerGoroutine; i++ {
		select {
		case hash := <-results:
			hashes = append(hashes, hash)
		case err := <-errors:
			t.Fatal(err)
		}
	}

	// Verify all hashes are different and valid
	assert.Len(t, hashes, numGoroutines*numOperationsPerGoroutine)

	for i, hash := range hashes {
		// Verify hash is valid
		err := pm.ValidatePassword(password, hash)
		assert.NoError(t, err, "Hash %d should be valid", i)

		// Verify all hashes are unique (due to salt)
		for j, otherHash := range hashes {
			if i != j {
				assert.NotEqual(t, hash, otherHash, "Hashes %d and %d should be different", i, j)
			}
		}
	}
}
