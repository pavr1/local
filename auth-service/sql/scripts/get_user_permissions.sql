-- Get all permissions for a given role
SELECT id, permission_name, description, role_id, created_at, updated_at
FROM permissions
WHERE role_id = $1
ORDER BY permission_name; 