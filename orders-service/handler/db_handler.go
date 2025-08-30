package handler

import (
	"database/sql"
	"fmt"
	"time"

	"orders-service/config"
	"orders-service/models"
	ordersql "orders-service/sql"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// pvillalobos - crete unit tests for this
// DBHandler handles database operations for orders
type DBHandler struct {
	db     *sql.DB
	config *config.Config
	repo   *ordersql.Repository
}

// NewDBHandler creates a new order database handler
func NewDBHandler(db *sql.DB, cfg *config.Config) (*DBHandler, error) {
	repo, err := ordersql.NewRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &DBHandler{
		db:     db,
		config: cfg,
		repo:   repo,
	}, nil
}

// CreateOrder creates a new order in the database
// Handles all tax calculations and compliance requirements
func (h *DBHandler) CreateOrder(req models.CreateOrderRequest, logger *logrus.Logger) (*models.OrderWithItems, error) {
	// Calculate subtotal from recipe final prices (from existences - no taxes included)
	subtotalAmount := 0.0
	for _, item := range req.Items {
		subtotalAmount += float64(item.Quantity) * item.UnitPrice
	}

	//pvillalobos - hardcoded tax rates
	// Calculate taxes on the subtotal (tax handling responsibility)
	ivaAmount := subtotalAmount * (h.config.DefaultTaxRate / 100) // 13% IVA
	serviceTaxAmount := subtotalAmount * 0.10                     // 10% service tax

	// Generate order number
	orderNumber := generateOrderNumber()

	// Create order
	order := &models.Order{
		ID:                    uuid.New(),
		OrderNumber:           orderNumber,
		CustomerID:            req.CustomerID,
		SalesRepresentativeID: nil, // Will be set from authenticated user
		Status:                models.OrderStatusPending,
		PaymentMethod:         req.PaymentMethod,
		TransactionReference:  req.TransactionReference,
		SinpeScreenshotURL:    req.SinpeScreenshotURL,
		SubtotalAmount:        subtotalAmount,
		DiscountAmount:        req.DiscountAmount,
		IvaAmount:             ivaAmount,
		ServiceTaxAmount:      serviceTaxAmount,
		TotalAmount:           subtotalAmount + ivaAmount + serviceTaxAmount - req.DiscountAmount,
		InvoiceNumber:         nil, // Will be set after invoice creation
		InvoiceURL:            nil, // Will be set after invoice creation
		TransactionTimestamp:  time.Now(),
		CompletedAt:           nil,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Create ordered recipes
	var items []models.OrderedRecipe
	for _, reqItem := range req.Items {
		subtotal := float64(reqItem.Quantity) * reqItem.UnitPrice
		item := models.OrderedRecipe{
			ID:           uuid.New(),
			OrderID:      order.ID,
			RecipeID:     reqItem.RecipeID,
			ProductName:  reqItem.RecipeName, // Use recipe name from request
			Quantity:     reqItem.Quantity,
			ReceipePrice: reqItem.UnitPrice,
			Subtotal:     subtotal,
		}
		items = append(items, item)
	}

	// Start database transaction
	tx, err := h.db.Begin()
	if err != nil {
		logger.WithError(err).Error("Failed to begin transaction for order creation")
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Save order to database within transaction
	if err := h.repo.CreateOrderWithTx(tx, order, items); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": order.ID,
		}).Error("Failed to create order in database")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.WithError(err).Error("Failed to commit order creation transaction")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get the complete order with calculated final_amount
	createdOrder, err := h.repo.GetOrderWithItems(order.ID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": order.ID,
		}).Error("Failed to retrieve created order")
		return nil, fmt.Errorf("failed to retrieve created order: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"total_amount": order.TotalAmount,
		"iva_amount":   order.IvaAmount,
		"service_tax":  order.ServiceTaxAmount,
	}).Info("Order created successfully in database")

	return createdOrder, nil
}

// GetOrder retrieves an order by ID
func (h *DBHandler) GetOrder(id uuid.UUID, logger *logrus.Logger) (*models.OrderWithItems, error) {
	order, err := h.repo.GetOrderWithItems(id)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": id,
		}).Error("Failed to retrieve order from database")
		return nil, err
	}

	return order, nil
}

// UpdateOrder updates an existing order
func (h *DBHandler) UpdateOrder(id uuid.UUID, req *models.UpdateOrderRequest, logger *logrus.Logger) (*models.OrderWithItems, error) {
	// Update order
	if err := h.repo.UpdateOrder(id, req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": id,
		}).Error("Failed to update order in database")
		return nil, err
	}

	// Get updated order
	updatedOrder, err := h.repo.GetOrderWithItems(id)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": id,
		}).Error("Failed to retrieve updated order from database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"order_id": id,
	}).Info("Order updated successfully in database")

	return updatedOrder, nil
}

// CancelOrder cancels an order
func (h *DBHandler) CancelOrder(id uuid.UUID, logger *logrus.Logger) error {
	if err := h.repo.CancelOrder(id); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"order_id": id,
		}).Error("Failed to cancel order in database")
		return err
	}

	logger.WithFields(logrus.Fields{
		"order_id": id,
	}).Info("Order cancelled successfully in database")

	return nil
}

// ListOrders retrieves orders with filtering and pagination
func (h *DBHandler) ListOrders(filter *models.OrderFilter, logger *logrus.Logger) ([]models.Order, int, error) {
	orders, totalCount, err := h.repo.ListOrders(filter)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve orders from database")
		return nil, 0, err
	}

	return orders, totalCount, nil
}

// GetOrderSummary retrieves order statistics
func (h *DBHandler) GetOrderSummary(logger *logrus.Logger) (*models.OrderSummary, error) {
	summary, err := h.repo.GetOrderSummary()
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve order summary from database")
		return nil, err
	}

	return summary, nil
}

// GetPaymentMethodStats retrieves payment method statistics
func (h *DBHandler) GetPaymentMethodStats(logger *logrus.Logger) ([]models.PaymentMethodStats, error) {
	stats, err := h.repo.GetPaymentMethodStats()
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve payment method stats from database")
		return nil, err
	}

	return stats, nil
}

// HealthCheck checks the health of the database connection
func (h *DBHandler) HealthCheck() error {
	return h.repo.HealthCheck()
}
