UPDATE recipe_categories 
SET name = COALESCE($2, name),
    description = COALESCE($3, description),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, created_at, updated_at; 