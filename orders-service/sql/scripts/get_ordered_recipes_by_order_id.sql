-- Get ordered recipes by order ID
SELECT id, order_id, recipe_id, product_name, quantity, receipe_price, subtotal
FROM ordered_receipes 
WHERE order_id = $1
ORDER BY id; 