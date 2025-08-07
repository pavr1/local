INSERT INTO expense_categories (category_name, description, is_active)
VALUES ($1, $2, $3)
RETURNING id, category_name, description, is_active, created_at, updated_at; 