UPDATE ingredients 
SET 
    ingredient_name = COALESCE($2, ingredient_name),
    ingredient_type = COALESCE($3, ingredient_type),
    unit_of_measure = COALESCE($4, unit_of_measure),
    cost_per_unit = COALESCE($5, cost_per_unit),
    supplier_id = COALESCE($6, supplier_id),
    minimum_stock_level = COALESCE($7, minimum_stock_level),
    notes = COALESCE($8, notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, ingredient_name, ingredient_type, unit_of_measure, cost_per_unit, supplier_id, minimum_stock_level, notes, created_at, updated_at; 