SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, 
       i.name as ingredient_name, COALESCE(e.final_price, 0) as final_price,
       ri.created_at, ri.updated_at 
FROM recipe_ingredients ri
LEFT JOIN ingredients i ON ri.ingredient_id = i.id
LEFT JOIN LATERAL (
    SELECT final_price, created_at
    FROM existences 
    WHERE ingredient_id = ri.ingredient_id 
    ORDER BY created_at DESC 
    LIMIT 1
) e ON true
WHERE ($1::uuid IS NULL OR ri.recipe_id = $1)
  AND ($2::uuid IS NULL OR ri.ingredient_id = $2)
ORDER BY ri.recipe_id ASC, i.name ASC
LIMIT $3 OFFSET $4; 