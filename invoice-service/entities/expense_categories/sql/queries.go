package sql

import _ "embed"

// ExpenseCategory SQL queries
//
//go:embed scripts/create_expense_category.sql
var CreateExpenseCategoryQuery string

//go:embed scripts/get_expense_category_by_id.sql
var GetExpenseCategoryByIDQuery string

//go:embed scripts/list_expense_categories.sql
var ListExpenseCategoriesQuery string

//go:embed scripts/update_expense_category.sql
var UpdateExpenseCategoryQuery string

//go:embed scripts/delete_expense_category.sql
var DeleteExpenseCategoryQuery string 