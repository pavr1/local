package sql

import _ "embed"

// Ingredient SQL queries
//
//go:embed scripts/create_ingredient.sql
var CreateIngredientQuery string

//go:embed scripts/get_ingredient_by_id.sql
var GetIngredientByIDQuery string

//go:embed scripts/list_ingredients.sql
var ListIngredientsQuery string

//go:embed scripts/update_ingredient.sql
var UpdateIngredientQuery string

//go:embed scripts/delete_ingredient.sql
var DeleteIngredientQuery string
