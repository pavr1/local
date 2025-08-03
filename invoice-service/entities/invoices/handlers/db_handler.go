package handlers

import (
	"database/sql"
	"fmt"

	"invoice-service/entities/invoices/models"
	invoiceSQL "invoice-service/entities/invoices/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for invoices and invoice details
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for invoices and invoice details
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// Invoice Methods

// CreateInvoiceWithDetails creates a new invoice with details in a transaction
func (h *DBHandler) CreateInvoiceWithDetails(req models.CreateInvoiceRequest) (*models.Invoice, error) {
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

	// Create the invoice (without total initially) - only get ID back
	var invoiceID string
	err = tx.QueryRow(invoiceSQL.CreateInvoiceQuery,
		req.InvoiceNumber, req.TransactionDate, req.TransactionType, req.SupplierID,
		req.ExpenseCategoryID, nil, req.ImageURL, req.Notes).
		Scan(&invoiceID)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number": req.InvoiceNumber,
		}).Error("Failed to create invoice in database")
		return nil, err
	}

	// Create all invoice details and calculate total
	var totalAmount float64
	for _, itemReq := range req.Items {
		itemTotal := itemReq.Count * itemReq.Price
		totalAmount += itemTotal

		// Create invoice detail
		_, err = tx.Exec(invoiceSQL.CreateInvoiceDetailQuery,
			invoiceID, itemReq.IngredientID, itemReq.Detail,
			itemReq.Count, itemReq.UnitType, itemReq.Price, itemTotal, itemReq.ExpirationDate)

		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"invoice_id": invoiceID,
				"detail":     itemReq.Detail,
			}).Error("Failed to create invoice detail in database")
			return nil, err
		}
	}

	// Update invoice with total amount
	_, err = tx.Exec(invoiceSQL.UpdateInvoiceTotalQuery, invoiceID, totalAmount)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to update invoice total in database")
		return nil, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		h.logger.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	// Fetch the complete invoice data
	invoice, err := h.GetInvoiceByID(invoiceID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to fetch created invoice")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"items_count":    len(req.Items),
		"total_amount":   totalAmount,
	}).Info("Invoice with details created successfully")

	return invoice, nil
}

// CreateInvoice creates a new invoice without details
func (h *DBHandler) CreateInvoice(req models.CreateInvoiceRequest) (*models.Invoice, error) {
	var invoiceID string

	err := h.db.QueryRow(invoiceSQL.CreateInvoiceQuery,
		req.InvoiceNumber, req.TransactionDate, req.TransactionType, req.SupplierID,
		req.ExpenseCategoryID, nil, req.ImageURL, req.Notes).
		Scan(&invoiceID)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number": req.InvoiceNumber,
		}).Error("Failed to create invoice in database")
		return nil, err
	}

	// Fetch the complete invoice data
	invoice, err := h.GetInvoiceByID(invoiceID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to fetch created invoice")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
	}).Info("Invoice created successfully")

	return invoice, nil
}

// GetInvoiceByID retrieves an invoice by its ID
func (h *DBHandler) GetInvoiceByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice

	err := h.db.QueryRow(invoiceSQL.GetInvoiceByIDQuery, id).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType,
		&invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount,
		&invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"invoice_id": id,
			}).Warn("Invoice not found")
			return nil, fmt.Errorf("invoice not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to get invoice from database")
		return nil, err
	}

	return &invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by its number
func (h *DBHandler) GetInvoiceByNumber(number string) (*models.Invoice, error) {
	var invoice models.Invoice

	err := h.db.QueryRow(invoiceSQL.GetInvoiceByNumberQuery, number).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType,
		&invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount,
		&invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"invoice_number": number,
			}).Warn("Invoice not found")
			return nil, fmt.Errorf("invoice not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number": number,
		}).Error("Failed to get invoice from database")
		return nil, err
	}

	return &invoice, nil
}

// ListInvoices retrieves a list of invoices with optional filtering
func (h *DBHandler) ListInvoices(req models.ListInvoicesRequest) ([]models.Invoice, int, error) {
	// Set defaults
	limit := 50
	offset := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	var invoices []models.Invoice
	var totalCount int

	// Build query based on filters
	var query string
	var countQuery string
	var args []interface{}

	if req.ExpenseCategoryID != nil {
		query = invoiceSQL.ListInvoicesByExpenseCategoryQuery
		countQuery = invoiceSQL.CountInvoicesByExpenseCategoryQuery
		args = append(args, *req.ExpenseCategoryID, limit, offset)
	} else if req.SupplierID != nil {
		query = invoiceSQL.ListInvoicesBySupplierQuery
		countQuery = invoiceSQL.CountInvoicesBySupplierQuery
		args = append(args, *req.SupplierID, limit, offset)
	} else {
		query = invoiceSQL.ListInvoicesBaseQuery
		countQuery = invoiceSQL.CountInvoicesQuery
		args = append(args, limit, offset)
	}

	// Get total count
	var countArgs []interface{}
	if req.ExpenseCategoryID != nil {
		countArgs = append(countArgs, *req.ExpenseCategoryID)
	} else if req.SupplierID != nil {
		countArgs = append(countArgs, *req.SupplierID)
	}

	err := h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get invoice count from database")
		return nil, 0, err
	}

	// Get invoices
	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list invoices from database")
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var invoice models.Invoice
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType,
			&invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount,
			&invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan invoice row")
			return nil, 0, err
		}
		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error iterating invoice rows")
		return nil, 0, err
	}

	return invoices, totalCount, nil
}

// UpdateInvoice updates an existing invoice
func (h *DBHandler) UpdateInvoice(id string, req models.UpdateInvoiceRequest) (*models.Invoice, error) {
	// Get current invoice
	currentInvoice, err := h.GetInvoiceByID(id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}
	argCount := 1

	if req.InvoiceNumber != nil {
		updateFields = append(updateFields, fmt.Sprintf("invoice_number = $%d", argCount))
		args = append(args, *req.InvoiceNumber)
		argCount++
	}
	if req.TransactionDate != nil {
		updateFields = append(updateFields, fmt.Sprintf("transaction_date = $%d", argCount))
		args = append(args, *req.TransactionDate)
		argCount++
	}
	if req.TransactionType != nil {
		updateFields = append(updateFields, fmt.Sprintf("transaction_type = $%d", argCount))
		args = append(args, *req.TransactionType)
		argCount++
	}
	if req.SupplierID != nil {
		updateFields = append(updateFields, fmt.Sprintf("supplier_id = $%d", argCount))
		args = append(args, *req.SupplierID)
		argCount++
	}
	if req.ExpenseCategoryID != nil {
		updateFields = append(updateFields, fmt.Sprintf("expense_category_id = $%d", argCount))
		args = append(args, *req.ExpenseCategoryID)
		argCount++
	}
	if req.ImageURL != nil {
		updateFields = append(updateFields, fmt.Sprintf("image_url = $%d", argCount))
		args = append(args, *req.ImageURL)
		argCount++
	}
	if req.Notes != nil {
		updateFields = append(updateFields, fmt.Sprintf("notes = $%d", argCount))
		args = append(args, *req.Notes)
		argCount++
	}

	if len(updateFields) == 0 {
		return currentInvoice, nil // No fields to update
	}

	// Add updated_at field
	updateFields = append(updateFields, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))

	// Add ID parameter for WHERE clause
	args = append(args, id)

	// Build and execute query
	query := fmt.Sprintf(`
		UPDATE invoice 
		SET %s
		WHERE id = $%d
		RETURNING id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at`,
		fmt.Sprintf("%s", updateFields[0]), argCount)

	for i := 1; i < len(updateFields); i++ {
		query = fmt.Sprintf("%s, %s", query, updateFields[i])
	}

	var updatedInvoice models.Invoice
	err = h.db.QueryRow(query, args...).Scan(
		&updatedInvoice.ID, &updatedInvoice.InvoiceNumber, &updatedInvoice.TransactionDate, &updatedInvoice.TransactionType,
		&updatedInvoice.SupplierID, &updatedInvoice.ExpenseCategoryID, &updatedInvoice.TotalAmount,
		&updatedInvoice.ImageURL, &updatedInvoice.Notes, &updatedInvoice.CreatedAt, &updatedInvoice.UpdatedAt,
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to update invoice in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":     updatedInvoice.ID,
		"invoice_number": updatedInvoice.InvoiceNumber,
	}).Info("Invoice updated successfully")

	return &updatedInvoice, nil
}

// DeleteInvoice deletes an invoice by ID
func (h *DBHandler) DeleteInvoice(id string) error {
	result, err := h.db.Exec(invoiceSQL.DeleteInvoiceQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to delete invoice from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"invoice_id": id,
		}).Warn("No invoice found to delete")
		return fmt.Errorf("invoice not found")
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id": id,
	}).Info("Invoice deleted successfully")

	return nil
}

// Invoice Detail Methods

// CreateInvoiceDetail creates a new invoice detail
func (h *DBHandler) CreateInvoiceDetail(req models.CreateInvoiceDetailRequest) (*models.InvoiceDetail, error) {
	var invoiceDetail models.InvoiceDetail

	// Calculate total
	total := req.Count * req.Price

	err := h.db.QueryRow(invoiceSQL.CreateInvoiceDetailQuery,
		req.InvoiceID, req.IngredientID, req.Detail,
		req.Count, req.UnitType, req.Price, total, req.ExpirationDate).Scan(
		&invoiceDetail.ID, &invoiceDetail.InvoiceID, &invoiceDetail.IngredientID,
		&invoiceDetail.Detail, &invoiceDetail.Count, &invoiceDetail.UnitType,
		&invoiceDetail.Price, &invoiceDetail.Total, &invoiceDetail.ExpirationDate,
		&invoiceDetail.CreatedAt, &invoiceDetail.UpdatedAt,
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": req.InvoiceID,
			"detail":     req.Detail,
		}).Error("Failed to create invoice detail in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": invoiceDetail.ID,
		"invoice_id":        invoiceDetail.InvoiceID,
		"detail":            invoiceDetail.Detail,
	}).Info("Invoice detail created successfully")

	return &invoiceDetail, nil
}

// GetInvoiceDetailByID retrieves an invoice detail by its ID
func (h *DBHandler) GetInvoiceDetailByID(id string) (*models.InvoiceDetail, error) {
	var invoiceDetail models.InvoiceDetail

	err := h.db.QueryRow(invoiceSQL.GetInvoiceDetailByIDQuery, id).Scan(
		&invoiceDetail.ID, &invoiceDetail.InvoiceID, &invoiceDetail.IngredientID,
		&invoiceDetail.Detail, &invoiceDetail.Count, &invoiceDetail.UnitType,
		&invoiceDetail.Price, &invoiceDetail.Total, &invoiceDetail.ExpirationDate,
		&invoiceDetail.CreatedAt, &invoiceDetail.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"invoice_detail_id": id,
			}).Warn("Invoice detail not found")
			return nil, fmt.Errorf("invoice detail not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to get invoice detail from database")
		return nil, err
	}

	return &invoiceDetail, nil
}

// ListInvoiceDetails retrieves a list of invoice details with optional filtering
func (h *DBHandler) ListInvoiceDetails(req models.ListInvoiceDetailsRequest) ([]models.InvoiceDetail, int, error) {
	// Set defaults
	limit := 50
	offset := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	var invoiceDetails []models.InvoiceDetail
	var totalCount int

	// Build query based on filters
	var query string
	var countQuery string
	var args []interface{}

	if req.InvoiceID != nil {
		query = invoiceSQL.ListInvoiceDetailsByInvoiceIDQuery
		countQuery = invoiceSQL.CountInvoiceDetailsByInvoiceQuery
		args = append(args, *req.InvoiceID, limit, offset)
	} else if req.IngredientID != nil {
		query = invoiceSQL.ListInvoiceDetailsByIngredientQuery
		countQuery = invoiceSQL.CountInvoiceDetailsByIngredientQuery
		args = append(args, *req.IngredientID, limit, offset)
	} else {
		query = invoiceSQL.ListInvoiceDetailsBaseQuery
		countQuery = invoiceSQL.CountInvoiceDetailsQuery
		args = append(args, limit, offset)
	}

	// Get total count
	var countArgs []interface{}
	if req.InvoiceID != nil {
		countArgs = append(countArgs, *req.InvoiceID)
	} else if req.IngredientID != nil {
		countArgs = append(countArgs, *req.IngredientID)
	}

	err := h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get invoice detail count from database")
		return nil, 0, err
	}

	// Get invoice details
	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list invoice details from database")
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var invoiceDetail models.InvoiceDetail
		err := rows.Scan(
			&invoiceDetail.ID, &invoiceDetail.InvoiceID, &invoiceDetail.IngredientID,
			&invoiceDetail.Detail, &invoiceDetail.Count, &invoiceDetail.UnitType,
			&invoiceDetail.Price, &invoiceDetail.Total, &invoiceDetail.ExpirationDate,
			&invoiceDetail.CreatedAt, &invoiceDetail.UpdatedAt,
		)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan invoice detail row")
			return nil, 0, err
		}
		invoiceDetails = append(invoiceDetails, invoiceDetail)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error iterating invoice detail rows")
		return nil, 0, err
	}

	return invoiceDetails, totalCount, nil
}

// UpdateInvoiceDetail updates an existing invoice detail
func (h *DBHandler) UpdateInvoiceDetail(id string, req models.UpdateInvoiceDetailRequest) (*models.InvoiceDetail, error) {
	var updatedInvoiceDetail models.InvoiceDetail

	err := h.db.QueryRow(invoiceSQL.UpdateInvoiceDetailQuery,
		id, req.IngredientID, req.Detail, req.Count, req.UnitType, req.Price, req.ExpirationDate).Scan(
		&updatedInvoiceDetail.ID, &updatedInvoiceDetail.InvoiceID, &updatedInvoiceDetail.IngredientID,
		&updatedInvoiceDetail.Detail, &updatedInvoiceDetail.Count, &updatedInvoiceDetail.UnitType,
		&updatedInvoiceDetail.Price, &updatedInvoiceDetail.Total, &updatedInvoiceDetail.ExpirationDate,
		&updatedInvoiceDetail.CreatedAt, &updatedInvoiceDetail.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"invoice_detail_id": id,
			}).Warn("Invoice detail not found")
			return nil, fmt.Errorf("invoice detail not found")
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to update invoice detail in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": updatedInvoiceDetail.ID,
		"invoice_id":        updatedInvoiceDetail.InvoiceID,
	}).Info("Invoice detail updated successfully")

	return &updatedInvoiceDetail, nil
}

// DeleteInvoiceDetail deletes an invoice detail by ID
func (h *DBHandler) DeleteInvoiceDetail(id string) error {
	result, err := h.db.Exec(invoiceSQL.DeleteInvoiceDetailQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to delete invoice detail from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Warn("No invoice detail found to delete")
		return fmt.Errorf("invoice detail not found")
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": id,
	}).Info("Invoice detail deleted successfully")

	return nil
}
