# 🍦📦 Ice Cream Store - Inventory Service

The **Inventory Service** is a comprehensive microservice for managing all inventory-related operations in the Ice Cream Store system. It handles suppliers, ingredients, existences (stock), runout reports, recipe categories, recipes, and recipe ingredients.

## 🏗️ Architecture Overview

This service follows the microservices architecture pattern and manages **7 core entities**:

1. **🏪 Suppliers** - Vendor/supplier management
2. **🥄 Ingredients** - Raw materials and ingredients
3. **📦 Existences** - Individual ingredient purchase batches with inventory tracking
4. **📊 Runout Reports** - Employee-reported ingredient usage
5. **📋 Recipe Categories** - Recipe categorization (Postres, Helados, Batidos, etc.)
6. **🍨 Recipes** - Product recipes with cost calculations
7. **🔗 Recipe Ingredients** - Junction table linking recipes to ingredients

## 📋 Features

### ✅ Fully Implemented
- **Suppliers Management**: Full CRUD operations
- **Ingredients Management**: Full CRUD operations
- **Health Monitoring**: Service health checks and status monitoring
- **Database Integration**: PostgreSQL with connection pooling
- **Logging**: Structured JSON logging with configurable levels
- **Docker Support**: Full containerization with Docker Compose
- **Testing**: Comprehensive API testing scripts

### 🚧 Ready for Implementation
- **Existences Management**: Stock tracking with expiration dates and pricing
- **Runout Reports**: Employee usage reporting with inventory updates
- **Recipe Categories**: Product categorization system
- **Recipes**: Product definitions with cost calculations
- **Recipe Ingredients**: Recipe composition management

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- PostgreSQL database (icecream_store)
- Data service running (for database)

### Installation

```bash
# Start the inventory service
make start

# Check service status
make status

# Run API tests
make test

# View logs
make logs
```

### Available Commands

```bash
make help           # Show all available commands
make start          # Start the service
make stop           # Stop the service
make restart        # Restart the service
make test           # Run API tests
make logs           # View service logs
make build          # Build service locally
make clean          # Clean up containers
make fresh          # Fresh start (clean + build + test)
make status         # Check service health
make health         # Quick health check
```

## 🔌 API Endpoints

### Service Information
- `GET /` - Service information and available endpoints
- `GET /api/v1/inventory/health` - Health check

### Suppliers (✅ Implemented)
- `POST /api/v1/suppliers` - Create supplier
- `GET /api/v1/suppliers` - List all suppliers
- `GET /api/v1/suppliers/{id}` - Get supplier by ID
- `PUT /api/v1/suppliers/{id}` - Update supplier
- `DELETE /api/v1/suppliers/{id}` - Delete supplier

### Ingredients (✅ Implemented)
- `POST /api/v1/ingredients` - Create ingredient
- `GET /api/v1/ingredients` - List all ingredients
- `GET /api/v1/ingredients/{id}` - Get ingredient by ID
- `PUT /api/v1/ingredients/{id}` - Update ingredient
- `DELETE /api/v1/ingredients/{id}` - Delete ingredient

### Existences (🚧 Ready to Implement)
- `POST /api/v1/existences` - Create existence (stock entry)
- `GET /api/v1/existences` - List all existences
- `GET /api/v1/existences/{id}` - Get existence by ID
- `PUT /api/v1/existences/{id}` - Update existence
- `DELETE /api/v1/existences/{id}` - Delete existence
- `GET /api/v1/existences/low-stock` - List low stock items
- `GET /api/v1/existences/expiring-soon` - List items expiring soon

### Other Endpoints (🚧 Ready to Implement)
- Runout Reports: `/api/v1/runout-reports/*`
- Recipe Categories: `/api/v1/recipe-categories/*`
- Recipes: `/api/v1/recipes/*`
- Recipe Ingredients: `/api/v1/recipe-ingredients/*`

## 📊 Database Schema

The service manages the following database tables:

```sql
-- Suppliers: Store supplier/vendor information
suppliers (id, supplier_name, contact_number, email, address, notes, created_at, updated_at)

-- Ingredients: Raw materials and ingredients
ingredients (id, name, supplier_id, created_at, updated_at)

-- Existences: Individual ingredient purchases with inventory tracking
existences (id, existence_reference_code, ingredient_id, expense_receipt_id, 
           units_purchased, units_available, unit_type, items_per_unit,
           cost_per_item, cost_per_unit, total_purchase_cost, remaining_value,
           expiration_date, income_margin_percentage, iva_percentage, etc.)

-- Runout Reports: Employee-reported ingredient usage
runout_ingredient_report (id, existence_id, employee_id, quantity, unit_type, 
                         report_date, created_at, updated_at)

-- Recipe Categories: Product categorization
recipe_categories (id, name, description, created_at, updated_at)

-- Recipes: Product definitions
recipes (id, recipe_name, recipe_description, picture_url, 
        recipe_category_id, total_recipe_cost, created_at, updated_at)

-- Recipe Ingredients: Recipe composition
recipe_ingredients (id, recipe_id, ingredient_id, number_of_units, 
                   created_at, updated_at)
```

## 🧪 Testing

### Manual Testing
```bash
# Test suppliers
curl -X GET http://localhost:8082/api/v1/suppliers
curl -X POST http://localhost:8082/api/v1/suppliers \
  -H "Content-Type: application/json" \
  -d '{"supplier_name": "Test Supplier", "email": "test@example.com"}'

# Test ingredients
curl -X GET http://localhost:8082/api/v1/ingredients
curl -X POST http://localhost:8082/api/v1/ingredients \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Ingredient"}'
```

### Automated Testing
```bash
make test           # Run full API test suite
make suppliers-test # Test suppliers endpoints
make ingredients-test # Test ingredients endpoints
```

## 📁 Project Structure

```
inventory-service/
├── config/                    # Configuration management
│   └── config.go
├── handlers/                  # HTTP request handlers
│   └── inventory_handler.go
├── models/                    # Data models and DTOs
│   └── models.go
├── sql/                       # SQL queries and scripts
│   ├── queries.go
│   └── scripts/
│       ├── suppliers/
│       ├── ingredients/
│       ├── existences/
│       ├── runout_reports/
│       ├── recipe_categories/
│       ├── recipes/
│       └── recipe_ingredients/
├── entities/                  # Entity-specific logic (ready for implementation)
│   ├── suppliers/
│   ├── ingredients/
│   ├── existences/
│   ├── runout_reports/
│   ├── recipe_categories/
│   ├── recipes/
│   └── recipe_ingredients/
├── docker/                    # Docker configuration
│   ├── Dockerfile
│   └── docker-compose.yml
├── scripts/                   # Management scripts
│   ├── start.sh
│   ├── stop.sh
│   ├── logs.sh
│   ├── reset.sh
│   └── test.sh
├── main.go                    # Application entry point
├── Makefile                   # Build and management commands
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── config.env.example         # Environment variables template
└── README.md                  # This file
```

## ⚙️ Configuration

### Environment Variables

Copy `config.env.example` to `.env` and customize:

```env
# Server Configuration
INVENTORY_SERVER_HOST=0.0.0.0
INVENTORY_SERVER_PORT=8082

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=icecream_store
DB_SSLMODE=disable

# Logging Configuration
LOG_LEVEL=info
```

## 🐳 Docker Support

The service includes full Docker support following the project standards:

```bash
# Build and start with Docker
make start

# Build Docker image only
make docker-build

# Open shell in container
make docker-shell

# View container stats
make top
```

## 🔧 Development

### Local Development
```bash
# Build locally
make build

# Run locally (requires database)
./inventory-service

# Run tests
make test
```

### Adding New Features
1. Add models to `models/models.go`
2. Add SQL queries to `sql/scripts/{entity}/`
3. Add handlers to `handlers/inventory_handler.go`
4. Update routes in `main.go`
5. Add tests to `scripts/test.sh`

## 📈 Monitoring

### Health Monitoring
```bash
make health         # Quick health check
make status         # Detailed status
make db-status      # Database connection check
```

### Logs
```bash
make logs           # View logs
make logs -f        # Follow logs in real-time
```

## 🔗 Integration

This service integrates with:
- **Data Service**: PostgreSQL database
- **Gateway Service**: Authentication and routing (auth handled by gateway)
- **Orders Service**: Recipe information for order processing
- **Expenses Service**: Expense receipts for existence tracking

## 📝 Notes

- **Authentication**: Handled by the gateway service
- **CORS**: Handled by the gateway service
- **Database**: Uses the shared `icecream_store` PostgreSQL database
- **Port**: Runs on port 8082 (configurable)
- **Logging**: Structured JSON logging with timestamps

## 🚧 Future Enhancements

1. **Complete Implementation**: Finish implementing all entity handlers
2. **Business Logic**: Add recipe cost calculation automation
3. **Inventory Alerts**: Low stock and expiration notifications
4. **FIFO Logic**: Implement First-In-First-Out inventory management
5. **Reporting**: Add comprehensive inventory reports
6. **File Upload**: Support for receipt images and product pictures

## 📞 Support

For questions or issues with the Inventory Service:
1. Check the logs: `make logs`
2. Check service status: `make status`
3. Run health check: `make health`
4. Review this README and the API documentation 