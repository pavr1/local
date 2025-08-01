UPDATE ingredients 
SET 
    name = COALESCE($2, name),
    supplier_id = COALESCE($3, supplier_id)
WHERE id = $1
RETURNING id, name, supplier_id; 