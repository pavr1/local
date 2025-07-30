package sql

import _ "embed"

// Suppliers queries
//
//go:embed scripts/suppliers/create_supplier.sql
var CreateSupplierQuery string

//go:embed scripts/suppliers/get_supplier_by_id.sql
var GetSupplierByIDQuery string

//go:embed scripts/suppliers/list_suppliers.sql
var ListSuppliersQuery string

//go:embed scripts/suppliers/update_supplier.sql
var UpdateSupplierQuery string

//go:embed scripts/suppliers/delete_supplier.sql
var DeleteSupplierQuery string

//go:embed scripts/suppliers/get_supplier_by_name.sql
var GetSupplierByNameQuery string

// Ingredients queries
//
//go:embed scripts/ingredients/create_ingredient.sql
var CreateIngredientQuery string

//go:embed scripts/ingredients/get_ingredient_by_id.sql
var GetIngredientByIDQuery string

//go:embed scripts/ingredients/list_ingredients.sql
var ListIngredientsQuery string

//go:embed scripts/ingredients/update_ingredient.sql
var UpdateIngredientQuery string

//go:embed scripts/ingredients/delete_ingredient.sql
var DeleteIngredientQuery string

//go:embed scripts/ingredients/get_ingredient_by_name.sql
var GetIngredientByNameQuery string

//go:embed scripts/ingredients/list_ingredients_by_supplier.sql
var ListIngredientsBySupplierQuery string

// Existences queries
//
//go:embed scripts/existences/create_existence.sql
var CreateExistenceQuery string

//go:embed scripts/existences/get_existence_by_id.sql
var GetExistenceByIDQuery string

//go:embed scripts/existences/list_existences.sql
var ListExistencesQuery string

//go:embed scripts/existences/update_existence.sql
var UpdateExistenceQuery string

//go:embed scripts/existences/delete_existence.sql
var DeleteExistenceQuery string

//go:embed scripts/existences/list_existences_by_ingredient.sql
var ListExistencesByIngredientQuery string

//go:embed scripts/existences/list_low_stock.sql
var ListLowStockQuery string

//go:embed scripts/existences/list_expiring_soon.sql
var ListExpiringSoonQuery string

//go:embed scripts/existences/update_units_available.sql
var UpdateUnitsAvailableQuery string

// Runout Reports queries
//
//go:embed scripts/runout_reports/create_runout_report.sql
var CreateRunoutReportQuery string

//go:embed scripts/runout_reports/get_runout_report_by_id.sql
var GetRunoutReportByIDQuery string

//go:embed scripts/runout_reports/list_runout_reports.sql
var ListRunoutReportsQuery string

//go:embed scripts/runout_reports/list_runout_reports_by_existence.sql
var ListRunoutReportsByExistenceQuery string

//go:embed scripts/runout_reports/list_runout_reports_by_employee.sql
var ListRunoutReportsByEmployeeQuery string

// Recipe Categories queries
//
//go:embed scripts/recipe_categories/create_recipe_category.sql
var CreateRecipeCategoryQuery string

//go:embed scripts/recipe_categories/get_recipe_category_by_id.sql
var GetRecipeCategoryByIDQuery string

//go:embed scripts/recipe_categories/list_recipe_categories.sql
var ListRecipeCategoriesQuery string

//go:embed scripts/recipe_categories/update_recipe_category.sql
var UpdateRecipeCategoryQuery string

//go:embed scripts/recipe_categories/delete_recipe_category.sql
var DeleteRecipeCategoryQuery string

//go:embed scripts/recipe_categories/get_recipe_category_by_name.sql
var GetRecipeCategoryByNameQuery string

// Recipes queries
//
//go:embed scripts/recipes/create_recipe.sql
var CreateRecipeQuery string

//go:embed scripts/recipes/get_recipe_by_id.sql
var GetRecipeByIDQuery string

//go:embed scripts/recipes/list_recipes.sql
var ListRecipesQuery string

//go:embed scripts/recipes/update_recipe.sql
var UpdateRecipeQuery string

//go:embed scripts/recipes/delete_recipe.sql
var DeleteRecipeQuery string

//go:embed scripts/recipes/get_recipe_by_name.sql
var GetRecipeByNameQuery string

//go:embed scripts/recipes/list_recipes_by_category.sql
var ListRecipesByCategoryQuery string

//go:embed scripts/recipes/update_recipe_cost.sql
var UpdateRecipeCostQuery string

// Recipe Ingredients queries
//
//go:embed scripts/recipe_ingredients/create_recipe_ingredient.sql
var CreateRecipeIngredientQuery string

//go:embed scripts/recipe_ingredients/get_recipe_ingredient_by_id.sql
var GetRecipeIngredientByIDQuery string

//go:embed scripts/recipe_ingredients/list_recipe_ingredients.sql
var ListRecipeIngredientsQuery string

//go:embed scripts/recipe_ingredients/update_recipe_ingredient.sql
var UpdateRecipeIngredientQuery string

//go:embed scripts/recipe_ingredients/delete_recipe_ingredient.sql
var DeleteRecipeIngredientQuery string

//go:embed scripts/recipe_ingredients/list_recipe_ingredients_by_recipe.sql
var ListRecipeIngredientsByRecipeQuery string

//go:embed scripts/recipe_ingredients/list_recipe_ingredients_by_ingredient.sql
var ListRecipeIngredientsByIngredientQuery string

//go:embed scripts/recipe_ingredients/delete_recipe_ingredients_by_recipe.sql
var DeleteRecipeIngredientsByRecipeQuery string

// Health check
//
//go:embed scripts/health_check.sql
var HealthCheckQuery string
