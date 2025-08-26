-- Migration: Restore tax fields to existences table (ROLLBACK)
-- Description: Add back IVA and Service Tax fields to existences table
-- Date: 2025-08-26

-- Add back tax-related columns to existences table
ALTER TABLE existences ADD COLUMN iva_percentage DECIMAL(5,2) DEFAULT 13.00;
ALTER TABLE existences ADD COLUMN iva_amount DECIMAL(10,2) DEFAULT 0.00;
ALTER TABLE existences ADD COLUMN service_tax_percentage DECIMAL(5,2) DEFAULT 10.00;
ALTER TABLE existences ADD COLUMN service_tax_amount DECIMAL(10,2) DEFAULT 0.00;

-- Restore original pricing business logic comments
COMMENT ON COLUMN existences.minimum_price IS 'Minimum acceptable price for income (includes cost + margins + taxes)';
COMMENT ON COLUMN existences.maximum_price IS 'Maximum price ceiling (includes cost + margins + taxes)';
COMMENT ON COLUMN existences.final_price IS 'User-editable final price (must be between minimum_price and maximum_price)';

-- Restore original table comment
COMMENT ON TABLE existences IS 'Track ingredient purchases with invoice traceability and pricing including taxes.';
