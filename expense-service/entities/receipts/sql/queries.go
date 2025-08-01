package sql

import _ "embed"

// Receipt SQL queries

//go:embed scripts/create_receipt.sql
var CreateReceiptQuery string

//go:embed scripts/get_receipt_by_id.sql
var GetReceiptByIDQuery string

//go:embed scripts/get_receipt_by_number.sql
var GetReceiptByNumberQuery string

//go:embed scripts/list_receipts.sql
var ListReceiptsBaseQuery string

//go:embed scripts/list_receipts_by_category.sql
var ListReceiptsByExpenseCategoryQuery string

//go:embed scripts/count_receipts.sql
var CountReceiptsQuery string

//go:embed scripts/update_receipt.sql
var UpdateReceiptQuery string

//go:embed scripts/update_receipt_total.sql
var UpdateReceiptTotalQuery string

//go:embed scripts/delete_receipt.sql
var DeleteReceiptQuery string

// Receipt Items SQL queries

//go:embed scripts/create_receipt_item.sql
var CreateReceiptItemQuery string

//go:embed scripts/get_receipt_item_by_id.sql
var GetReceiptItemByIDQuery string

//go:embed scripts/get_receipt_items_by_receipt_id.sql
var ListReceiptItemsByReceiptIDQuery string

//go:embed scripts/list_receipt_items.sql
var ListReceiptItemsBaseQuery string

//go:embed scripts/get_receipt_total_from_items.sql
var GetReceiptTotalFromItemsQuery string

// Placeholder queries for missing ones (to be created as needed)
const ListReceiptsBySupplierQuery = `
SELECT id, receipt_number, purchase_date, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at
FROM receipts
WHERE supplier_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const CountReceiptsByExpenseCategoryQuery = `SELECT COUNT(*) FROM receipts WHERE expense_category_id = $1`
const CountReceiptsBySupplierQuery = `SELECT COUNT(*) FROM receipts WHERE supplier_id = $1`
const CountReceiptItemsByReceiptQuery = `SELECT COUNT(*) FROM receipt_items WHERE receipt_id = $1`
const CountReceiptItemsQuery = `SELECT COUNT(*) FROM receipt_items`
const ListReceiptItemsByIngredientQuery = `
SELECT id, receipt_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at
FROM receipt_items
WHERE ingredient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
const CountReceiptItemsByIngredientQuery = `SELECT COUNT(*) FROM receipt_items WHERE ingredient_id = $1`

const UpdateReceiptItemQuery = `
UPDATE receipt_items 
SET ingredient_id = COALESCE($2, ingredient_id),
    detail = COALESCE($3, detail),
    count = COALESCE($4, count),
    unit_type = COALESCE($5, unit_type),
    price = COALESCE($6, price),
    total = COALESCE($4, count) * COALESCE($6, price),
    expiration_date = COALESCE($7, expiration_date),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, receipt_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at`

const DeleteReceiptItemQuery = `DELETE FROM receipt_items WHERE id = $1`

// Health check query
const HealthCheckQuery = `SELECT 1`
