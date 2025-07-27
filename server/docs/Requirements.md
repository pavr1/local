# Ice Cream Store - Invoice and Inventory Application
## Server-Side Requirements Document

**Ticket ID:** 1  
**Project:** Ice cream store management system  
**Focus:** Server-side implementation for inventory management, invoicing, and business analytics

---

## Table of Contents
1. [Inventory Management](#inventory-management)
   - [Suppliers](#suppliers)
   - [Ingredients](#ingredients)
   - [Existences](#existences)
   - [Recipe Categories](#recipe-categories)
   - [Recipes](#recipes)
2. [Expenses Management](#expenses-management)
   - [Expense Receipts](#expense-receipts)
3. [Customer Management](#customer-management)
4. [Income Managements (Orders)](#income-managements-orders)
5. [Promotions & Loyalty System](#promotions--loyalty-system)
   - [Promotions](#promotions)
   - [Customer Points System](#customer-points-system)
6. [Equipment Management](#equipment-management)
   - [Equipment Tracking](#equipment-tracking)
   - [Mechanic Management](#mechanic-management)
7. [Waste & Loss Tracking](#waste--loss-tracking)
8. [Administration Panel](#administration-panel)
   - [Configuration](#configuration)
   - [Ingredients](#ingredients-1)
   - [Stock](#stock)
   - [Employees](#employees)
   - [Business Metrics & Analytics](#business-metrics--analytics)
9. [Authentication & Authorization](#authentication--authorization)
   - [Internal Authentication](#internal-authentication)
   - [Permission System](#permission-system)
   - [User Access Control](#user-access-control)
10. [Audit & Security](#audit--security)
11. [Technical Stack](#technical-stack)
    - [Backend Requirements](#backend-requirements)
    - [File Management](#file-management)
    - [Database Design](#database-design)
    - [Future Enhancements](#future-enhancements)

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

### Recipe Categories
**Description:** Product categorization system to organize recipes by type for better management and customer browsing.

- **Attributes:**
    - **Category ID** (UUID): Unique identifier for each recipe category
    - **Name** (string): Category name (unique)
    - **Description** (text): Description of the category type

- **Predefined Categories:**
    - **Postres**: Desserts and sweet treats
    - **Helados**: Traditional ice cream products
    - **Batidos**: Milkshakes and blended drinks
    - **Gelato**: Italian-style gelato products
    - **Artesanales**: Artisan and handcrafted specialty items

- **Business Logic:**
    - Each recipe must belong to exactly one category
    - Categories help organize menu displays and reporting
    - Used for filtering and searching recipes in administration
    - Supports promotional campaigns targeting specific product categories

### Recipes
**Description:** Product recipes that define combinations of raw materials with specific quantities needed to create finished products.

- **Attributes:**
    - **Recipe Name** (string): Name of the product/recipe
    - **Recipe Description** (string): Description of the product
    - **Recipe Category ID** (foreign key): Link to recipe category (required)
    - **Price**: Price of the product. This price can be calculated based on costs, margin and taxes.
    - **Picture**: Picture of the product to be use as reference.
    
    - **Recipe Ingredients (List):**
        - **Material Reference** (foreign key): Link to specific raw material/ingredient
        - **Number of Units** (decimal): Quantity of the raw material needed
        - **Cost per Unit**: Individual material cost × number of units needed

        - **Total Recipe Cost**: Sum of all material costs in the recipe  
  
- **Business Logic:**
  - Multiple ingredients can be part of one recipe
  - Each recipe must be assigned to a valid category
  - Total cost automatically calculated based on current ingredient prices
  - Used for cost analysis and recipe planning
  - Supplier information is managed at the ingredient level
  - Pricing (margins, taxes) is calculated at existence level, not recipe level
  - Category assignment enables better organization and promotional targeting

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

## Customer Management
**Description:** Customer relationship management system for tracking customer information, enabling targeted marketing, and supporting loyalty programs.

- **Customer Attributes:**
  - **Customer ID** (UUID): Unique identifier for each customer
  - **Name** (string): Customer's full name (required)
  - **Phone** (string, nullable): Customer's phone number for contact and promotions
  - **Email** (string, nullable): Customer's email address for marketing and notifications

- **Business Logic:**
  - Customer information is optional during order creation (walk-in customers)
  - Phone and email are collected for future marketing campaigns and promotions
  - Supports customer loyalty point accumulation when linked to orders
  - Enables customer purchase history tracking and analytics
  - Customer data can be used for targeted promotions and communication

- **Privacy and Marketing:**
  - Optional customer data collection respects privacy preferences
  - Email and phone used for promotional campaigns and special offers
  - Customer consent tracked for marketing communications
  - Supports customer segmentation for targeted advertising

## Income Managements (Orders)
**Description:** Comprehensive sales tracking system that records all customer transactions with detailed product information, payment methods, and supporting documentation for accurate income analysis.

- **Transaction Attributes:**
  - **Product Details**: Complete list of products sold in each transaction (snapshots taken at order creation time)
    - Product name and quantity
    - Recipe price (receipe_price) - captured when order is created
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

- **Transaction Timestamp**: Date and time of sale (tracked at order level)
- **Customer Information** (optional): Link to customer record if provided for loyalty points and marketing
- **Sales Representative**: Employee who processed the sale
- **Transaction Status**: Completed, pending, cancelled

- **Promotional Integration:**
  - **Discount Application**: System automatically applies available promotions during order creation
  - **Customer Loyalty Points**: Points awarded to customers based on active promotions
  - **Promotional Validation**: Check time-based, recipe-specific, and customer point promotions
  - **Discount Calculation**: Apply percentage discounts or point redemptions to order total

**Note**: All timing information (creation, updates, completion) is tracked at the order level. Individual product line items (ordered recipes) do not have separate timestamps as they are part of the same transaction and inherit the order's timing.

* Orders are created with pending status, it can be cancelled and nothing happens, but once payed the order should be set as completed and an invoice should be created.
* During order creation, system checks for available promotions and applies them automatically
* Customer points are awarded upon order completion if customer is linked to the order

## Promotions & Loyalty System
**Description:** Comprehensive promotional and customer loyalty system that enables targeted discounts, time-based promotions, and points-based rewards to drive customer engagement and repeat business.

### Promotions
**Description:** Flexible promotional system supporting multiple types of discounts and incentives.

- **Promotion Attributes:**
  - **Promotion ID** (UUID): Unique identifier for each promotion
  - **Name** (string): Promotional campaign name
  - **Description** (text): Detailed description of the promotion
  - **Recipe ID** (foreign key, nullable): Specific recipe/product the promotion applies to (null = applies to all)
  - **Time From** (datetime, not-null): Start date/time for time-limited promotions
  - **Time To** (datetime, nullable): End date/time for time-limited promotions
  - **Promotion Type** (enum): Type of promotion (percentage_discount, points_reward)
  - **Value** (decimal): Promotion value (percentage for discounts, points awarded for loyalty)
  - **Minimum Purchase Amount** (decimal, nullable): For points_reward only - minimum purchase amount required to qualify for points (e.g., 5000 colones)
  - **Points Expiration Duration** (string, nullable): For points_reward only - expiration format: 1d/3w/7m/2y (null = no expiration)
  - **Is Active** (boolean): Whether the promotion is currently active

- **Promotion Types:**
  - **Recipe-Specific Discounts**: Percentage discounts applied to specific products
  - **Time-Based Promotions**: Limited-time offers with start and end dates
  - **Customer Points Rewards**: Points awarded to customers for loyalty program (with configurable minimum purchase conditions and expiration rules)
  - **General Discounts**: Store-wide percentage discounts (no recipe specified)

- **Business Logic:**
  - Promotions are automatically validated during order creation
  - Time-based promotions only apply within specified date ranges
  - Recipe-specific promotions only apply to designated products
  - For points_reward promotions: validate minimum purchase amount condition before awarding points
  - Points expiration calculated automatically based on promotion's expiration duration format
  - Multiple promotions can be active simultaneously (admin configurable stacking rules)
  - Only active promotions are considered during order processing

### Customer Points System
**Description:** Customer loyalty program that awards points for purchases and enables point-based promotions.

- **Customer Points Attributes:**
  - **Points Record ID** (UUID): Unique identifier for each points transaction
  - **Customer ID** (foreign key): Reference to customer who earned/spent points
  - **Points Earned** (integer): Number of points added to customer account
  - **Points Source** (enum): How points were earned (purchase, promotion_bonus, manual_adjustment)
  - **Order ID** (foreign key, nullable): Order that generated the points (if applicable)
  - **Date Earned** (date): When the points were awarded
  - **Expiration Date** (date, nullable): When points expire (calculated based on promotion's expiration duration)

- **Points Management:**
  - **Points Accumulation**: Customers earn points from qualifying orders and promotions
  - **Points Redemption**: Points can be used for discounts on future orders
  - **Points Balance**: Real-time calculation of available points per customer
  - **Points History**: Complete audit trail of points earned and spent
  - **Points Expiration**: Automatic expiration calculation based on promotion's duration format (1d/3w/7m/2y)

- **Integration with Orders:**
  - Points automatically awarded upon order completion
  - Available points displayed during order creation for redemption
  - Point redemption applies as discount to order total
  - Points conversion rate configurable (e.g., 100 points = ₡1000 discount)

## Equipment Management
**Description:** Comprehensive equipment and asset management system for tracking store equipment, maintenance schedules, and mechanic relationships.

### Equipment Tracking
**Description:** Track all store equipment with maintenance scheduling and cost management.

- **Equipment Attributes:**
  - **Equipment ID** (UUID): Unique identifier for each equipment item
  - **Name** (string): Equipment name/model
  - **Description** (text): Detailed description of the equipment
  - **Purchase Date** (date): When the equipment was purchased
  - **Mechanic ID** (foreign key): Reference to assigned mechanic for maintenance
  - **Maintenance Schedule** (integer): Days between scheduled maintenance (e.g., 90 days)
  - **Purchase Cost** (decimal): Original purchase cost of equipment
  - **Current Status** (enum): Equipment status (operational, maintenance_required, out_of_service, retired)

- **Maintenance Management:**
  - **Scheduled Maintenance**: Automatic alerts based on maintenance schedule
  - **Maintenance History**: Track all maintenance performed on equipment
  - **Cost Tracking**: Monitor maintenance costs and equipment total cost of ownership
  - **Downtime Tracking**: Track equipment downtime for operational analysis

### Mechanic Management
**Description:** Contact management for equipment maintenance professionals.

- **Mechanic Attributes:**
  - **Mechanic ID** (UUID): Unique identifier for each mechanic
  - **Name** (string): Mechanic or company name
  - **Email** (string, nullable): Email contact for scheduling and communication
  - **Phone** (string): Primary phone contact for emergency repairs
  - **Specialization** (text, nullable): Equipment types or brands they specialize in
  - **Notes** (text, nullable): Additional notes about the mechanic

- **Business Logic:**
  - Equipment can be assigned to specific mechanics for consistency
  - Mechanic contact information used for maintenance scheduling
  - Supports multiple mechanics for different equipment types
  - Emergency contact capability for urgent equipment failures

## Waste & Loss Tracking
**Description:** Comprehensive waste and loss tracking system to monitor expired ingredients, calculate financial losses, and improve inventory management efficiency.

- **Waste/Loss Attributes:**
  - **Waste Record ID** (UUID): Unique identifier for each waste incident
  - **Existence ID** (foreign key): Reference to specific existence/batch that was wasted
  - **Waste Type** (enum): Type of waste (expired, damaged, spoiled, theft, other)
  - **Items Wasted** (decimal): Amount of items in a unit that were wasted
  - **Unit Type** (enum): Unit of measurement (Liters, Gallons, Units, Bag)
  - **Financial Loss** (decimal): Calculated as items_wasted × existence price per unit
  - **Waste Date** (date): When the waste was discovered/reported
  - **Reported By** (foreign key): Employee who reported the waste
  - **Reason** (text): Detailed explanation of why the waste occurred
  - **Prevention Notes** (text, nullable): Notes on how to prevent similar waste

- **Financial Impact Calculation:**
  - **Cost per Unit**: Retrieved from existence record (original purchase cost)
  - **Total Loss Value**: Items Wasted × Cost per Unit
  - **Percentage Loss**: (Items Wasted ÷ Original Purchase Quantity) × 100
  - **Monthly Waste Reports**: Aggregate waste costs by category and time period

- **Business Logic:**
  - Waste tracking helps identify patterns in ingredient loss
  - Expired ingredient alerts can prevent waste through early consumption
  - Waste cost analysis supports better purchasing decisions
  - Employee training opportunities identified through waste pattern analysis
  - Integration with inventory to automatically update available quantities
  - **Automatic Inventory Updates**: When waste is reported, system automatically decreases `units_available` in existences table by items_wasted amount
  - **Validation Logic**: System validates that waste amounts do not exceed available existence quantities

- **Waste Prevention:**
  - **Expiration Monitoring**: Automated alerts for approaching expiration dates
  - **FIFO Enforcement**: First-in-first-out consumption to minimize expiration waste
  - **Portion Control**: Tracking helps identify over-portioning issues
  - **Storage Optimization**: Waste patterns inform better storage practices

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
  - User identification by username (not email address)
  - Internal password-based authentication
  - Secure password hashing and validation
- Employee salary management
  - Track individual employee salaries and compensation
  - Link salary records to expense management system
  - Support for additional expenses and bonuses
  - Automatic total compensation calculation
  - Admin-only access to salary information
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

### Internal Authentication
- **Implementation Requirements:**
  - Username and password-based authentication
  - Secure password hashing (bcrypt or similar)
  - Session management for logged-in users
  - Login/logout functionality
  - Password strength requirements
  - Default admin user creation for initial system setup
  - Password change functionality for users

- **Authentication Workflow:**
  1. User submits username/password via login form
  2. Server validates credentials against users table
  3. On success, create authenticated session
  4. Session used for subsequent API requests
  5. Permissions checked against user's roles and permissions tables

### Permission System
- **Simplified Role-Permission Structure:**
  - Permissions are directly assigned to roles (no junction table required)
  - Each permission belongs to one role and is identified by role-entity-action pattern
  - Permission naming follows [entity]-[action] format (e.g., "Ingredients-Create")
- **Server-Side Validation:**
  - Session-based authentication on all API endpoints with middleware for session verification
  - Permission-based access control with direct role lookup
  - Role-based endpoint restrictions with simplified permission checking
- **Client-Side Validation:**
  - Session validation
  - UI component access control based on user's role permissions
  - Login form with username/password
  - UI client side will be worked in incoming documentation once server-side is implemented

### User Access Control
- **Direct Role Assignment:**
  - Each user is assigned one role (admin or employee)
  - Permissions are directly linked to roles for simplified access control
  - No complex many-to-many relationships required
- **Employee Restrictions:**
  - Limited access to sensitive financial data (no salary information access)
  - Restricted administrative functions
  - Read-only access to inventory information
  - Full access to runout reporting and order management
  - Limited reporting access (no sensitive financial data)
- **Admin Privileges:**
  - Full system access with all permissions
  - User and salary management capabilities
  - Complete financial data access
  - System configuration control

---

## Audit & Security
**Description:** Comprehensive audit trail and security monitoring system to track critical operations, maintain data integrity, and ensure regulatory compliance.

- **Audit Log Attributes:**
  - **Audit Log ID** (UUID): Unique identifier for each audit record
  - **User ID** (foreign key): Reference to user who performed the action
  - **Action Type** (enum): Type of operation (create, update, delete, login, logout, etc.)
  - **Entity Type** (string): Type of entity affected (users, orders, inventory, etc.)
  - **Entity ID** (UUID): Specific record that was affected
  - **Old Values** (JSON, nullable): Previous values before change (for updates/deletes)
  - **New Values** (JSON, nullable): New values after change (for creates/updates)
  - **IP Address** (string): Client IP address where action originated
  - **User Agent** (string): Browser/client information
  - **Timestamp** (datetime): When the action occurred
  - **Success** (boolean): Whether the action was successful
  - **Error Message** (text, nullable): Error details if action failed

- **Critical Operations to Audit:**
  - **User Management**: User creation, deletion, password changes, role modifications
  - **Financial Data**: Order creation/modification, pricing changes, payment processing
  - **Inventory Management**: Ingredient additions/modifications, existence updates, waste reporting
  - **System Configuration**: Config changes, promotion creation/modification
  - **Authentication**: Login attempts, logout, session timeouts, failed authentications
  - **Sensitive Data Access**: Salary information access, financial reports generation

- **Security Monitoring:**
  - **Failed Login Attempts**: Track and alert on suspicious login patterns
  - **Unusual Activity**: Monitor for abnormal data access or modification patterns
  - **Data Integrity**: Verify critical calculations and data consistency
  - **Session Management**: Track active sessions and detect concurrent logins
  - **Privilege Escalation**: Monitor for unauthorized access attempts

- **Compliance and Reporting:**
  - **Audit Trail Reports**: Generate audit reports for specific time periods or users
  - **Data Retention**: Configurable audit log retention periods
  - **Search and Filter**: Ability to search audit logs by user, action, entity, or time range
  - **Export Capabilities**: Export audit logs for external compliance reviews
  - **Alert System**: Automated alerts for critical security events

- **Data Protection:**
  - **Sensitive Data Masking**: Mask sensitive information in audit logs (passwords, payment details)
  - **Tamper Protection**: Ensure audit logs cannot be modified or deleted by users
  - **Backup Integration**: Include audit logs in regular backup procedures
  - **Access Control**: Restrict audit log access to authorized administrators only

---

## Technical Stack

### Backend Requirements
- Go-based API server
- RESTful API design
- Session-based authentication middleware
- Secure password hashing (bcrypt)
- File upload handling for invoices
- Database integration for data persistence

### File Management
- Monthly directory structure for invoice storage
- File upload validation and security
- Storage optimization and organization

### Database Design
- Relational database structure
- Inventory tracking tables
- Financial transaction records
- Customer management and loyalty system
- Promotions and discount management
- Equipment and mechanic tracking
- Waste and loss monitoring
- Authentication & authorization entities (users, roles, permissions)
- Simplified role-permission management with direct foreign key relationships
- Employee salary tracking linked to expense management
- Comprehensive audit trails for critical operations

### Future Enhancements
- **Automated Backup Strategy**:
  - Scheduled database backups with configurable retention periods
  - Automated backup verification and restoration testing
  - Cloud backup integration for disaster recovery
  - Point-in-time recovery capabilities
  - Backup monitoring and alert system
- **Advanced Analytics Dashboard**:
  - Real-time business intelligence and reporting
  - Predictive analytics for inventory management
  - Customer behavior analysis and segmentation
- **Mobile Application Support**:
  - Mobile-optimized API endpoints
  - Employee mobile app for inventory management
  - Customer mobile app for loyalty program
- **Integration Capabilities**:
  - Accounting software integration (QuickBooks, etc.)
  - Payment processor APIs
  - Email marketing platform integration
  - SMS notification services

---

## Next Steps
This document will be reviewed and expanded with additional details and feedback. Areas requiring further specification:
- Detailed API endpoint specifications
- Database schema design for new entities
- File storage architecture
- Security protocols and validation rules
- Error handling and logging requirements
- Customer data privacy and GDPR compliance
- Promotional campaign management workflows
- Equipment maintenance scheduling algorithms 