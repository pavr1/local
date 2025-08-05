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

-- Runout Ingredient Report Table
CREATE TABLE runout_ingredient_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL, -- References users table
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity >= 0),
    unit_type VARCHAR(50) NOT NULL,
    report_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    description TEXT,
    instructions TEXT,
    preparation_time INTEGER, -- in minutes
    serving_size INTEGER,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    category_id UUID REFERENCES recipe_categories(id) ON DELETE SET NULL,
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
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    transaction_date DATE NOT NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('income', 'outcome')),
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    total_amount DECIMAL(10,2),
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
    detail VARCHAR(255) NOT NULL,
    count DECIMAL(10,2) NOT NULL CHECK (count > 0),
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('Liters', 'Gallons', 'Units', 'Bag')),
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    total DECIMAL(10,2) GENERATED ALWAYS AS (count * price) STORED,
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
    income_margin_amount DECIMAL(10,2) GENERATED ALWAYS AS (total_purchase_cost * income_margin_percentage / 100) STORED,
    iva_percentage DECIMAL(5,2) DEFAULT 13.00, -- grabbed from config
    iva_amount DECIMAL(10,2) GENERATED ALWAYS AS ((total_purchase_cost + income_margin_amount) * iva_percentage / 100) STORED,
    service_tax_percentage DECIMAL(5,2) DEFAULT 10.00,
    service_tax_amount DECIMAL(10,2) GENERATED ALWAYS AS ((total_purchase_cost + income_margin_amount) * service_tax_percentage / 100) STORED,
    calculated_price DECIMAL(10,2) GENERATED ALWAYS AS (total_purchase_cost + income_margin_amount + iva_amount + service_tax_amount) STORED, -- round to top next 100
    final_price DECIMAL(10,2),
    --dates
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
-- INCOME MANAGEMENT (ORDERS) ENTITIES
-- =============================================================================

-- Orders Table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number INTEGER DEFAULT nextval('order_number_seq'),
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    discount_amount DECIMAL(10,2) DEFAULT 0 CHECK (discount_amount >= 0),
    final_amount DECIMAL(10,2) GENERATED ALWAYS AS (total_amount - discount_amount) STORED,
    order_status VARCHAR(50) DEFAULT 'pending' CHECK (order_status IN ('pending', 'confirmed', 'completed', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_number)
);

-- Ordered Receipes Table
CREATE TABLE ordered_receipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    receipe_price DECIMAL(10,2) NOT NULL CHECK (receipe_price >= 0)
);

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
-- WASTE & LOSS TRACKING ENTITIES
-- =============================================================================

-- Waste Loss Table
CREATE TABLE waste_loss (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL, -- References users table
    items_wasted DECIMAL(10,2) NOT NULL CHECK (items_wasted > 0), -- amount of items in a unit wasted
    reason VARCHAR(255) NOT NULL,
    financial_loss DECIMAL(10,2) NOT NULL, -- Calculated by application: items_wasted * existence.price_per_unit
    waste_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

-- User Salary Table
CREATE TABLE user_salary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, -- References users table
    expense_id UUID REFERENCES expenses(id) ON DELETE SET NULL,
    salary DECIMAL(10,2) NOT NULL CHECK (salary >= 0),
    additional_expenses DECIMAL(10,2) DEFAULT 0 CHECK (additional_expenses >= 0),
    total DECIMAL(10,2) GENERATED ALWAYS AS (salary + additional_expenses) STORED,
    payment_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- AUTHENTICATION & AUTHORIZATION ENTITIES
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
CREATE TABLE sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    role_name VARCHAR(255) NOT NULL,
    permissions TEXT[], -- Array of permission strings
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- =============================================================================
-- AUDIT & SECURITY ENTITIES
-- =============================================================================

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
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);
CREATE INDEX idx_sessions_user_active ON sessions(user_id, is_active);

-- Indexes for Invoice Tables
CREATE INDEX idx_invoice_transaction_type ON invoice(transaction_type);
CREATE INDEX idx_invoice_supplier_id ON invoice(supplier_id);
CREATE INDEX idx_invoice_expense_category_id ON invoice(expense_category_id);
CREATE INDEX idx_invoice_transaction_date ON invoice(transaction_date);
CREATE INDEX idx_invoice_details_invoice_id ON invoice_details(invoice_id);
CREATE INDEX idx_invoice_details_ingredient_id ON invoice_details(ingredient_id);

-- =============================================================================
-- ADD FOREIGN KEY CONSTRAINTS THAT WERE DEFERRED
-- =============================================================================

-- Add foreign key constraints for user references
ALTER TABLE runout_ingredient_report ADD CONSTRAINT fk_runout_employee_id 
    FOREIGN KEY (employee_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE waste_loss ADD CONSTRAINT fk_waste_employee_id 
    FOREIGN KEY (employee_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_salary ADD CONSTRAINT fk_salary_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

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
('Ingredients', 'Raw materials and ingredients for ice cream production'),
('Equipment', 'Machinery and equipment purchases'),
('Utilities', 'Electricity, water, gas bills'),
('Rent', 'Store rent and property costs'),
('Salaries', 'Employee wages and benefits'),
('Marketing', 'Advertising and promotional expenses'),
('Maintenance', 'Equipment maintenance and repairs'),
('Packaging', 'Containers, cups, spoons, napkins'),
('Transportation', 'Delivery and transportation costs'),
('Other', 'Miscellaneous expenses');

-- Insert default ingredient categories
INSERT INTO ingredient_categories (name, description) VALUES
('dairy_products', 'Milk, cream, butter, eggs, cheese, yogurt'),
('sweeteners', 'Sugar, honey, artificial sweeteners, syrups, agave'),
('flavorings_extracts', 'Vanilla extract, almond extract, food coloring, artificial flavors'),
('fruits_fresh', 'Fresh strawberries, bananas, berries, seasonal fruits'),
('fruits_preserved', 'Dried fruits, fruit purees, jams, frozen fruits'),
('nuts_seeds', 'Almonds, walnuts, pistachios, coconut, seeds'),
('chocolate_cocoa', 'Cocoa powder, chocolate chips, chocolate bars, white chocolate'),
('stabilizers_emulsifiers', 'Xanthan gum, lecithin, agar, gelatin'),
('candies_confections', 'Gummy bears, chocolate chips, candy pieces, marshmallows'),
('cookies_baked_goods', 'Cookie crumbs, brownie pieces, cake chunks'),
('cereals_grains', 'Granola, cereal pieces, oats, rice crisps'),
('sauces_syrups', 'Chocolate sauce, caramel, fruit syrups, hot fudge'),
('toppings_garnishes', 'Sprinkles, whipped cream, cherries, nuts for topping'),
('specialty_toppings', 'Edible glitter, candy decorations, specialty sauces'),
('containers_cups', 'Ice cream cups, takeaway containers, tubs, bowls'),
('cones_wafers', 'Ice cream cones, waffle cones, crepes, wafer cookies'),
('utensils_serving', 'Spoons, napkins, straws, stirrers'),
('packaging_materials', 'Bags, lids, labels, boxes, wrapping'),
('cleaning_supplies', 'Sanitizers, detergents, cleaning cloths, brushes'),
('equipment_parts', 'Machine parts, filters, maintenance supplies'),
('office_supplies', 'Receipt paper, pens, markers, tags'),
('beverages', 'Coffee, tea, soft drinks, water'),
('snacks_sides', 'Chips, crackers, other snacks'),
('bread_pastry', 'Bread for sandwiches, pastries, muffins');

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

CREATE TRIGGER update_expense_categories_updated_at BEFORE UPDATE ON expense_categories 
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