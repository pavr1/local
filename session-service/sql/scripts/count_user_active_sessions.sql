-- Count active sessions for a user
SELECT COUNT(*) as active_count
FROM sessions 
WHERE user_id = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP; 