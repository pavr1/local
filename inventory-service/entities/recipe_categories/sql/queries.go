package sql

import _ "embed"

// Recipe Category SQL queries
//
//go:embed scripts/create_recipe_category.sql
var CreateRecipeCategoryQuery string

//go:embed scripts/get_recipe_category_by_id.sql
var GetRecipeCategoryByIDQuery string

//go:embed scripts/list_recipe_categories.sql
var ListRecipeCategoriesQuery string

//go:embed scripts/update_recipe_category.sql
var UpdateRecipeCategoryQuery string

//go:embed scripts/delete_recipe_category.sql
var DeleteRecipeCategoryQuery string
