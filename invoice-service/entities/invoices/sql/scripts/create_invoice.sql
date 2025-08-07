INSERT INTO invoice (invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, image_url, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at; 