INSERT INTO recipes (recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, status) 
VALUES ($1, $2, $3, $4, $5, 'pending') 
RETURNING id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, status, created_at, updated_at; 