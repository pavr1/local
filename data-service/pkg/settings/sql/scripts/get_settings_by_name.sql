SELECT 
    setting_id,
    service,
    key,
    value,
    description,
    created_at,
    updated_at
FROM settings 
WHERE key = $1
ORDER BY service;
