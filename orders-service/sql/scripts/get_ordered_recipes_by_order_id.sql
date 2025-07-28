-- Get ordered recipes by order ID
SELECT id, order_id, recipe_id, quantity, unit_price, total_price,
       special_instructions, created_at
FROM ordered_receipes 
WHERE order_id = $1
ORDER BY created_at; 