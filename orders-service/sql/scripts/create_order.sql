-- Create a new order
INSERT INTO orders (
    id, customer_id, order_date, total_amount, tax_amount, 
    discount_amount, final_amount, payment_method, order_status, notes,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
); 