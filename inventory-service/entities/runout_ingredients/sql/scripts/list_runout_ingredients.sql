SELECT id, existence_id, employee_id, quantity, unit_type, report_date, created_at, updated_at 
FROM runout_ingredient_report 
WHERE ($1::uuid IS NULL OR existence_id = $1)
  AND ($2::uuid IS NULL OR employee_id = $2)
  AND ($3::varchar IS NULL OR unit_type = $3)
  AND ($4::date IS NULL OR report_date = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6; 