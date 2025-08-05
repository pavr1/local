SELECT id, name, description, supplier_id, created_at, updated_at
FROM ingredients
WHERE id = $1; 