package handlers

import (
	"database/sql"
	"fmt"

	"expense-service/entities/receipts/models"
	receiptSQL "expense-service/entities/receipts/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for receipts and receipt items
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for receipts and receipt items
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// Receipt Methods

// CreateReceiptWithItems creates a new receipt with items in a transaction
func (h *DBHandler) CreateReceiptWithItems(req models.CreateReceiptRequest) (*models.Receipt, error) {
	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create the receipt (without total initially) - only get ID back
	var receiptID string
	err = tx.QueryRow(receiptSQL.CreateReceiptQuery,
		req.ReceiptNumber, req.PurchaseDate, req.SupplierID,
		req.ExpenseCategoryID, nil, req.ImageURL, req.Notes).
		Scan(&receiptID)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_number": req.ReceiptNumber,
		}).Error("Failed to create receipt in database")
		return nil, err
	}

	// Create receipt items and calculate total
	var totalAmount float64
	for _, itemReq := range req.Items {
		itemTotal := itemReq.Count * itemReq.Price
		totalAmount += itemTotal

		// Insert receipt item
		_, err = tx.Exec(receiptSQL.CreateReceiptItemQuery,
			receiptID, itemReq.IngredientID, itemReq.Detail, itemReq.Count,
			itemReq.UnitType, itemReq.Price, itemTotal, itemReq.ExpirationDate)

		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"receipt_id": receiptID,
				"detail":     itemReq.Detail,
			}).Error("Failed to create receipt item in database")
			return nil, err
		}
	}

	// Update receipt with total amount
	_, err = tx.Exec(receiptSQL.UpdateReceiptTotalQuery, receiptID, totalAmount)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": receiptID,
		}).Error("Failed to update receipt total in database")
		return nil, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		h.logger.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	// Fetch the complete receipt data
	receipt, err := h.GetReceiptByID(receiptID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": receiptID,
		}).Error("Failed to fetch created receipt")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_id":     receipt.ID,
		"receipt_number": receipt.ReceiptNumber,
		"items_count":    len(req.Items),
		"total_amount":   totalAmount,
	}).Info("Receipt with items created successfully")

	return receipt, nil
}

// CreateReceipt creates a new receipt without items
func (h *DBHandler) CreateReceipt(req models.CreateReceiptRequest) (*models.Receipt, error) {
	var receiptID string

	err := h.db.QueryRow(receiptSQL.CreateReceiptQuery,
		req.ReceiptNumber, req.PurchaseDate, req.SupplierID,
		req.ExpenseCategoryID, nil, req.ImageURL, req.Notes).
		Scan(&receiptID)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_number": req.ReceiptNumber,
		}).Error("Failed to create receipt in database")
		return nil, err
	}

	// Fetch the complete receipt data
	receipt, err := h.GetReceiptByID(receiptID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": receiptID,
		}).Error("Failed to fetch created receipt")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_id":     receipt.ID,
		"receipt_number": receipt.ReceiptNumber,
	}).Info("Receipt created successfully")

	return receipt, nil
}

// GetReceiptByID retrieves a receipt by ID
func (h *DBHandler) GetReceiptByID(id string) (*models.Receipt, error) {
	var receipt models.Receipt

	err := h.db.QueryRow(receiptSQL.GetReceiptByIDQuery, id).
		Scan(&receipt.ID, &receipt.ReceiptNumber, &receipt.PurchaseDate,
			&receipt.SupplierID, &receipt.ExpenseCategoryID, &receipt.TotalAmount,
			&receipt.ImageURL, &receipt.Notes, &receipt.CreatedAt, &receipt.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"receipt_id": id,
			}).Warn("Receipt not found")
			return nil, fmt.Errorf("receipt not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": id,
		}).Error("Failed to get receipt from database")
		return nil, err
	}

	return &receipt, nil
}

// GetReceiptByNumber retrieves a receipt by receipt number
func (h *DBHandler) GetReceiptByNumber(receiptNumber string) (*models.Receipt, error) {
	var receipt models.Receipt

	err := h.db.QueryRow(receiptSQL.GetReceiptByNumberQuery, receiptNumber).
		Scan(&receipt.ID, &receipt.ReceiptNumber, &receipt.PurchaseDate,
			&receipt.SupplierID, &receipt.ExpenseCategoryID, &receipt.TotalAmount,
			&receipt.ImageURL, &receipt.Notes, &receipt.CreatedAt, &receipt.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"receipt_number": receiptNumber,
			}).Warn("Receipt not found by number")
			return nil, fmt.Errorf("receipt not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_number": receiptNumber,
		}).Error("Failed to get receipt by number from database")
		return nil, err
	}

	return &receipt, nil
}

// ListReceipts retrieves a list of receipts with optional filtering
func (h *DBHandler) ListReceipts(req models.ListReceiptsRequest) ([]models.Receipt, int, error) {
	limit := 50 // default limit
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0 // default offset
	if req.Offset != nil {
		offset = *req.Offset
	}

	var receipts []models.Receipt
	var rows *sql.Rows
	var err error
	var countQuery string
	var countArgs []interface{}

	// Determine which query to use based on filters
	if req.ExpenseCategoryID != nil {
		rows, err = h.db.Query(receiptSQL.ListReceiptsByExpenseCategoryQuery, *req.ExpenseCategoryID, limit, offset)
		countQuery = receiptSQL.CountReceiptsByExpenseCategoryQuery
		countArgs = []interface{}{*req.ExpenseCategoryID}
	} else if req.SupplierID != nil {
		rows, err = h.db.Query(receiptSQL.ListReceiptsBySupplierQuery, *req.SupplierID, limit, offset)
		countQuery = receiptSQL.CountReceiptsBySupplierQuery
		countArgs = []interface{}{*req.SupplierID}
	} else {
		rows, err = h.db.Query(receiptSQL.ListReceiptsBaseQuery, limit, offset)
		countQuery = receiptSQL.CountReceiptsQuery
		countArgs = []interface{}{}
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list receipts from database")
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var receipt models.Receipt
		err = rows.Scan(&receipt.ID, &receipt.ReceiptNumber, &receipt.PurchaseDate,
			&receipt.SupplierID, &receipt.ExpenseCategoryID, &receipt.TotalAmount,
			&receipt.ImageURL, &receipt.Notes, &receipt.CreatedAt, &receipt.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan receipt row")
			return nil, 0, err
		}
		receipts = append(receipts, receipt)
	}

	// Get total count
	var totalCount int
	err = h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get receipts count")
		return nil, 0, err
	}

	return receipts, totalCount, nil
}

// UpdateReceipt updates a receipt in the database
func (h *DBHandler) UpdateReceipt(id string, req models.UpdateReceiptRequest) (*models.Receipt, error) {
	var receipt models.Receipt

	err := h.db.QueryRow(receiptSQL.UpdateReceiptQuery,
		id, req.ReceiptNumber, req.PurchaseDate, req.SupplierID,
		req.ExpenseCategoryID, req.ImageURL, req.Notes).
		Scan(&receipt.ID, &receipt.ReceiptNumber, &receipt.PurchaseDate,
			&receipt.SupplierID, &receipt.ExpenseCategoryID, &receipt.TotalAmount,
			&receipt.ImageURL, &receipt.Notes, &receipt.CreatedAt, &receipt.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"receipt_id": id,
			}).Warn("Receipt not found for update")
			return nil, fmt.Errorf("receipt not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": id,
		}).Error("Failed to update receipt in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_id": receipt.ID,
	}).Info("Receipt updated successfully")

	return &receipt, nil
}

// DeleteReceipt deletes a receipt from the database
func (h *DBHandler) DeleteReceipt(id string) error {
	result, err := h.db.Exec(receiptSQL.DeleteReceiptQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": id,
		}).Error("Failed to delete receipt from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": id,
		}).Error("Failed to get rows affected after receipt deletion")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"receipt_id": id,
		}).Warn("Receipt not found for deletion")
		return fmt.Errorf("receipt not found")
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_id": id,
	}).Info("Receipt deleted successfully")

	return nil
}

// Receipt Items Methods

// CreateReceiptItem creates a new receipt item in the database
func (h *DBHandler) CreateReceiptItem(req models.CreateReceiptItemRequest) (*models.ReceiptItem, error) {
	var receiptItem models.ReceiptItem
	itemTotal := req.Count * req.Price

	err := h.db.QueryRow(receiptSQL.CreateReceiptItemQuery,
		req.ReceiptID, req.IngredientID, req.Detail, req.Count,
		req.UnitType, req.Price, itemTotal, req.ExpirationDate).
		Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.IngredientID,
			&receiptItem.Detail, &receiptItem.Count, &receiptItem.UnitType,
			&receiptItem.Price, &receiptItem.Total, &receiptItem.ExpirationDate,
			&receiptItem.CreatedAt, &receiptItem.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": req.ReceiptID,
			"detail":     req.Detail,
		}).Error("Failed to create receipt item in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_item_id": receiptItem.ID,
		"receipt_id":      receiptItem.ReceiptID,
		"detail":          receiptItem.Detail,
	}).Info("Receipt item created successfully")

	return &receiptItem, nil
}

// GetReceiptItemByID retrieves a receipt item by ID
func (h *DBHandler) GetReceiptItemByID(id string) (*models.ReceiptItem, error) {
	var receiptItem models.ReceiptItem

	err := h.db.QueryRow(receiptSQL.GetReceiptItemByIDQuery, id).
		Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.IngredientID,
			&receiptItem.Detail, &receiptItem.Count, &receiptItem.UnitType,
			&receiptItem.Price, &receiptItem.Total, &receiptItem.ExpirationDate,
			&receiptItem.CreatedAt, &receiptItem.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"receipt_item_id": id,
			}).Warn("Receipt item not found")
			return nil, fmt.Errorf("receipt item not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_item_id": id,
		}).Error("Failed to get receipt item from database")
		return nil, err
	}

	return &receiptItem, nil
}

// GetReceiptItemsByReceiptID retrieves all receipt items for a specific receipt
func (h *DBHandler) GetReceiptItemsByReceiptID(receiptID string) ([]models.ReceiptItem, error) {
	rows, err := h.db.Query(receiptSQL.ListReceiptItemsByReceiptIDQuery, receiptID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": receiptID,
		}).Error("Failed to get receipt items from database")
		return nil, err
	}
	defer rows.Close()

	var receiptItems []models.ReceiptItem
	for rows.Next() {
		var receiptItem models.ReceiptItem
		err = rows.Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.IngredientID,
			&receiptItem.Detail, &receiptItem.Count, &receiptItem.UnitType,
			&receiptItem.Price, &receiptItem.Total, &receiptItem.ExpirationDate,
			&receiptItem.CreatedAt, &receiptItem.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan receipt item row")
			return nil, err
		}
		receiptItems = append(receiptItems, receiptItem)
	}

	return receiptItems, nil
}

// ListReceiptItems retrieves a list of receipt items with optional filtering
func (h *DBHandler) ListReceiptItems(req models.ListReceiptItemsRequest) ([]models.ReceiptItem, int, error) {
	limit := 50 // default limit
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0 // default offset
	if req.Offset != nil {
		offset = *req.Offset
	}

	var receiptItems []models.ReceiptItem
	var rows *sql.Rows
	var err error
	var countQuery string
	var countArgs []interface{}

	// Determine which query to use based on filters
	if req.ReceiptID != nil {
		rows, err = h.db.Query(receiptSQL.ListReceiptItemsByReceiptIDQuery, *req.ReceiptID)
		countQuery = receiptSQL.CountReceiptItemsByReceiptQuery
		countArgs = []interface{}{*req.ReceiptID}
	} else if req.IngredientID != nil {
		rows, err = h.db.Query(receiptSQL.ListReceiptItemsByIngredientQuery, *req.IngredientID, limit, offset)
		countQuery = receiptSQL.CountReceiptItemsByIngredientQuery
		countArgs = []interface{}{*req.IngredientID}
	} else {
		rows, err = h.db.Query(receiptSQL.ListReceiptItemsBaseQuery, limit, offset)
		countQuery = receiptSQL.CountReceiptItemsQuery
		countArgs = []interface{}{}
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list receipt items from database")
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var receiptItem models.ReceiptItem
		err = rows.Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.IngredientID,
			&receiptItem.Detail, &receiptItem.Count, &receiptItem.UnitType,
			&receiptItem.Price, &receiptItem.Total, &receiptItem.ExpirationDate,
			&receiptItem.CreatedAt, &receiptItem.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan receipt item row")
			return nil, 0, err
		}
		receiptItems = append(receiptItems, receiptItem)
	}

	// Get total count
	var totalCount int
	err = h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get receipt items count")
		return nil, 0, err
	}

	return receiptItems, totalCount, nil
}

// UpdateReceiptItem updates a receipt item in the database
func (h *DBHandler) UpdateReceiptItem(id string, req models.UpdateReceiptItemRequest) (*models.ReceiptItem, error) {
	var receiptItem models.ReceiptItem

	err := h.db.QueryRow(receiptSQL.UpdateReceiptItemQuery,
		id, req.IngredientID, req.Detail, req.Count, req.UnitType,
		req.Price, req.ExpirationDate).
		Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.IngredientID,
			&receiptItem.Detail, &receiptItem.Count, &receiptItem.UnitType,
			&receiptItem.Price, &receiptItem.Total, &receiptItem.ExpirationDate,
			&receiptItem.CreatedAt, &receiptItem.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"receipt_item_id": id,
			}).Warn("Receipt item not found for update")
			return nil, fmt.Errorf("receipt item not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_item_id": id,
		}).Error("Failed to update receipt item in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_item_id": receiptItem.ID,
	}).Info("Receipt item updated successfully")

	return &receiptItem, nil
}

// DeleteReceiptItem deletes a receipt item from the database
func (h *DBHandler) DeleteReceiptItem(id string) error {
	result, err := h.db.Exec(receiptSQL.DeleteReceiptItemQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_item_id": id,
		}).Error("Failed to delete receipt item from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_item_id": id,
		}).Error("Failed to get rows affected after receipt item deletion")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"receipt_item_id": id,
		}).Warn("Receipt item not found for deletion")
		return fmt.Errorf("receipt item not found")
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_item_id": id,
	}).Info("Receipt item deleted successfully")

	return nil
}

// GetReceiptTotal calculates the total amount for a receipt from its items
func (h *DBHandler) GetReceiptTotal(receiptID string) (float64, error) {
	var total float64

	err := h.db.QueryRow(receiptSQL.GetReceiptTotalFromItemsQuery, receiptID).Scan(&total)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"receipt_id": receiptID,
		}).Error("Failed to calculate receipt total from items")
		return 0, err
	}

	return total, nil
}
