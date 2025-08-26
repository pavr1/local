INSERT INTO existences (
    ingredient_id,
    invoice_detail_id,
    units_purchased,
    units_available,
    unit_type,
    items_per_unit,
    cost_per_unit,
    expiration_date,
    income_margin_percentage,
    minimum_price,
    maximum_price,
    final_price
) VALUES (
    $1,  -- ingredient_id
    $2,  -- invoice_detail_id
    $3,  -- units_purchased
    $4,  -- units_available
    $5,  -- unit_type
    $6,  -- items_per_unit
    $7,  -- cost_per_unit
    $8,  -- expiration_date
    COALESCE($9, 30.00),   -- income_margin_percentage (default 30%)
    $10, -- minimum_price
    $11, -- maximum_price
    $12  -- final_price
) RETURNING id, existence_reference_code, ingredient_id, invoice_detail_id, 
           units_purchased, units_available, unit_type, items_per_unit,
           cost_per_item, cost_per_unit, total_purchase_cost, remaining_value,
           expiration_date, income_margin_percentage, income_margin_amount,
           minimum_price, maximum_price, final_price, created_at, updated_at; 