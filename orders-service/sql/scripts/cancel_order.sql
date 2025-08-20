-- Cancel an order (set status to cancelled)
UPDATE orders 
SET status = 'cancelled', updated_at = $1 
WHERE id = $2 AND status != 'completed'; 