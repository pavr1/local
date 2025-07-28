package sql

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"orders-service/models"

	"github.com/google/uuid"
)

//go:embed scripts/*.sql
var sqlFiles embed.FS

// SQLQueries holds all SQL queries loaded from files as a map
// Key: filename without .sql extension
// Value: SQL query content
type SQLQueries map[string]string

// LoadQueries dynamically loads all SQL queries from embedded files
func LoadQueries() (SQLQueries, error) {
	queries := make(SQLQueries)

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
func (q SQLQueries) Get(name string) (string, error) {
	query, exists := q[name]
	if !exists {
		return "", fmt.Errorf("query '%s' not found", name)
	}
	return query, nil
}

// MustGet retrieves a query by name, panics if not found (use for required queries)
func (q SQLQueries) MustGet(name string) string {
	query, err := q.Get(name)
	if err != nil {
		panic(err)
	}
	return query
}

// List returns all available query names
func (q SQLQueries) List() []string {
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

// Repository struct holds database connection and loaded queries
type Repository struct {
	db      *sql.DB
	queries SQLQueries
}

// NewRepository creates a new repository instance with loaded queries
func NewRepository(db *sql.DB) (*Repository, error) {
	queries, err := LoadQueries()
	if err != nil {
		return nil, fmt.Errorf("failed to load SQL queries: %w", err)
	}

	return &Repository{
		db:      db,
		queries: queries,
	}, nil
}

// === ORDER QUERIES ===

// CreateOrder creates a new order with its items in a transaction
func (r *Repository) CreateOrder(order *models.Order, items []models.OrderedRecipe) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := r.queries.MustGet("create_order")
	_, err = tx.Exec(orderQuery,
		order.ID, order.CustomerID, order.OrderDate, order.TotalAmount,
		order.TaxAmount, order.DiscountAmount, order.FinalAmount, order.PaymentMethod,
		order.OrderStatus, order.Notes, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert ordered recipes
	if len(items) > 0 {
		itemQuery := r.queries.MustGet("create_ordered_recipe")
		for _, item := range items {
			_, err = tx.Exec(itemQuery,
				item.ID, item.OrderID, item.RecipeID, item.Quantity,
				item.UnitPrice, item.TotalPrice, item.SpecialInstructions, item.CreatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert ordered recipe: %w", err)
			}
		}
	}

	return tx.Commit()
}

// GetOrderByID retrieves an order by its ID
func (r *Repository) GetOrderByID(id uuid.UUID) (*models.Order, error) {
	query := r.queries.MustGet("get_order_by_id")

	var order models.Order
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.CustomerID, &order.OrderDate, &order.TotalAmount,
		&order.TaxAmount, &order.DiscountAmount, &order.FinalAmount,
		&order.PaymentMethod, &order.OrderStatus, &order.Notes,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetOrderWithItems retrieves an order with its ordered recipes
func (r *Repository) GetOrderWithItems(id uuid.UUID) (*models.OrderWithItems, error) {
	order, err := r.GetOrderByID(id)
	if err != nil {
		return nil, err
	}

	items, err := r.GetOrderedRecipesByOrderID(id)
	if err != nil {
		return nil, err
	}

	return &models.OrderWithItems{
		Order: *order,
		Items: items,
	}, nil
}

// GetOrderedRecipesByOrderID retrieves all ordered recipes for an order
func (r *Repository) GetOrderedRecipesByOrderID(orderID uuid.UUID) ([]models.OrderedRecipe, error) {
	query := r.queries.MustGet("get_ordered_recipes_by_order_id")

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ordered recipes: %w", err)
	}
	defer rows.Close()

	var items []models.OrderedRecipe
	for rows.Next() {
		var item models.OrderedRecipe
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.RecipeID, &item.Quantity,
			&item.UnitPrice, &item.TotalPrice, &item.SpecialInstructions,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ordered recipe: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateOrder updates an order
func (r *Repository) UpdateOrder(id uuid.UUID, updates *models.UpdateOrderRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.PaymentMethod != nil {
		setParts = append(setParts, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, *updates.PaymentMethod)
		argIndex++
	}

	if updates.OrderStatus != nil {
		setParts = append(setParts, fmt.Sprintf("order_status = $%d", argIndex))
		args = append(args, *updates.OrderStatus)
		argIndex++
	}

	if updates.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *updates.Notes)
		argIndex++
	}

	if updates.DiscountAmount != nil {
		setParts = append(setParts, fmt.Sprintf("discount_amount = $%d", argIndex))
		args = append(args, *updates.DiscountAmount)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add the ID for the WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE orders 
		SET %s 
		WHERE id = $%d`,
		strings.Join(setParts, ", "), argIndex)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// CancelOrder sets an order status to cancelled
func (r *Repository) CancelOrder(id uuid.UUID) error {
	query := r.queries.MustGet("cancel_order")

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found or already completed")
	}

	return nil
}

// ListOrders retrieves orders with filtering and pagination
func (r *Repository) ListOrders(filter *models.OrderFilter) ([]models.Order, int, error) {
	// Build WHERE conditions
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.CustomerID != nil {
		whereParts = append(whereParts, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.OrderStatus != nil {
		whereParts = append(whereParts, fmt.Sprintf("order_status = $%d", argIndex))
		args = append(args, *filter.OrderStatus)
		argIndex++
	}

	if filter.PaymentMethod != nil {
		whereParts = append(whereParts, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, *filter.PaymentMethod)
		argIndex++
	}

	if filter.DateFrom != nil {
		whereParts = append(whereParts, fmt.Sprintf("order_date >= $%d", argIndex))
		args = append(args, *filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != nil {
		whereParts = append(whereParts, fmt.Sprintf("order_date <= $%d", argIndex))
		args = append(args, *filter.DateTo)
		argIndex++
	}

	if filter.MinAmount != nil {
		whereParts = append(whereParts, fmt.Sprintf("final_amount >= $%d", argIndex))
		args = append(args, *filter.MinAmount)
		argIndex++
	}

	if filter.MaxAmount != nil {
		whereParts = append(whereParts, fmt.Sprintf("final_amount <= $%d", argIndex))
		args = append(args, *filter.MaxAmount)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// Get total count
	baseCountQuery := r.queries.MustGet("count_orders_base")
	countQuery := fmt.Sprintf("%s %s", baseCountQuery, whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "order_date DESC"
	if filter.SortBy != "" {
		sortOrder := "ASC"
		if filter.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		orderBy = filter.SortBy + " " + sortOrder
	}

	// Build LIMIT and OFFSET
	limit := 50 // default
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	// Main query
	baseQuery := r.queries.MustGet("list_orders_base")
	query := fmt.Sprintf(`
		%s 
		%s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`,
		baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.CustomerID, &order.OrderDate, &order.TotalAmount,
			&order.TaxAmount, &order.DiscountAmount, &order.FinalAmount,
			&order.PaymentMethod, &order.OrderStatus, &order.Notes,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, totalCount, rows.Err()
}

// GetOrderSummary retrieves order statistics
func (r *Repository) GetOrderSummary() (*models.OrderSummary, error) {
	query := r.queries.MustGet("get_order_summary")

	var summary models.OrderSummary
	err := r.db.QueryRow(query).Scan(
		&summary.TotalOrders, &summary.PendingOrders, &summary.CompletedOrders,
		&summary.CancelledOrders, &summary.TotalRevenue, &summary.AverageOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order summary: %w", err)
	}

	return &summary, nil
}

// GetPaymentMethodStats retrieves payment method statistics
func (r *Repository) GetPaymentMethodStats() ([]models.PaymentMethodStats, error) {
	query := r.queries.MustGet("get_payment_method_stats")

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment method stats: %w", err)
	}
	defer rows.Close()

	var stats []models.PaymentMethodStats
	for rows.Next() {
		var stat models.PaymentMethodStats
		err := rows.Scan(
			&stat.PaymentMethod, &stat.Count, &stat.TotalAmount, &stat.Percentage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment method stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// === HEALTH CHECK ===

// HealthCheck verifies database connectivity
func (r *Repository) HealthCheck() error {
	query := r.queries.MustGet("health_check")
	var result int
	err := r.db.QueryRow(query).Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}
