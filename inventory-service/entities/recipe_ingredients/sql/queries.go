package sql

import _ "embed"

// Recipe Ingredient SQL queries
//
//go:embed scripts/create_recipe_ingredient.sql
var CreateRecipeIngredientQuery string

//go:embed scripts/get_recipe_ingredient_by_id.sql
var GetRecipeIngredientByIDQuery string

//go:embed scripts/list_recipe_ingredients.sql
var ListRecipeIngredientsQuery string

//go:embed scripts/update_recipe_ingredient.sql
var UpdateRecipeIngredientQuery string

//go:embed scripts/delete_recipe_ingredient.sql
var DeleteRecipeIngredientQuery string
