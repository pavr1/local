SELECT id, name, description, ingredient_category_id, supplier_id, created_at, updated_at
FROM ingredients
WHERE id = $1; 