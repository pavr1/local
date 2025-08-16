-- Migration 002: Update existences table pricing fields
-- Description: Rename calculatedPrice to minimumPrice, finalPrice to maximumPrice, and add new finalPrice field
-- Date: 2025-08-16

-- Start transaction
BEGIN;

-- Step 1: Add new columns (only if they don't exist)
DO $$ 
BEGIN
    -- Add minimum_price column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'existences' AND column_name = 'minimum_price') THEN
        ALTER TABLE existences ADD COLUMN minimum_price DECIMAL(10,2) DEFAULT 0.00;
    END IF;
    
    -- Add maximum_price column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'existences' AND column_name = 'maximum_price') THEN
        ALTER TABLE existences ADD COLUMN maximum_price DECIMAL(10,2);
    END IF;
    
    -- Add new final_price column if it doesn't exist (we'll rename the old one later)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'existences' AND column_name = 'final_price_new') THEN
        ALTER TABLE existences ADD COLUMN final_price_new DECIMAL(10,2);
    END IF;
END $$;

-- Step 2: Migrate existing data
-- Move calculated_price to minimum_price (if calculated_price exists)
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'calculated_price') THEN
        UPDATE existences 
        SET minimum_price = calculated_price 
        WHERE calculated_price IS NOT NULL;
    END IF;
END $$;

-- Move final_price to maximum_price (if final_price exists)
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'final_price') THEN
        UPDATE existences 
        SET maximum_price = final_price 
        WHERE final_price IS NOT NULL;
        
        -- Also copy to the new final_price column
        UPDATE existences 
        SET final_price_new = final_price 
        WHERE final_price IS NOT NULL;
    END IF;
END $$;

-- Set new final_price to maximum_price (default behavior) if not already set
UPDATE existences 
SET final_price_new = maximum_price 
WHERE maximum_price IS NOT NULL AND final_price_new IS NULL;

-- Step 3: Add constraints for the new final_price field
-- Ensure final_price_new is between minimum_price and maximum_price
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE table_name = 'existences' AND constraint_name = 'check_final_price_range') THEN
        ALTER TABLE existences 
        ADD CONSTRAINT check_final_price_range 
        CHECK (final_price_new >= minimum_price AND final_price_new <= maximum_price);
    END IF;
END $$;

-- Step 4: Drop old columns (only if they exist)
DO $$ 
BEGIN
    -- Drop calculated_price column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'calculated_price') THEN
        ALTER TABLE existences DROP COLUMN calculated_price;
    END IF;
    
    -- Drop old final_price column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'final_price') THEN
        ALTER TABLE existences DROP COLUMN final_price;
    END IF;
END $$;

-- Step 5: Rename final_price_new to final_price
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'final_price_new') THEN
        ALTER TABLE existences RENAME COLUMN final_price_new TO final_price;
    END IF;
END $$;

-- Step 6: Update comments for clarity
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'minimum_price') THEN
        COMMENT ON COLUMN existences.minimum_price IS 'Minimum price (previously calculated_price) - represents the minimum acceptable price for income';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'maximum_price') THEN
        COMMENT ON COLUMN existences.maximum_price IS 'Maximum price (previously final_price) - represents the maximum price ceiling';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'existences' AND column_name = 'final_price') THEN
        COMMENT ON COLUMN existences.final_price IS 'User-editable final price - must be between minimum_price and maximum_price for income';
    END IF;
END $$;

-- Commit transaction
COMMIT;
