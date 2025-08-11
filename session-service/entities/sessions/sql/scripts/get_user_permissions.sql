SELECT p.id, p.permission_name, p.description, p.role_id, p.created_at, p.updated_at
FROM permissions p
WHERE p.role_id = $1 