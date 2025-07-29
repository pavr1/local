# Ice Cream Store - Service Architecture Document

**Project:** Ice cream store management system  
**Focus:** Microservices architecture with proper separation of concerns  
**Last Updated:** $(date +'%Y-%m-%d')

---

## 🏗️ System Overview

The Ice Cream Store management system implements a **microservices architecture** with **9 specialized services**, each handling specific business domains with clear boundaries and dependencies.

### 🎯 Design Principles

- **Single Responsibility**: Each service handles one business domain
- **Separation of Concerns**: Authentication vs. Administration clearly separated  
- **Dependency Management**: Services ordered by dependencies to minimize coupling
- **Security First**: Centralized authentication with distributed authorization
- **Scalability**: Independent deployment and scaling per service

---

## 📊 Service Dependency Hierarchy

```mermaid
graph TD
    %% Level 0 - Foundation
    A1[🔐 Session Service<br/>Security & Session Management] --> B1[Level 1 Services]
    A2[📋 Audit Service<br/>LogAuditEntry() & RetrieveAuditLogs()<br/>Activity Logging] --> B1

    %% Level 1 - Administrative
    B1 --> B2[⚙️ Administration Service<br/>👑 Admin Only<br/>User/Role/Permission CRUD<br/>Equipment Management]
    B1 --> B3[👥 Customer Service<br/>Customer Management]
    B1 --> B4[💰 Expenses Service<br/>Financial Management]

    %% Level 2 - Business Logic
    B2 --> C1[📦 Inventory Service<br/>Core Business Logic]
    B4 --> C1

    %% Level 3 - Advanced Logic
    C1 --> D1[🎉 Promotions Service<br/>Loyalty & Discounts]
    B3 --> D1

    %% Level 4 - Operations
    D1 --> E1[🛒 Orders Service<br/>Sales & Invoicing]
    B3 --> E1
    C1 --> E1

    %% Level 5 - Analytics
    C1 --> F1[🗑️ Waste Service<br/>Loss Analysis]

    classDef foundation fill:#ffebee,stroke:#d32f2f,stroke-width:3px
    classDef admin fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef business fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef advanced fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    classDef operations fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    classDef analytics fill:#fce4ec,stroke:#c2185b,stroke-width:2px

    class A1,A2 foundation
    class B2,B3,B4 admin
    class C1 business
    class D1 advanced
    class E1 operations
    class F1 analytics
```

---

## 🔐 Authentication vs Administration Separation

### **Session Service** 🔐
**Purpose**: Security and session management **ONLY**

- **Core Responsibilities**:
  - User credential validation (login/logout)
  - JWT token generation, validation, and refresh
  - Password hashing and verification
  - Session management and tracking
  - Security middleware for route protection
  - Authorization checks (validates JWT tokens)

- **What it DOESN'T do**:
  - ❌ User creation, modification, or deletion
  - ❌ Role management or assignment
  - ❌ Permission management
  - ❌ User profile management

- **Database Tables**: None (reads from Administration Service)

- **API Integration**: 
  - Calls Administration Service for user/role/permission data
  - Provides authentication middleware for all other services

### **Administration Service** ⚙️
**Purpose**: User, role, and system management with **ADMIN-ONLY** access

- **Core Responsibilities**:
  - **User Management**: Complete CRUD operations
  - **Role Management**: Role creation, modification, deletion
  - **Permission Management**: Permission assignment and validation
  - **System Configuration**: Global settings and business parameters
  - **Salary Management**: Employee payroll and compensation

- **Security Model**: 
  - 🔒 **ALL operations require admin authorization**
  - Validates admin permissions before any CRUD operation
  - Provides user/role/permission data to other services

- **Database Tables**: `users`, `roles`, `permissions`, `system_config`, `user_salary`

- **Authorization Flow**:
  ```
  Request → Session Service (validates JWT) → Administration Service (checks admin permissions) → Execute operation
  ```

---

## 📋 Complete Service Breakdown

### **🟥 Level 0: Foundation Services (No Dependencies)**

#### 1. **Session Service** 🔐
- **Port**: 8081
- **Tables**: None (reads from Administration Service)
- **Functions**:
  - Login/logout endpoint (`POST /api/v1/auth/login`)
  - JWT token management (`POST /api/v1/auth/refresh`)
  - Token validation middleware
  - Password operations
  - Session tracking

#### 2. **Audit Service** 📋
- **Tables**: `audit_logs`
- **Core API Methods**:
  - **`LogAuditEntry()`**: Insert audit records with severity levels (info, warning, error)
  - **`RetrieveAuditLogs()`**: Query audit data with flexible filtering capabilities
- **Functions**:
  - **Audit Logging**:
    - Operation logging (CREATE, READ, UPDATE, DELETE)
    - Security event monitoring with severity classification
    - Login/logout tracking with IP and user agent
    - Failed operation logging with error details
    - Cross-service correlation tracking
  - **Audit Retrieval**:
    - Multi-criteria filtering (user, date range, severity, entity type, etc.)
    - Pagination and sorting capabilities
    - Export functionality for compliance reporting
    - Real-time audit monitoring

### **🔍 Audit Service API Specification**

#### **LogAuditEntry() Method**
```go
func LogAuditEntry(
    userID *uuid.UUID,           // User who performed action (nullable for system)
    severity string,             // "info", "warning", "error"
    actionType string,           // "create", "update", "delete", "login", etc.
    entityType string,           // "users", "orders", "inventory", etc.
    entityID *uuid.UUID,         // Specific record affected (nullable)
    description string,          // Human-readable description
    oldValues map[string]interface{}, // Previous values (nullable)
    newValues map[string]interface{}, // New values (nullable)
    ipAddress string,            // Client IP address
    userAgent string,            // Browser/client info
    correlationID *uuid.UUID,    // For tracking related operations
    serviceName string           // Service generating the log
) error
```

#### **RetrieveAuditLogs() Method with Variadic Filtering**
```go
type AuditFilter struct {
    UserIDs        []uuid.UUID
    SeverityLevels []string    // ["info", "warning", "error"]
    ActionTypes    []string
    EntityTypes    []string
    EntityIDs      []uuid.UUID
    DateFrom       *time.Time
    DateTo         *time.Time
    IPAddresses    []string
    ServiceNames   []string
    CorrelationIDs []uuid.UUID
    SuccessOnly    *bool       // Filter by success/failure
    Limit          int         // Pagination
    Offset         int         // Pagination
    SortBy         string      // "timestamp", "severity", "user_id"
    SortOrder      string      // "asc", "desc"
}

func RetrieveAuditLogs(filters ...AuditFilter) ([]AuditLog, int, error)
```

### **🔄 Audit Service Integration Patterns**

#### **Usage Examples:**

**1. Logging a User Creation (Info Level):**
```go
auditService.LogAuditEntry(
    &adminUserID,           // Admin who created the user
    "info",                 // Severity level
    "create",               // Action type
    "users",                // Entity type
    &newUserID,             // Newly created user ID
    "Created new employee user account", // Description
    nil,                    // No old values for creation
    userDataMap,            // New user data
    clientIP,               // Client IP
    userAgent,              // Browser info
    &correlationID,         // Request correlation ID
    "administration-service" // Service name
)
```

**2. Logging a Failed Login (Error Level):**
```go
auditService.LogAuditEntry(
    nil,                    // No user ID (failed login)
    "error",                // Severity level
    "login",                // Action type
    "authentication",       // Entity type
    nil,                    // No specific entity
    "Failed login attempt for username: invalid_user", // Description
    nil,                    // No old values
    map[string]interface{}{"username": "invalid_user"}, // Attempted username
    clientIP,               // Client IP
    userAgent,              // Browser info
    &correlationID,         // Request correlation ID
    "session-service" // Service name
)
```

**3. Retrieving Audit Logs with Multiple Filters:**
```go
// Get all error-level audit logs from last 24 hours
logs, count, err := auditService.RetrieveAuditLogs(AuditFilter{
    SeverityLevels: []string{"error"},
    DateFrom:       &yesterday,
    DateTo:         &now,
    Limit:          50,
    SortBy:         "timestamp",
    SortOrder:      "desc",
})

// Get all operations for specific user
userLogs, _, err := auditService.RetrieveAuditLogs(AuditFilter{
    UserIDs:   []uuid.UUID{userID},
    DateFrom:  &startOfMonth,
    Limit:     100,
})
```

### **🟧 Level 1: Administrative & Basic Services**

#### 3. **Administration Service** ⚙️
- **Authorization**: 🔒 **Admin-only for ALL operations**
- **Tables**: `users`, `roles`, `permissions`, `system_config`, `user_salary`, `mechanics`, `equipment`
- **Functions**:
  - **User Management**:
    - `POST /admin/users` - Create user (Admin only)
    - `GET /admin/users` - List users (Admin only)
    - `PUT /admin/users/{id}` - Update user (Admin only)
    - `DELETE /admin/users/{id}` - Delete user (Admin only)
  - **Role Management**: Complete CRUD (Admin only)
  - **Permission Management**: Permission assignment (Admin only)
  - **System Configuration**: Global settings management
  - **Salary Management**: Employee payroll
  - **Equipment Management**:
    - Equipment inventory tracking and status management
    - Maintenance scheduling and alerts
    - Mechanic contact management and specializations
    - Equipment cost tracking and reporting

#### 4. **Customer Service** 👥
- **Tables**: `customers`
- **Functions**:
  - Customer registration and profile management
  - Contact information management
  - Customer search and filtering
  - Customer analytics and reporting

#### 5. **Expenses Service** 💰
- **Tables**: `expense_categories`, `expenses`, `expense_receipts`
- **Functions**:
  - **Expense Categories**: Budget classification
  - **Expenses**: Record management and categorization
  - **Expense Receipts**: Invoice management with image uploads

### **🟪 Level 2: Business Logic Services**

#### 6. **Inventory Service** 📦
- **Tables**: `suppliers`, `ingredients`, `existences`, `runout_ingredient_report`, `recipe_categories`, `recipes`, `recipe_ingredients`
- **Functions**:
  - **Suppliers**: Vendor management
  - **Ingredients**: Raw materials management
  - **Existences**: Batch tracking, FIFO logic, cost calculation
  - **Runout Reports**: Usage tracking and stock updates
  - **Recipe Management**: Product recipes and pricing
  - **Recipe Categories**: Product categorization

### **🟢 Level 3: Advanced Business Logic**

#### 7. **Promotions Service** 🎉
- **Tables**: `promotions`, `customer_points`
- **Functions**:
  - **Promotions**: Discount campaigns, time-based offers
  - **Customer Points**: Loyalty tracking, point redemption

### **🔵 Level 4: Complex Operations**

#### 8. **Orders Service** 🛒
- **Tables**: `orders`, `ordered_receipes`
- **Functions**:
  - **Orders**: Sales processing, payment handling, invoice generation
  - **Order Items**: Line item management, quantity tracking

### **🟩 Level 5: Analytics Services**

#### 9. **Waste Service** 🗑️
- **Tables**: `waste_loss`
- **Functions**:
  - Waste incident reporting
  - Financial loss calculation
  - Prevention analysis and optimization

---

## 🔄 Inter-Service Communication

### **Authentication Flow**
```
1. User Login Request → Session Service
2. Session Service → Administration Service (fetch user/role data)
3. Session Service → Issues JWT with user context
4. Subsequent Requests → JWT validated by Session Service
5. Business Services → Call Administration Service for permission validation
```

### **Authorization Flow**
```
1. Request with JWT → Service Endpoint
2. Service → Session Service (validate JWT)
3. Service → Administration Service (check specific permissions)
4. Service → Execute operation if authorized
5. Service → Audit Service (log operation)
```

### **Service Communication Patterns**

- **Synchronous HTTP**: Direct service-to-service API calls
- **Authentication Middleware**: Shared across all services
- **Database Access**: Each service owns its tables
- **Shared Database**: All services use same PostgreSQL instance with proper table ownership

---

## 🛡️ Security Model

### **JWT Token Structure**
```json
{
  "user_id": "uuid",
  "username": "string", 
  "role_id": "uuid",
  "role_name": "string",
  "iat": "timestamp",
  "exp": "timestamp"
}
```

### **Permission Validation**
```
1. Extract user context from JWT
2. Call Administration Service: GET /admin/permissions/validate
3. Check specific permission (e.g., "Ingredients-Create")
4. Allow/deny operation based on response
```

### **Admin Protection**
- All Administration Service endpoints require admin role
- User/role/permission operations are strictly controlled
- System configuration changes are admin-only
- Audit logging for all administrative actions

---

## 🚀 Implementation Guidelines

### **Service Development Order**
1. ✅ **Session Service** (Completed)
2. ✅ **Data Service** (Completed)
3. ✅ **Gateway Service** (Completed)
4. 🔄 **Administration Service** (Next - Critical for other services, includes equipment management)
5. 🔄 **Audit Service** (Independent implementation)
6. 🔄 **Customer & Expenses Services** (Basic CRUD)
7. 🔄 **Inventory Service** (Core business logic)
8. 🔄 **Promotions Service** (Advanced features)
9. 🔄 **Orders Service** (Complex integrations)
10. 🔄 **Waste Service** (Analytics and optimization)

### **Development Standards**
- **API Design**: RESTful conventions with proper HTTP status codes
- **Error Handling**: Consistent error responses across services
- **Logging**: Structured logging with correlation IDs
- **Testing**: Unit tests, integration tests, and API tests
- **Documentation**: OpenAPI/Swagger specifications per service

### **Database Strategy**
- **Shared Database**: Single PostgreSQL instance
- **Table Ownership**: Each service owns specific tables
- **Migrations**: Database-first approach with SQL migrations
- **Transactions**: Service-level transaction management

---

## 📁 Directory Structure per Service

```
<service-name>/
├── cmd/
│   └── main.go                 # Service entry point
├── internal/
│   ├── handlers/               # HTTP handlers
│   │   ├── <table1>_handler.go
│   │   └── <table2>_handler.go
│   ├── services/               # Business logic
│   │   ├── <table1>_service.go
│   │   └── <table2>_service.go
│   ├── repositories/           # Data access
│   │   ├── <table1>_repository.go
│   │   └── <table2>_repository.go
│   ├── models/                 # Data models
│   │   ├── <table1>.go
│   │   └── <table2>.go
│   └── middleware/             # Service middleware
│       ├── auth.go
│       └── logging.go
├── pkg/                        # Shared packages
│   ├── auth/                   # Authentication utilities
│   ├── database/               # Database utilities
│   └── utils/                  # Common utilities
├── docs/                       # Service documentation
│   ├── api.yaml               # OpenAPI specification
│   └── README.md              # Service-specific docs
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── scripts/                    # Service scripts
├── tests/                      # Test files
├── Makefile                    # Service management
└── go.mod                      # Go module
```

---

## 🎯 Architecture Benefits

### **Maintainability**
- Clear service boundaries reduce complexity
- Single responsibility makes services easier to understand
- Independent deployment and scaling per service

### **Security**
- Centralized authentication with distributed authorization
- Admin-only access for sensitive operations
- Comprehensive audit logging

### **Scalability**
- Services can be scaled independently based on load
- Database can be partitioned if needed
- Microservices enable team specialization

### **Flexibility**
- Services can be updated independently
- Technology stack can vary per service if needed
- Easy to add new features as new services

---

**This architecture provides a solid foundation for building a comprehensive, secure, and scalable ice cream store management system.** 🍦🚀 