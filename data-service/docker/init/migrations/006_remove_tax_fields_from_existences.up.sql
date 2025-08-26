-- Migration: Remove tax fields from existences table
-- Description: Remove IVA and Service Tax fields from existences to separate tax handling to orders
-- Date: 2025-08-26

-- Remove tax-related columns from existences table
ALTER TABLE existences DROP COLUMN IF EXISTS iva_percentage;
ALTER TABLE existences DROP COLUMN IF EXISTS iva_amount;
ALTER TABLE existences DROP COLUMN IF EXISTS service_tax_percentage;
ALTER TABLE existences DROP COLUMN IF EXISTS service_tax_amount;

-- Update the pricing business logic comment
COMMENT ON COLUMN existences.minimum_price IS 'Minimum acceptable price for income (cost + margin only, no taxes)';
COMMENT ON COLUMN existences.maximum_price IS 'Maximum price ceiling (cost + margin only, no taxes)';
COMMENT ON COLUMN existences.final_price IS 'User-editable final price before taxes (must be between minimum_price and maximum_price)';

-- Add comment to table explaining the architectural change
COMMENT ON TABLE existences IS 'Track ingredient purchases with base pricing strategy (cost + margin only). Tax calculations handled in orders system.';
