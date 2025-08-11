-- Get session by token hash (most recent first)
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
WHERE token_hash = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT 1; 