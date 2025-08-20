-- Revert expense_category_id to NOT NULL
ALTER TABLE invoice ALTER COLUMN expense_category_id SET NOT NULL;
