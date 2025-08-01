-- Deactivate all sessions for a user
UPDATE sessions 
SET 
    is_active = false,
    last_activity = CURRENT_TIMESTAMP
WHERE user_id = $1 AND is_active = true; 