package handlers

import (
	"database/sql"

	"invoice-service/entities/expense_categories/models"
	expenseCategorySQL "invoice-service/entities/expense_categories/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for expense categories
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for expense categories
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// CreateExpenseCategory creates a new expense category in the database
func (h *DBHandler) CreateExpenseCategory(req models.CreateExpenseCategoryRequest) (*models.ExpenseCategory, error) {
	var expenseCategory models.ExpenseCategory

	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	err := h.db.QueryRow(expenseCategorySQL.CreateExpenseCategoryQuery,
		req.CategoryName, req.Description, isActive).
		Scan(&expenseCategory.ID, &expenseCategory.CategoryName, &expenseCategory.Description, &expenseCategory.IsActive, &expenseCategory.CreatedAt, &expenseCategory.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"category_name": req.CategoryName,
		}).Error("Failed to create expense category in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id":   expenseCategory.ID,
		"expense_category_name": expenseCategory.CategoryName,
	}).Info("Expense category created successfully")

	return &expenseCategory, nil
}

// GetExpenseCategoryByID retrieves an expense category by ID from the database
func (h *DBHandler) GetExpenseCategoryByID(id string) (*models.ExpenseCategory, error) {
	var expenseCategory models.ExpenseCategory

	err := h.db.QueryRow(expenseCategorySQL.GetExpenseCategoryByIDQuery, id).
		Scan(&expenseCategory.ID, &expenseCategory.CategoryName, &expenseCategory.Description, &expenseCategory.IsActive, &expenseCategory.CreatedAt, &expenseCategory.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"expense_category_id": id,
		}).Error("Failed to retrieve expense category from database")
		return nil, err
	}

	return &expenseCategory, nil
}

// ListExpenseCategories retrieves all expense categories from the database
func (h *DBHandler) ListExpenseCategories() ([]models.ExpenseCategory, error) {
	rows, err := h.db.Query(expenseCategorySQL.ListExpenseCategoriesQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute expense categories list query")
		return nil, err
	}
	defer rows.Close()

	var expenseCategories []models.ExpenseCategory
	for rows.Next() {
		var expenseCategory models.ExpenseCategory
		err := rows.Scan(&expenseCategory.ID, &expenseCategory.CategoryName, &expenseCategory.Description, &expenseCategory.IsActive, &expenseCategory.CreatedAt, &expenseCategory.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to scan expense category row, skipping")
			continue
		}
		expenseCategories = append(expenseCategories, expenseCategory)
	}

	// Ensure we return an empty slice instead of nil for consistency
	if expenseCategories == nil {
		expenseCategories = []models.ExpenseCategory{}
	}

	h.logger.WithFields(logrus.Fields{
		"expense_categories_count": len(expenseCategories),
	}).Info("Listed expense categories successfully")

	return expenseCategories, nil
}

// UpdateExpenseCategory updates an expense category in the database
func (h *DBHandler) UpdateExpenseCategory(id string, req models.UpdateExpenseCategoryRequest) (*models.ExpenseCategory, error) {
	var expenseCategory models.ExpenseCategory

	err := h.db.QueryRow(expenseCategorySQL.UpdateExpenseCategoryQuery,
		id, req.CategoryName, req.Description, req.IsActive).
		Scan(&expenseCategory.ID, &expenseCategory.CategoryName, &expenseCategory.Description, &expenseCategory.IsActive, &expenseCategory.CreatedAt, &expenseCategory.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		h.logger.WithError(err).WithFields(logrus.Fields{
			"expense_category_id": id,
		}).Error("Failed to update expense category in database")
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id":   expenseCategory.ID,
		"expense_category_name": expenseCategory.CategoryName,
	}).Info("Expense category updated successfully")

	return &expenseCategory, nil
}

// DeleteExpenseCategory deletes an expense category from the database
func (h *DBHandler) DeleteExpenseCategory(id string) error {
	result, err := h.db.Exec(expenseCategorySQL.DeleteExpenseCategoryQuery, id)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"expense_category_id": id,
		}).Error("Failed to execute expense category delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected for expense category delete")
		return err
	}

	if rowsAffected == 0 {
		h.logger.WithFields(logrus.Fields{
			"expense_category_id": id,
		}).Warn("No expense category found to delete")
		return sql.ErrNoRows
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id": id,
		"rows_affected":       rowsAffected,
	}).Info("Expense category deleted successfully")

	return nil
} 