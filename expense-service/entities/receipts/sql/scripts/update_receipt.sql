UPDATE receipts 
SET receipt_number = COALESCE($2, receipt_number),
    purchase_date = COALESCE($3, purchase_date),
    supplier_id = COALESCE($4, supplier_id),
    expense_category_id = COALESCE($5, expense_category_id),
    image_url = COALESCE($6, image_url),
    notes = COALESCE($7, notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, receipt_number, purchase_date, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at 