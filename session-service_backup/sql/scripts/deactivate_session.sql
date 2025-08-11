-- Deactivate a session (soft delete)
UPDATE sessions 
SET 
    is_active = false,
    last_activity = CURRENT_TIMESTAMP
WHERE session_id = $1; 