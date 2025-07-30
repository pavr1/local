package sql

import _ "embed"

// Supplier SQL queries
//
//go:embed scripts/create_supplier.sql
var CreateSupplierQuery string

//go:embed scripts/get_supplier_by_id.sql
var GetSupplierByIDQuery string

//go:embed scripts/list_suppliers.sql
var ListSuppliersQuery string

//go:embed scripts/update_supplier.sql
var UpdateSupplierQuery string

//go:embed scripts/delete_supplier.sql
var DeleteSupplierQuery string
