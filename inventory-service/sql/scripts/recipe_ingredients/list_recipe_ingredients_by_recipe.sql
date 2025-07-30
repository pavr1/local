SELECT id, recipe_id, ingredient_id, number_of_units, created_at, updated_at
FROM recipe_ingredients WHERE recipe_id = $1 ORDER BY created_at ASC; 