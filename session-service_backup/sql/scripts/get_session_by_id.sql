-- Get session by session ID
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
WHERE session_id = $1; 