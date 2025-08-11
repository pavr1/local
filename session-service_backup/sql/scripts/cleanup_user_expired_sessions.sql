-- Cleanup expired sessions for a specific user (deactivate them)
UPDATE sessions 
SET is_active = false 
WHERE user_id = $1 AND expires_at < CURRENT_TIMESTAMP AND is_active = true; 