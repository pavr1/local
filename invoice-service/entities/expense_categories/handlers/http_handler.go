package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"invoice-service/entities/expense_categories/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// pvillalobos - crete unit tests for this
// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	CreateExpenseCategory(req models.CreateExpenseCategoryRequest, logger *logrus.Logger) (*models.ExpenseCategory, error)
	GetExpenseCategoryByID(id string, logger *logrus.Logger) (*models.ExpenseCategory, error)
	ListExpenseCategories(logger *logrus.Logger) ([]models.ExpenseCategory, error)
	UpdateExpenseCategory(id string, req models.UpdateExpenseCategoryRequest, logger *logrus.Logger) (*models.ExpenseCategory, error)
	DeleteExpenseCategory(id string, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// HttpHandler handles HTTP requests for existence operations
type HttpHandler struct {
	dbHandler DBHandlerInterface
	logger    *logrus.Logger
}

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface, logger *logrus.Logger) *HttpHandler {
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
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	expenseCategory, err := h.dbHandler.CreateExpenseCategory(req, h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.ExpenseCategory{},
			Message: "Failed to create expense category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *expenseCategory,
		Message: "Expense category created successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id":   expenseCategory.ID,
		"expense_category_name": expenseCategory.CategoryName,
	}).Info("Expense category created successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// GetExpenseCategory handles GET /expense-categories/{id}
func (h *HttpHandler) GetExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in get request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Expense category ID is required",
		})
		return
	}

	expenseCategory, err := h.dbHandler.GetExpenseCategoryByID(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.ExpenseCategory{},
				Message: "Expense category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.ExpenseCategory{},
			Message: "Failed to retrieve expense category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *expenseCategory,
		Message: "Expense category retrieved successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// ListExpenseCategories handles GET /expense-categories
func (h *HttpHandler) ListExpenseCategories(w http.ResponseWriter, r *http.Request) {
	expenseCategories, err := h.dbHandler.ListExpenseCategories(h.logger)
	if err != nil {
		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.ExpenseCategory{},
			Message: "Failed to list expense categories: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    expenseCategories,
		Message: "Expense categories listed successfully",
	}
	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Expense categories listed successfully",
		Data:    response,
	})
}

// UpdateExpenseCategory handles PUT /expense-categories/{id}
func (h *HttpHandler) UpdateExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in update request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Expense category ID is required",
		})
		return
	}

	var req models.UpdateExpenseCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid JSON in update expense category request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	expenseCategory, err := h.dbHandler.UpdateExpenseCategory(id, req, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Expense category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update expense category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Expense category updated successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id":   expenseCategory.ID,
		"expense_category_name": expenseCategory.CategoryName,
	}).Info("Expense category updated successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}

// DeleteExpenseCategory handles DELETE /expense-categories/{id}
func (h *HttpHandler) DeleteExpenseCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.logger.Warn("Missing expense category ID in delete request")
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Expense category ID is required"})
		return
	}

	err := h.dbHandler.DeleteExpenseCategory(id, h.logger)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is expected behavior, don't log as error
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Expense category not found",
			}
			httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
			return
		}

		// DBHandler already logged the error, don't duplicate
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete expense category: " + err.Error(),
		}
		httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Expense category deleted successfully",
	}

	h.logger.WithFields(logrus.Fields{
		"expense_category_id": id,
	}).Info("Expense category deleted successfully")

	httpresponse.WriteResponse(w, h.logger, sharedLogger.SERVICE_INVOICE_SERVICE, response)
}
