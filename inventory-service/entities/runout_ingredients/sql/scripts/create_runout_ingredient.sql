INSERT INTO runout_ingredient_report (existence_id, employee_id, quantity, unit_type, report_date) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at; 