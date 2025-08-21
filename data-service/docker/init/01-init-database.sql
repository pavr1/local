-- Ice Cream Store Database Schema
-- Database: icecream_store
-- Version: 1.0

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create sequences
CREATE SEQUENCE IF NOT EXISTS existence_reference_seq START 1;
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;
CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START 1;

-- =============================================================================
-- INVENTORY MANAGEMENT ENTITIES
-- =============================================================================

-- Suppliers Table
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_name VARCHAR(255) NOT NULL UNIQUE,
    contact_number VARCHAR(20),
    email VARCHAR(255),
    address TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ingredient Categories Table
CREATE TABLE ingredient_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ingredients Table
CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    ingredient_category_id UUID REFERENCES ingredient_categories(id) ON DELETE SET NULL,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recipe Categories Table
CREATE TABLE recipe_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recipes Table
CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_name VARCHAR(255) NOT NULL UNIQUE,
    recipe_description TEXT,
    picture_url VARCHAR(500),
    recipe_category_id UUID REFERENCES recipe_categories(id) ON DELETE SET NULL,
    total_recipe_cost DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recipe Ingredients Table
CREATE TABLE recipe_ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity > 0),
    unit_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(recipe_id, ingredient_id)
);

-- =============================================================================
-- INVOICES MANAGEMENT ENTITIES
-- =============================================================================

-- Expense Categories Table
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Expenses Table
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE CASCADE,
    description TEXT,
    amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
    expense_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Invoice Table (modernized expense tracking)
CREATE TABLE invoice (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    transaction_date DATE NOT NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('income', 'outcome')),
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    total_amount DECIMAL(12,2),
    image_url VARCHAR(500) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Invoice Details Table (line items for invoices)
CREATE TABLE invoice_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    ingredient_id UUID REFERENCES ingredients(id) ON DELETE SET NULL,
    detail TEXT NOT NULL,
    count DECIMAL(10,2) NOT NULL CHECK (count > 0),
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('Liters', 'Gallons', 'Units', 'Bag')),
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    items_per_unit INTEGER NOT NULL DEFAULT 1 CHECK (items_per_unit > 0),
    total DECIMAL(12,2) GENERATED ALWAYS AS (count * price) STORED,
    expiration_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Existences Table
CREATE TABLE existences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_reference_code INTEGER UNIQUE NOT NULL DEFAULT nextval('existence_reference_seq'),
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    invoice_detail_id UUID NOT NULL, -- TODO: Add REFERENCES invoice_details(id) ON DELETE CASCADE after table order is fixed
    --units
    units_purchased DECIMAL(10,2) NOT NULL, -- get this from invoice detail
    units_available DECIMAL(10,2) NOT NULL, -- same as unit purchased, update when running out
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('Liters', 'Gallons', 'Units', 'Bag')), -- get from invoice detail
    --items
    items_per_unit INTEGER NOT NULL, --ie. Galon has 31 ice-cream balls
    cost_per_item DECIMAL(10,2) GENERATED ALWAYS AS (cost_per_unit / items_per_unit) STORED,
    cost_per_unit DECIMAL(10,2) NOT NULL, -- get from invoice detail
    --costs
    total_purchase_cost DECIMAL(12,2) GENERATED ALWAYS AS (units_purchased * cost_per_unit) STORED,
    remaining_value DECIMAL(12,2) GENERATED ALWAYS AS (units_available * cost_per_unit) STORED,
    --expiry
    expiration_date DATE, -- get from invoice detail
    --incomes & taxes
    income_margin_percentage DECIMAL(5,2) DEFAULT 30.00, -- grabbed from config
    income_margin_amount DECIMAL(10,2) DEFAULT 0.00, -- Will be calculated by application
    iva_percentage DECIMAL(5,2) DEFAULT 13.00, -- grabbed from config
    iva_amount DECIMAL(10,2) DEFAULT 0.00, -- Will be calculated by application
    service_tax_percentage DECIMAL(5,2) DEFAULT 10.00,
    service_tax_amount DECIMAL(10,2) DEFAULT 0.00, -- Will be calculated by application
    minimum_price DECIMAL(10,2) DEFAULT 0.00, -- Minimum acceptable price for income (previously calculated_price)
    maximum_price DECIMAL(10,2), -- Maximum price ceiling (previously final_price)
    final_price DECIMAL(10,2), -- User-editable final price (must be between minimum_price and maximum_price)
    --dates
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    --constraints
    CONSTRAINT check_final_price_range CHECK (final_price >= minimum_price AND final_price <= maximum_price)
);

-- =============================================================================
-- CUSTOMER MANAGEMENT ENTITIES
-- =============================================================================

-- Customers Table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- AUTHENTICATION & AUTHORIZATION ENTITIES (Moved before orders for dependency)
-- =============================================================================

-- Roles Table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions Table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions Table (for database-backed session management)
-- Minimal schema: only session_id and token
-- All session metadata (expiration, user data, etc.) is stored in the JWT token itself
CREATE TABLE sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    token TEXT NOT NULL UNIQUE
);

-- =============================================================================
-- INCOME MANAGEMENT (ORDERS) ENTITIES
-- =============================================================================

-- Orders Table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    sales_representative_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled', 'sinpe_pending')) DEFAULT 'pending',
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'card', 'sinpe')),
    transaction_reference VARCHAR(100),
    sinpe_screenshot_url VARCHAR(500),
    subtotal_amount DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) DEFAULT 0.00,
    iva_amount DECIMAL(12,2) NOT NULL,
    service_tax_amount DECIMAL(12,2) NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    invoice_number VARCHAR(50) UNIQUE,
    invoice_url VARCHAR(500),
    transaction_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ordered Receipes Table
CREATE TABLE ordered_receipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    receipe_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(12,2) GENERATED ALWAYS AS (quantity * receipe_price) STORED
);

-- Indexes for orders table
CREATE INDEX idx_orders_number ON orders(order_number);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_method ON orders(payment_method);
CREATE INDEX idx_orders_sales_rep ON orders(sales_representative_id);
CREATE INDEX idx_orders_transaction_timestamp ON orders(transaction_timestamp);
CREATE INDEX idx_orders_invoice_number ON orders(invoice_number);

-- Indexes for ordered_receipes table
CREATE INDEX idx_ordered_receipes_order ON ordered_receipes(order_id);
CREATE INDEX idx_ordered_receipes_recipe ON ordered_receipes(recipe_id);
CREATE INDEX idx_ordered_receipes_product_name ON ordered_receipes(product_name);

-- =============================================================================
-- PROMOTIONS & LOYALTY SYSTEM ENTITIES
-- =============================================================================

-- Promotions Table
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    promotion_type VARCHAR(50) NOT NULL CHECK (promotion_type IN ('percentage', 'fixed_amount', 'points_reward')),
    value DECIMAL(10,2) NOT NULL CHECK (value >= 0),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    minimum_purchase_amount DECIMAL(10,2) CHECK (minimum_purchase_amount >= 0),
    points_expiration_duration VARCHAR(20), -- e.g., '1d', '3w', '7m', '2y'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customer Points Table
CREATE TABLE customer_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    points_earned INTEGER NOT NULL CHECK (points_earned >= 0),
    earned_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    promotion_id UUID REFERENCES promotions(id) ON DELETE SET NULL
);

-- =============================================================================
-- EQUIPMENT MANAGEMENT ENTITIES
-- =============================================================================

-- Mechanics Table
CREATE TABLE mechanics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Equipment Table
CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    purchase_date DATE,
    mechanic_id UUID REFERENCES mechanics(id) ON DELETE SET NULL,
    maintenance_schedule_days INTEGER DEFAULT 30,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- ADMINISTRATION PANEL ENTITIES
-- =============================================================================

-- System Configuration Table
CREATE TABLE system_configuration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- AUTHENTICATION & AUTHORIZATION ENTITIES (Already moved above)
-- =============================================================================

-- =============================================================================
-- USER-DEPENDENT ENTITIES (Moved after users table)
-- =============================================================================

-- Runout Ingredient Report Table
CREATE TABLE runout_ingredient_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity >= 0),
    unit_type VARCHAR(50) NOT NULL,
    report_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Waste Loss Table
CREATE TABLE waste_loss (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    items_wasted DECIMAL(10,2) NOT NULL CHECK (items_wasted > 0), -- amount of items in a unit wasted
    reason VARCHAR(255) NOT NULL,
    financial_loss DECIMAL(10,2) NOT NULL, -- Calculated by application: items_wasted * existence.price_per_unit
    waste_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Salary Table
CREATE TABLE user_salary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expense_id UUID REFERENCES expenses(id) ON DELETE SET NULL,
    salary DECIMAL(10,2) NOT NULL CHECK (salary >= 0),
    additional_expenses DECIMAL(10,2) DEFAULT 0 CHECK (additional_expenses >= 0),
    total DECIMAL(10,2) GENERATED ALWAYS AS (salary + additional_expenses) STORED,
    payment_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit Logs Table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id UUID,
    old_values JSONB,
    new_values JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address INET
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Inventory indexes
CREATE INDEX idx_ingredient_categories_name ON ingredient_categories(name);
CREATE INDEX idx_ingredient_categories_active ON ingredient_categories(is_active);
CREATE INDEX idx_ingredients_name ON ingredients(name);
CREATE INDEX idx_ingredients_category ON ingredients(ingredient_category_id);
CREATE INDEX idx_ingredients_supplier ON ingredients(supplier_id);
CREATE INDEX idx_existences_ingredient ON existences(ingredient_id);
CREATE INDEX idx_existences_reference_code ON existences(existence_reference_code);
CREATE INDEX idx_existences_invoice_detail ON existences(invoice_detail_id);
CREATE INDEX idx_existences_available ON existences(units_available);
CREATE INDEX idx_existences_cost_per_item ON existences(cost_per_item);
CREATE INDEX idx_existences_minimum_price ON existences(minimum_price);
CREATE INDEX idx_existences_maximum_price ON existences(maximum_price);
CREATE INDEX idx_existences_final_price ON existences(final_price);
CREATE INDEX idx_existences_expiration_date ON existences(expiration_date);
CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_ingredient_id ON recipe_ingredients(ingredient_id);

-- Orders indexes
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_ordered_receipes_order_id ON ordered_receipes(order_id);
CREATE INDEX idx_ordered_receipes_recipe_id ON ordered_receipes(recipe_id);

-- Expenses indexes
CREATE INDEX idx_expenses_category_id ON expenses(expense_category_id);
CREATE INDEX idx_expenses_expense_date ON expenses(expense_date);

-- Promotions indexes
CREATE INDEX idx_promotions_recipe_id ON promotions(recipe_id);
CREATE INDEX idx_promotions_dates ON promotions(start_date, end_date);
CREATE INDEX idx_customer_points_customer_id ON customer_points(customer_id);

-- Audit indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_table_name ON audit_logs(table_name);

-- Session indexes for performance
CREATE INDEX idx_sessions_token ON sessions(token);

-- Indexes for Invoice Tables
CREATE INDEX idx_invoice_number ON invoice(invoice_number);
CREATE INDEX idx_invoice_supplier ON invoice(supplier_id);
CREATE INDEX idx_invoice_category ON invoice(expense_category_id);
CREATE INDEX idx_invoice_transaction_date ON invoice(transaction_date);
CREATE INDEX idx_invoice_transaction_type ON invoice(transaction_type);
CREATE INDEX idx_invoice_details_invoice ON invoice_details(invoice_id);
CREATE INDEX idx_invoice_details_ingredient ON invoice_details(ingredient_id);
CREATE INDEX idx_invoice_details_total ON invoice_details(total);
CREATE INDEX idx_invoice_details_unit_type ON invoice_details(unit_type);
CREATE INDEX idx_invoice_details_items_per_unit ON invoice_details(items_per_unit);
CREATE INDEX idx_invoice_details_expiration ON invoice_details(expiration_date);

-- =============================================================================
-- DEFAULT DATA
-- =============================================================================

-- Insert default system configuration
INSERT INTO system_configuration (config_key, config_value, description) VALUES
('store_name', 'Ice Cream Paradise', 'Name of the ice cream store'),
('store_address', '123 Sweet Street, Flavor City', 'Physical address of the store'),
('store_phone', '+1-555-ICE-CREAM', 'Contact phone number'),
('currency', 'USD', 'Default currency for transactions'),
('tax_rate', '0.08', 'Default tax rate as decimal (8%)'),
('loyalty_points_rate', '0.01', 'Points earned per dollar spent'),
('max_order_items', '50', 'Maximum items allowed per order');

-- Insert default roles
INSERT INTO roles (role_name, description) VALUES
('super_admin', 'Full system access and control'),
('admin', 'Administrative access to most features'),
('manager', 'Store management and operational oversight'),
('employee', 'Basic operational access'),
('cashier', 'Point of sale and order management only');

-- Insert default permissions for super_admin
INSERT INTO permissions (permission_name, description, role_id) VALUES
('inventory-read', 'View inventory data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('inventory-write', 'Modify inventory data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('inventory-delete', 'Delete inventory data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('expenses-read', 'View expense data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('expenses-write', 'Modify expense data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('expenses-delete', 'Delete expense data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('orders-read', 'View order data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('orders-write', 'Modify order data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('orders-delete', 'Delete order data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('customers-read', 'View customer data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('customers-write', 'Modify customer data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('customers-delete', 'Delete customer data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('promotions-read', 'View promotion data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('promotions-write', 'Modify promotion data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('promotions-delete', 'Delete promotion data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('equipment-read', 'View equipment data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('equipment-write', 'Modify equipment data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('equipment-delete', 'Delete equipment data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('waste-read', 'View waste data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('waste-write', 'Modify waste data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('waste-delete', 'Delete waste data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('admin-read', 'View admin data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('admin-write', 'Modify admin data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('admin-delete', 'Delete admin data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('auth-read', 'View auth data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('auth-write', 'Modify auth data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('auth-delete', 'Delete auth data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('audit-read', 'View audit data', (SELECT id FROM roles WHERE role_name = 'super_admin')),
('system-config', 'Manage system configuration', (SELECT id FROM roles WHERE role_name = 'super_admin'));

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password_hash, full_name, role_id) VALUES
('admin', '$2a$12$04xNgahgyY9qDgv7goYUVenjgTHF7.ei9GVkp.uYixLs.ebrJxw6u', 'System Administrator', 
 (SELECT id FROM roles WHERE role_name = 'super_admin'));

-- Insert default expense categories
INSERT INTO expense_categories (category_name, description) VALUES
('Salary payments', 'Employee salaries and wages'),
('Service payments', 'Utility services, maintenance, subscriptions'),
('Rent payments', 'Property rent and lease payments'),
('Ingredients', 'Ingredient and supply purchases'),
('Other operational expenses', 'Miscellaneous business expenses');

-- Insert default ingredient categories
INSERT INTO ingredient_categories (name, description) VALUES
('Productos Lácteos', 'Leche, crema, mantequilla, huevos, queso, yogur'),
('Endulzantes', 'Azúcar, miel, endulzantes artificiales, jarabes, agave'),
('Sabores y Extractos', 'Extracto de vainilla, extracto de almendra, colorantes alimentarios, sabores artificiales'),
('Frutas Frescas', 'Fresas frescas, plátanos, bayas, frutas de temporada'),
('Frutas Conservadas', 'Frutas secas, purés de frutas, mermeladas, frutas congeladas'),
('Semillas', 'Almendras, nueces, pistachos, coco, semillas'),
('Chocolate y Cacao', 'Polvo de cacao, chips de chocolate, barras de chocolate, chocolate blanco'),
('Gelatinas', 'Gelatina, agar, goma xantana, lecitina'),
('Dulces', 'Ositos de goma, chips de chocolate, piezas de caramelo, malvaviscos'),
('Galletas', 'Migas de galletas, piezas de brownie, barquillos'),
('Cereales y Granos', 'Granola, piezas de cereal, avena, crisps de arroz'),
('Toppings', 'Salsas de chocolate, caramelo, jarabes de frutas, fudge, sprinkles, crema batida, cerezas, nueces'),
('Contenedores Desechables', 'Tazas de helado, contenedores para llevar, tarrinas, tazones desechables'),
('Contenedores Comestibles', 'Conos de helado, conos waffle, crepes, galletas oblea comestibles'),
('Helados', 'Helados de paleta, helados de cono, helados de taza, helados de vaso'),
('Utensilios de Servicio', 'Cucharas, servilletas, pajillas, agitadores'),
('Empaques', 'Bolsas, tapas, etiquetas, cajas, envolturas'),
('Suministros de Limpieza', 'Sanitizantes, detergentes, paños de limpieza, cepillos'),
('Bebidas', 'Café, té, refrescos, agua'),
('Pasteleria', 'Pasteles, muffins, cupcakes, productos horneados');

-- Insert default recipe categories
INSERT INTO recipe_categories (name, description) VALUES
('Postres', 'Dessert ice creams and frozen treats'),
('Helados', 'Traditional ice cream flavors'),
('Batidos', 'Milkshakes and blended drinks'),
('Gelato', 'Italian-style gelato flavors'),
('Artesanales', 'Artisanal and specialty flavors');

-- =============================================================================
-- TRIGGERS FOR AUTOMATIC UPDATES
-- =============================================================================

-- Update timestamps trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables with updated_at
CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ingredient_categories_updated_at BEFORE UPDATE ON ingredient_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ingredients_updated_at BEFORE UPDATE ON ingredients 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_existences_updated_at BEFORE UPDATE ON existences 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recipe_categories_updated_at BEFORE UPDATE ON recipe_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recipes_updated_at BEFORE UPDATE ON recipes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recipe_ingredients_updated_at BEFORE UPDATE ON recipe_ingredients 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_runout_ingredient_report_updated_at BEFORE UPDATE ON runout_ingredient_report 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_expense_categories_updated_at BEFORE UPDATE ON expense_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invoice_updated_at BEFORE UPDATE ON invoice 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invoice_details_updated_at BEFORE UPDATE ON invoice_details 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_expenses_updated_at BEFORE UPDATE ON expenses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_promotions_updated_at BEFORE UPDATE ON promotions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mechanics_updated_at BEFORE UPDATE ON mechanics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_equipment_updated_at BEFORE UPDATE ON equipment 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_configuration_updated_at BEFORE UPDATE ON system_configuration 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_salary_updated_at BEFORE UPDATE ON user_salary 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- SETTINGS TABLE FOR CENTRALIZED CONFIGURATION
-- =============================================================================

-- Create settings table
CREATE TABLE settings (
    setting_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service, key)
);

-- Insert default settings for all services
INSERT INTO settings (service, key, value, description) VALUES
-- General Database Settings
('General', 'DB_HOST', 'icecream_postgres', 'Database host address'),
('General', 'DB_PORT', '5432', 'Database port number'),
('General', 'DB_USER', 'postgres', 'Database username'),
('General', 'DB_PASSWORD', 'postgres123', 'Database password'),
('General', 'DB_NAME', 'icecream_store', 'Database name'),
('General', 'DB_SSL_MODE', 'disable', 'Database SSL mode'),

-- General Logging Settings
('General', 'LOG_LEVEL', 'info', 'Logging level (debug, info, warn, error)'),

-- JWT Settings
('General', 'JWT_SECRET', 'icecream-super-secret-jwt-key-change-in-production-2024', 'JWT signing secret'),
('General', 'JWT_EXPIRATION_TIME', '30m', 'JWT token expiration time'),

-- Business Settings
('General', 'DEFAULT_TAX_RATE', '13.0', 'Default tax rate percentage (Costa Rica IVA)'),
('General', 'DEFAULT_SERVICE_RATE', '10.0', 'Default service charge rate percentage'),
('General', 'ORDER_TIMEOUT', '30', 'Order timeout in minutes'),

-- Gateway Service Settings
('Gateway', 'GATEWAY_PORT', '8082', 'Port for the gateway service'),
('Gateway', 'SESSION_SERVICE_URL', 'http://icecream_session:8081', 'URL for session service'),
('Gateway', 'ORDERS_SERVICE_URL', 'http://icecream_orders:8083', 'URL for orders service'),
('Gateway', 'INVENTORY_SERVICE_URL', 'http://icecream_inventory:8084', 'URL for inventory service'),
('Gateway', 'INVOICE_SERVICE_URL', 'http://icecream_invoice:8085', 'URL for invoice service'),

-- Session Service Settings
('Session', 'SESSION_SERVER_PORT', '8081', 'Port for the session service'),

-- Orders Service Settings
('Orders', 'SERVER_PORT', '8083', 'Port for the orders service'),

-- Inventory Service Settings
('Inventory', 'INVENTORY_SERVER_PORT', '8084', 'Port for the inventory service'),

-- Invoice Service Settings
('Invoice', 'INVOICE_SERVER_PORT', '8085', 'Port for the invoice service'),
('Invoice', 'INVENTORY_SERVICE_URL', 'http://icecream_inventory:8084', 'URL for inventory service'),

-- Data Service Settings
('Data', 'DATA_SERVER_PORT', '8086', 'Port for the data service'),
('Data', 'DB_MAX_OPEN_CONNS', '25', 'Maximum number of open connections'),
('Data', 'DB_MAX_IDLE_CONNS', '5', 'Maximum number of idle connections'),
('Data', 'DB_CONN_MAX_LIFETIME', '5m', 'Maximum lifetime of connections'),
('Data', 'DB_CONN_MAX_IDLE_TIME', '5m', 'Maximum idle time of connections'),
('Data', 'DB_CONNECT_TIMEOUT', '10s', 'Database connection timeout'),
('Data', 'DB_QUERY_TIMEOUT', '30s', 'Database query timeout'),
('Data', 'DB_MAX_RETRIES', '3', 'Maximum number of retry attempts'),
('Data', 'DB_RETRY_INTERVAL', '1s', 'Interval between retry attempts'),

-- UI Service Settings
('UI', 'UI_PORT', '3000', 'Port for the UI service'),
('UI', 'GATEWAY_URL', 'http://icecream_gateway:8082', 'Gateway service URL');

-- Create indexes for settings table
CREATE INDEX idx_settings_service ON settings(service);
CREATE INDEX idx_settings_key ON settings(key);
CREATE INDEX idx_settings_service_key ON settings(service, key);

-- Create trigger for settings table
CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 