UPDATE existences 
SET 
    units_available = COALESCE($2, units_available),
    cost_per_unit = COALESCE($3, cost_per_unit),
    expiration_date = COALESCE($4, expiration_date),
    final_price = COALESCE($5, final_price),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING 
    id, existence_reference_code, ingredient_id, expense_receipt_id,
    units_purchased, units_available, unit_type, items_per_unit,
    cost_per_item, cost_per_unit, total_purchase_cost, remaining_value,
    expiration_date, income_margin_percentage, income_margin_amount,
    iva_percentage, iva_amount, service_tax_percentage, service_tax_amount,
    calculated_price, final_price, created_at, updated_at; 