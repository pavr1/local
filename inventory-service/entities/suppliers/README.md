# Suppliers API

The Suppliers API manages supplier information for the Ice Cream Store Inventory System. This module handles CRUD operations for suppliers who provide ingredients and materials.

## 🚀 Quick Start

The service runs on `http://localhost:8084` when started locally.
**Gateway**: Available through the gateway service at `http://localhost:8082`

### Health Check
```bash
# Direct to inventory service
curl http://localhost:8084/api/v1/inventory/p/health

# Through gateway (requires authentication)
curl http://localhost:8082/api/v1/inventory/p/health
```

## 📋 API Endpoints

### Base URL
```
# Direct to inventory service
http://localhost:8084/api/v1/inventory/suppliers

# Through gateway (recommended - includes session management)
http://localhost:8082/api/v1/inventory/suppliers
```

### Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/inventory/suppliers` | List all suppliers |
| `POST` | `/inventory/suppliers` | Create a new supplier |
| `GET` | `/inventory/suppliers/{id}` | Get supplier by ID |
| `PUT` | `/inventory/suppliers/{id}` | Update supplier |
| `DELETE` | `/inventory/suppliers/{id}` | Delete supplier |

---

## 🔍 Detailed API Reference

### 1. List All Suppliers

**Request:**
```http
GET /api/v1/inventory/suppliers
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-here",
      "name": "Fresh Dairy Co.",
      "contact_person": "John Smith",
      "email": "john@freshdairy.com",
      "phone": "+1-555-0123",
      "address": "123 Dairy Lane, Farm City, FC 12345",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "message": "Suppliers retrieved successfully"
}
```

**Example:**
```bash
# Through gateway (recommended)
curl -X GET http://localhost:8082/api/v1/inventory/suppliers \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Direct to service
curl -X GET http://localhost:8084/api/v1/inventory/suppliers
```

---

### 2. Create New Supplier

**Request:**
```http
POST /api/v1/inventory/suppliers
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Premium Ingredients Inc.",
  "contact_person": "Jane Doe",
  "email": "jane@premium-ingredients.com",
  "phone": "+1-555-0456",
  "address": "456 Supply Street, Business City, BC 67890"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-uuid-here",
    "name": "Premium Ingredients Inc.",
    "contact_person": "Jane Doe", 
    "email": "jane@premium-ingredients.com",
    "phone": "+1-555-0456",
    "address": "456 Supply Street, Business City, BC 67890",
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  },
  "message": "Supplier created successfully"
}
```

**Example:**
```bash
# Through gateway (recommended)
curl -X POST http://localhost:8082/api/v1/inventory/suppliers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Premium Ingredients Inc.",
    "contact_person": "Jane Doe",
    "email": "jane@premium-ingredients.com",
    "phone": "+1-555-0456",
    "address": "456 Supply Street, Business City, BC 67890"
  }'

# Direct to service
curl -X POST http://localhost:8084/api/v1/inventory/suppliers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Ingredients Inc.",
    "contact_person": "Jane Doe",
    "email": "jane@premium-ingredients.com",
    "phone": "+1-555-0456",
    "address": "456 Supply Street, Business City, BC 67890"
  }'
```

---

### 3. Get Supplier by ID

**Request:**
```http
GET /api/v1/inventory/suppliers/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "supplier-uuid",
    "name": "Fresh Dairy Co.",
    "contact_person": "John Smith",
    "email": "john@freshdairy.com",
    "phone": "+1-555-0123",
    "address": "123 Dairy Lane, Farm City, FC 12345",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "message": "Supplier retrieved successfully"
}
```

**Example:**
```bash
# Through gateway (recommended)
curl -X GET http://localhost:8082/api/v1/inventory/suppliers/your-supplier-id-here \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Direct to service
curl -X GET http://localhost:8084/api/v1/inventory/suppliers/your-supplier-id-here
```

---

### 4. Update Supplier

**Request:**
```http
PUT /api/v1/inventory/suppliers/{id}
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Fresh Dairy Co. Updated",
  "contact_person": "John Smith Jr.",
  "email": "john.jr@freshdairy.com",
  "phone": "+1-555-0124",
  "address": "124 Dairy Lane, Farm City, FC 12345"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "supplier-uuid",
    "name": "Fresh Dairy Co. Updated",
    "contact_person": "John Smith Jr.",
    "email": "john.jr@freshdairy.com", 
    "phone": "+1-555-0124",
    "address": "124 Dairy Lane, Farm City, FC 12345",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:30:00Z"
  },
  "message": "Supplier updated successfully"
}
```

**Example:**
```bash
# Through gateway (recommended)
curl -X PUT http://localhost:8082/api/v1/inventory/suppliers/your-supplier-id-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Fresh Dairy Co. Updated",
    "contact_person": "John Smith Jr.",
    "email": "john.jr@freshdairy.com",
    "phone": "+1-555-0124",
    "address": "124 Dairy Lane, Farm City, FC 12345"
  }'

# Direct to service
curl -X PUT http://localhost:8084/api/v1/inventory/suppliers/your-supplier-id-here \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fresh Dairy Co. Updated",
    "contact_person": "John Smith Jr.",
    "email": "john.jr@freshdairy.com",
    "phone": "+1-555-0124",
    "address": "124 Dairy Lane, Farm City, FC 12345"
  }'
```

---

### 5. Delete Supplier

**Request:**
```http
DELETE /api/v1/inventory/suppliers/{id}
```

**Response:**
```json
{
  "success": true,
  "message": "Supplier deleted successfully"
}
```

**Example:**
```bash
# Through gateway (recommended)
curl -X DELETE http://localhost:8082/api/v1/inventory/suppliers/your-supplier-id-here \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Direct to service
curl -X DELETE http://localhost:8084/api/v1/inventory/suppliers/your-supplier-id-here
```

---

## 📝 Data Models

### Supplier Model
```json
{
  "id": "string (UUID)",
  "name": "string (required, max 255 chars)",
  "contact_person": "string (optional, max 255 chars)",
  "email": "string (optional, max 255 chars)",
  "phone": "string (optional, max 50 chars)",
  "address": "string (optional, max 500 chars)",
  "created_at": "string (ISO 8601 timestamp)",
  "updated_at": "string (ISO 8601 timestamp)"
}
```

### Create/Update Request Model
```json
{
  "name": "string (required, max 255 chars)",
  "contact_person": "string (optional, max 255 chars)",
  "email": "string (optional, max 255 chars)",  
  "phone": "string (optional, max 50 chars)",
  "address": "string (optional, max 500 chars)"
}
```

---

## ⚠️ Error Responses

### Common Error Formats

**400 Bad Request:**
```json
{
  "success": false,
  "error": "Bad Request",
  "message": "Invalid JSON format"
}
```

**404 Not Found:**
```json
{
  "success": false,
  "data": {},
  "message": "Supplier not found"
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "data": {},
  "message": "Failed to create supplier: database error details"
}
```

---

## 🧪 Testing

### Running Unit Tests
```bash
# From inventory-service directory
go test -v ./entities/suppliers/...
```

### Manual Testing Examples

1. **Create a supplier:**
```bash
# Through gateway (recommended)
curl -X POST http://localhost:8082/api/v1/inventory/suppliers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Supplier",
    "contact_person": "Test Person",
    "email": "test@example.com"
  }'

# Direct to service
curl -X POST http://localhost:8084/api/v1/inventory/suppliers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Supplier",
    "contact_person": "Test Person",
    "email": "test@example.com"
  }'
```

2. **List suppliers:**
```bash
# Through gateway (recommended)
curl http://localhost:8082/api/v1/inventory/suppliers \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Direct to service
curl http://localhost:8084/api/v1/inventory/suppliers
```

3. **Health check:**
```bash
# Through gateway
curl http://localhost:8082/api/v1/inventory/p/health

# Direct to service
curl http://localhost:8084/api/v1/inventory/p/health
```

---

## 🏗️ Architecture

### Components

- **HTTP Handler** (`http_handler.go`): Handles HTTP requests and responses
- **DB Handler** (`db_handler.go`): Manages database operations
- **Models** (`models.go`): Defines data structures and validation
- **SQL Queries** (`sql/`): Contains SQL scripts for database operations

### Request Flow
```
HTTP Request → HTTP Handler → DB Handler → Database → Response
```

---

## 🔧 Development

### Adding New Endpoints

1. Add method to `DBHandlerInterface` in `http_handler.go`
2. Implement method in `db_handler.go`
3. Add HTTP handler method in `http_handler.go`
4. Register route in `main.go` 
5. Add tests for new functionality

### Database Schema

The suppliers table includes:
- `id` (UUID, Primary Key)
- `name` (VARCHAR(255), NOT NULL)
- `contact_person` (VARCHAR(255))
- `email` (VARCHAR(255))
- `phone` (VARCHAR(50))
- `address` (VARCHAR(500))
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

---

## 🚀 Next Steps

This suppliers module is designed to scale alongside other inventory entities:
- **Ingredients**: Raw materials from suppliers
- **Existences**: Current stock levels
- **Recipes**: Ice cream formulations
- **Orders**: Purchase orders from suppliers

Each entity will follow the same architectural pattern established by suppliers. 