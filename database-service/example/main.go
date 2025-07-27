package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"database-service/pkg/database"

	"github.com/sirupsen/logrus"
)

func main() {
	// Create a logger with custom configuration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     true,
	})

	// Create database configuration
	config := &database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres123",
		DBName:   "icecream_store",
		SSLMode:  "disable",

		// Connection pool settings
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		// Timeout settings
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		// Retry settings
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}

	// Create database handler
	db := database.New(config, logger)

	// Connect to database
	fmt.Println("🍦 Connecting to Ice Cream Store Database...")
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Perform health check
	if err := db.HealthCheck(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	fmt.Println("✅ Database connection established successfully")

	// Example: Query system configuration
	if err := querySystemConfig(db); err != nil {
		logger.WithError(err).Error("Failed to query system configuration")
	}

	// Example: Query available roles
	if err := queryRoles(db); err != nil {
		logger.WithError(err).Error("Failed to query roles")
	}

	// Example: Query expense categories
	if err := queryExpenseCategories(db); err != nil {
		logger.WithError(err).Error("Failed to query expense categories")
	}

	// Example: Query recipe categories
	if err := queryRecipeCategories(db); err != nil {
		logger.WithError(err).Error("Failed to query recipe categories")
	}

	// Print connection statistics
	stats := db.GetStats()
	fmt.Printf("\n📊 Database Connection Statistics:\n")
	fmt.Printf("   Open connections: %d\n", stats.OpenConnections)
	fmt.Printf("   In use: %d\n", stats.InUse)
	fmt.Printf("   Idle: %d\n", stats.Idle)
	fmt.Printf("   Wait count: %d\n", stats.WaitCount)
	fmt.Printf("   Wait duration: %v\n", stats.WaitDuration)

	fmt.Println("\n🎉 Database service is working correctly!")
}

// querySystemConfig demonstrates querying system configuration
func querySystemConfig(db database.DatabaseHandler) error {
	query := `
		SELECT config_key, config_value, description
		FROM system_configuration
		ORDER BY config_key
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query system configuration: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n⚙️  System Configuration:")
	fmt.Println("==========================================")

	for rows.Next() {
		var key, value, description string

		err := rows.Scan(&key, &value, &description)
		if err != nil {
			return fmt.Errorf("failed to scan system config row: %w", err)
		}

		fmt.Printf("%-20s: %s\n", key, value)
		if description != "" {
			fmt.Printf("%-20s  (%s)\n", "", description)
		}
	}

	return rows.Err()
}

// queryRoles demonstrates querying roles
func queryRoles(db database.DatabaseHandler) error {
	query := `
		SELECT role_name, description, created_at
		FROM roles
		ORDER BY role_name
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n👥 Available Roles:")
	fmt.Println("==========================================")

	for rows.Next() {
		var roleName, description string
		var createdAt time.Time

		err := rows.Scan(&roleName, &description, &createdAt)
		if err != nil {
			return fmt.Errorf("failed to scan role row: %w", err)
		}

		fmt.Printf("Role: %s\n", roleName)
		fmt.Printf("Description: %s\n", description)
		fmt.Printf("Created: %s\n", createdAt.Format("2006-01-02 15:04:05"))
		fmt.Println("------------------------------------------")
	}

	return rows.Err()
}

// queryExpenseCategories demonstrates querying expense categories
func queryExpenseCategories(db database.DatabaseHandler) error {
	query := `
		SELECT category_name, description
		FROM expense_categories
		ORDER BY category_name
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n💰 Expense Categories:")
	fmt.Println("==========================================")

	for rows.Next() {
		var categoryName, description string

		err := rows.Scan(&categoryName, &description)
		if err != nil {
			return fmt.Errorf("failed to scan expense category row: %w", err)
		}

		fmt.Printf("%-15s: %s\n", categoryName, description)
	}

	return rows.Err()
}

// queryRecipeCategories demonstrates querying recipe categories
func queryRecipeCategories(db database.DatabaseHandler) error {
	query := `
		SELECT name, description
		FROM recipe_categories
		ORDER BY name
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query recipe categories: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n🍨 Recipe Categories:")
	fmt.Println("==========================================")

	for rows.Next() {
		var name, description string

		err := rows.Scan(&name, &description)
		if err != nil {
			return fmt.Errorf("failed to scan recipe category row: %w", err)
		}

		fmt.Printf("%-12s: %s\n", name, description)
	}

	return rows.Err()
}
