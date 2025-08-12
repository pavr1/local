-- Migration: Add items_per_unit column to invoice_details table
-- UP Migration: Adds the items_per_unit column to existing invoice_details tables

-- Add the new column with a default value for existing records
ALTER TABLE invoice_details 
ADD COLUMN items_per_unit INTEGER NOT NULL DEFAULT 1 CHECK (items_per_unit > 0);

-- Add comment to explain the column purpose
COMMENT ON COLUMN invoice_details.items_per_unit IS 'Number of individual items contained in one unit (e.g., 31 ice cream balls per gallon, 12 units per bag)';

-- Create index for the new column
CREATE INDEX idx_invoice_details_items_per_unit ON invoice_details(items_per_unit); 