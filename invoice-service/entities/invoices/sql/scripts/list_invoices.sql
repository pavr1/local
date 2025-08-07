SELECT id, invoice_number, transaction_date, transaction_type, supplier_id, expense_category_id, total_amount, image_url, notes, created_at, updated_at
FROM invoice
ORDER BY transaction_date DESC, created_at DESC; 