SELECT id, receipt_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at
FROM receipt_items
WHERE id = $1 