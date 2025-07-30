INSERT INTO recipe_ingredients (id, recipe_id, ingredient_id, number_of_units, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, recipe_id, ingredient_id, number_of_units, created_at, updated_at; 