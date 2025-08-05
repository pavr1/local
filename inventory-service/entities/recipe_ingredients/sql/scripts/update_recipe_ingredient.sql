UPDATE recipe_ingredients 
SET recipe_id = COALESCE($2, recipe_id),
    ingredient_id = COALESCE($3, ingredient_id),
    quantity = COALESCE($4, quantity),
    unit_type = COALESCE($5, unit_type),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at; 