SELECT 
    i.id, 
    i.name, 
    i.description, 
    i.ingredient_category_id, 
    i.created_at, 
    i.updated_at,
    ic.name as category_name
FROM ingredients i
LEFT JOIN ingredient_categories ic ON i.ingredient_category_id = ic.id
ORDER BY i.name ASC; 