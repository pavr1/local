UPDATE invoice_details 
SET ingredient_id = COALESCE($2, ingredient_id),
    detail = COALESCE($3, detail),
    count = COALESCE($4, count),
    unit_type = COALESCE($5, unit_type),
    price = COALESCE($6, price),
    expiration_date = COALESCE($7, expiration_date),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at; 