package handlers

import (
	"database/sql"
	"invoice-service/entities/invoices/models"
	invoiceSQL "invoice-service/entities/invoices/sql"
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for invoices
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for invoices
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// getExpenseCategoryName retrieves the expense category name by ID
func (h *DBHandler) getExpenseCategoryName(tx *sql.Tx, categoryID string) (string, error) {
	var categoryName string
	err := tx.QueryRow("SELECT category_name FROM expense_categories WHERE id = $1", categoryID).Scan(&categoryName)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"expense_category_id": categoryID,
		}).Error("Failed to get expense category name")
		return "", err
	}
	return categoryName, nil
}

// CreateInvoice creates a new invoice in the database
func (h *DBHandler) CreateInvoice(req models.CreateInvoiceRequest) (*models.Invoice, error) {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction for invoice creation")
		return nil, err
	}
	//will rollback if no commit done
	defer tx.Rollback()

	var invoice models.Invoice

	// Use provided transaction date or current time
	transactionDate := time.Now()
	if req.TransactionDate != nil {
		transactionDate = *req.TransactionDate
	}

	// Create the invoice
	err = tx.QueryRow(invoiceSQL.CreateInvoiceQuery,
		req.InvoiceNumber, transactionDate, req.TransactionType, req.SupplierID, req.ExpenseCategoryID, req.ImageURL, req.Notes).
		Scan(&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType, &invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount, &invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number": req.InvoiceNumber,
		}).Error("Failed to create invoice in database")
		return nil, err
	}

	// Get expense category name to check if it's "Ingredients"
	expenseCategoryName, err := h.getExpenseCategoryName(tx, req.ExpenseCategoryID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"expense_category_id": req.ExpenseCategoryID,
		}).Error("Failed to get expense category name")
		return nil, err
	}

	// Create invoice details
	var totalAmount float64 = 0
	for _, item := range req.Items {
		var detail models.InvoiceDetail
		err = tx.QueryRow(invoiceSQL.CreateInvoiceDetailQuery,
			invoice.ID, item.IngredientID, item.Detail, item.Count, item.UnitType, item.Price, item.ExpirationDate).
			Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)

		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"invoice_id": invoice.ID,
				"detail":     item.Detail,
			}).Error("Failed to create invoice detail in database")
			return nil, err
		}

		totalAmount += detail.Total

		// Create existence if this is an ingredient item AND expense category is "Ingredients"
		//pvillalobos - get rid of hardcoded values
		if item.IngredientID != nil && expenseCategoryName == "Ingredients" {
			existenceReq := models.CreateExistenceRequest{
				IngredientID:           *item.IngredientID,
				InvoiceDetailID:        detail.ID,
				UnitsPurchased:         item.Count,
				UnitType:               item.UnitType,
				CostPerUnit:            item.Price,
				ExpirationDate:         item.ExpirationDate,
				IncomeMarginPercentage: 30.0, // Default 30%
				IvaPercentage:          13.0, // Default 13%
				ServiceTaxPercentage:   10.0, // Default 10%
			}

			err = h.CreateInventoryExistence(tx, existenceReq)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"invoice_detail_id": detail.ID,
					"ingredient_id":     *item.IngredientID,
				}).Error("Failed to create existence for ingredient")
				return nil, err
			}
		}
	}

	// Update invoice total
	_, err = tx.Exec(invoiceSQL.UpdateInvoiceTotalQuery, invoice.ID, totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoice.ID,
		}).Error("Failed to update invoice total")
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit invoice creation transaction")
		return nil, err
	}

	// Update the invoice object with the total
	invoice.TotalAmount = &totalAmount

	h.logger.WithFields(logrus.Fields{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"total_amount":   totalAmount,
	}).Info("Invoice created successfully")

	return &invoice, nil
}

// GetInvoiceByID retrieves an invoice by ID from the database
func (h *DBHandler) GetInvoiceByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice

	err := h.db.QueryRow(invoiceSQL.GetInvoiceByIDQuery, id).
		Scan(&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType, &invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount, &invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to retrieve invoice from database")
		return nil, err
	}

	return &invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by number from the database
func (h *DBHandler) GetInvoiceByNumber(number string) (*models.Invoice, error) {
	var invoice models.Invoice

	err := h.db.QueryRow(invoiceSQL.GetInvoiceByNumberQuery, number).
		Scan(&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType, &invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount, &invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_number": number,
		}).Error("Failed to retrieve invoice by number from database")
		return nil, err
	}

	return &invoice, nil
}

// ListInvoices retrieves all invoices from the database
func (h *DBHandler) ListInvoices() ([]models.Invoice, error) {
	rows, err := h.db.Query(invoiceSQL.ListInvoicesQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute invoices list query")
		return nil, err
	}
	defer rows.Close()

	var invoices []models.Invoice
	for rows.Next() {
		var invoice models.Invoice
		err := rows.Scan(&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType, &invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount, &invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan invoice row, skipping")
			continue
		}
		invoices = append(invoices, invoice)
	}

	// Ensure we return an empty slice instead of nil for consistency
	if invoices == nil {
		invoices = []models.Invoice{}
	}

	h.logger.WithFields(logrus.Fields{
		"invoices_count": len(invoices),
	}).Info("Listed invoices successfully")

	return invoices, nil
}

// UpdateInvoice updates an invoice in the database
func (h *DBHandler) UpdateInvoice(id string, req models.UpdateInvoiceRequest) (*models.Invoice, error) {
	var invoice models.Invoice

	err := h.db.QueryRow(invoiceSQL.UpdateInvoiceQuery,
		id, req.InvoiceNumber, req.TransactionDate, req.TransactionType, req.SupplierID, req.ExpenseCategoryID, req.ImageURL, req.Notes).
		Scan(&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionDate, &invoice.TransactionType, &invoice.SupplierID, &invoice.ExpenseCategoryID, &invoice.TotalAmount, &invoice.ImageURL, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to update invoice in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
	}).Info("Invoice updated successfully")

	return &invoice, nil
}

// DeleteInvoice deletes an invoice from the database
func (h *DBHandler) DeleteInvoice(id string) error {
	result, err := h.db.Exec(invoiceSQL.DeleteInvoiceQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": id,
		}).Error("Failed to execute invoice delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected for invoice delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"invoice_id": id,
		}).Warn("No invoice found to delete")
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":    id,
		"rows_affected": rowsAffected,
	}).Info("Invoice deleted successfully")

	return nil
}

// CreateInvoiceDetail creates a new invoice detail in the database
func (h *DBHandler) CreateInvoiceDetail(req models.CreateInvoiceDetailRequest) (*models.InvoiceDetail, error) {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction for invoice detail creation")
		return nil, err
	}
	defer tx.Rollback()

	var detail models.InvoiceDetail

	// Create the invoice detail
	err = tx.QueryRow(invoiceSQL.CreateInvoiceDetailQuery,
		req.InvoiceID, req.IngredientID, req.Detail, req.Count, req.UnitType, req.Price, req.ExpirationDate).
		Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": req.InvoiceID,
			"detail":     req.Detail,
		}).Error("Failed to create invoice detail in database")
		return nil, err
	}

	// Update invoice total
	var totalAmount float64
	err = tx.QueryRow(invoiceSQL.GetInvoiceTotalFromDetailsQuery, req.InvoiceID).Scan(&totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": req.InvoiceID,
		}).Error("Failed to get invoice total from details")
		return nil, err
	}

	_, err = tx.Exec(invoiceSQL.UpdateInvoiceTotalQuery, req.InvoiceID, totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": req.InvoiceID,
		}).Error("Failed to update invoice total")
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit invoice detail creation transaction")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": detail.ID,
		"invoice_id":        detail.InvoiceID,
		"total":             detail.Total,
	}).Info("Invoice detail created successfully")

	return &detail, nil
}

// GetInvoiceDetailByID retrieves an invoice detail by ID from the database
func (h *DBHandler) GetInvoiceDetailByID(id string) (*models.InvoiceDetail, error) {
	var detail models.InvoiceDetail

	err := h.db.QueryRow(invoiceSQL.GetInvoiceDetailByIDQuery, id).
		Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to retrieve invoice detail from database")
		return nil, err
	}

	return &detail, nil
}

// GetInvoiceDetailsByInvoiceID retrieves all invoice details for a specific invoice
func (h *DBHandler) GetInvoiceDetailsByInvoiceID(invoiceID string) ([]models.InvoiceDetail, error) {
	rows, err := h.db.Query(invoiceSQL.GetInvoiceDetailsByInvoiceIDQuery, invoiceID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to execute invoice details query")
		return nil, err
	}
	defer rows.Close()

	var details []models.InvoiceDetail
	for rows.Next() {
		var detail models.InvoiceDetail
		err := rows.Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan invoice detail row, skipping")
			continue
		}
		details = append(details, detail)
	}

	// Ensure we return an empty slice instead of nil for consistency
	if details == nil {
		details = []models.InvoiceDetail{}
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_id":            invoiceID,
		"invoice_details_count": len(details),
	}).Info("Listed invoice details successfully")

	return details, nil
}

// ListInvoiceDetails retrieves all invoice details from the database
func (h *DBHandler) ListInvoiceDetails() ([]models.InvoiceDetail, error) {
	rows, err := h.db.Query(invoiceSQL.ListInvoiceDetailsQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute invoice details list query")
		return nil, err
	}
	defer rows.Close()

	var details []models.InvoiceDetail
	for rows.Next() {
		var detail models.InvoiceDetail
		err := rows.Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan invoice detail row, skipping")
			continue
		}
		details = append(details, detail)
	}

	// Ensure we return an empty slice instead of nil for consistency
	if details == nil {
		details = []models.InvoiceDetail{}
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_details_count": len(details),
	}).Info("Listed invoice details successfully")

	return details, nil
}

// UpdateInvoiceDetail updates an invoice detail in the database
func (h *DBHandler) UpdateInvoiceDetail(id string, req models.UpdateInvoiceDetailRequest) (*models.InvoiceDetail, error) {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction for invoice detail update")
		return nil, err
	}
	defer tx.Rollback()

	var detail models.InvoiceDetail

	err = tx.QueryRow(invoiceSQL.UpdateInvoiceDetailQuery,
		id, req.IngredientID, req.Detail, req.Count, req.UnitType, req.Price, req.ExpirationDate).
		Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to update invoice detail in database")
		return nil, err
	}

	// Update invoice total
	var totalAmount float64
	err = tx.QueryRow(invoiceSQL.GetInvoiceTotalFromDetailsQuery, detail.InvoiceID).Scan(&totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": detail.InvoiceID,
		}).Error("Failed to get invoice total from details")
		return nil, err
	}

	_, err = tx.Exec(invoiceSQL.UpdateInvoiceTotalQuery, detail.InvoiceID, totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": detail.InvoiceID,
		}).Error("Failed to update invoice total")
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit invoice detail update transaction")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": detail.ID,
		"invoice_id":        detail.InvoiceID,
		"total":             detail.Total,
	}).Info("Invoice detail updated successfully")

	return &detail, nil
}

// DeleteInvoiceDetail deletes an invoice detail from the database
func (h *DBHandler) DeleteInvoiceDetail(id string) error {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin transaction for invoice detail deletion")
		return err
	}
	defer tx.Rollback()

	// Get the invoice ID before deleting
	var invoiceID string
	err = tx.QueryRow("SELECT invoice_id FROM invoice_details WHERE id = $1", id).Scan(&invoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithFields(logrus.Fields{
				"invoice_detail_id": id,
			}).Warn("No invoice detail found to delete")
			return sql.ErrNoRows
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to get invoice ID for detail deletion")
		return err
	}

	// Delete the invoice detail
	result, err := tx.Exec(invoiceSQL.DeleteInvoiceDetailQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Error("Failed to execute invoice detail delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected for invoice detail delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"invoice_detail_id": id,
		}).Warn("No invoice detail found to delete")
		return sql.ErrNoRows
	}

	// Update invoice total
	var totalAmount float64
	err = tx.QueryRow(invoiceSQL.GetInvoiceTotalFromDetailsQuery, invoiceID).Scan(&totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to get invoice total from details")
		return err
	}

	_, err = tx.Exec(invoiceSQL.UpdateInvoiceTotalQuery, invoiceID, totalAmount)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"invoice_id": invoiceID,
		}).Error("Failed to update invoice total")
		return err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.WithError(err).Error("Failed to commit invoice detail deletion transaction")
		return err
	}

	h.logger.WithFields(logrus.Fields{
		"invoice_detail_id": id,
		"invoice_id":        invoiceID,
		"rows_affected":     rowsAffected,
	}).Info("Invoice detail deleted successfully")

	return nil
}

// CreateInventoryExistence creates an existence record from an invoice detail
func (h *DBHandler) CreateInventoryExistence(tx *sql.Tx, req models.CreateExistenceRequest) error {
	// Calculate derived fields
	itemsPerUnit := 1 //pvillalobos - we would have to request this in the invoice item
	costPerItem := req.CostPerUnit / float64(itemsPerUnit)

	// Calculate margins and taxes
	incomeMarginAmount := costPerItem * req.IncomeMarginPercentage / 100
	ivaAmount := (costPerItem + incomeMarginAmount) * req.IvaPercentage / 100
	serviceTaxAmount := (costPerItem + incomeMarginAmount) * req.ServiceTaxPercentage / 100

	// Calculate final price
	calculatedPrice := costPerItem + incomeMarginAmount + ivaAmount + serviceTaxAmount
	// Round up to nearest 100
	finalPrice := math.Ceil(calculatedPrice/100) * 100

	// Log calculations for debugging
	h.logger.WithFields(logrus.Fields{
		"cost_per_item":            costPerItem,
		"income_margin_percentage": req.IncomeMarginPercentage,
		"income_margin_amount":     incomeMarginAmount,
		"iva_percentage":           req.IvaPercentage,
		"iva_amount":               ivaAmount,
		"service_tax_percentage":   req.ServiceTaxPercentage,
		"service_tax_amount":       serviceTaxAmount,
		"calculated_price":         calculatedPrice,
		"final_price":              finalPrice,
	}).Debug("Existence calculations completed")

	_, err := tx.Exec(invoiceSQL.CreateExistenceQuery,
		req.IngredientID,
		req.InvoiceDetailID,
		req.UnitsPurchased,
		req.UnitType,
		req.CostPerUnit,
		req.ExpirationDate,
		req.IncomeMarginPercentage,
		incomeMarginAmount,
		req.IvaPercentage,
		ivaAmount,
		req.ServiceTaxPercentage,
		serviceTaxAmount,
		calculatedPrice,
		finalPrice,
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id":     req.IngredientID,
			"invoice_detail_id": req.InvoiceDetailID,
		}).Error("Failed to create existence in database")
		return err
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id":     req.IngredientID,
		"invoice_detail_id": req.InvoiceDetailID,
		"units_purchased":   req.UnitsPurchased,
	}).Info("Existence created successfully")

	return nil
}
