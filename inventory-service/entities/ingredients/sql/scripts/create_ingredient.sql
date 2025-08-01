INSERT INTO ingredients (id, name, supplier_id)
VALUES (gen_random_uuid(), $1, $2)
RETURNING id, name, supplier_id; 