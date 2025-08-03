SELECT COALESCE(SUM(total), 0) as total_amount
FROM invoice_details
WHERE invoice_id = $1 