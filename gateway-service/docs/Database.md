# Ice Cream Store - Database Schema
## Server-Side Database Design Document

**Ticket ID:** 1  
**Project:** Ice cream store management system  
**Focus:** Database schema design for all application entities

---

## Table of Contents
1. [Inventory Management Entities](#inventory-management-entities)
   - [Suppliers Table](#suppliers-table)
   - [Ingredients Table](#ingredients-table)
   - [Existences Table](#existences-table)
   - [Runout Ingredient Report Table](#runout-ingredient-report-table)
   - [Recipe Categories Table](#recipe-categories-table)
   - [Recipes Table](#recipes-table)
   - [Recipe Ingredients Table](#recipe-ingredients-table)
2. [Expenses Management Entities](#expenses-management-entities)
   - [Expense Categories Table](#expense-categories-table)
   - [Expenses Table](#expenses-table)
   - [Expense Receipts Table](#expense-receipts-table)
3. [Customer Management Entities](#customer-management-entities)
   - [Customers Table](#customers-table)
4. [Income Management (Orders) Entities](#income-management-orders-entities)
   - [Orders Table](#orders-table)
   - [Ordered Receipes Table](#ordered-receipes-table)
5. [Promotions & Loyalty System Entities](#promotions--loyalty-system-entities)
   - [Promotions Table](#promotions-table)
   - [Customer Points Table](#customer-points-table)
6. [Equipment Management Entities](#equipment-management-entities)
   - [Mechanics Table](#mechanics-table)
   - [Equipment Table](#equipment-table)
7. [Waste & Loss Tracking Entities](#waste--loss-tracking-entities)
   - [Waste Loss Table](#waste-loss-table)
8. [Administration Panel Entities](#administration-panel-entities)
   - [System Configuration Table](#system-configuration-table)
   - [User Salary Table](#user-salary-table)
9. [Authentication & Authorization Entities](#authentication--authorization-entities)
   - [Users Table](#users-table)
   - [Roles Table](#roles-table)
   - [Permissions Table](#permissions-table)
10. [Audit & Security Entities](#audit--security-entities)
    - [Audit Logs Table](#audit-logs-table)

---

## Prerequisites
```sql
-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

## 🏗️ Service Architecture Integration

This database schema supports a **microservices architecture** with **9 specialized services**:

- **🔐 Session Service**: JWT tokens, login/logout (reads user data from Administration Service)
- **📋 Audit Service**: Activity logging (`audit_logs` table)
- **⚙️ Administration Service**: User/role/permission management, equipment tracking (`users`, `roles`, `permissions`, `system_config`, `user_salary`, `mechanics`, `equipment` tables) - **Admin only**
- **👥 Customer Service**: Customer management (`customers` table)
- **💰 Expenses Service**: Financial management (`expense_categories`, `expenses`, `expense_receipts` tables)
- **📦 Inventory Service**: Core business logic (`suppliers`, `ingredients`, `existences`, `runout_ingredient_report`, `recipe_categories`, `recipes`, `recipe_ingredients` tables)
- **🎉 Promotions Service**: Loyalty programs (`promotions`, `customer_points` tables)
- **🛒 Orders Service**: Sales processing (`orders`, `ordered_receipes` tables)
- **🗑️ Waste Service**: Loss analysis (`waste_loss` table)

> **Security Model**: Session Service handles login/JWT tokens, while Administration Service manages all user/role/permission CRUD operations with admin-only access.

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

### Recipe Categories Table
**Purpose:** Categorize recipes by product type for better organization and filtering

```sql
CREATE TABLE recipe_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default categories
INSERT INTO recipe_categories (name, description) VALUES
('Postres', 'Desserts and sweet treats'),
('Helados', 'Traditional ice cream products'),
('Batidos', 'Milkshakes and blended drinks'),
('Gelato', 'Italian-style gelato products'),
('Artesanales', 'Artisan and handcrafted specialty items');

-- Indexes
CREATE INDEX idx_recipe_categories_name ON recipe_categories(name);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Category name (unique)
- `description`: Description of the category type

### Recipes Table
**Purpose:** Store product recipes with pricing information and categorization

```sql
CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_name VARCHAR(255) NOT NULL UNIQUE,
    recipe_description TEXT,
    picture_url VARCHAR(500),
    recipe_category_id UUID NOT NULL REFERENCES recipe_categories(id) ON DELETE RESTRICT,
    total_recipe_cost DECIMAL(10,2) DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_recipes_name ON recipes(recipe_name);
CREATE INDEX idx_recipes_category ON recipes(recipe_category_id);
CREATE INDEX idx_recipes_final_price ON recipes(final_price);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `recipe_name`: Name of the product/recipe (unique)
- `recipe_description`: Description of the product
- `picture_url`: Picture of the product for reference
- `recipe_category_id`: Foreign key reference to recipe_categories table (UUID)
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
- **recipe_categories** ← **recipes** (One-to-Many: One category can contain multiple recipes)
- **recipes** ← **recipe_ingredients** → **ingredients** (Many-to-Many: Recipes contain multiple ingredients, ingredients can be in multiple recipes)

### Expenses Management
- **expense_categories** ← **expenses** (One-to-Many: One category can have multiple expenses)
- **expenses** ← **expense_receipts** (One-to-Many: One expense can have multiple receipts)

### Customer Management
- **customers** ← **orders** (One-to-Many: One customer can have multiple orders)
- **customers** ← **customer_points** (One-to-Many: One customer can have multiple point transactions)

### Income Management
- **users** ← **orders** (One-to-Many: One sales representative can create multiple orders)
- **customers** ← **orders** (One-to-Many: One customer can have multiple orders)
- **orders** ← **ordered_receipes** (One-to-Many: One order can have multiple recipe line items)
- **orders** ← **customer_points** (One-to-Many: One order can generate customer points)
- **recipes** ← **ordered_receipes** (One-to-Many: One recipe can be ordered multiple times)

### Promotions & Loyalty System
- **recipe_categories** ← **recipes** ← **promotions** (One-to-Many chain: Categories contain recipes, recipes can have promotions)
- **recipes** ← **promotions** (One-to-Many: One recipe can have multiple promotions)
- **customers** ← **customer_points** (One-to-Many: One customer can have multiple point records)
- **orders** ← **customer_points** (One-to-Many: One order can award customer points)

### Administration & Equipment Management
- **mechanics** ← **equipment** (One-to-Many: One mechanic can service multiple equipment)

### Waste & Loss Tracking
- **existences** ← **waste_loss** (One-to-Many: One existence can have multiple waste records)
- **users** ← **waste_loss** (One-to-Many: One employee can report multiple waste incidents)

### Authentication & Authorization
- **roles** ← **users** (One-to-Many: One role can be assigned to multiple users)
- **users** ← **user_salary** (One-to-Many: One user can have multiple salary records)
- **users** ← **audit_logs** (One-to-Many: One user can generate multiple audit log entries)
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

### Customer Management & Loyalty
- Optional customer linking during order creation for marketing and loyalty programs
- Track customer purchase history and preferences through order relationships
- Maintain customer contact information for promotional campaigns

### Promotions & Discounts
- Automatically validate and apply promotions during order creation based on time, recipe, and customer eligibility
- Calculate discount amounts and apply to order total
- Award customer loyalty points based on active promotions upon order completion
- For points_reward promotions: validate minimum purchase amount condition before awarding points
- Calculate point expiration dates based on promotion's points_expiration_duration (1d/3w/7m/2y format)
- Track point accumulation and redemption for customer loyalty program
- Enforce promotion time limits and recipe-specific rules

### Equipment Management
- Schedule equipment maintenance based on maintenance_schedule intervals
- Generate maintenance alerts when next_maintenance_date approaches
- Track equipment downtime and maintenance costs
- Update equipment status based on maintenance requirements
- Maintain mechanic contact information for emergency repairs

### Waste & Loss Tracking
- Calculate financial loss automatically when waste is reported (items_wasted × existence price per unit)
- Update existence quantities when waste is recorded: decrease `units_available` in existences table by items_wasted amount
- Track waste patterns by type, date, and employee for analysis
- Generate waste reports for cost analysis and prevention strategies
- Integrate waste tracking with inventory management to maintain accurate stock levels
- Validate that waste amounts do not exceed available existence quantities

### Audit & Security
- Automatically log all critical operations (user management, financial transactions, inventory changes)
- Track user authentication events (login, logout, failed attempts)
- Record IP addresses and user agents for security monitoring
- Maintain tamper-proof audit trail with timestamps
- Enable audit log searching and reporting for compliance
- Alert on suspicious activity patterns or failed operations

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

---

## Customer Management Entities

### Customers Table
**Purpose:** Store customer information for marketing, loyalty programs, and sales tracking.

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_customers_name ON customers(name);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_email ON customers(email);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Customer's full name (required)
- `phone`: Customer's phone number (optional, for marketing and contact)
- `email`: Customer's email address (optional, for marketing and notifications)
- `created_at`: When the customer record was created
- `updated_at`: When the customer record was last modified

---

## Income Management (Orders) Entities

### Orders Table
**Purpose:** Track all customer transactions/sales with complete product and payment information for accurate income analysis.

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    sales_representative_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled')) DEFAULT 'pending',
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'card', 'sinpe')),
    transaction_reference VARCHAR(100), -- For card and sinpe payments
    sinpe_screenshot_url VARCHAR(500), -- Required for sinpe payments
    subtotal_amount DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) DEFAULT 0.00, -- Total discount applied from promotions
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
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_method ON orders(payment_method);
CREATE INDEX idx_orders_sales_rep ON orders(sales_representative_id);
CREATE INDEX idx_orders_transaction_timestamp ON orders(transaction_timestamp);
CREATE INDEX idx_orders_invoice_number ON orders(invoice_number);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `order_number`: Unique order identifier (auto-generated: ORD-000001, ORD-000002, etc.)
- `customer_id`: Foreign key reference to customers table (nullable for walk-in customers)
- `sales_representative_id`: Foreign key reference to users table (employee who processed sale)
- `status`: Order status (pending → completed/cancelled)
- `payment_method`: Payment method used (cash, card, sinpe)
- `transaction_reference`: Transaction reference for card/sinpe payments (required for non-cash)
- `sinpe_screenshot_url`: Required screenshot URL for sinpe payments
- `subtotal_amount`: Order subtotal before taxes
- `discount_amount`: Total discount applied from promotions (defaults to 0.00)
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

## Promotions & Loyalty System Entities

### Promotions Table
**Purpose:** Manage promotional campaigns, discounts, and loyalty programs.

```sql
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE, -- null = applies to all recipes
    time_from TIMESTAMP NOT NULL, -- Start date/time for promotion
    time_to TIMESTAMP, -- End date/time for promotion (null = no end date)
    promotion_type VARCHAR(20) NOT NULL CHECK (promotion_type IN ('percentage_discount', 'points_reward')),
    value DECIMAL(10,2) NOT NULL, -- Percentage for discounts, points for loyalty
    -- Fields for points_reward promotions only
    minimum_purchase_amount DECIMAL(12,2), -- Minimum purchase amount to qualify for points (e.g., 5000 colones)
    points_expiration_duration VARCHAR(10), -- Duration format: 1d/3w/7m/2y (null = no expiration)
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_promotions_recipe ON promotions(recipe_id);
CREATE INDEX idx_promotions_type ON promotions(promotion_type);
CREATE INDEX idx_promotions_active ON promotions(is_active);
CREATE INDEX idx_promotions_time_range ON promotions(time_from, time_to);
CREATE INDEX idx_promotions_min_purchase ON promotions(minimum_purchase_amount);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Promotional campaign name
- `description`: Detailed description of the promotion
- `recipe_id`: Specific recipe the promotion applies to (nullable for store-wide promotions)
- `time_from`: Start date/time for time-limited promotions (required)
- `time_to`: End date/time for time-limited promotions (nullable for ongoing)
- `promotion_type`: Type of promotion (percentage_discount, points_reward)
- `value`: Promotion value (percentage for discounts, points awarded for loyalty)
- `minimum_purchase_amount`: Minimum purchase amount to qualify for points (only for points_reward promotions)
- `points_expiration_duration`: Duration format for point expiration: 1d/3w/7m/2y (only for points_reward, null = no expiration)
- `is_active`: Whether the promotion is currently active

### Customer Points Table
**Purpose:** Track customer loyalty points earned and spent for rewards program.

```sql
CREATE TABLE customer_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    points_earned INTEGER NOT NULL,
    points_source VARCHAR(20) NOT NULL CHECK (points_source IN ('purchase', 'promotion_bonus', 'manual_adjustment')),
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL, -- Order that generated points (if applicable)
    date_earned DATE NOT NULL DEFAULT CURRENT_DATE,
    expiration_date DATE, -- this will be calculated based on promotion expiration field
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_customer_points_customer ON customer_points(customer_id);
CREATE INDEX idx_customer_points_order ON customer_points(order_id);
CREATE INDEX idx_customer_points_source ON customer_points(points_source);
CREATE INDEX idx_customer_points_date ON customer_points(date_earned);
CREATE INDEX idx_customer_points_expiration ON customer_points(expiration_date);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `customer_id`: Foreign key reference to customers table
- `points_earned`: Number of points added to customer account
- `points_source`: How points were earned (purchase, promotion_bonus, manual_adjustment)
- `order_id`: Order that generated the points (nullable for non-purchase points)
- `date_earned`: When the points were awarded
- `expiration_date`: When points expire (calculated based on promotion's points_expiration_duration field)

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

### Mechanics Table
**Purpose:** Store contact information for equipment maintenance professionals.

```sql
CREATE TABLE mechanics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50) NOT NULL,
    specialization TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_mechanics_name ON mechanics(name);
CREATE INDEX idx_mechanics_phone ON mechanics(phone);
CREATE INDEX idx_mechanics_email ON mechanics(email);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Mechanic or company name
- `email`: Email contact for scheduling and communication (optional)
- `phone`: Primary phone contact for emergency repairs
- `specialization`: Equipment types or brands they specialize in (optional)
- `notes`: Additional notes about the mechanic (optional)

### Equipment Table
**Purpose:** Track store equipment with maintenance scheduling and cost management.

```sql
CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    purchase_date DATE NOT NULL,
    mechanic_id UUID REFERENCES mechanics(id) ON DELETE SET NULL,
    maintenance_schedule INTEGER NOT NULL, -- Days between maintenance (e.g., 90)
    purchase_cost DECIMAL(12,2) NOT NULL,
    current_status VARCHAR(20) NOT NULL CHECK (current_status IN ('operational', 'maintenance_required', 'out_of_service', 'retired')) DEFAULT 'operational',
    last_maintenance_date DATE,
    next_maintenance_date DATE, -- Calculated based on last maintenance + schedule
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_equipment_name ON equipment(name);
CREATE INDEX idx_equipment_mechanic ON equipment(mechanic_id);
CREATE INDEX idx_equipment_status ON equipment(current_status);
CREATE INDEX idx_equipment_next_maintenance ON equipment(next_maintenance_date);
CREATE INDEX idx_equipment_purchase_date ON equipment(purchase_date);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `name`: Equipment name/model
- `description`: Detailed description of the equipment
- `purchase_date`: When the equipment was purchased
- `mechanic_id`: Foreign key reference to assigned mechanic for maintenance
- `maintenance_schedule`: Days between scheduled maintenance (e.g., 90 days)
- `purchase_cost`: Original purchase cost of equipment
- `current_status`: Equipment status (operational, maintenance_required, out_of_service, retired)
- `last_maintenance_date`: Date of last maintenance performed
- `next_maintenance_date`: Calculated next maintenance due date



## Waste & Loss Tracking Entities

### Waste Loss Table
**Purpose:** Track expired ingredients and calculate financial losses for inventory management efficiency. 

```sql
CREATE TABLE waste_loss (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    existence_id UUID NOT NULL REFERENCES existences(id) ON DELETE CASCADE,
    waste_type VARCHAR(20) NOT NULL CHECK (waste_type IN ('expired', 'damaged', 'spoiled', 'theft', 'other')),
    items_wasted DECIMAL(10,2) NOT NULL, -- amount of items in a unit wasted
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('Liters', 'Gallons', 'Units', 'Bag')),
    financial_loss DECIMAL(12,2) NOT NULL, -- Calculated as: items_wasted * existence price per unit
    waste_date DATE NOT NULL DEFAULT CURRENT_DATE,
    reported_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    reason TEXT NOT NULL,
    prevention_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_waste_loss_existence ON waste_loss(existence_id);
CREATE INDEX idx_waste_loss_type ON waste_loss(waste_type);
CREATE INDEX idx_waste_loss_date ON waste_loss(waste_date);
CREATE INDEX idx_waste_loss_reported_by ON waste_loss(reported_by);
CREATE INDEX idx_waste_loss_financial ON waste_loss(financial_loss);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `existence_id`: Foreign key reference to specific existence/batch that was wasted
- `waste_type`: Type of waste (expired, damaged, spoiled, theft, other)
- `items_wasted`: Amount of items in a unit that were wasted
- `unit_type`: Unit of measurement (Liters, Gallons, Units, Bag)
- `financial_loss`: Calculated as items_wasted * existence price per unit
- `waste_date`: When the waste was discovered/reported
- `reported_by`: Foreign key reference to employee who reported the waste
- `reason`: Detailed explanation of why the waste occurred
- `prevention_notes`: Notes on how to prevent similar waste (optional)

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
UNION ALL SELECT r.id, 'Reports-Read', 'View business reports and analytics', 'Reports', 'Read' FROM roles r WHERE r.role_name = 'admin'
-- Customer Management
UNION ALL SELECT r.id, 'Customers-Create', 'Create new customers', 'Customers', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Customers-Read', 'View customer information', 'Customers', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Customers-Update', 'Update customer information', 'Customers', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Customers-Delete', 'Delete customers', 'Customers', 'Delete' FROM roles r WHERE r.role_name = 'admin'
-- Promotions & Loyalty
UNION ALL SELECT r.id, 'Promotions-Create', 'Create new promotions', 'Promotions', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Promotions-Read', 'View promotion information', 'Promotions', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Promotions-Update', 'Update promotion information', 'Promotions', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Promotions-Delete', 'Delete promotions', 'Promotions', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'CustomerPoints-Create', 'Create customer points', 'CustomerPoints', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'CustomerPoints-Read', 'View customer points', 'CustomerPoints', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'CustomerPoints-Update', 'Update customer points', 'CustomerPoints', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'CustomerPoints-Delete', 'Delete customer points', 'CustomerPoints', 'Delete' FROM roles r WHERE r.role_name = 'admin'
-- Equipment Management
UNION ALL SELECT r.id, 'Equipment-Create', 'Create new equipment', 'Equipment', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Equipment-Read', 'View equipment information', 'Equipment', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Equipment-Update', 'Update equipment information', 'Equipment', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Equipment-Delete', 'Delete equipment', 'Equipment', 'Delete' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Mechanics-Create', 'Create new mechanics', 'Mechanics', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Mechanics-Read', 'View mechanic information', 'Mechanics', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Mechanics-Update', 'Update mechanic information', 'Mechanics', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'Mechanics-Delete', 'Delete mechanics', 'Mechanics', 'Delete' FROM roles r WHERE r.role_name = 'admin'
-- Waste & Loss Tracking
UNION ALL SELECT r.id, 'WasteLoss-Create', 'Create waste loss records', 'WasteLoss', 'Create' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'WasteLoss-Read', 'View waste loss information', 'WasteLoss', 'Read' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'WasteLoss-Update', 'Update waste loss records', 'WasteLoss', 'Update' FROM roles r WHERE r.role_name = 'admin'
UNION ALL SELECT r.id, 'WasteLoss-Delete', 'Delete waste loss records', 'WasteLoss', 'Delete' FROM roles r WHERE r.role_name = 'admin'
-- Audit & Security
UNION ALL SELECT r.id, 'AuditLogs-Read', 'View audit logs', 'AuditLogs', 'Read' FROM roles r WHERE r.role_name = 'admin';

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
UNION ALL SELECT r.id, 'Reports-Read', 'View business reports and analytics', 'Reports', 'Read' FROM roles r WHERE r.role_name = 'employee'
-- Customer Management (limited access)
UNION ALL SELECT r.id, 'Customers-Read', 'View customer information', 'Customers', 'Read' FROM roles r WHERE r.role_name = 'employee'
-- Promotions (read-only to apply during order creation)
UNION ALL SELECT r.id, 'Promotions-Read', 'View promotion information', 'Promotions', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'CustomerPoints-Read', 'View customer points', 'CustomerPoints', 'Read' FROM roles r WHERE r.role_name = 'employee'
-- Equipment (read-only for status checking)
UNION ALL SELECT r.id, 'Equipment-Read', 'View equipment information', 'Equipment', 'Read' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'Mechanics-Read', 'View mechanic information', 'Mechanics', 'Read' FROM roles r WHERE r.role_name = 'employee'
-- Waste & Loss (full access for reporting)
UNION ALL SELECT r.id, 'WasteLoss-Create', 'Create waste loss records', 'WasteLoss', 'Create' FROM roles r WHERE r.role_name = 'employee'
UNION ALL SELECT r.id, 'WasteLoss-Read', 'View waste loss information', 'WasteLoss', 'Read' FROM roles r WHERE r.role_name = 'employee';

---

## Audit & Security Entities

### Audit Logs Table
**Purpose:** Comprehensive audit trail for tracking critical operations, maintaining data integrity, and ensuring regulatory compliance.

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    severity_level VARCHAR(20) NOT NULL CHECK (severity_level IN ('info', 'warning', 'error')) DEFAULT 'info',
    action_type VARCHAR(50) NOT NULL, -- create, update, delete, login, logout, etc.
    entity_type VARCHAR(50) NOT NULL, -- users, orders, inventory, etc.
    entity_id UUID, -- Specific record that was affected
    old_values JSONB, -- Previous values before change (for updates/deletes)
    new_values JSONB, -- New values after change (for creates/updates)
    description TEXT, -- Human-readable description of the action
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    correlation_id UUID, -- For tracking related operations across services
    service_name VARCHAR(50), -- Which service generated this log
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_severity ON audit_logs(severity_level);
CREATE INDEX idx_audit_logs_action ON audit_logs(action_type);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_success ON audit_logs(success);
CREATE INDEX idx_audit_logs_ip ON audit_logs(ip_address);
CREATE INDEX idx_audit_logs_correlation ON audit_logs(correlation_id);
CREATE INDEX idx_audit_logs_service ON audit_logs(service_name);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `user_id`: Foreign key reference to user who performed the action (nullable for system actions)
- `severity_level`: Severity of the audit entry (info, warning, error) - defaults to 'info'
- `action_type`: Type of operation (create, update, delete, login, logout, etc.)
- `entity_type`: Type of entity affected (users, orders, inventory, etc.)
- `entity_id`: Specific record that was affected (nullable for general actions)
- `old_values`: Previous values before change (JSON, for updates/deletes)
- `new_values`: New values after change (JSON, for creates/updates)
- `description`: Human-readable description of the action performed
- `ip_address`: Client IP address where action originated
- `user_agent`: Browser/client information
- `timestamp`: When the action occurred
- `success`: Whether the action was successful
- `error_message`: Error details if action failed (nullable)
- `correlation_id`: UUID for tracking related operations across multiple services
- `service_name`: Name of the service that generated this audit log entry

---

**Database Schema Complete!**
All sections from Requirements.md have been implemented in the database design. 