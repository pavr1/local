SELECT r.id, r.recipe_name, r.recipe_description, r.picture_url, r.recipe_category_id, 
       rc.name as recipe_category_name, r.total_recipe_cost, r.created_at, r.updated_at 
FROM recipes r
LEFT JOIN recipe_categories rc ON r.recipe_category_id = rc.id
WHERE r.id = $1; 