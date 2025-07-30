UPDATE ingredients 
SET 
    name = COALESCE($2, name),
    supplier_id = COALESCE($3, supplier_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, supplier_id, created_at, updated_at; 