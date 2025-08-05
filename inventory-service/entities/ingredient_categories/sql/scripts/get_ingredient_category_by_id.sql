SELECT id, name, description, is_active, created_at, updated_at
FROM ingredient_categories
WHERE id = $1; 