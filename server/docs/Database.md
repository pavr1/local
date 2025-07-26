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

## Relationships
- **suppliers** ← **expense_receipts** (One-to-Many: One supplier can have multiple expense receipts)
- **suppliers** ← **ingredients** (One-to-Many: One supplier can provide multiple ingredients)
- **expense_receipts** ← **existences** (One-to-Many: One expense receipt can contain multiple existences/line items)
- **ingredients** ← **existences** (One-to-Many: One ingredient can have multiple purchase batches/existences)
- **recipes** ← **recipe_ingredients** → **ingredients** (Many-to-Many: Recipes contain multiple ingredients, ingredients can be in multiple recipes)

## Business Logic Triggers
- Update `total_recipe_cost` in recipes table when recipe_ingredients change
- Recalculate pricing fields in existences table when cost components change
- Track ingredient consumption by updating `units_available` in existences table
- Implement FIFO logic: use oldest expense receipts first (by purchase_date from expense_receipts table)
- Alert when existences are near expiry (based on expiration_date from existences table)
- Calculate final pricing (margins, taxes) at existence level for inventory items
- Maintain expense receipt totals when existences are added/modified
- Link expense receipts to expense management system for accounting integration

---

**Next Sections to Review:**
- Expenses Management Entities
- Income Management (Orders) Entities  
- Administration Panel Entities
- Authentication & Authorization Entities 