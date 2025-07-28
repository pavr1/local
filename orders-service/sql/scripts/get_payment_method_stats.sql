-- Get payment method statistics
SELECT 
    payment_method,
    COUNT(*) as count,
    COALESCE(SUM(final_amount), 0) as total_amount,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM orders 
WHERE order_status = 'completed'
GROUP BY payment_method
ORDER BY count DESC; 