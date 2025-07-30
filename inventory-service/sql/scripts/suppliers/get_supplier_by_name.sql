SELECT id, supplier_name, contact_number, email, address, notes, created_at, updated_at
FROM suppliers
WHERE supplier_name = $1; 