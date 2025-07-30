INSERT INTO suppliers (id, supplier_name, contact_number, email, address, notes, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, supplier_name, contact_number, email, address, notes, created_at, updated_at; 