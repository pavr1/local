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
    income_margin_amount,
    iva_percentage,
    iva_amount,
    service_tax_percentage,
    service_tax_amount,
    calculated_price,
    final_price
) VALUES (
    $1,  -- ingredient_id
    $2,  -- invoice_detail_id
    $3,  -- units_purchased
    $3,  -- units_available (same as units_purchased initially)
    $4,  -- unit_type
    1,   -- items_per_unit (set to 1 for now)
    $5,  -- cost_per_unit
    $6,  -- expiration_date
    $7,  -- income_margin_percentage
    $8,  -- income_margin_amount
    $9,  -- iva_percentage
    $10, -- iva_amount
    $11, -- service_tax_percentage
    $12, -- service_tax_amount
    $13, -- calculated_price
    $14  -- final_price
); 