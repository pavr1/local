UPDATE ingredients 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    ingredient_category_id = COALESCE($4, ingredient_category_id),
    supplier_id = COALESCE($5, supplier_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, ingredient_category_id, supplier_id, created_at, updated_at; 