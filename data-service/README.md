# Ice Cream Store Data Service

This service provides PostgreSQL database infrastructure for the Ice Cream Store management system. It includes Docker containerization, database schema initialization, and a Go database handler that other services can import.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Go 1.21 or later

### Starting the Database

```bash
./scripts/start.sh
```

This will:
- Start PostgreSQL and PgAdmin containers
- Initialize the database with the complete schema
- Wait for the database to be ready
- Display connection information

### Stopping the Database

```bash
./scripts/stop.sh
```

### Resetting the Database (⚠️ Deletes all data)

```bash
./scripts/reset.sh
```

## 📊 Database Access

### Connection Details
- **Host:** localhost
- **Port:** 5432
- **Database:** icecream_store
- **Username:** postgres
- **Password:** postgres123

### PgAdmin Web Interface
- **URL:** http://localhost:8080
- **Email:** admin@icecream.local
- **Password:** admin123

### Command Line Access

```bash
./scripts/connect.sh
```

## 🏗️ Architecture

### Database Schema

The database includes the following management areas:

- **Inventory Management:** Suppliers, Ingredients, Existences, Recipes, Recipe Categories
- **Expenses Management:** Expense Categories, Expenses, Expense Receipts
- **Customer Management:** Customers, Customer Points
- **Income Management:** Orders, Ordered Recipes
- **Promotions & Loyalty:** Promotions system with points and discounts
- **Equipment Management:** Equipment tracking with maintenance scheduling
- **Waste & Loss Tracking:** Waste reporting with financial loss calculation
- **Administration:** System Configuration, User Salary management
- **Authentication & Authorization:** Users, Roles, Permissions with RBAC
- **Audit & Security:** Comprehensive audit logging

### Key Features

- **UUID Primary Keys:** All tables use UUID for better scalability
- **Audit Logging:** Comprehensive audit trail for all critical operations
- **Data Integrity:** Foreign key constraints and check constraints
- **Performance Optimized:** Proper indexing for common query patterns
- **RBAC Security:** Role-based access control with granular permissions

## 🔧 Using the Database Handler

Other services can import and use the database handler:

### Installation

```go
import "data-service/pkg/database"
```

### Basic Usage

```go
package main

import (
    "log"
    "time"
    "data-service/pkg/database"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)

    // Create configuration
    config := &database.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "postgres123",
        DBName:   "icecream_store",
        SSLMode:  "disable",
    }

    // Create database handler
    db := database.New(config, logger)

    // Connect to database
    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer db.Close()

    // Use the database
    rows, err := db.Query("SELECT * FROM suppliers LIMIT 10")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    defer rows.Close()

    // Process results...
}
```

### Configuration Options

```go
type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
    
    // Connection pool settings
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    
    // Timeout settings
    ConnectTimeout time.Duration
    QueryTimeout   time.Duration
    
    // Retry settings
    MaxRetries    int
    RetryInterval time.Duration
}
```

### Environment Variables

The handler supports loading configuration from environment variables:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=icecream_store
DB_SSLMODE=disable
```

## 📁 Project Structure

```
data-service/
├── docker/
│   ├── docker-compose.yml      # Container orchestration
│   ├── config/
│   │   └── postgresql.conf     # PostgreSQL configuration
│   └── init/
│       └── 01-init-database.sql # Database schema initialization
├── scripts/
│   ├── start.sh               # Start database containers
│   ├── stop.sh                # Stop database containers
│   ├── reset.sh               # Reset database (delete all data)
│   ├── logs.sh                # View container logs
│   └── connect.sh             # Connect to database CLI
├── pkg/
│   └── database/
│       └── db_handler.go      # Database handler implementation
├── docs/                      # Documentation files
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## 🛠️ Management Scripts

### View Logs

```bash
# All logs
./scripts/logs.sh

# PostgreSQL only
./scripts/logs.sh postgres

# PgAdmin only
./scripts/logs.sh pgadmin
```

### Health Checks

The database handler includes built-in health checking:

```go
if err := db.HealthCheck(); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

### Connection Pool Monitoring

```go
stats := db.GetStats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

## 🔒 Security Features

- **Password Hashing:** bcrypt for user passwords
- **Role-Based Access Control:** Granular permissions system
- **Audit Logging:** All critical operations are logged
- **SQL Injection Protection:** Parameterized queries only
- **Connection Security:** Configurable SSL modes

## 🚨 Troubleshooting

### Container Won't Start

```bash
# Check Docker status
docker info

# View detailed logs
./scripts/logs.sh postgres

# Reset everything
./scripts/reset.sh
```

### Connection Issues

1. Ensure containers are running: `docker ps`
2. Check port availability: `lsof -i :5432`
3. Verify configuration in `docker-compose.yml`
4. Check logs: `./scripts/logs.sh`

### Permission Errors

```bash
# Fix script permissions
chmod +x scripts/*.sh

# Check Docker permissions
docker ps
```

## 📈 Performance Tuning

The PostgreSQL configuration is optimized for development. For production:

1. Adjust `shared_buffers` in `postgresql.conf`
2. Tune connection pool settings in application config
3. Monitor query performance with `log_min_duration_statement`
4. Consider read replicas for heavy read workloads

## 🔄 Backup and Recovery

### Manual Backup

```bash
docker exec icecream_postgres pg_dump -U postgres icecream_store > backup.sql
```

### Restore from Backup

```bash
# Reset database first
./scripts/reset.sh
./scripts/start.sh

# Restore
docker exec -i icecream_postgres psql -U postgres icecream_store < backup.sql
```

## 🤝 Contributing

When making schema changes:

1. Update `docker/init/01-init-database.sql`
2. Test with `./scripts/reset.sh` and `./scripts/start.sh`
3. Update this README if needed
4. Ensure other services are compatible

## 📞 Support

For issues with the data service:

1. Check logs: `./scripts/logs.sh`
2. Verify container status: `docker ps`
3. Test connection: `./scripts/connect.sh`
4. Reset if needed: `./scripts/reset.sh` 