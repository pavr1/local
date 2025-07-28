-- Create an ordered recipe item
INSERT INTO ordered_receipes (
    id, order_id, recipe_id, quantity, unit_price, total_price,
    special_instructions, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
); 