SELECT id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at
FROM invoice
WHERE expense_category_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3 