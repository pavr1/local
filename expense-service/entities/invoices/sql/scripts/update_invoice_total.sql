UPDATE invoice 
SET total_amount = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at 