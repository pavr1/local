-- Migration: Remove supplier_id from ingredients table
-- Version: 003
-- Description: Remove supplier_id column and foreign key constraint from ingredients table
--              Suppliers will be handled at the invoice level instead

-- Step 1: Drop the foreign key constraint (if it exists)
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'ingredients' 
        AND constraint_name = 'ingredients_supplier_id_fkey'
    ) THEN
        ALTER TABLE ingredients DROP CONSTRAINT ingredients_supplier_id_fkey;
    END IF;
END $$;

-- Step 2: Drop the supplier_id column (if it exists)
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'ingredients' 
        AND column_name = 'supplier_id'
    ) THEN
        ALTER TABLE ingredients DROP COLUMN supplier_id;
    END IF;
END $$;

-- Step 3: Update comments for clarity
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'ingredients' 
        AND column_name = 'ingredient_category_id'
    ) THEN
        COMMENT ON COLUMN ingredients.ingredient_category_id IS 'Category classification for the ingredient (suppliers handled at invoice level)';
    END IF;
END $$;

