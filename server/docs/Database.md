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
   - [Recipes Table](#recipes-table)
   - [Recipe Ingredients Table](#recipe-ingredients-table)
2. [Expenses Management Entities](#expenses-management-entities)
   - [Expense Categories Table](#expense-categories-table)
   - [Expenses Table](#expenses-table)
3. [Income Management (Orders) Entities](#income-management-orders-entities)
   - [Orders Table](#orders-table)
   - [Order Items Table](#order-items-table)

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
**Purpose:** Store purchase receipt/invoice information from suppliers or supermarkets. Each expense receipt can contain multiple ingredient purchases (existences).

```sql
CREATE TABLE expense_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_number VARCHAR(100) UNIQUE NOT NULL,
    purchase_date DATE NOT NULL,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    total_amount DECIMAL(12,2), -- get all existences for that recipt number to get total amount
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_expense_receipts_number ON expense_receipts(receipt_number);
CREATE INDEX idx_expense_receipts_supplier ON expense_receipts(supplier_id);
CREATE INDEX idx_expense_receipts_purchase_date ON expense_receipts(purchase_date);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `receipt_number`: Receipt/invoice number (unique)
- `purchase_date`: When the purchase was made
- `supplier_id`: Foreign key reference to suppliers table (UUID, nullable for supermarket purchases)
- `total_amount`: Total amount of the expense receipt/invoice
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
- **recipes** ← **recipe_ingredients** → **ingredients** (Many-to-Many: Recipes contain multiple ingredients, ingredients can be in multiple recipes)

### Expenses Management
- **expense_categories** ← **expenses** (One-to-Many: One category can have multiple expenses)

## Overall Business Logic Triggers

### Inventory Management
- Update `total_recipe_cost` in recipes table when recipe_ingredients change
- Recalculate pricing fields in existences table when cost components change
- Track ingredient consumption by updating `units_available` in existences table
- Implement FIFO logic: use oldest expense receipts first (by purchase_date from expense_receipts table)
- Alert when existences are near expiry (based on expiration_date from existences table)
- Calculate final pricing (margins, taxes) at existence level for inventory items
- Maintain expense receipt totals when existences are added/modified
- Link expense receipts to expense management system for accounting integration

### Expenses Management
- Validate `payment_receipt_url` is required when `is_paid` = true
- Generate monthly expense records when `is_recurring` = true
- Organize invoice documents in monthly directories (MM-yyyy format)
- Calculate monthly expense totals for financial analysis
- Validate timeline format matches MM/yyyy pattern

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
('Raw material purchases', 'Ingredient and supply purchases'),
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
**Purpose:** Track all business expenses with complete documentation, categorization, and payment scheduling.

```sql
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    description TEXT NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    timeline VARCHAR(7) NOT NULL, -- MM/yyyy format
    expense_payment_date DATE,
    is_paid BOOLEAN DEFAULT FALSE,
    payment_receipt_url VARCHAR(500),
    invoice_document_url VARCHAR(500) NOT NULL,
    is_recurring BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_timeline_format CHECK (timeline ~ '^[0-1][0-9]/[0-9]{4}$')
);

-- Indexes
CREATE INDEX idx_expenses_category ON expenses(expense_category_id);
CREATE INDEX idx_expenses_timeline ON expenses(timeline);
CREATE INDEX idx_expenses_payment_date ON expenses(expense_payment_date);
CREATE INDEX idx_expenses_is_paid ON expenses(is_paid);
CREATE INDEX idx_expenses_is_recurring ON expenses(is_recurring);
CREATE INDEX idx_expenses_amount ON expenses(amount);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `expense_category_id`: Foreign key reference to expense_categories table
- `description`: Brief description of the expense
- `amount`: Monetary amount of the expense
- `timeline`: Month/year the expense is valid for (MM/yyyy format)
- `expense_payment_date`: Scheduled payment date for recurring expenses (nullable)
- `is_paid`: Boolean indicating if expense has been paid for the timeline
- `payment_receipt_url`: URL/path to payment receipt screenshot (required when is_paid = true)
- `invoice_document_url`: URL/path to uploaded invoice image (mandatory)
- `is_recurring`: Whether this expense recurs monthly (creates new records)
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

### Order Items Table
**Purpose:** Track individual products sold in each order with quantities and pricing snapshots.

```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL, -- Snapshot of recipe name at time of sale
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL, -- Snapshot of recipe price at time of sale
    subtotal DECIMAL(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_recipe ON order_items(recipe_id);
CREATE INDEX idx_order_items_product_name ON order_items(product_name);
```

**Field Descriptions:**
- `id`: Primary key, UUID (auto-generated)
- `order_id`: Foreign key reference to orders table
- `recipe_id`: Foreign key reference to recipes table
- `product_name`: Snapshot of product name at time of sale (for historical accuracy)
- `quantity`: Number of items ordered
- `unit_price`: Snapshot of recipe price at time of sale (for historical accuracy)
- `subtotal`: Calculated subtotal for this line item (quantity × unit_price)

---

**Next Sections to Review:**
- Administration Panel Entities
- Authentication & Authorization Entities 