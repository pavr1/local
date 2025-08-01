-- Get all sessions for a user
SELECT 
    session_id,
    user_id,
    username,
    role_name,
    permissions,
    token_hash,
    created_at,
    expires_at,
    last_activity,
    is_active
FROM sessions 
WHERE user_id = $1
ORDER BY created_at DESC; 