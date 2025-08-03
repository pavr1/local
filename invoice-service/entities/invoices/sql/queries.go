package sql

import _ "embed"

// Invoice SQL queries

//go:embed scripts/create_invoice.sql
var CreateInvoiceQuery string

//go:embed scripts/get_invoice_by_id.sql
var GetInvoiceByIDQuery string

//go:embed scripts/get_invoice_by_number.sql
var GetInvoiceByNumberQuery string

//go:embed scripts/list_invoices.sql
var ListInvoicesBaseQuery string

//go:embed scripts/list_invoices_by_category.sql
var ListInvoicesByExpenseCategoryQuery string

//go:embed scripts/count_invoices.sql
var CountInvoicesQuery string

//go:embed scripts/update_invoice.sql
var UpdateInvoiceQuery string

//go:embed scripts/update_invoice_total.sql
var UpdateInvoiceTotalQuery string

//go:embed scripts/delete_invoice.sql
var DeleteInvoiceQuery string

// Invoice Details SQL queries

//go:embed scripts/create_invoice_detail.sql
var CreateInvoiceDetailQuery string

//go:embed scripts/get_invoice_detail_by_id.sql
var GetInvoiceDetailByIDQuery string

//go:embed scripts/get_invoice_details_by_invoice_id.sql
var ListInvoiceDetailsByInvoiceIDQuery string

//go:embed scripts/list_invoice_details.sql
var ListInvoiceDetailsBaseQuery string

//go:embed scripts/get_invoice_total_from_details.sql
var GetInvoiceTotalFromDetailsQuery string

// Placeholder queries for missing ones (to be created as needed)
const ListInvoicesBySupplierQuery = `
SELECT id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at
FROM invoice
WHERE supplier_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const CountInvoicesByExpenseCategoryQuery = `SELECT COUNT(*) FROM invoice WHERE expense_category_id = $1`
const CountInvoicesBySupplierQuery = `SELECT COUNT(*) FROM invoice WHERE supplier_id = $1`
const CountInvoiceDetailsByInvoiceQuery = `SELECT COUNT(*) FROM invoice_details WHERE invoice_id = $1`
const CountInvoiceDetailsQuery = `SELECT COUNT(*) FROM invoice_details`
const ListInvoiceDetailsByIngredientQuery = `
SELECT id, invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at
FROM invoice_details
WHERE ingredient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
const CountInvoiceDetailsByIngredientQuery = `SELECT COUNT(*) FROM invoice_details WHERE ingredient_id = $1`

const UpdateInvoiceDetailQuery = `
UPDATE invoice_details 
SET ingredient_id = COALESCE($2, ingredient_id),
    detail = COALESCE($3, detail),
    count = COALESCE($4, count),
    unit_type = COALESCE($5, unit_type),
    price = COALESCE($6, price),
    total = COALESCE($4, count) * COALESCE($6, price),
    expiration_date = COALESCE($7, expiration_date),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at`

const DeleteInvoiceDetailQuery = `DELETE FROM invoice_details WHERE id = $1`

// Health check query
const HealthCheckQuery = `SELECT 1`
