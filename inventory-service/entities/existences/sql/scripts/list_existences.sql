SELECT 
    id,
    existence_reference_code,
    ingredient_id,
    invoice_detail_id,
    units_purchased,
    units_available,
    unit_type,
    items_per_unit,
    cost_per_item,
    cost_per_unit,
    total_purchase_cost,
    remaining_value,
    expiration_date,
    income_margin_percentage,
    income_margin_amount,
    iva_percentage,
    iva_amount,
    service_tax_percentage,
    service_tax_amount,
    calculated_price,
    final_price,
    created_at,
    updated_at
FROM existences 
WHERE 1=1
    AND ($1::uuid IS NULL OR ingredient_id = $1)
    AND ($2::varchar IS NULL OR unit_type = $2)
    AND ($3::boolean IS NULL OR ($3 = true AND expiration_date < CURRENT_DATE) OR ($3 = false AND (expiration_date IS NULL OR expiration_date >= CURRENT_DATE)))
    AND ($4::boolean IS NULL OR ($4 = true AND units_available <= (units_purchased * 0.1)) OR ($4 = false AND units_available > (units_purchased * 0.1)))
ORDER BY created_at DESC
LIMIT COALESCE($5, 50) OFFSET COALESCE($6, 0); 