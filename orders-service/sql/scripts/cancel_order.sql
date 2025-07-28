-- Cancel an order (set status to cancelled)
UPDATE orders 
SET order_status = 'cancelled', updated_at = $1 
WHERE id = $2 AND order_status != 'completed'; 