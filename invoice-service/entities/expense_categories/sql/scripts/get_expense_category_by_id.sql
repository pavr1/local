SELECT id, category_name, description, is_active, created_at, updated_at
FROM expense_categories
WHERE id = $1; 