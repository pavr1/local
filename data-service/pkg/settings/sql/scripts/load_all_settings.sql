SELECT 
    setting_id,
    service,
    key,
    value,
    description,
    created_at,
    updated_at
FROM settings 
ORDER BY service, key;
