-- Migration: Rollback orders table tax handling enhancements (ROLLBACK)
-- Description: Remove comments and constraints added for tax handling
-- Date: 2025-08-26

-- Remove constraints added for tax calculations
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_iva_amount_positive;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_service_tax_amount_positive;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_total_amount_positive;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_total_amount_calculation;

-- Remove comments from orders table columns
COMMENT ON COLUMN orders.subtotal_amount IS NULL;
COMMENT ON COLUMN orders.iva_amount IS NULL;
COMMENT ON COLUMN orders.service_tax_amount IS NULL;
COMMENT ON COLUMN orders.total_amount IS NULL;

-- Remove comment from orders table
COMMENT ON TABLE orders IS NULL;
