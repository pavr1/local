SELECT id, recipe_id, ingredient_id, quantity, unit_type, created_at, updated_at 
FROM recipe_ingredients 
WHERE ($1::uuid IS NULL OR recipe_id = $1)
  AND ($2::uuid IS NULL OR ingredient_id = $2)
ORDER BY recipe_id ASC, ingredient_id ASC
LIMIT $3 OFFSET $4; 