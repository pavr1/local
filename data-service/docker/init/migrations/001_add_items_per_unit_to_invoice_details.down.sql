-- Migration: Add items_per_unit column to invoice_details table
-- DOWN Migration: Removes the items_per_unit column from invoice_details table

-- Drop the index first
DROP INDEX IF EXISTS idx_invoice_details_items_per_unit;

-- Remove the column
ALTER TABLE invoice_details DROP COLUMN IF EXISTS items_per_unit; 