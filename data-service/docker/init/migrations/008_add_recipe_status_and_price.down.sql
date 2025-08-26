-- Migration: Remove status field from recipes table (ROLLBACK)
-- Description: Remove status field added to recipes table
-- Date: 2025-08-26

-- Remove status column from recipes table
ALTER TABLE recipes DROP COLUMN IF EXISTS status;

-- Restore original table comment
COMMENT ON TABLE recipes IS 'Track ice cream recipes with ingredients and instructions.';
