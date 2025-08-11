SELECT u.id, u.username, u.password_hash, u.full_name, u.role_id, u.is_active, u.last_login, u.created_at, u.updated_at,
       r.id as role_id, r.role_name, r.created_at as role_created_at, r.updated_at as role_updated_at
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.username = $1 AND u.is_active = true 