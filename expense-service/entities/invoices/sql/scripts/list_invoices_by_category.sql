SELECT id, receipt_number, purchase_date, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at
FROM receipts
WHERE expense_category_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3 