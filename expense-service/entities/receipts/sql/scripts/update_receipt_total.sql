UPDATE receipts 
SET total_amount = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, receipt_number, purchase_date, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at 