SELECT 
    e.id,
    e.existence_reference_code,
    e.ingredient_id,
    i.name as ingredient_name,
    ic.name as ingredient_category,
    e.invoice_detail_id,
    e.units_purchased,
    e.units_available,
    e.unit_type,
    e.items_per_unit,
    e.cost_per_item,
    e.cost_per_unit,
    e.total_purchase_cost,
    e.remaining_value,
    e.expiration_date,
    e.income_margin_percentage,
    e.income_margin_amount,
    e.iva_percentage,
    e.iva_amount,
    e.service_tax_percentage,
    e.service_tax_amount,
    e.calculated_price,
    e.final_price,
    e.created_at,
    e.updated_at
FROM existences e
LEFT JOIN ingredients i ON e.ingredient_id = i.id
LEFT JOIN ingredient_categories ic ON i.ingredient_category_id = ic.id
WHERE 1=1
    AND ($1::uuid IS NULL OR e.ingredient_id = $1)
    AND ($2::varchar IS NULL OR e.unit_type = $2)
    AND ($3::boolean IS NULL OR ($3 = true AND e.expiration_date < CURRENT_DATE) OR ($3 = false AND (e.expiration_date IS NULL OR e.expiration_date >= CURRENT_DATE)))
    AND ($4::boolean IS NULL OR ($4 = true AND e.units_available <= (e.units_purchased * 0.1)) OR ($4 = false AND e.units_available > (e.units_purchased * 0.1)))
ORDER BY e.created_at DESC
LIMIT COALESCE($5, 50) OFFSET COALESCE($6, 0); 