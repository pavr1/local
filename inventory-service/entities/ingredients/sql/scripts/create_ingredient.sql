INSERT INTO ingredients (id, name, description, supplier_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, description, supplier_id, created_at, updated_at; 