package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"inventory-service/entities/suppliers/models"
	supplierSQL "inventory-service/entities/suppliers/sql"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SupplierHandler handles HTTP requests for supplier operations
type SupplierHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewSupplierHandler creates a new supplier handler
func NewSupplierHandler(db *sql.DB, logger *logrus.Logger) *SupplierHandler {
	return &SupplierHandler{
		db:     db,
		logger: logger,
	}
}

// CreateSupplier handles POST /suppliers
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.CreateSupplierQuery,
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

// GetSupplier handles GET /suppliers/{id}
func (h *SupplierHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.GetSupplierByIDQuery, id).
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

// ListSuppliers handles GET /suppliers
func (h *SupplierHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(supplierSQL.ListSuppliersQuery)
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

// UpdateSupplier handles PUT /suppliers/{id}
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.UpdateSupplierQuery,
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

// DeleteSupplier handles DELETE /suppliers/{id}
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.db.Exec(supplierSQL.DeleteSupplierQuery, id)
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
