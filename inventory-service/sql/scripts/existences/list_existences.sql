SELECT 
    id, existence_reference_code, ingredient_id, expense_receipt_id,
    units_purchased, units_available, unit_type, items_per_unit,
    cost_per_item, cost_per_unit, total_purchase_cost, remaining_value,
    expiration_date, income_margin_percentage, income_margin_amount,
    iva_percentage, iva_amount, service_tax_percentage, service_tax_amount,
    calculated_price, final_price, created_at, updated_at
FROM existences
ORDER BY existence_reference_code DESC; 