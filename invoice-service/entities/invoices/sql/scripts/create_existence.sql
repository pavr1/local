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
    $5,  -- items_per_unit
    $6,  -- cost_per_unit
    $7,  -- expiration_date
    $8,  -- income_margin_percentage
    $9,  -- income_margin_amount
    $10, -- iva_percentage
    $11, -- iva_amount
    $12, -- service_tax_percentage
    $13, -- service_tax_amount
    $14, -- calculated_price
    $15  -- final_price
); 