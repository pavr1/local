-- Get order by ID
SELECT id, customer_id, order_date, total_amount, tax_amount,
       discount_amount, final_amount, payment_method, order_status,
       notes, created_at, updated_at
FROM orders 
WHERE id = $1; 