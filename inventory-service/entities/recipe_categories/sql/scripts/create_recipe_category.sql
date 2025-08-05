INSERT INTO recipe_categories (name, description) 
VALUES ($1, $2) 
RETURNING id, name, description, created_at, updated_at; 