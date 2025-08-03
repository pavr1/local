UPDATE invoice 
SET invoice_number = COALESCE($2, invoice_number),
    transaction_date = COALESCE($3, transaction_date),
    transaction_type = COALESCE($4, transaction_type),
    supplier_id = COALESCE($5, supplier_id),
    expense_category_id = COALESCE($6, expense_category_id),
    image_url = COALESCE($7, image_url),
    notes = COALESCE($8, notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at 