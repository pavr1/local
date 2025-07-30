package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"inventory-service/config"
	"inventory-service/models"
	inventorySQL "inventory-service/sql"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type InventoryHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logrus.Logger
}

func NewInventoryHandler(db *sql.DB, cfg *config.Config, logger *logrus.Logger) *InventoryHandler {
	return &InventoryHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// Health check endpoint
func (h *InventoryHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Test database connection
	if err := h.db.Ping(); err != nil {
		h.logger.WithError(err).Error("Health check failed - database connection error")
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Service:   "inventory-service",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Supplier handlers
func (h *InventoryHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var supplier models.Supplier
	err := h.db.QueryRow(inventorySQL.CreateSupplierQuery,
		req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create supplier")
		http.Error(w, "Failed to create supplier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(supplier)
}

func (h *InventoryHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var supplier models.Supplier
	err := h.db.QueryRow(inventorySQL.GetSupplierByIDQuery, id).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Supplier not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get supplier")
		http.Error(w, "Failed to get supplier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

func (h *InventoryHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(inventorySQL.ListSuppliersQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list suppliers")
		http.Error(w, "Failed to list suppliers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		err := rows.Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan supplier")
			continue
		}
		suppliers = append(suppliers, supplier)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *InventoryHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var supplier models.Supplier
	err := h.db.QueryRow(inventorySQL.UpdateSupplierQuery,
		id, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Supplier not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to update supplier")
		http.Error(w, "Failed to update supplier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

func (h *InventoryHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.db.Exec(inventorySQL.DeleteSupplierQuery, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete supplier")
		http.Error(w, "Failed to delete supplier", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Supplier not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Ingredient handlers
func (h *InventoryHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var ingredient models.Ingredient
	err := h.db.QueryRow(inventorySQL.CreateIngredientQuery,
		req.Name, req.SupplierID).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID,
			&ingredient.CreatedAt, &ingredient.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create ingredient")
		http.Error(w, "Failed to create ingredient", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ingredient)
}

func (h *InventoryHandler) GetIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ingredient models.Ingredient
	err := h.db.QueryRow(inventorySQL.GetIngredientByIDQuery, id).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID,
			&ingredient.CreatedAt, &ingredient.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Ingredient not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to get ingredient")
		http.Error(w, "Failed to get ingredient", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredient)
}

func (h *InventoryHandler) ListIngredients(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(inventorySQL.ListIngredientsQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list ingredients")
		http.Error(w, "Failed to list ingredients", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ingredient models.Ingredient
		err := rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID,
			&ingredient.CreatedAt, &ingredient.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan ingredient")
			continue
		}
		ingredients = append(ingredients, ingredient)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredients)
}

func (h *InventoryHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var ingredient models.Ingredient
	err := h.db.QueryRow(inventorySQL.UpdateIngredientQuery,
		id, req.Name, req.SupplierID).
		Scan(&ingredient.ID, &ingredient.Name, &ingredient.SupplierID,
			&ingredient.CreatedAt, &ingredient.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Ingredient not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("Failed to update ingredient")
		http.Error(w, "Failed to update ingredient", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredient)
}

func (h *InventoryHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.db.Exec(inventorySQL.DeleteIngredientQuery, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete ingredient")
		http.Error(w, "Failed to delete ingredient", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Placeholder handlers for other entities (will implement in detail if needed)
func (h *InventoryHandler) CreateExistence(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) GetExistence(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListExistences(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) UpdateExistence(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) DeleteExistence(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListLowStock(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListExpiringSoon(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) CreateRunoutReport(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) GetRunoutReport(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListRunoutReports(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) CreateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) GetRecipeCategory(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListRecipeCategories(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) UpdateRecipeCategory(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) DeleteRecipeCategory(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) CreateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) GetRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) ListRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *InventoryHandler) DeleteRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}
