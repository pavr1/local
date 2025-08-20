-- Migration: Remove unit_type column from recipe_ingredients table
-- Description: The unit_type field is redundant as it's already stored in the ingredients table
-- Date: 2024-12-19
-- Author: System

-- Remove the unit_type column from recipe_ingredients table
ALTER TABLE recipe_ingredients DROP COLUMN IF EXISTS unit_type;
