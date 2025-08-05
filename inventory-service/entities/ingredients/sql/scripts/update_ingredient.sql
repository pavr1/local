UPDATE ingredients 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    supplier_id = COALESCE($4, supplier_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, supplier_id, created_at, updated_at; 