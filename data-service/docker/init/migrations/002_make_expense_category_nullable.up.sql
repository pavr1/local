-- Make expense_category_id nullable for income invoices
ALTER TABLE invoice ALTER COLUMN expense_category_id DROP NOT NULL;
