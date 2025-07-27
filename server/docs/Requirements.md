# Ice Cream Store - Invoice and Inventory Application
## Server-Side Requirements Document

**Ticket ID:** 1  
**Project:** Ice cream store management system  
**Focus:** Server-side implementation for inventory management, invoicing, and business analytics

---

## Table of Contents
1. [Inventory Management](#inventory-management)
   - [Suppliers](#suppliers)
   - [Expense Receipts](#expense-receipts)
   - [Ingredients](#ingredients)
   - [Existences](#existences)
   - [Recipes](#recipes)
2. [Expenses Management](#expenses-management)
3. [Income Managements (Orders)](#income-managements-orders)
4. [Administration Panel](#administration-panel)
   - [Configuration](#configuration)
   - [Ingredients](#ingredients-1)
   - [Stock](#stock)
   - [Employees](#employees)
   - [Business Metrics & Analytics](#business-metrics--analytics)
5. [Authentication & Authorization](#authentication--authorization)
   - [Auth0 Integration](#auth0-integration)
   - [Permission System](#permission-system)
   - [User Access Control](#user-access-control)
6. [Technical Stack](#technical-stack)
   - [Backend Requirements](#backend-requirements)
   - [File Management](#file-management)
   - [Database Design](#database-design)

---

## Inventory Management

### Suppliers
**Description:** Vendor/supplier management system for ingredient procurement and relationship management.

- **Attributes:**
  - **Supplier Name** (string): Name of the supplier/vendor (unique)
  - **Contact Number** (string, nullable): Phone number for supplier contact
  - **Email** (string, nullable): Email address for supplier communication
  - **Address** (text, nullable): Physical address of supplier
  - **Notes** (text, nullable): Additional notes about the supplier

- **Business Logic:**
  - Centralized supplier information management
  - Supports multiple ingredients per supplier
  - Supplier information is optional for local store purchases
  - Used for reordering and supplier relationship management

### Expense Receipts
**Description:** Purchase receipt/invoice management system that tracks purchases from suppliers or supermarkets. Each expense receipt can contain multiple ingredient purchases.

- **Attributes:**
  - **Receipt Number** (string): Receipt/invoice number (unique)
  - **Purchase Date** (date): When the purchase was made
  - **Supplier Reference** (foreign key, nullable): Reference to supplier (nullable for supermarket purchases)
  - **Total Amount** (decimal, nullable): Total amount of the receipt/invoice
  - **Notes** (text, nullable): Additional notes about the purchase

- **Business Logic:**
  - One expense receipt can contain multiple ingredient existences (line items)
  - Centralizes purchase date and supplier information
  - Links to expense management for accounting purposes
  - Supports both supplier and supermarket purchases
  - Provides audit trail for all ingredient acquisitions

### Ingredients
**Description:** Raw materials required to prepare different products in the ice cream store.

- **Attributes:**
  - **Name** (string): Ingredient identifier/name
  - **Description** (text, nullable): Additional details about the ingredient
  - **Purchase Price** (decimal): Cost per unit when purchasing the ingredient (expense)
  - **Unit** (enum): Unit of measurement
    - Liters
    - Gallons  
    - Units
    - Bag
  - **Items per Unit** (integer): Number of individual items that can be produced from one unit
    - Example: 1 Gallon of ice cream → 31 ice cream balls
    - Example: 1 Bag of sugar → 200 servings
  - **Item Cost** (decimal, read-only): Calculated cost per individual item
    - **Formula:** `Item Cost = Purchase Price ÷ Items per Unit`
    - **Purpose:** Used for calculating total product costs and pricing
    - **Purchase Date** (date): When the raw material was acquired
  - **Expiring Date** (date): Expiration date to control material freshness
  - **Supplier Reference** (foreign key, nullable): Reference to supplier in suppliers table
    - Optional for items purchased from local stores

  - **Business Logic:**
  - Each ingredient must have all required fields configured
  - Item Cost is automatically calculated and cannot be manually edited
  - Used as foundation for product cost calculation and profit margin analysis

### Existences
**Description:** Track individual ingredient purchases/acquisitions from suppliers or supermarkets. Each existence represents a specific purchase batch with receipt traceability and expiration tracking.

- **Attributes:**
  - **Existence Reference Code** (integer, auto-increment): Simple numeric consecutive code for easy identification
  - **Ingredient Reference** (foreign key): Link to ingredient
  - **Expense Receipt ID** (foreign key): Reference to expense receipt/invoice table
  - **Units Purchased** (decimal): Original quantity purchased
  - **Units Available** (decimal): Current quantity available (at creation same as units_purchased, decreases as used)
  - **Unit Type** (enum): Unit of measurement for this existence (Liters, Gallons, Units, Bag)
  - **Items per Unit** (integer): Number of individual items produced from one unit (e.g., 1 Gallon = 31 ice cream balls)
  - **Cost per Item** (decimal, read-only): Calculated field (Cost per Unit ÷ Items per Unit) - cost per individual item
  - **Cost per Unit** (decimal): Cost per unit for this specific purchase (e.g., Gallon costs ₡12,000)
  - **Total Purchase Cost** (decimal, read-only): Calculated field (Units Purchased × Cost per Unit)
  - **Remaining Value** (decimal, read-only): Calculated field (Units Available × Cost per Unit)
  - **Expiration Date** (date, nullable): Expiration date for this specific ingredient batch
  - **Income Margin Percentage** (decimal): Configurable margin percentage (default 30%, from config)
  - **Income Margin Amount** (decimal, read-only): Calculated margin amount
  - **IVA Percentage** (decimal): IVA tax percentage (default 13%, from config)
  - **IVA Amount** (decimal, read-only): IVA tax amount (auto-generated)
  - **Service Tax Percentage** (decimal): Service tax percentage (default 10%, from config)
  - **Service Tax Amount** (decimal, read-only): Service tax amount (auto-generated)
  - **Calculated Price** (decimal, read-only): Auto-calculated total price with margins and taxes
  - **Final Price** (decimal): Final price (can be rounded up to next 100)
  
- **Automatic Notifications:**
  - **Low Stock Alerts**: 
    - Triggered when items have only one unit remaining in stock
    - Threshold configurable via system configuration
  - **Expiration Warnings**:
    - Notifications for materials approaching expiration date
    - Warning timeframe configurable (e.g., 7 days, 3 days before expiration)
  
- **Configuration Management:**
  - **Configurable Parameters** (stored in config database table):
    - Minimum stock level threshold for alerts
    - Expiration warning timeframe (days before expiration)
    - Notification frequency settings
  
- **Business Logic:**
  - Multiple existences can exist for the same ingredient (different purchase batches)
  - Track ingredient usage by reducing "Units Available" from specific existences
  - **Runout Reporting Process**: When ingredients run out, employees report usage through runout ingredient reports
    - Employee creates runout report specifying existence, quantity used, and date
    - System automatically updates "Units Available" in existences table based on reported usage
    - Maintains audit trail of who reported usage and when
    - Validates reported quantities against available stock
  - Support FIFO (First In, First Out) consumption by using oldest batches first
  - Prevent usage of expired materials by checking expiration dates at existence level
  - Expense receipt traceability for audit and accounting purposes (links to expense receipt table)
  - Each purchase batch maintains its own cost, pricing, and expiration tracking
  - Pricing calculations (margins, taxes) happen at existence level for inventory items
  - Final pricing can be adjusted from calculated price (rounded up to next 100)
  - Purchase date and supplier information accessed through expense receipt relationship
  - Different ingredients on same receipt can have different expiration dates

### Recipes
**Description:** Product recipes that define combinations of raw materials with specific quantities needed to create finished products.

- **Attributes:**
    - **Recipe Name** (string): Name of the product/recipe
    - **Recipe Description** (string): Description of the product
    - **Price**: Price of the product. This price can be calculated based on costs, margin and taxes.
    - **Picture**: Picture of the product to be use as reference.
    
    - **Recipe Ingredients (List):**
        - **Material Reference** (foreign key): Link to specific raw material/ingredient
        - **Number of Units** (decimal): Quantity of the raw material needed
        - **Cost per Unit**: Individual material cost × number of units needed

        - **Total Recipe Cost**: Sum of all material costs in the recipe  
  
- **Business Logic:**
  - Multiple ingredients can be part of one recipe
  - Total cost automatically calculated based on current ingredient prices
  - Used for cost analysis and recipe planning
  - Supplier information is managed at the ingredient level
  - Pricing (margins, taxes) is calculated at existence level, not recipe level

## Expenses Management
**Description:** Comprehensive expense tracking system that requires digital invoice documentation for all business expenses. The system automatically organizes invoice images by month to facilitate accounting processes and provides clear visibility into operational costs for profit margin analysis.

- **Expense Attributes:**
  - **Expense Categories:**
    - Salary payments
    - Service payments
    - Rent payments
    - Ingredients
    - Other operational expenses
  - **Description:** Brief description of expense.

- **Expense Receipt Attributes:**
  - **Receipt Number:** Unique identifier for the receipt/invoice
  - **Purchase Date:** When the purchase was made
  - **Total Amount:** Monetary amount of the receipt/invoice
  - **Image Upload:** Mandatory receipt/invoice image documentation
  - **Supplier Reference:** Optional reference to supplier (nullable for supermarket purchases)
  
- **Digital Invoice Requirements:**
  - All expense receipts must include digital invoice image upload
  - Supported formats: JPG, PNG, PDF
  - Mandatory documentation for expense validation
  - Each expense receipt is linked to an expense category through the parent expense record
  
- **Automatic File Organization:**
  - **Monthly Directory Creation**: System automatically creates `.../invoices/MM-yyyy` directories
  - **Purpose**: Organized structure enables easy zip file creation for accountant submission
  - **Example Structure**: 
    ```
    invoices/
    ├── 01-2024/
    ├── 02-2024/
    └── 03-2024/
    ```

- **Monthly Expense Management:**
  - **Create Monthly Expense Lists**: Predefined monthly expenses that need to be paid
  - **Track Recurring Expenses**: Identify and manage regular monthly payments
  - **Expense Scheduling**: Set reminders for upcoming payments
  
- **Financial Analysis:**
  - **Total Monthly Expenses**: Calculate sum of all monthly expenses
  - **Expense vs Income Comparison**: Compare total expenses against total income
  - **Income Margin Calculation**: Calculate percentage income margin by month/year
    - **Formula**: `Income Margin % = ((Total Income - Total Expenses) / Total Income) × 100`
  - **Expense Category Breakdown**: Analyze spending patterns by category

- **Accounting Integration:**
  - **Monthly Zip Export**: Generate zip files with all monthly invoices for accountant
  - **Expense Reports**: Generate detailed monthly/yearly expense reports
  - **Audit Trail**: Complete record of all expense transactions with supporting documentation

## Income Managements (Orders)
**Description:** Comprehensive sales tracking system that records all customer transactions with detailed product information, payment methods, and supporting documentation for accurate income analysis.

- **Transaction Attributes:**
  - **Product Details**: Complete list of products sold in each transaction
    - Product name and quantity
    - Recipe price (receipe_price)
    - Product subtotal
  - **Amount Details**:
    - **Total Amount** (decimal): Final transaction total
    - **Individual Recipe Amounts**: Uses pre-calculated prices from recipes (includes cost + margin + taxes)
    - **Tax Breakdown**: Display IVA (13%) and Service Tax (10%) amounts for transparency
      - Taxes are calculated at recipe level, not during order creation
  - **Invoice Integration**:
    - **Invoice Link/Reference**: Connection to generated invoice document
    - **Invoice Number**: Sequential invoice numbering system
  
- **Payment Method Documentation:**
  - **Cash Payments**:
    - Payment method marked as "cash"
    - No additional documentation required
  - **Card Payments**:
    - Payment method marked as "card" 
    - **Transaction Reference Number**: Transaction reference from card terminal
  - **Sinpe Payments**:
    - Payment method marked as "sinpe"
    - **Screenshot Upload**: Mandatory screenshot of Sinpe transaction
    - **Transaction Reference**: Sinpe transaction ID

- **Transaction Timestamp**: Date and time of sale
- **Customer Information** (optional): Basic customer details if provided
- **Sales Representative**: Employee who processed the sale
- **Transaction Status**: Completed, pending, cancelled

* Orders are created with pending status, it can be cancelled and nothing happens, but once payed the order should be set as completed and an invoice should be created.


## Administration Panel
### Configuration
- System-wide settings
- Business parameters configuration
- Pricing margin settings (configurable ~30%)

### Ingredients
- CRUD operations for ingredients
- Price management per unit type
- Ingredient categorization

### Stock
- Stock level monitoring
- Stock adjustment capabilities
- Low stock alert configuration

### Employees
- **User Roles:**
  - Admin (full access)
  - Employee (restricted access)
- Employee account management
- Role-based permissions

### Business Metrics & Analytics
#### Financial Reports
- **Expenses/Income Analysis:**
  - Daily reports
  - Monthly reports
  - Yearly reports
- **Payment Method Breakdown:**
  - Cash transactions
  - Card transactions
  - Sinpe transactions

#### Product Analytics
- **Sales Performance:**
  - Most selling products identification
  - Product popularity trends
- **Profit Margin Analysis:**
  - Cost calculation per product (sum of ingredient costs)
  - Profit margin calculation (cost + configurable margin ~30%)
  - Suggested pricing recommendations

#### Business Projections
- **Monthly Projections:**
  - Income goals vs actual income
  - Expense-based income requirements
  - Break-even analysis
  - Profitability forecasting

---

## Authentication & Authorization

### Auth0 Integration
- **Implementation Requirements:**
  - Auth0 account setup and configuration
  - JWT token-based authentication
  - Permission claims management through Auth0 Actions

### Permission System
- **Server-Side Validation:**
  - JWT token verification on all API endpoints. We can use middleware to implement jtw verification
  - Permission-based access control. Permissions would be named after [entity]-[action]. For instance "Ingredients-Create"
  - Role-based endpoint restrictions
- **Client-Side Validation:**
  - Token validation
  - UI component access control based on permissions
  - UI client side will be worked in incoming documentation once server-side is implemented

### User Access Control
- **Employee Restrictions:**
  - Limited access to sensitive financial data
  - Restricted administrative functions
  - Read-only access to certain reports
- **Admin Privileges:**
  - Full system access
  - User management capabilities
  - Financial data access

---

## Technical Stack

### Backend Requirements
- Go-based API server
- RESTful API design
- JWT authentication middleware
- File upload handling for invoices
- Database integration for data persistence
- Auth0 SDK integration

### File Management
- Monthly directory structure for invoice storage
- File upload validation and security
- Storage optimization and organization

### Database Design
- Relational database structure
- Inventory tracking tables
- Financial transaction records
- User and permission management
- Audit trails for critical operations

---

## Next Steps
This document will be reviewed and expanded with additional details and feedback. Areas requiring further specification:
- Detailed API endpoint specifications
- Database schema design
- File storage architecture
- Security protocols and validation rules
- Error handling and logging requirements 