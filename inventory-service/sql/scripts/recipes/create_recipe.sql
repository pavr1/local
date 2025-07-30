INSERT INTO recipes (id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at; 