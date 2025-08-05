INSERT INTO ingredient_categories (id, name, description, is_active, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, COALESCE($3, TRUE), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, description, is_active, created_at, updated_at; 