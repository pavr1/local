INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) 
VALUES ($1, $2, $3) 
RETURNING id, recipe_id, ingredient_id, quantity, created_at, updated_at; 