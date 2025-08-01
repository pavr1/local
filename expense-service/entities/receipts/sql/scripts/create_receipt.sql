INSERT INTO receipts (receipt_number, purchase_date, supplier_id, expense_category_id, total_amount, image_url, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id 