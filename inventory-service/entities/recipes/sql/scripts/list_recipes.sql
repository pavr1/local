SELECT r.id, r.recipe_name, r.recipe_description, r.picture_url, r.recipe_category_id, 
       rc.name as recipe_category_name, r.total_recipe_cost, r.created_at, r.updated_at 
FROM recipes r
LEFT JOIN recipe_categories rc ON r.recipe_category_id = rc.id
WHERE ($1::varchar IS NULL OR r.recipe_name ILIKE '%' || $1 || '%')
  AND ($2::uuid IS NULL OR r.recipe_category_id = $2)
ORDER BY r.recipe_name ASC
LIMIT $3 OFFSET $4; 