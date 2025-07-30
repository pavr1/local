SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at
FROM runout_ingredient_report
WHERE employee_id = $1
ORDER BY report_date DESC; 