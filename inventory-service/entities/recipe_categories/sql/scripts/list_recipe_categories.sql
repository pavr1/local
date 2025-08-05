SELECT id, name, description, created_at, updated_at 
FROM recipe_categories 
WHERE ($1::varchar IS NULL OR name ILIKE '%' || $1 || '%')
ORDER BY name ASC
LIMIT $2 OFFSET $3; 