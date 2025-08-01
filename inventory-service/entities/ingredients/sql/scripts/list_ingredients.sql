SELECT id, ingredient_name, ingredient_type, unit_of_measure, cost_per_unit, supplier_id, minimum_stock_level, notes, created_at, updated_at
FROM ingredients
ORDER BY ingredient_name ASC; 