UPDATE expense_categories 
SET category_name = COALESCE($2, category_name),
    description = COALESCE($3, description),
    is_active = COALESCE($4, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, category_name, description, is_active, created_at, updated_at; 