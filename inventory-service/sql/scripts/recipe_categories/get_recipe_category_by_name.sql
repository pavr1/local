SELECT id, name, description, created_at, updated_at
FROM recipe_categories
WHERE name = $1; 