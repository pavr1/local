UPDATE ingredient_categories 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    is_active = COALESCE($4, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, is_active, created_at, updated_at; 