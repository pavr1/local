UPDATE runout_ingredient_report 
SET quantity = COALESCE($2, quantity),
    unit_type = COALESCE($3, unit_type),
    report_date = COALESCE($4, report_date),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at; 