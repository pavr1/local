INSERT INTO ingredients (id, name, supplier_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, supplier_id, created_at, updated_at; 