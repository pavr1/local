-- Get all recipes that use a specific ingredient
-- This query returns recipe IDs that need to be recalculated when an ingredient gets its first existence

SELECT DISTINCT r.id
FROM recipes r
INNER JOIN recipe_ingredients ri ON r.id = ri.recipe_id
WHERE ri.ingredient_id = $1;
