SELECT id, name, supplier_id, created_at, updated_at
FROM ingredients
WHERE supplier_id = $1
ORDER BY name ASC; 