INSERT INTO runout_ingredient_report (id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, COALESCE($5, CURRENT_DATE), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at; 