-- Migration: Restore supplier_id to ingredients table (ROLLBACK)
-- Version: 003
-- Description: Add back supplier_id column and foreign key constraint to ingredients table

-- Step 1: Add the supplier_id column back
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'ingredients' 
        AND column_name = 'supplier_id'
    ) THEN
        ALTER TABLE ingredients ADD COLUMN supplier_id UUID;
    END IF;
END $$;

-- Step 2: Add the foreign key constraint back
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'ingredients' 
        AND constraint_name = 'ingredients_supplier_id_fkey'
    ) THEN
        ALTER TABLE ingredients 
        ADD CONSTRAINT ingredients_supplier_id_fkey 
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Step 3: Restore original comment
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'ingredients' 
        AND column_name = 'ingredient_category_id'
    ) THEN
        COMMENT ON COLUMN ingredients.ingredient_category_id IS 'Category classification for the ingredient';
    END IF;
END $$;

