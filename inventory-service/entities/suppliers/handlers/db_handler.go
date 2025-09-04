package handlers

import (
	"database/sql"

	"inventory-service/entities/suppliers/models"
	supplierSQL "inventory-service/entities/suppliers/sql"

	"github.com/sirupsen/logrus"
)

// DBHandler handles database operations for suppliers
type DBHandler struct {
	db *sql.DB
}

// NewDBHandler creates a new database handler for suppliers
func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{
		db: db,
	}
}

// CreateSupplier creates a new supplier in the database
func (h *DBHandler) CreateSupplier(req models.CreateSupplierRequest, logger *logrus.Logger) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.CreateSupplierQuery,
		req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_name": req.SupplierName,
		}).Error("Failed to create supplier in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"supplier_id":   supplier.ID,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier created successfully")

	return &supplier, nil
}

// GetSupplierByID retrieves a supplier by ID from the database
func (h *DBHandler) GetSupplierByID(id string, logger *logrus.Logger) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.GetSupplierByIDQuery, id).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_id": id,
		}).Error("Failed to retrieve supplier from database")
		return nil, err
	}

	return &supplier, nil
}

// ListSuppliers retrieves all suppliers from the database
func (h *DBHandler) ListSuppliers(logger *logrus.Logger) ([]models.Supplier, error) {
	rows, err := h.db.Query(supplierSQL.ListSuppliersQuery)
	if err != nil {
		logger.WithError(err).Error("Failed to execute suppliers list query")
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
			logger.WithError(err).Warn("Failed to scan supplier row, skipping")
			continue
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, nil
}

// UpdateSupplier updates a supplier in the database
func (h *DBHandler) UpdateSupplier(id string, req models.UpdateSupplierRequest, logger *logrus.Logger) (*models.Supplier, error) {
	var supplier models.Supplier

	err := h.db.QueryRow(supplierSQL.UpdateSupplierQuery,
		id, req.SupplierName, req.ContactNumber, req.Email, req.Address, req.Notes).
		Scan(&supplier.ID, &supplier.SupplierName, &supplier.ContactNumber,
			&supplier.Email, &supplier.Address, &supplier.Notes,
			&supplier.CreatedAt, &supplier.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't log as error since "not found" is a normal business case
			return nil, err
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_id": id,
		}).Error("Failed to update supplier in database")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"supplier_id":   supplier.ID,
		"supplier_name": supplier.SupplierName,
	}).Info("Supplier updated successfully")

	return &supplier, nil
}

// DeleteSupplier deletes a supplier from the database
func (h *DBHandler) DeleteSupplier(id string, logger *logrus.Logger) error {
	result, err := h.db.Exec(supplierSQL.DeleteSupplierQuery, id)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_id": id,
		}).Error("Failed to execute supplier delete query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"supplier_id": id,
		}).Error("Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		// Don't log as error since "not found" is a normal business case
		return sql.ErrNoRows
	}

	logger.WithFields(logrus.Fields{
		"supplier_id": id,
	}).Info("Supplier deleted successfully")

	return nil
}
