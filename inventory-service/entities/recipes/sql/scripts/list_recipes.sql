SELECT id, recipe_name, recipe_description, picture_url, recipe_category_id, total_recipe_cost, created_at, updated_at 
FROM recipes 
WHERE ($1::varchar IS NULL OR recipe_name ILIKE '%' || $1 || '%')
  AND ($2::uuid IS NULL OR recipe_category_id = $2)
ORDER BY recipe_name ASC
LIMIT $3 OFFSET $4; 