-- Migration 002 Down: Rollback existences table pricing fields
-- Description: Revert pricing field changes back to original structure
-- Date: 2025-08-16

-- Start transaction
BEGIN;

-- Step 1: Add back the original columns
ALTER TABLE existences 
ADD COLUMN calculated_price DECIMAL(10,2) DEFAULT 0.00,
ADD COLUMN final_price_old DECIMAL(10,2);

-- Step 2: Migrate data back
-- Move minimum_price back to calculated_price
UPDATE existences 
SET calculated_price = minimum_price 
WHERE minimum_price IS NOT NULL;

-- Move maximum_price back to final_price_old
UPDATE existences 
SET final_price_old = maximum_price 
WHERE maximum_price IS NOT NULL;

-- Step 3: Drop the constraint
ALTER TABLE existences 
DROP CONSTRAINT IF EXISTS check_final_price_range;

-- Step 4: Drop the new columns
ALTER TABLE existences 
DROP COLUMN minimum_price,
DROP COLUMN maximum_price,
DROP COLUMN final_price;

-- Step 5: Rename final_price_old back to final_price
ALTER TABLE existences 
RENAME COLUMN final_price_old TO final_price;

-- Step 6: Restore original comments
COMMENT ON COLUMN existences.calculated_price IS 'Will be calculated by application';
COMMENT ON COLUMN existences.final_price IS 'Final price for the existence';

-- Commit transaction
COMMIT;
