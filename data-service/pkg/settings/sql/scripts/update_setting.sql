UPDATE settings 
SET value = $1, updated_at = NOW() 
WHERE service = $2 AND key = $3;
