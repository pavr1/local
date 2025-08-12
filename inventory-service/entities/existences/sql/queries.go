package sql

import _ "embed"

// SQL queries for existences operations

//go:embed scripts/create_existence.sql
var CreateExistenceQuery string

//go:embed scripts/get_existence_by_id.sql
var GetExistenceByIDQuery string

//go:embed scripts/list_existences.sql
var ListExistencesQuery string

//go:embed scripts/update_existence.sql
var UpdateExistenceQuery string

//go:embed scripts/delete_existence.sql
var DeleteExistenceQuery string

//go:embed scripts/get_most_recent_existence_by_ingredient_and_unit_type.sql
var GetMostRecentExistenceByIngredientAndUnitTypeQuery string
