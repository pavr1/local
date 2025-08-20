-- Migration: Restore unit_type column to recipe_ingredients table (ROLLBACK)
-- Description: Restores the unit_type column that was removed in migration 005
-- Date: 2024-12-19
-- Author: System

-- Add the unit_type column back to recipe_ingredients table
ALTER TABLE recipe_ingredients ADD COLUMN unit_type VARCHAR(50);
