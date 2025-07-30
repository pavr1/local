package handlers

import (
	"database/sql"

	"inventory-service/entities/suppliers/models"
	supplierSQL "inventory-service/entities/suppliers/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for supplier operations
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// CreateSupplier creates a new supplier in the database
func (h *DBHandler) CreateSupplier(req models.CreateSupplierRequest) models.SupplierResponse {
	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.CreateSupplierQuery,
		req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create supplier")
		return models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to create supplier: " + err.Error(),
		}
	}

	return models.SupplierResponse{
		Success: true,
		Data:    supplier,
		Message: "Supplier created successfully",
	}
}

// GetSupplier retrieves a supplier by ID from the database
func (h *DBHandler) GetSupplier(req models.GetSupplierRequest) models.SupplierResponse {
	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.GetSupplierByIDQuery, req.ID).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.SupplierResponse{
				Success: false,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
		}
		h.logger.WithError(err).Error("Failed to get supplier")
		return models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to get supplier: " + err.Error(),
		}
	}

	return models.SupplierResponse{
		Success: true,
		Data:    supplier,
		Message: "Supplier retrieved successfully",
	}
}

// ListSuppliers retrieves all suppliers from the database
func (h *DBHandler) ListSuppliers(req models.ListSuppliersRequest) models.SuppliersListResponse {
	rows, err := h.db.Query(supplierSQL.ListSuppliersQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list suppliers")
		return models.SuppliersListResponse{
			Success: false,
			Data:    []models.Supplier{},
			Count:   0,
			Message: "Failed to list suppliers: " + err.Error(),
		}
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

	return models.SuppliersListResponse{
		Success: true,
		Data:    suppliers,
		Count:   len(suppliers),
		Message: "Suppliers retrieved successfully",
	}
}

// UpdateSupplier updates a supplier in the database
func (h *DBHandler) UpdateSupplier(id string, req models.UpdateSupplierRequest) models.SupplierResponse {
	var supplier models.Supplier
	err := h.db.QueryRow(supplierSQL.UpdateSupplierQuery,
		id, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.SupplierResponse{
				Success: false,
				Data:    models.Supplier{},
				Message: "Supplier not found",
			}
		}
		h.logger.WithError(err).Error("Failed to update supplier")
		return models.SupplierResponse{
			Success: false,
			Data:    models.Supplier{},
			Message: "Failed to update supplier: " + err.Error(),
		}
	}

	return models.SupplierResponse{
		Success: true,
		Data:    supplier,
		Message: "Supplier updated successfully",
	}
}

// DeleteSupplier deletes a supplier from the database
func (h *DBHandler) DeleteSupplier(req models.DeleteSupplierRequest) models.SupplierDeleteResponse {
	result, err := h.db.Exec(supplierSQL.DeleteSupplierQuery, req.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete supplier")
		return models.SupplierDeleteResponse{
			Success: false,
			Message: "Failed to delete supplier: " + err.Error(),
		}
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.SupplierDeleteResponse{
			Success: false,
			Message: "Supplier not found",
		}
	}

	return models.SupplierDeleteResponse{
		Success: true,
		Message: "Supplier deleted successfully",
	}
}
