package suppliers

import (
	"database/sql"

	"inventory-service/entities/suppliers/models"
	supplierSQL "inventory-service/entities/suppliers/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for suppliers
type DBHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewDBHandler creates a new database handler for suppliers
func NewDBHandler(db *sql.DB, logger *logrus.Logger) *DBHandler {
	return &DBHandler{
		db:     db,
		logger: logger,
	}
}

// CreateSupplier creates a new supplier in the database
func (h *DBHandler) CreateSupplier(req models.CreateSupplierRequest) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.CreateSupplierQuery,
		req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create supplier in database")
		return nil, err
	}

	return &supplier, nil
}

// GetSupplierByID retrieves a supplier by ID from the database
func (h *DBHandler) GetSupplierByID(id string) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.GetSupplierByIDQuery, id).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Let the caller handle the not found case
		}
		h.logger.WithError(err).Error("Failed to get supplier from database")
		return nil, err
	}

	return &supplier, nil
}

// ListSuppliers retrieves all suppliers from the database
func (h *DBHandler) ListSuppliers() ([]models.Supplier, error) {
	rows, err := h.db.Query(supplierSQL.ListSuppliersQuery)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list suppliers from database")
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		err := rows.Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan supplier from database")
			continue
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, nil
}

// UpdateSupplier updates a supplier in the database
func (h *DBHandler) UpdateSupplier(id string, req models.UpdateSupplierRequest) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.UpdateSupplierQuery,
		id, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Let the caller handle the not found case
		}
		h.logger.WithError(err).Error("Failed to update supplier in database")
		return nil, err
	}

	return &supplier, nil
}

// DeleteSupplier deletes a supplier from the database
func (h *DBHandler) DeleteSupplier(id string) error {
	result, err := h.db.Exec(supplierSQL.DeleteSupplierQuery, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete supplier from database")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Supplier not found
	}

	return nil
}
