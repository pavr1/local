-- Create an ordered recipe item
INSERT INTO ordered_receipes (
    id, order_id, recipe_id, product_name, quantity, receipe_price, subtotal
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
); 