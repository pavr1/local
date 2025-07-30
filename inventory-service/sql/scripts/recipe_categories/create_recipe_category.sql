INSERT INTO recipe_categories (id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, description, created_at, updated_at; 