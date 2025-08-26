-- Migration: Add status field to recipes table and update total_recipe_cost calculation
-- Description: Add status field for recipe availability and update total_recipe_cost to be calculated from ingredient final prices
-- Date: 2025-08-26

-- Add status column to recipes table
ALTER TABLE recipes ADD COLUMN status VARCHAR(20) DEFAULT 'pending';

-- Update comments to explain the fields
COMMENT ON COLUMN recipes.status IS 'Recipe status: pending (no valid ingredient costs), active (all ingredients have valid costs)';
COMMENT ON COLUMN recipes.total_recipe_cost IS 'Current recipe cost calculated from ingredient final prices (sum of all recipe ingredients final prices)';

-- Update table comment
COMMENT ON TABLE recipes IS 'Track ice cream recipes with status management and cost calculation. Recipes are pending until all ingredients have valid costs.';
