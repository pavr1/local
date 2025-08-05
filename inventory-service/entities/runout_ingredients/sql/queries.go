package sql

import _ "embed"

// Runout Ingredient SQL queries
//
//go:embed scripts/create_runout_ingredient.sql
var CreateRunoutIngredientQuery string

//go:embed scripts/get_runout_ingredient_by_id.sql
var GetRunoutIngredientByIDQuery string

//go:embed scripts/list_runout_ingredients.sql
var ListRunoutIngredientsQuery string

//go:embed scripts/update_runout_ingredient.sql
var UpdateRunoutIngredientQuery string

//go:embed scripts/delete_runout_ingredient.sql
var DeleteRunoutIngredientQuery string
