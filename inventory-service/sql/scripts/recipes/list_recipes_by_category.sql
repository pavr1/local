SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at
FROM recipes WHERE recipe_category_id = $1 ORDER BY recipe_name ASC; 