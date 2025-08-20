SELECT id, recipe_id, ingredient_id, quantity, created_at, updated_at 
FROM recipe_ingredients 
WHERE id = $1; 