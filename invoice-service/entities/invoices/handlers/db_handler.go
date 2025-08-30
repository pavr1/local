package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"invoice-service/config"
	"invoice-service/entities/invoices/models"
	invoiceSQL "invoice-service/entities/invoices/sql"
	"math"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// pvillalobos - crete unit tests for this
// DBHandler handles database operations for invoices
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
	config *config.Config
}

// NewDBHandler creates a new database handler for invoices
func NewDBHandler(db *sql.DB, logger *logrus.Logger, cfg *config.Config) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
		config: cfg,
	}
}

// getExpenseCategoryName retrieves the expense category name by ID
func (h *DBHandler) getExpenseCategoryName(tx *sql.Tx, categoryID string) (string, error) {
	var categoryName string
	err := tx.QueryRow(invoiceSQL.GetCategoryNameByIDQuery, categoryID).Scan(&categoryName)
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

	// Use provided transaction date (always required, stored as UTC)
	transactionDate := req.TransactionDate

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

	// Get expense category name to check if it's "Ingredients" (only if expense category is provided)
	var expenseCategoryName string
	if req.ExpenseCategoryID != nil {
		expenseCategoryName, err = h.getExpenseCategoryName(tx, *req.ExpenseCategoryID)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"expense_category_id": *req.ExpenseCategoryID,
			}).Error("Failed to get expense category name")
			return nil, err
		}
	}

	// Create invoice details
	var totalAmount float64 = 0

	for _, item := range req.Items {

		var detail models.InvoiceDetail
		err = tx.QueryRow(invoiceSQL.CreateInvoiceDetailQuery,
			invoice.ID, item.IngredientID, item.Detail, item.Count, item.UnitType, item.ItemsPerUnit, item.Price, item.ExpirationDate).
			Scan(&detail.ID, &detail.InvoiceID, &detail.IngredientID, &detail.Detail, &detail.Count, &detail.UnitType, &detail.ItemsPerUnit, &detail.Price, &detail.Total, &detail.ExpirationDate, &detail.CreatedAt, &detail.UpdatedAt)

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
				IngredientID:    *item.IngredientID,
				InvoiceDetailID: detail.ID,
				UnitsPurchased:  item.Count,
				UnitType:        item.UnitType,
				ItemsPerUnit:    item.ItemsPerUnit,
				CostPerUnit:     item.Price,
				ExpirationDate:  item.ExpirationDate,
				//pvillalobos - hardcoded values
				IncomeMarginPercentage: 30.0, // Default 30%
			}

			err = h.CreateInventoryExistence(tx, existenceReq)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"ingredient_id": *item.IngredientID,
					"detail_id":     detail.ID,
				}).Error("Failed to create inventory existence")
				return nil, err
			}

			// Recalculate all recipes that use this ingredient within the same transaction
			err = h.RecalculateRecipesForIngredient(tx, *item.IngredientID)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"ingredient_id": *item.IngredientID,
				}).Warn("Failed to recalculate recipes for ingredient, continuing with invoice creation")
				// Don't fail the entire invoice creation if recipe recalculation fails
			} else {
				h.logger.WithFields(logrus.Fields{
					"ingredient_id": *item.IngredientID,
				}).Info("Recipe recalculation completed successfully")
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

// CreateInventoryExistence creates an existence record from an invoice detail
func (h *DBHandler) CreateInventoryExistence(tx *sql.Tx, req models.CreateExistenceRequest) error {
	// Debug logging to verify items_per_unit value
	h.logger.WithFields(logrus.Fields{
		"ingredient_id":     req.IngredientID,
		"invoice_detail_id": req.InvoiceDetailID,
		"items_per_unit":    req.ItemsPerUnit,
		"cost_per_unit":     req.CostPerUnit,
		"unit_type":         req.UnitType,
	}).Info("CreateInventoryExistence called with items_per_unit")

	if req.ItemsPerUnit == 0 {
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": req.IngredientID,
			"unit_type":     req.UnitType,
		}).Error("Items per unit is 0")
		return fmt.Errorf("items per unit is 0")
	}

	// Call the inventory service to get the most recent existence
	mostRecentExistence, err := h.getMostRecentExistenceFromInventoryService(req.IngredientID, req.UnitType)
	if err != nil {
		if err == sql.ErrNoRows {
			mostRecentExistence = nil // Ensure it's nil when no rows found
		} else {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"ingredient_id": req.IngredientID,
				"unit_type":     req.UnitType,
			}).Error("Failed to call inventory service for existence")
			return err
		}
	}

	// Calculate derived fields
	costPerItem := req.CostPerUnit / float64(req.ItemsPerUnit)

	// Log the cost per item calculation
	h.logger.WithFields(logrus.Fields{
		"cost_per_unit":  req.CostPerUnit,
		"items_per_unit": req.ItemsPerUnit,
		"cost_per_item":  costPerItem,
	}).Info("Cost per item calculation")

	var finalPrice, incomeMarginAmount, calculatedPrice float64
	var incomeMarginPercentage float64

	// Calculate income margin (base pricing only - no taxes)
	incomeMarginAmount = costPerItem * req.IncomeMarginPercentage / 100

	if mostRecentExistence != nil && mostRecentExistence.FinalPrice != nil {
		// Previous existence found, maintain pricing consistency
		h.logger.WithFields(logrus.Fields{
			"existence_id": mostRecentExistence.ID,
		}).Info("Previous existence found, maintaining pricing consistency")

		// Use the existing final price to maintain consistency
		finalPrice = *mostRecentExistence.FinalPrice

		// Calculate income margin = final price - cost per item (base pricing only)
		incomeMarginAmount = finalPrice - costPerItem
		incomeMarginPercentage = (incomeMarginAmount / finalPrice) * 100

		// For existing existences, minimum price is cost + margin (base pricing)
		calculatedPrice = costPerItem + incomeMarginAmount

		h.logger.WithFields(logrus.Fields{
			"existing_final_price":         finalPrice,
			"calculated_income_margin":     incomeMarginAmount,
			"calculated_income_percentage": incomeMarginPercentage,
			"calculated_base_price":        calculatedPrice,
		}).Info("Using existing final price for consistency")
	} else {
		// No previous existence found, use default calculation
		h.logger.WithFields(logrus.Fields{
			"ingredient_id": req.IngredientID,
			"unit_type":     req.UnitType,
		}).Info("No previous existence found, using default calculation")

		// Log the input values for debugging
		h.logger.WithFields(logrus.Fields{
			"cost_per_item":            costPerItem,
			"income_margin_percentage": req.IncomeMarginPercentage,
		}).Info("Input values for first existence calculation")

		// Log intermediate calculations
		h.logger.WithFields(logrus.Fields{
			"income_margin_amount": incomeMarginAmount,
		}).Info("Intermediate calculations for first existence")

		// Calculate final price (base pricing only - no taxes)
		calculatedPrice = costPerItem + incomeMarginAmount
		// Round up to nearest 100
		finalPrice = math.Ceil(calculatedPrice/100) * 100

		// Use the original margin percentage from the request
		incomeMarginPercentage = req.IncomeMarginPercentage
	}

	// Map to new pricing structure (base pricing only - no taxes)
	minimumPrice := calculatedPrice // cost + margin
	maximumPrice := finalPrice      // user-editable final price

	// Log calculations for debugging
	h.logger.WithFields(logrus.Fields{
		"cost_per_item":            costPerItem,
		"income_margin_percentage": incomeMarginPercentage,
		"income_margin_amount":     incomeMarginAmount,
		"calculated_price":         calculatedPrice,
		"minimum_price":            minimumPrice,
		"maximum_price":            maximumPrice,
		"final_price":              finalPrice,
		"has_previous_existence":   mostRecentExistence != nil,
	}).Info("Existence calculations completed")

	_, err = tx.Exec(invoiceSQL.CreateExistenceQuery,
		req.IngredientID,
		req.InvoiceDetailID,
		req.UnitsPurchased,
		req.UnitType,
		req.ItemsPerUnit,
		req.CostPerUnit,
		req.ExpirationDate,
		incomeMarginPercentage,
		incomeMarginAmount,
		minimumPrice,
		maximumPrice,
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

// getMostRecentExistenceFromInventoryService calls the inventory service to get the most recent existence
func (h *DBHandler) getMostRecentExistenceFromInventoryService(ingredientID, unitType string) (*models.Existence, error) {
	//pvillalobos - hardcoded values
	url := fmt.Sprintf("%s/api/v1/inventory/existences/ingredient/%s/unit-type/%s",
		h.config.InventoryServiceURL, ingredientID, unitType)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add gateway headers for internal service communication
	req.Header.Set("X-Gateway-Service", "ice-cream-gateway")
	req.Header.Set("X-Gateway-Session-Managed", "true")

	resp, err := client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to call inventory service")
		return nil, fmt.Errorf("failed to call inventory service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No existence found, return nil
		return nil, sql.ErrNoRows
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
		}).Error("Inventory service returned non-OK status")
		return nil, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response struct {
		Success bool             `json:"success"`
		Data    models.Existence `json:"data"`
		Message string           `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		h.logger.WithError(err).Error("Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		h.logger.WithFields(logrus.Fields{
			"message": response.Message,
		}).Error("Inventory service returned error")
		return nil, fmt.Errorf("inventory service returned error: %s", response.Message)
	}

	return &response.Data, nil
}

// RecalculateRecipesForIngredient recalculates all recipes that use a specific ingredient within a database transaction
func (h *DBHandler) RecalculateRecipesForIngredient(tx *sql.Tx, ingredientID string) error {
	h.logger.WithFields(logrus.Fields{
		"ingredient_id": ingredientID,
	}).Info("Starting recipe recalculation for ingredient")

	// Get all recipes that use this ingredient
	recipeIDs, err := h.GetRecipesByIngredient(tx, ingredientID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": ingredientID,
		}).Error("Failed to get recipes by ingredient")
		return err
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id": ingredientID,
		"recipe_count":  len(recipeIDs),
		"recipe_ids":    recipeIDs,
	}).Info("Found recipes to recalculate")

	// Recalculate each recipe within the transaction
	for i, recipeID := range recipeIDs {
		h.logger.WithFields(logrus.Fields{
			"recipe_index":  i + 1,
			"total_recipes": len(recipeIDs),
			"recipe_id":     recipeID,
			"ingredient_id": ingredientID,
		}).Info("Recalculating recipe price and status")

		err := h.RecalculateRecipePriceAndStatus(tx, recipeID)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"recipe_id":     recipeID,
				"ingredient_id": ingredientID,
			}).Error("Failed to recalculate recipe, continuing with others")
			// Continue with other recipes even if one fails
			continue
		}

		h.logger.WithFields(logrus.Fields{
			"recipe_id":     recipeID,
			"ingredient_id": ingredientID,
		}).Info("Recipe price and status updated successfully")
	}

	h.logger.WithFields(logrus.Fields{
		"ingredient_id":   ingredientID,
		"recipes_updated": len(recipeIDs),
	}).Info("Completed recipe recalculation for ingredient")

	return nil
}

// GetRecipesByIngredient gets all recipe IDs that use a specific ingredient within a transaction
func (h *DBHandler) GetRecipesByIngredient(tx *sql.Tx, ingredientID string) ([]string, error) {
	rows, err := tx.Query(invoiceSQL.GetRecipesByIngredientQuery, ingredientID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"ingredient_id": ingredientID,
		}).Error("Failed to get recipes by ingredient in transaction")
		return nil, err
	}
	defer rows.Close()

	var recipeIDs []string
	for rows.Next() {
		var recipeID string
		err := rows.Scan(&recipeID)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan recipe ID, skipping")
			continue
		}
		recipeIDs = append(recipeIDs, recipeID)
	}

	return recipeIDs, nil
}

// RecalculateRecipePriceAndStatus recalculates the price and status of a recipe within a transaction
func (h *DBHandler) RecalculateRecipePriceAndStatus(tx *sql.Tx, recipeID string) error {
	h.logger.WithFields(logrus.Fields{
		"recipe_id": recipeID,
		"sql_query": "RecalculateRecipePriceAndStatusQuery",
	}).Info("Executing recipe price and status recalculation SQL")

	result, err := tx.Exec(invoiceSQL.RecalculateRecipePriceAndStatusQuery, recipeID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": recipeID,
		}).Error("Failed to recalculate recipe price and status in transaction")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"recipe_id": recipeID,
		}).Warn("Failed to get rows affected for recipe recalculation")
	} else {
		h.logger.WithFields(logrus.Fields{
			"recipe_id":     recipeID,
			"rows_affected": rowsAffected,
		}).Info("Recipe price and status recalculation SQL executed successfully")
	}

	return nil
}
