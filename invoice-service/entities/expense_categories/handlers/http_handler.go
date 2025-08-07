package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"invoice-service/entities/expense_categories/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateExpenseCategory(req models.CreateExpenseCategoryRequest) (*models.ExpenseCategory, error)
	GetExpenseCategoryByID(id string) (*models.ExpenseCategory, error)
	ListExpenseCategories() ([]models.ExpenseCategory, error)
	UpdateExpenseCategory(id string, req models.UpdateExpenseCategoryRequest) (*models.ExpenseCategory, error)
	DeleteExpenseCategory(id string) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for expense category operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler *DBHandler, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// NewHttpHandlerWithInterface creates a new HTTP handler with interface (for testing)
func NewHttpHandlerWithInterface(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
		logger:    logger,
	}
}

// CreateExpenseCategory handles POST /expense-categories
func (h *HttpHandler) CreateExpenseCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateExpenseCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in create expense category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	expenseCategory, err := h.dbHandler.CreateExpenseCategory(req)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.ExpenseCategoryResponse{
			Success: false,
			Data:    models.ExpenseCategory{},
			Message: "Failed to create expense category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.ExpenseCategoryResponse{
		Success: true,
		Data:    *expenseCategory,
		Message: "Expense category created successfully",
	}
	h.writeJSONResponse(w, response, http.StatusCreated)
}

// GetExpenseCategory handles GET /expense-categories/{id}
func (h *HttpHandler) GetExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in get request")
		h.writeErrorResponse(w, "Expense category ID is required", http.StatusBadRequest)
		return
	}

	expenseCategory, err := h.dbHandler.GetExpenseCategoryByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.ExpenseCategoryResponse{
				Success: false,
				Data:    models.ExpenseCategory{},
				Message: "Expense category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.ExpenseCategoryResponse{
			Success: false,
			Data:    models.ExpenseCategory{},
			Message: "Failed to retrieve expense category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.ExpenseCategoryResponse{
		Success: true,
		Data:    *expenseCategory,
		Message: "Expense category retrieved successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListExpenseCategories handles GET /expense-categories
func (h *HttpHandler) ListExpenseCategories(w http.ResponseWriter, r *http.Request) {
	expenseCategories, err := h.dbHandler.ListExpenseCategories()
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := models.ExpenseCategoriesListResponse{
			Success: false,
			Data:    []models.ExpenseCategory{},
			Count:   0,
			Message: "Failed to list expense categories: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.ExpenseCategoriesListResponse{
		Success: true,
		Data:    expenseCategories,
		Count:   len(expenseCategories),
		Message: "Expense categories listed successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// UpdateExpenseCategory handles PUT /expense-categories/{id}
func (h *HttpHandler) UpdateExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in update request")
		h.writeErrorResponse(w, "Expense category ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateExpenseCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update expense category request")
		h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	expenseCategory, err := h.dbHandler.UpdateExpenseCategory(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.ExpenseCategoryResponse{
				Success: false,
				Data:    models.ExpenseCategory{},
				Message: "Expense category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.ExpenseCategoryResponse{
			Success: false,
			Data:    models.ExpenseCategory{},
			Message: "Failed to update expense category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.ExpenseCategoryResponse{
		Success: true,
		Data:    *expenseCategory,
		Message: "Expense category updated successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// DeleteExpenseCategory handles DELETE /expense-categories/{id}
func (h *HttpHandler) DeleteExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in delete request")
		h.writeErrorResponse(w, "Expense category ID is required", http.StatusBadRequest)
		return
	}

	err := h.dbHandler.DeleteExpenseCategory(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := models.ExpenseCategoryDeleteResponse{
				Success: false,
				Message: "Expense category not found",
			}
			h.writeJSONResponse(w, response, http.StatusNotFound)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := models.ExpenseCategoryDeleteResponse{
			Success: false,
			Message: "Failed to delete expense category: " + err.Error(),
		}
		h.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := models.ExpenseCategoryDeleteResponse{
		Success: true,
		Message: "Expense category deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

// writeJSONResponse writes a JSON response with the given status code
func (h *HttpHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// writeErrorResponse writes an error response with the given message and status code
func (h *HttpHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := models.ErrorResponse{
		Success: false,
		Error:   message,
		Message: message,
	}
	h.writeJSONResponse(w, response, statusCode)
} 