package sql

import _ "embed"

// Ingredient Category SQL queries
//
//go:embed scripts/create_ingredient_category.sql
var CreateIngredientCategoryQuery string

//go:embed scripts/get_ingredient_category_by_id.sql
var GetIngredientCategoryByIDQuery string

//go:embed scripts/list_ingredient_categories.sql
var ListIngredientCategoriesQuery string

//go:embed scripts/update_ingredient_category.sql
var UpdateIngredientCategoryQuery string

//go:embed scripts/delete_ingredient_category.sql
var DeleteIngredientCategoryQuery string
