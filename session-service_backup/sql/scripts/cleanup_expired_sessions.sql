-- Cleanup expired sessions (deactivate them)
UPDATE sessions 
SET is_active = false 
WHERE expires_at < CURRENT_TIMESTAMP AND is_active = true; 