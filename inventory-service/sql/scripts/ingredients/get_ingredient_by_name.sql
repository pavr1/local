SELECT id, name, supplier_id, created_at, updated_at
FROM ingredients
WHERE name = $1; 