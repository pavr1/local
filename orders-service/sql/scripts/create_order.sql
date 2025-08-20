-- Create a new order
INSERT INTO orders (
    id, order_number, customer_id, sales_representative_id, status, payment_method,
    transaction_reference, sinpe_screenshot_url, subtotal_amount, discount_amount,
    iva_amount, service_tax_amount, total_amount, invoice_number, invoice_url,
    transaction_timestamp, completed_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
); 