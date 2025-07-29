-- Get user profile with role information by username
SELECT u.id, u.username, u.password_hash, u.full_name, u.role_id, u.is_active, 
       u.last_login, u.created_at, u.updated_at,
       r.id, r.role_name, r.description, r.created_at, r.updated_at
FROM users u
INNER JOIN roles r ON u.role_id = r.id
WHERE u.username = $1 AND u.is_active = true; 