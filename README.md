# 🍦 Ice Cream Store - Microservices Management System

A comprehensive business management system built with Go microservices architecture for ice cream store operations.

## 🏗️ System Architecture

This project implements a **microservices architecture** with 9 specialized services, each handling specific business domains with clear separation of concerns.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Data Service  │ ──▶ │   Auth Service  │ ──▶ │ Gateway Service │
│   PostgreSQL    │     │   JWT Auth      │     │   API Gateway   │
│   Port: 5432    │     │   Port: 8081    │     │   Port: 8080    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## 📦 Services Overview

### 🟥 **Level 0: Foundation Services**
- **🔐 Authentication Service** (Port: 8081) - JWT tokens, login/logout, security middleware
- **📋 Audit Service** - Activity logging with `LogAuditEntry()` & `RetrieveAuditLogs()` API methods

### 🟧 **Level 1: Administrative Services**  
- **⚙️ Administration Service** - User/role/permission management, system config, equipment tracking (Admin only)
- **👥 Customer Service** - Customer management and profiles
- **💰 Invoice Service** - Invoice and transaction management

### 🟪 **Level 2: Business Logic Services**
- **📦 Inventory Service** - Suppliers, ingredients, recipes, stock management

### 🟢 **Level 3: Advanced Services**
- **🎉 Promotions Service** - Discounts, loyalty programs, customer points

### 🔵 **Level 4: Operations Services**
- **🛒 Orders Service** - Sales processing, invoicing, payment handling

### 🟩 **Level 5: Analytics Services**
- **🗑️ Waste Service** - Waste tracking and loss analysis

## 🚀 Quick Start

### Prerequisites
- **Docker & Docker Compose** - Container orchestration
- **Colima** - Docker runtime for macOS
- **Go 1.21+** - Backend development
- **PostgreSQL 15** - Database

### Installation

1. **Clone and Setup**
   ```bash
   git clone <repository>
   cd local
   ```

2. **Complete System Setup**
   ```bash
   make fresh  # 🔥 Fresh install of ALL services
   ```

3. **Individual Service Management**
   ```bash
   make fresh-data     # Fresh install data service only
   make fresh-auth     # Fresh install auth service only
   make fresh-gateway  # Fresh install gateway service only
   ```

## 🎛️ Service Management

### **System-Wide Commands**
```bash
make fresh          # Complete fresh installation of all services
make start-all      # Start all services in correct order
make stop-all       # Stop all services
make test-all       # Test all service endpoints
make status         # Check status of all services
make health-all     # Health check all services
make clean-all      # Clean all services
```

### **Individual Service Commands**
```bash
# Service-specific operations
make start-data     # Start data service
make start-auth     # Start auth service  
make start-gateway  # Start gateway service

# Testing and monitoring
make test-data      # Test data service
make health-auth    # Health check auth service
make logs-all       # View logs from all services
```

## 🔗 Service Endpoints

### **Core Services**
- **Database (PostgreSQL)**: `postgresql://postgres:postgres123@localhost:5432/icecream_store`
- **PgAdmin**: http://localhost:8080 (`admin@icecreamstore.com` / `admin123`)
- **Session Service**: http://localhost:8081/api/v1/sessions/p/health
- **Gateway Service**: http://localhost:8080/api/v1/gateway/p/health
- **Portainer (Docker UI)**: https://localhost:9443

### **Management Interfaces**
- **Database Management**: PgAdmin at http://localhost:8080
- **Container Management**: Portainer at https://localhost:9443
- **API Gateway**: http://localhost:8080

## 🏛️ Service Architecture Details

### **Authentication Flow**
```
1. Authentication Service → Validates credentials, issues JWT
2. JWT Contains → User ID, role, basic permissions  
3. Each Service → Validates JWT via Administration Service
4. Administration Service → Central authority for user/role/permission data
```

### **Service Dependencies**
```
Authentication Service (No dependencies)
    ↓
Administration Service (Admin-only user/role management)
    ↓  
Business Services (Customer, Equipment, Expenses)
    ↓
Inventory Service (Core business logic)
    ↓
Promotions Service (Advanced business logic)
    ↓
Orders Service (Complex integrations)
    ↓
Waste Service (Analytics)
```

## 📁 Project Structure

```
local/
├── session-service/        # 🔐 Session & JWT management
├── data-service/           # 🗄️ PostgreSQL database setup
├── gateway-service/        # 🌐 API Gateway and routing
├── audit-service/          # 📋 Audit logging (LogAuditEntry & RetrieveAuditLogs) (Future)
├── administration-service/ # ⚙️ User/role/config/equipment management (Future)
├── customer-service/       # 👥 Customer management (Future)
├── expenses-service/       # 💰 Invoice management (Future)
├── inventory-service/      # 📦 Core inventory logic (Future)
├── promotions-service/     # 🎉 Promotions & loyalty (Future)
├── orders-service/         # 🛒 Sales & orders (Future)
├── waste-service/          # 🗑️ Waste analytics (Future)
├── Makefile               # 🎯 Root orchestration
└── README.md              # 📚 This documentation
```

## 🛠️ Development Workflow

### **Service Implementation Order**
1. ✅ **Authentication Service** (Completed)
2. ✅ **Data Service** (Completed)  
3. ✅ **Gateway Service** (Completed)
4. 🔄 **Audit Service** (LogAuditEntry & RetrieveAuditLogs APIs)
5. 🔄 **Administration Service** (Next - Critical for other services, includes equipment management)
6. 🔄 **Customer Service**
7. 🔄 **Invoice Service**
8. 🔄 **Inventory Service**
9. 🔄 **Promotions Service**
10. 🔄 **Orders Service**
11. 🔄 **Waste Service**

### **Development Commands**
```bash
# Service development
cd <service-name>
make dev            # Run in development mode
make build          # Build service binary
make test           # Run service tests
make lint           # Run code linting

# Database operations
make connect        # Connect to PostgreSQL CLI
make reset          # Reset database (⚠️ DELETES ALL DATA)
```

## 🔐 Security & Authorization

### **Authentication Service**
- **Purpose**: Security & session management only
- **Functions**: Login/logout, JWT tokens, password validation
- **Integration**: Calls Administration Service for user data

### **Administration Service** 
- **Purpose**: User/role/permission management (Admin only)
- **Authorization**: Requires admin permissions for ALL operations
- **Functions**: User CRUD, role management, system configuration

### **Permission Model**
- **Format**: `Entity-Action` (e.g., "Ingredients-Create")
- **Validation**: Each service validates permissions via Administration Service
- **Roles**: Admin (full access), Employee (limited access)

## 📊 Business Domains

### **Inventory Management**
- Suppliers, ingredients, existences, recipes
- Stock tracking with FIFO logic
- Cost calculation with margins and taxes
- Expiration and runout reporting

### **Financial Management**
- Expense tracking and categorization
- Receipt management with image uploads
- Employee salary and payroll
- Order processing and invoicing

### **Customer Operations**
- Customer profiles and contact management
- Loyalty points and promotions
- Order history and analytics

### **Administrative Operations**
- User, role, and permission management
- System configuration and settings
- Equipment inventory and maintenance scheduling
- Employee salary and payroll management

## 📈 Monitoring & Analytics

### **System Health**
```bash
make final-status   # Complete system status check
make health-all     # Health check all services
make logs-all       # View logs from all services
```

### **Audit & Security Monitoring**
- **Audit Service API**: `LogAuditEntry()` for real-time activity logging
- **Advanced Filtering**: `RetrieveAuditLogs()` with variadic parameters
- **Severity Classification**: Info, warning, and error level tracking
- **Cross-Service Correlation**: Track operations across microservices
- **Compliance Reporting**: Flexible querying for regulatory requirements

### **Business Analytics**
- Waste tracking and loss analysis
- Sales reporting and trends
- Inventory optimization
- Customer behavior insights

## 🤝 Contributing

1. **Service Development**
   - Follow microservices patterns
   - Maintain clear service boundaries
   - Implement proper error handling
   - Add comprehensive tests

2. **Database Changes**
   - Update Database.md documentation
   - Create migration scripts
   - Test with sample data

3. **API Design**
   - Follow RESTful conventions
   - Implement proper status codes
   - Add request/response validation
   - Document endpoints

## 📚 Documentation

- **Database Schema**: `gateway-service/docs/Database.md`
- **Business Requirements**: `gateway-service/docs/Requirements.md`
- **ER Diagram**: `gateway-service/docs/Database-ER-Diagram.md`
- **Service APIs**: Each service's `README.md`

## 🎯 System Features

### **Operational Excellence**
- ✅ Complete microservices architecture
- ✅ Docker containerization  
- ✅ Database management with PgAdmin
- ✅ Centralized orchestration with Makefiles
- ✅ Comprehensive logging and monitoring

### **Business Capabilities**
- 🔄 Inventory management with FIFO logic
- 🔄 Financial tracking and reporting  
- 🔄 Customer loyalty programs
- 🔄 Equipment maintenance scheduling
- 🔄 Waste tracking and optimization

### **Security & Compliance**
- ✅ JWT-based authentication
- 🔄 Role-based access control
- 🔄 Comprehensive audit logging
- 🔄 Data encryption and security

---

**🍦 Ready to build the sweetest business management system!** 🚀

For detailed service documentation, see individual service directories and the docs folder. 