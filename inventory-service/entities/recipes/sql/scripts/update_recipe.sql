UPDATE recipes 
SET recipe_name = COALESCE($2, recipe_name),
    recipe_description = COALESCE($3, recipe_description),
    picture_url = COALESCE($4, picture_url),
    recipe_category_id = COALESCE($5, recipe_category_id),
    total_recipe_cost = COALESCE($6, total_recipe_cost),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at; 