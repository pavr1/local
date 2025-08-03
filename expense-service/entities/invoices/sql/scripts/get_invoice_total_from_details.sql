SELECT COALESCE(SUM(total), 0) as total_amount
FROM receipt_items
WHERE receipt_id = $1 