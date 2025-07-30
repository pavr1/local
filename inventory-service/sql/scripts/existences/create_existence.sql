INSERT INTO existences (
    id, ingredient_id, expense_receipt_id, units_purchased, units_available, 
    unit_type, items_per_unit, cost_per_unit, expiration_date,
    income_margin_percentage, iva_percentage, service_tax_percentage, final_price,
    created_at, updated_at
)
VALUES (
    gen_random_uuid(), $1, $2, $3, $3, $4, $5, $6, $7,
    COALESCE($8, 30.00), COALESCE($9, 13.00), COALESCE($10, 10.00), $11,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING 
    id, existence_reference_code, ingredient_id, expense_receipt_id,
    units_purchased, units_available, unit_type, items_per_unit,
    cost_per_item, cost_per_unit, total_purchase_cost, remaining_value,
    expiration_date, income_margin_percentage, income_margin_amount,
    iva_percentage, iva_amount, service_tax_percentage, service_tax_amount,
    calculated_price, final_price, created_at, updated_at; 