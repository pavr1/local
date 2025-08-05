SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at 
FROM recipe_ingredients 
WHERE id = $1; 