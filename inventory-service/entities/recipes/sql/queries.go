package sql

import _ "embed"

// Recipe SQL queries
//
//go:embed scripts/create_recipe.sql
var CreateRecipeQuery string

//go:embed scripts/get_recipe_by_id.sql
var GetRecipeByIDQuery string

//go:embed scripts/list_recipes.sql
var ListRecipesQuery string

//go:embed scripts/update_recipe.sql
var UpdateRecipeQuery string

//go:embed scripts/delete_recipe.sql
var DeleteRecipeQuery string

//go:embed scripts/update_recipe_price_and_status.sql
var UpdateRecipePriceAndStatusQuery string

//go:embed scripts/get_recipes_by_ingredient.sql
var GetRecipesByIngredientQuery string
