-- Get order summary statistics for today
SELECT 
    COUNT(*) as total_orders,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_orders,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_orders,
    COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_orders,
    COALESCE(SUM(CASE WHEN status = 'completed' THEN total_amount ELSE 0 END), 0) as total_revenue,
    COALESCE(AVG(CASE WHEN status = 'completed' THEN total_amount ELSE NULL END), 0) as average_order
FROM orders
WHERE DATE(transaction_timestamp) = CURRENT_DATE; 