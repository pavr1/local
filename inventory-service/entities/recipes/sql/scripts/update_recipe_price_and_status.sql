-- Update recipe total_recipe_cost and status based on ingredient final prices
-- This query calculates the total recipe cost from ingredient final prices
-- and updates the status to 'active' if all ingredients have valid final prices

UPDATE recipes 
SET 
    total_recipe_cost = (
        SELECT COALESCE(SUM(ri.quantity * COALESCE(e.final_price, 0)), 0)
        FROM recipe_ingredients ri
        LEFT JOIN existences e ON ri.ingredient_id = e.ingredient_id
        WHERE ri.recipe_id = recipes.id
    ),
    status = CASE 
        WHEN EXISTS (
            SELECT 1 FROM recipe_ingredients ri
            LEFT JOIN existences e ON ri.ingredient_id = e.ingredient_id
            WHERE ri.recipe_id = recipes.id 
            AND (e.final_price IS NULL OR e.final_price = 0)
        ) THEN 'pending'
        ELSE 'active'
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
