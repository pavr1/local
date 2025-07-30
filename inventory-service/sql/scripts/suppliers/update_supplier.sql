UPDATE suppliers 
SET 
    supplier_name = COALESCE($2, supplier_name),
    contact_number = COALESCE($3, contact_number),
    email = COALESCE($4, email),
    address = COALESCE($5, address),
    notes = COALESCE($6, notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, supplier_name, contact_number, email, address, notes, created_at, updated_at; 