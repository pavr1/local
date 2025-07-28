-- Get order summary statistics
SELECT 
    COUNT(*) as total_orders,
    COUNT(CASE WHEN order_status = 'pending' THEN 1 END) as pending_orders,
    COUNT(CASE WHEN order_status = 'completed' THEN 1 END) as completed_orders,
    COUNT(CASE WHEN order_status = 'cancelled' THEN 1 END) as cancelled_orders,
    COALESCE(SUM(CASE WHEN order_status = 'completed' THEN final_amount ELSE 0 END), 0) as total_revenue,
    COALESCE(AVG(CASE WHEN order_status = 'completed' THEN final_amount ELSE NULL END), 0) as average_order
FROM orders; 