SELECT id, invoice_id, ingredient_id, detail, count, unit_type, items_per_unit, price, total, expiration_date, created_at, updated_at
FROM invoice_details
WHERE id = $1; 