# Ice Cream Store - Database Schema
## Server-Side Database Design Document

**Ticket ID:** 1  
**Project:** Ice cream store management system  
**Focus:** Database schema design for all application entities

---

## Table of Contents
1. [Inventory Management Entities](#inventory-management-entities)
   - [Suppliers Table](#suppliers-table)
   - [Expense Receipts Table](#expense-receipts-table)
   - [Ingredients Table](#ingredients-table)
   - [Existences Table](#existences-table)
   - [Runout Ingredient Report Table](#runout-ingredient-report-table)
   - [Recipes Table](#recipes-table)
   - [Recipe Ingredients Table](#recipe-ingredients-table)
2. [Expenses Management Entities](#expenses-management-entities)
   - [Expense Categories Table](#expense-categories-table)
   - [Expenses Table](#expenses-table)
3. [Income Management (Orders) Entities](#income-management-orders-entities)
   - [Orders Table](#orders-table)
   - [Ordered Receipes Table](#ordered-receipes-table)
4. [Administration Panel Entities](#administration-panel-entities)
   - [System Configuration Table](#system-configuration-table)
   - [User Salary Table](#user-salary-table)
5. [Authentication & Authorization Entities](#authentication--authorization-entities)
   - [Users Table](#users-table)
   - [Roles Table](#roles-table)
   - [Permissions Table](#permissions-table)

---

## Prerequisites
```sql
-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

## Inventory Management Entities

### Suppliers Table
**Purpose:** Store supplier/vendor information for ingredient procurement

```sql
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_name VARCHAR(255) NOT NULL UNIQUE,
    contact_number VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_suppliers_name ON suppliers(supplier_name);
CREATE INDEX idx_suppliers_email ON suppliers(email);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `supplier_name`: Name of the supplier/vendor (unique)
- `contact_number`: Phone number for supplier contact
- `email`: Email address for supplier communication
- `address`: Physical address of supplier
- `notes`: Additional notes about the supplier

### Expense Receipts Table
**Purpose:** Store receipt/invoice documentation with images and amounts, linked to expense categories. Each expense receipt can contain multiple ingredient purchases (existences) and is categorized through the parent expense record.

```sql
CREATE TABLE expense_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    receipt_number VARCHAR(100) UNIQUE NOT NULL,
    purchase_date DATE NOT NULL,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    total_amount DECIMAL(12,2), -- get all existences for that recipt number to get total amount
    image_url VARCHAR(500) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_expense_receipts_expense ON expense_receipts(expense_id);
CREATE INDEX idx_expense_receipts_number ON expense_receipts(receipt_number);
CREATE INDEX idx_expense_receipts_supplier ON expense_receipts(supplier_id);
CREATE INDEX idx_expense_receipts_purchase_date ON expense_receipts(purchase_date);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `expense_id`: Foreign key reference to expenses table (UUID)
- `receipt_number`: Receipt/invoice number (unique)
- `purchase_date`: When the purchase was made
- `supplier_id`: Foreign key reference to suppliers table (UUID, nullable for supermarket purchases)
- `total_amount`: Total amount of the expense receipt/invoice
- `image_url`: URL/path to uploaded receipt/invoice image (mandatory)
- `notes`: Additional notes about the purchase

### Ingredients Table
**Purpose:** Store raw materials/ingredients information with pricing and supplier details

```sql
CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
);

-- Indexes
CREATE INDEX idx_ingredients_name ON ingredients(name);
CREATE INDEX idx_ingredients_supplier ON ingredients(supplier_id);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Ingredient identifier/name (unique)
- `supplier_id`: Foreign key reference to suppliers table (UUID, nullable for local store purchases)

### Existences Table
**Purpose:** Track individual ingredient purchases/acquisitions from suppliers or supermarkets. Each record represents a specific purchase batch with receipt traceability.

```sql
CREATE TABLE existences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_reference_code INTEGER UNIQUE NOT NULL,
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    expense_receipt_id UUID NOT NULL REFERENCES expense_receipts(id) ON DELETE CASCADE,
    --units
    units_purchased DECIMAL(10,2) NOT NULL,
    units_available DECIMAL(10,2) NOT NULL, -- at creation it will be the same as units_purchased, updated with running out
    unit_Type VARCHAR(20) NOT NULL CHECK (unit IN ('Liters', 'Gallons', 'Units', 'Bag')),
    --items
    items_per_unit INTEGER NOT NULL, --ie. Galon has 31 ice-cream balls
    cost_per_item DECIMAL(10,2) GENERATED ALWAYS AS (cost_per_unit / items_per_unit) STORED,
    cost_per_unit DECIMAL(10,2) NOT NULL, --ie. Galon costs 12,000
    --costs
    total_purchase_cost DECIMAL(12,2) GENERATED ALWAYS AS (units_purchased * cost_per_unit) STORED,
    remaining_value DECIMAL(12,2) GENERATED ALWAYS AS (units_available * cost_per_unit) STORED,
    --expiry
    expiration_date DATE,
    --incomes & taxes
    income_margin_percentage DECIMAL(5,2) DEFAULT 30.00, -- grabbed from config
    income_margin_amount DECIMAL(10,2) GENERATED ALWAYS AS (total_recipe_cost * income_margin_percentage / 100) STORED,
    iva_percentage DECIMAL(5,2) DEFAULT 13.00, -- grabbed from config
    iva_amount DECIMAL(10,2) GENERATED ALWAYS AS ((total_recipe_cost + income_margin_amount) * iva_percentage / 100) STORED,
    service_tax_percentage DECIMAL(5,2) DEFAULT 10.00,
    service_tax_amount DECIMAL(10,2) GENERATED ALWAYS AS ((total_recipe_cost + income_margin_amount) * service_tax_percentage / 100) STORED,
    calculated_price DECIMAL(10,2) GENERATED ALWAYS AS (total_recipe_cost + income_margin_amount + iva_amount + service_tax_amount) STORED,
    final_price DECIMAL(10,2),
    --dates
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Auto-increment sequence for existence_reference_code
CREATE SEQUENCE existence_reference_seq START 1;
ALTER TABLE existences ALTER COLUMN existence_reference_code SET DEFAULT nextval('existence_reference_seq');

-- Indexes
CREATE INDEX idx_existences_ingredient ON existences(ingredient_id);
CREATE INDEX idx_existences_reference_code ON existences(existence_reference_code);
CREATE INDEX idx_existences_receipt ON existences(expense_receipt_id);
CREATE INDEX idx_existences_available ON existences(units_available);
CREATE INDEX idx_existences_cost_per_item ON existences(cost_per_item);
CREATE INDEX idx_existences_expiration_date ON existences(expiration_date);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `existence_reference_code`: Simple numeric consecutive code for easy identification
- `ingredient_id`: Foreign key reference to ingredients table (UUID)
- `expense_receipt_id`: Foreign key reference to expense_receipts table (UUID)
- `units_purchased`: Original quantity purchased
- `units_available`: Current quantity available (at creation same as units_purchased, decreases as used)
- `unit_type`: Unit of measurement for this existence (Liters, Gallons, Units, Bag)
- `items_per_unit`: Number of individual items produced from one unit (e.g., 1 Gallon = 31 ice cream balls)
- `cost_per_item`: Calculated field (cost_per_unit ÷ items_per_unit) - cost per individual item
- `cost_per_unit`: Cost per unit for this specific purchase (e.g., Gallon costs ₡12,000)
- `total_purchase_cost`: Calculated field (units_purchased × cost_per_unit)
- `remaining_value`: Calculated field (units_available × cost_per_unit)
- `expiration_date`: Expiration date for this specific ingredient batch (nullable)
- `income_margin_percentage`: Configurable margin percentage (default 30%, from config)
- `income_margin_amount`: Calculated margin amount (read-only)
- `iva_percentage`: IVA tax percentage (default 13%, from config)
- `iva_amount`: IVA tax amount (read-only auto-generated)
- `service_tax_percentage`: Service tax percentage (default 10%, from config)
- `service_tax_amount`: Service tax amount (read-only auto-generated)
- `calculated_price`: Auto-calculated total price with margins and taxes
- `final_price`: Final price (can be rounded up to next 100)

### Runout Ingredient Report Table
**Purpose:** Track ingredient usage and runouts reported by employees. Updates existences table to reflect ingredient consumption.

```sql
CREATE TABLE runout_ingredient_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    quantity DECIMAL(10,2) NOT NULL,
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('Liters', 'Gallons', 'Units', 'Bag')),
    report_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_runout_report_existence ON runout_ingredient_report(existence_id);
CREATE INDEX idx_runout_report_employee ON runout_ingredient_report(employee_id);
CREATE INDEX idx_runout_report_date ON runout_ingredient_report(report_date);
CREATE INDEX idx_runout_report_unit_type ON runout_ingredient_report(unit_type);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `existence_id`: Foreign key reference to existences table (UUID)
- `employee_id`: Foreign key reference to users table (employee who reported the runout)
- `quantity`: Amount of ingredient that was used/ran out
- `unit_type`: Unit of measurement for the reported quantity (Liters, Gallons, Units, Bag)
- `report_date`: Date when the runout was reported (defaults to current date)
- `created_at`: When the runout report was created
- `updated_at`: When the runout report was last modified

### Recipes Table
**Purpose:** Store product recipes with pricing information

```sql
CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_name VARCHAR(255) NOT NULL UNIQUE,
    recipe_description TEXT,
    picture_url VARCHAR(500),

    total_recipe_cost DECIMAL(10,2) DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_recipes_name ON recipes(recipe_name);
CREATE INDEX idx_recipes_final_price ON recipes(final_price);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `recipe_name`: Name of the product/recipe (unique)
- `recipe_description`: Description of the product
- `picture_url`: Picture of the product for reference
- `total_recipe_cost`: Sum of all material costs in the recipe

### Recipe Ingredients Table
**Purpose:** Junction table linking recipes to ingredients with quantities

```sql
CREATE TABLE recipe_ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    number_of_units DECIMAL(10,3) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(recipe_id, ingredient_id)
);

-- Indexes
CREATE INDEX idx_recipe_ingredients_recipe ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_ingredient ON recipe_ingredients(ingredient_id);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `recipe_id`: Foreign key reference to recipes table (UUID)
- `ingredient_id`: Foreign key reference to ingredients table (UUID)
- `number_of_units`: Quantity of the raw material needed for the recipe

## Overall System Relationships

### Inventory Management
- **suppliers** ← **expense_receipts** (One-to-Many: One supplier can have multiple expense receipts)
- **suppliers** ← **ingredients** (One-to-Many: One supplier can provide multiple ingredients)
- **expense_receipts** ← **existences** (One-to-Many: One expense receipt can contain multiple existences/line items)
- **ingredients** ← **existences** (One-to-Many: One ingredient can have multiple purchase batches/existences)
- **existences** ← **runout_ingredient_report** (One-to-Many: One existence can have multiple runout reports)
- **users** ← **runout_ingredient_report** (One-to-Many: One employee can create multiple runout reports)
- **recipes** ← **recipe_ingredients** → **ingredients** (Many-to-Many: Recipes contain multiple ingredients, ingredients can be in multiple recipes)

### Expenses Management
- **expense_categories** ← **expenses** (One-to-Many: One category can have multiple expenses)
- **expenses** ← **expense_receipts** (One-to-Many: One expense can have multiple receipts)

### Income Management
- **users** ← **orders** (One-to-Many: One sales representative can create multiple orders)
- **orders** ← **ordered_receipes** (One-to-Many: One order can have multiple recipe line items)
- **recipes** ← **ordered_receipes** (One-to-Many: One recipe can be ordered multiple times)

### Authentication & Authorization
- **roles** ← **users** (One-to-Many: One role can be assigned to multiple users)
- **users** ← **user_salary** (One-to-Many: One user can have multiple salary records)
- **expenses** ← **user_salary** (One-to-Many: One expense can be linked to multiple salary records)
- **roles** ← **permissions** (One-to-Many: One role can have multiple permissions)

## Overall Business Logic Triggers

### Inventory Management
- Update `total_recipe_cost` in recipes table when recipe_ingredients change
- Recalculate pricing fields in existences table when cost components change
- Track ingredient consumption by updating `units_available` in existences table
- Process runout reports: when employees report ingredient usage, create runout_ingredient_report record and decrease `units_available` in existences table accordingly
- Validate runout report quantities against available stock in existences table
- Implement FIFO logic: use oldest expense receipts first (by purchase_date from expense_receipts table)
- Alert when existences are near expiry (based on expiration_date from existences table)
- Calculate final pricing (margins, taxes) at existence level for inventory items
- Maintain expense receipt totals when existences are added/modified
- Link expense receipts to expense management system for accounting integration

### Expenses Management
- Organize invoice documents in monthly directories (MM-yyyy format)
- Calculate monthly expense totals from expense receipts
- Link expense receipts to their parent expense categories through the expenses table
- Validate receipt image uploads for all expense receipts
- Track employee salaries through user_salary table linked to expense records
- Calculate total compensation automatically (salary + additional_expenses)
- Maintain salary audit trail with creation and update timestamps

---

## Expenses Management Entities

### Expense Categories Table
**Purpose:** Define and manage different categories of business expenses for classification and reporting purposes.

```sql
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_expense_categories_name ON expense_categories(category_name);
CREATE INDEX idx_expense_categories_active ON expense_categories(is_active);

-- Insert default categories
INSERT INTO expense_categories (category_name, description) VALUES
('Salary payments', 'Employee salaries and wages'),
('Service payments', 'Utility services, maintenance, subscriptions'),
('Rent payments', 'Property rent and lease payments'),
('Ingredients', 'Ingredient and supply purchases'),
('Other operational expenses', 'Miscellaneous business expenses');
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `category_name`: Name of the expense category (unique)
- `description`: Detailed description of the category
- `is_active`: Whether the category is currently active/available
- `created_at`: When the category was created
- `updated_at`: When the category was last modified

### Expenses Table
**Purpose:** Define business expense categories and descriptions for organizational purposes.

```sql
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_expenses_category ON expenses(expense_category_id);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `expense_category_id`: Foreign key reference to expense_categories table
- `description`: Brief description of the expense
- `created_at`: When the expense record was created
- `updated_at`: When the expense record was last modified



---

## Income Management (Orders) Entities

### Orders Table
**Purpose:** Track all customer transactions/sales with complete product and payment information for accurate income analysis.

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    sales_representative_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled')) DEFAULT 'pending',
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'card', 'sinpe')),
    transaction_reference VARCHAR(100), -- For card and sinpe payments
    sinpe_screenshot_url VARCHAR(500), -- Required for sinpe payments
    subtotal_amount DECIMAL(12,2) NOT NULL,
    iva_amount DECIMAL(12,2) NOT NULL, -- 13% IVA tax
    service_tax_amount DECIMAL(12,2) NOT NULL, -- 10% service tax
    total_amount DECIMAL(12,2) NOT NULL,
    invoice_number VARCHAR(50) UNIQUE,
    invoice_url VARCHAR(500),
    transaction_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Auto-increment sequence for order_number
CREATE SEQUENCE order_number_seq START 1;
ALTER TABLE orders ALTER COLUMN order_number SET DEFAULT 'ORD-' || LPAD(nextval('order_number_seq')::text, 6, '0');

-- Auto-increment sequence for invoice_number (generated when order completed)
CREATE SEQUENCE invoice_number_seq START 1;

-- Indexes
CREATE INDEX idx_orders_number ON orders(order_number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_method ON orders(payment_method);
CREATE INDEX idx_orders_sales_rep ON orders(sales_representative_id);
CREATE INDEX idx_orders_transaction_timestamp ON orders(transaction_timestamp);
CREATE INDEX idx_orders_invoice_number ON orders(invoice_number);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `order_number`: Unique order identifier (auto-generated: ORD-000001, ORD-000002, etc.)
- `sales_representative_id`: Foreign key reference to users table (employee who processed sale)
- `status`: Order status (pending → completed/cancelled)
- `payment_method`: Payment method used (cash, card, sinpe)
- `transaction_reference`: Transaction reference for card/sinpe payments (required for non-cash)
- `sinpe_screenshot_url`: Required screenshot URL for sinpe payments
- `subtotal_amount`: Order subtotal before taxes
- `iva_amount`: IVA tax amount (13%)
- `service_tax_amount`: Service tax amount (10%)
- `total_amount`: Final total amount
- `invoice_number`: Sequential invoice number (generated when order completed)
- `invoice_url`: URL to generated invoice document
- `transaction_timestamp`: When the transaction occurred
- `completed_at`: When the order was completed (nullable)

### Ordered Receipes Table
**Purpose:** Track individual products sold in each order with quantities and pricing snapshots.

```sql
CREATE TABLE ordered_receipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL, -- Snapshot of recipe name at time of sale
    quantity INTEGER NOT NULL,
    receipe_price DECIMAL(10,2) NOT NULL, -- Snapshot of recipe price at time of sale
    subtotal DECIMAL(12,2) GENERATED ALWAYS AS (quantity * receipe_price) STORED
);

-- Indexes
CREATE INDEX idx_ordered_receipes_order ON ordered_receipes(order_id);
CREATE INDEX idx_ordered_receipes_recipe ON ordered_receipes(recipe_id);
CREATE INDEX idx_ordered_receipes_product_name ON ordered_receipes(product_name);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `order_id`: Foreign key reference to orders table
- `recipe_id`: Foreign key reference to recipes table
- `product_name`: Snapshot of product name at time of sale (for historical accuracy)
- `quantity`: Number of items ordered
- `receipe_price`: Snapshot of recipe price at time of sale (for historical accuracy)
- `subtotal`: Calculated subtotal for this line item (quantity × receipe_price)

---

## Administration Panel Entities

### System Configuration Table
**Purpose:** Store system-wide configuration parameters and business settings for centralized management.

```sql
CREATE TABLE system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) NOT NULL CHECK (config_type IN ('string', 'number', 'boolean', 'decimal')),
    description TEXT,
    is_editable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default configuration values
INSERT INTO system_config (config_key, config_value, config_type, description) VALUES
('default_income_margin_percentage', '30.00', 'decimal', 'Default income margin percentage for pricing calculations'),
('default_iva_percentage', '13.00', 'decimal', 'Default IVA tax percentage'),
('default_service_tax_percentage', '10.00', 'decimal', 'Default service tax percentage'),
('low_stock_threshold', '1', 'number', 'Minimum stock level threshold for alerts'),
('expiration_warning_days', '7', 'number', 'Days before expiration to show warnings'),
('business_name', 'Ice Cream Store', 'string', 'Business name for invoices and reports'),
('business_address', '', 'string', 'Business address for invoices'),
('business_phone', '', 'string', 'Business phone number'),
('business_email', '', 'string', 'Business email address');

-- Indexes
CREATE INDEX idx_system_config_key ON system_config(config_key);
CREATE INDEX idx_system_config_editable ON system_config(is_editable);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `config_key`: Unique configuration parameter key
- `config_value`: Configuration value (stored as text, cast based on type)
- `config_type`: Data type of the configuration value (string, number, boolean, decimal)
- `description`: Human-readable description of the parameter
- `is_editable`: Whether this config can be modified through the administration UI

### User Salary Table
**Purpose:** Track employee salaries and link them to expense management for payroll processing.

```sql
CREATE TABLE user_salary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    salary DECIMAL(12,2) NOT NULL,
    additional_expenses DECIMAL(12,2) DEFAULT 0.00,
    total DECIMAL(12,2) GENERATED ALWAYS AS (salary + additional_expenses) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_user_salary_user ON user_salary(user_id);
CREATE INDEX idx_user_salary_expense ON user_salary(expense_id);
CREATE INDEX idx_user_salary_total ON user_salary(total);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `user_id`: Foreign key reference to users table (employee)
- `expense_id`: Foreign key reference to expenses table (links salary to expense tracking)
- `salary`: Base salary amount for the employee
- `additional_expenses`: Extra expenses or bonuses (defaults to 0.00)
- `total`: Calculated total compensation (salary + additional_expenses)
- `created_at`: When the salary record was created
- `updated_at`: When the salary record was last modified

---

## Authentication & Authorization Entities

### Users Table
**Purpose:** Store user accounts for employees and administrators with internal authentication.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- Hashed password for authentication
    full_name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role_id);
CREATE INDEX idx_users_active ON users(is_active);

-- Insert default admin user (password should be changed on first login)
-- Note: Replace 'hashed_password_here' with actual bcrypt hash of temporary password
INSERT INTO users (username, password_hash, full_name, role_id, is_active)
SELECT 'admin', 'hashed_password_here', 'System Administrator', r.id, true
FROM roles r WHERE r.role_name = 'admin';
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `username`: User username (unique)
- `password_hash`: Hashed password for secure authentication
- `full_name`: User's full name
- `role_id`: Foreign key reference to roles table (UUID)
- `is_active`: Whether the user account is active
- `last_login`: Timestamp of last successful login

### Roles Table
**Purpose:** Define user roles in the system for access control.

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO roles (role_name, description) VALUES
('admin', 'Full system access with administrative privileges'),
('employee', 'Limited access for regular employees');

-- Indexes
CREATE INDEX idx_roles_name ON roles(role_name);
CREATE INDEX idx_roles_active ON roles(is_active);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `role_name`: Name of the role (unique)
- `description`: Description of role responsibilities and access level
- `is_active`: Whether the role is currently active

### Permissions Table
**Purpose:** Define granular permissions for system operations using [entity]-[action] naming convention.

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_name VARCHAR(100) NOT NULL, -- Format: Entity-Action (e.g., "Ingredients-Create")
    description TEXT,
    entity_name VARCHAR(50) NOT NULL,
    action_name VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(role_id, permission_name)
);



-- Indexes
CREATE INDEX idx_permissions_role ON permissions(role_id);
CREATE INDEX idx_permissions_name ON permissions(permission_name);
CREATE INDEX idx_permissions_entity ON permissions(entity_name);
CREATE INDEX idx_permissions_action ON permissions(action_name);
CREATE INDEX idx_permissions_active ON permissions(is_active);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `role_id`: Foreign key reference to roles table (UUID)
- `permission_name`: Permission identifier following [Entity]-[Action] format
- `description`: Human-readable description of the permission
- `entity_name`: The entity/resource being accessed (e.g., Ingredients, Orders)
- `action_name`: The action being performed (Create, Read, Update, Delete)
- `is_active`: Whether the permission is currently active

-- Insert admin permissions (admin gets all permissions)
INSERT INTO permissions (role_id, permission_name, description, entity_name, action_name)
SELECT r.id, 'Suppliers-Create', 'Create new suppliers', 'Suppliers', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Suppliers-Read', 'View supplier information', 'Suppliers', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Suppliers-Update', 'Update supplier information', 'Suppliers', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Suppliers-Delete', 'Delete suppliers', 'Suppliers', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Ingredients-Create', 'Create new ingredients', 'Ingredients', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Ingredients-Read', 'View ingredient information', 'Ingredients', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Ingredients-Update', 'Update ingredient information', 'Ingredients', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Ingredients-Delete', 'Delete ingredients', 'Ingredients', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Existences-Create', 'Create new existences', 'Existences', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Existences-Read', 'View existence information', 'Existences', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Existences-Update', 'Update existence information', 'Existences', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Existences-Delete', 'Delete existences', 'Existences', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'RunoutReports-Create', 'Create new runout ingredient reports', 'RunoutReports', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'RunoutReports-Read', 'View runout ingredient reports', 'RunoutReports', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'RunoutReports-Update', 'Update runout ingredient reports', 'RunoutReports', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'RunoutReports-Delete', 'Delete runout ingredient reports', 'RunoutReports', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Recipes-Create', 'Create new recipes', 'Recipes', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Recipes-Read', 'View recipe information', 'Recipes', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Recipes-Update', 'Update recipe information', 'Recipes', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Recipes-Delete', 'Delete recipes', 'Recipes', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Expenses-Create', 'Create new expenses', 'Expenses', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Expenses-Read', 'View expense information', 'Expenses', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Expenses-Update', 'Update expense information', 'Expenses', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Expenses-Delete', 'Delete expenses', 'Expenses', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Orders-Create', 'Create new orders', 'Orders', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Orders-Read', 'View order information', 'Orders', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Orders-Update', 'Update order information', 'Orders', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Orders-Delete', 'Cancel orders', 'Orders', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Users-Create', 'Create new users', 'Users', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Users-Read', 'View user information', 'Users', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Users-Update', 'Update user information', 'Users', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Users-Delete', 'Delete users', 'Users', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Salaries-Create', 'Create new salary records', 'Salaries', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Salaries-Read', 'View salary information', 'Salaries', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Salaries-Update', 'Update salary records', 'Salaries', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Salaries-Delete', 'Delete salary records', 'Salaries', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Config-Read', 'View system configuration', 'Config', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Config-Update', 'Update system configuration', 'Config', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Reports-Read', 'View business reports and analytics', 'Reports', 'Read' FROM roles r WHERE r.role_name = 'admin';

-- Insert employee permissions (limited access)
INSERT INTO permissions (role_id, permission_name, description, entity_name, action_name)
SELECT r.id, 'Suppliers-Read', 'View supplier information', 'Suppliers', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Ingredients-Read', 'View ingredient information', 'Ingredients', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Existences-Read', 'View existence information', 'Existences', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Recipes-Read', 'View recipe information', 'Recipes', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'RunoutReports-Create', 'Create new runout ingredient reports', 'RunoutReports', 'Create' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'RunoutReports-Read', 'View runout ingredient reports', 'RunoutReports', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Orders-Create', 'Create new orders', 'Orders', 'Create' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Orders-Read', 'View order information', 'Orders', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Orders-Update', 'Update order information', 'Orders', 'Update' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Reports-Read', 'View business reports and analytics', 'Reports', 'Read' FROM roles r WHERE r.role_name = 'employee';

---

**Database Schema Complete!**
All sections from Requirements.md have been implemented in the database design. 