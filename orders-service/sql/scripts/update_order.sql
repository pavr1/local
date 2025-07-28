-- Base update order query (will be built dynamically)
-- This file is kept for reference, actual query is built in Go code
UPDATE orders 
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1; 