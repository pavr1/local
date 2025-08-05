INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit_type) 
VALUES ($1, $2, $3, $4) 
RETURNING id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at; 