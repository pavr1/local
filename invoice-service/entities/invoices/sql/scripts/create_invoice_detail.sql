INSERT INTO invoice_details (invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, invoice_id, ingredient_id, detail, count, unit_type, price, total, expiration_date, created_at, updated_at 