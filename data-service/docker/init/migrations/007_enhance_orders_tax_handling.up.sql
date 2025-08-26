-- Migration: Enhance orders table tax handling
-- Description: Add comments and ensure proper tax handling structure in orders table
-- Date: 2025-08-26

-- Add comments to orders table columns to clarify tax handling
COMMENT ON COLUMN orders.subtotal_amount IS 'Order subtotal before taxes (sum of all recipe final prices from existences)';
COMMENT ON COLUMN orders.iva_amount IS 'IVA tax amount (13% of subtotal) - calculated and applied at order level';
COMMENT ON COLUMN orders.service_tax_amount IS 'Service tax amount (10% of subtotal) - calculated and applied at order level';
COMMENT ON COLUMN orders.total_amount IS 'Final total amount (subtotal + iva_amount + service_tax_amount - discount_amount)';

-- Add comment to orders table explaining the tax handling responsibility
COMMENT ON TABLE orders IS 'Track customer transactions with complete tax calculations and compliance. Handles all IVA and Service Tax calculations.';

-- Ensure proper constraints for tax calculations
ALTER TABLE orders ADD CONSTRAINT check_iva_amount_positive CHECK (iva_amount >= 0);
ALTER TABLE orders ADD CONSTRAINT check_service_tax_amount_positive CHECK (service_tax_amount >= 0);
ALTER TABLE orders ADD CONSTRAINT check_total_amount_positive CHECK (total_amount >= 0);

-- Add constraint to ensure total_amount calculation is correct
ALTER TABLE orders ADD CONSTRAINT check_total_amount_calculation 
    CHECK (total_amount = subtotal_amount + iva_amount + service_tax_amount - discount_amount);
