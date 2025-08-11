SELECT 
    i.id, 
    i.name, 
    i.description, 
    i.ingredient_category_id, 
    i.supplier_id, 
    i.created_at, 
    i.updated_at,
    ic.name as category_name,
    s.supplier_name as supplier_name
FROM ingredients i
LEFT JOIN ingredient_categories ic ON i.ingredient_category_id = ic.id
LEFT JOIN suppliers s ON i.supplier_id = s.id
ORDER BY i.name ASC; 