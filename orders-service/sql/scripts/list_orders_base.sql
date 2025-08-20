-- Base query for listing orders (filters will be added dynamically)
SELECT id, order_number, customer_id, sales_representative_id, status, payment_method,
       transaction_reference, sinpe_screenshot_url, subtotal_amount, discount_amount,
       iva_amount, service_tax_amount, total_amount, invoice_number, invoice_url,
       transaction_timestamp, completed_at, created_at, updated_at
FROM orders 