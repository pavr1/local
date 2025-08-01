INSERT INTO ingredients (id, ingredient_name, ingredient_type, unit_of_measure, cost_per_unit, supplier_id, minimum_stock_level, notes, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, ingredient_name, ingredient_type, unit_of_measure, cost_per_unit, supplier_id, minimum_stock_level, notes, created_at, updated_at; 