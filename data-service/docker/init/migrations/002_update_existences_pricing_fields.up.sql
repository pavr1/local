-- Migration 002: Update existences table pricing fields
-- Description: Rename calculatedPrice to minimumPrice, finalPrice to maximumPrice, and add new finalPrice field
-- Date: 2025-08-16

-- Start transaction
BEGIN;

-- Step 1: Add new columns
ALTER TABLE existences 
ADD COLUMN minimum_price DECIMAL(10,2) DEFAULT 0.00,
ADD COLUMN maximum_price DECIMAL(10,2),
ADD COLUMN final_price DECIMAL(10,2);

-- Step 2: Migrate existing data
-- Move calculated_price to minimum_price
UPDATE existences 
SET minimum_price = calculated_price 
WHERE calculated_price IS NOT NULL;

-- Move final_price to maximum_price
UPDATE existences 
SET maximum_price = final_price 
WHERE final_price IS NOT NULL;

-- Set new final_price to maximum_price (default behavior)
UPDATE existences 
SET final_price = maximum_price 
WHERE maximum_price IS NOT NULL;

-- Step 3: Add constraints for the new final_price field
-- Ensure final_price is between minimum_price and maximum_price
ALTER TABLE existences 
ADD CONSTRAINT check_final_price_range 
CHECK (final_price >= minimum_price AND final_price <= maximum_price);

-- Step 4: Drop old columns
ALTER TABLE existences 
DROP COLUMN calculated_price,
DROP COLUMN final_price;

-- Step 5: Update comments for clarity
COMMENT ON COLUMN existences.minimum_price IS 'Minimum price (previously calculated_price) - represents the minimum acceptable price for income';
COMMENT ON COLUMN existences.maximum_price IS 'Maximum price (previously final_price) - represents the maximum price ceiling';
COMMENT ON COLUMN existences.final_price IS 'User-editable final price - must be between minimum_price and maximum_price for income';

-- Commit transaction
COMMIT;
