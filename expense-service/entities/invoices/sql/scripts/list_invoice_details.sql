SELECT id, invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at
FROM invoice_details
ORDER BY created_at DESC
LIMIT $1 OFFSET $2 