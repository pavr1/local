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
    iva_percentage,
    service_tax_percentage,
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
    $8,  -- iva_percentage
    $9,  -- service_tax_percentage
    $10  -- final_price
); 