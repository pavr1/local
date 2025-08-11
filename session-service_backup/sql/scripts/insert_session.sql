-- Insert a new session into the database
INSERT INTO sessions (
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
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
); 