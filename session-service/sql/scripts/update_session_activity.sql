-- Update session last activity and expiration
UPDATE sessions 
SET 
    last_activity = $2,
    expires_at = COALESCE($3, expires_at)
WHERE session_id = $1; 