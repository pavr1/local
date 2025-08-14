# 🍦 Local Development Setup

This guide helps you run the Ice Cream Store system locally for optimal performance. Only the database runs in Docker - all other services run natively on your machine.

## 🚀 Quick Start

### Prerequisites
- **Docker** (for database only)
- **Go 1.19+** (for backend services)
- **Node.js 16+** (for UI service)
- **curl** (for health checks)

### Start Everything
```bash
./start-local.sh
```

That's it! 🎉 All services will start automatically.

## 📋 Available Commands

| Command | Description |
|---------|-------------|
| `./start-local.sh` | Start all services (default) |
| `./start-local.sh stop` | Stop all services |
| `./start-local.sh restart` | Restart all services |
| `./start-local.sh status` | Show service status |
| `./start-local.sh health` | Check service health |
| `./start-local.sh logs [service]` | View service logs |
| `./start-local.sh help` | Show help message |

## 🔗 Service URLs

| Service | Local URL | Description |
|---------|-----------|-------------|
| **UI Application** | http://localhost:3000 | Main web interface |
| **Gateway API** | http://localhost:8082 | API Gateway (main entry point) |
| **Session API** | http://localhost:8081 | Authentication service |
| **Orders API** | http://localhost:8083 | Order management |
| **Inventory API** | http://localhost:8084 | Inventory management |
| **Database Admin** | http://localhost:8080 | PgAdmin interface |

## 🗄️ Database Access

- **Host**: localhost
- **Port**: 5432
- **Database**: icecream_store
- **Username**: postgres
- **Password**: postgres123
- **PgAdmin**: http://localhost:8080
  - Email: admin@icecreamstore.com
  - Password: admin123

## 📊 Service Status

Check what's running:
```bash
./start-local.sh status
```

Example output:
```
📊 Service Status:
==================================
🗄️  Database:        RUNNING (Docker)
🔧 session-service:  RUNNING (Port: 8081)
🔧 orders-service:   RUNNING (Port: 8083)
🔧 gateway-service:  RUNNING (Port: 8082)
🔧 inventory-service:RUNNING (Port: 8084)
🎨 UI Service:        RUNNING (Port: 3000)
```

## 📋 Viewing Logs

View all available logs:
```bash
./start-local.sh logs
```

View specific service logs:
```bash
./start-local.sh logs gateway-service
./start-local.sh logs session-service
./start-local.sh logs orders-service
./start-local.sh logs inventory-service
./start-local.sh logs ui-service
```

## 🧪 Testing the System

### Health Checks
```bash
# Check all services
./start-local.sh health

# Individual service health
curl http://localhost:8082/api/v1/gateway/p/health                    # Gateway
curl http://localhost:8081/api/v1/sessions/p/health      # Session
curl http://localhost:8083/api/v1/orders/p/health        # Orders
curl http://localhost:8084/api/v1/inventory/p/health     # Inventory
```

### API Testing Examples

**Suppliers API (through Gateway):**
```bash
# List suppliers
curl http://localhost:8082/api/v1/inventory/suppliers

# Create supplier (requires authentication)
curl -X POST http://localhost:8082/api/v1/inventory/suppliers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Supplier",
    "contact_person": "John Doe",
    "email": "john@test.com"
  }'
```

**Direct Service Access:**
```bash
# Bypass gateway (no auth required)
curl http://localhost:8084/api/v1/inventory/suppliers
```

## 🛑 Stopping Services

Stop everything:
```bash
./start-local.sh stop
```

Stop specific service manually:
```bash
# Find PID
cat .local-pids/gateway-service.pid

# Kill process
kill $(cat .local-pids/gateway-service.pid)
```

## 🔧 Development Workflow

### Making Changes

1. **Edit code** in any service
2. **Restart specific service**:
   ```bash
   # Stop all
   ./start-local.sh stop
   
   # Start all (will rebuild changed services)
   ./start-local.sh start
   ```

3. **Or rebuild manually**:
   ```bash
   cd gateway-service
   go build -o gateway-service .
   ./gateway-service &
   ```

### Hot Reload (UI)
The UI service supports hot reload automatically. Just edit files in `ui/` and refresh your browser.

### Database Changes
```bash
# Reset database
cd data-service/docker
docker-compose down -v
docker-compose up -d
```

## 🚨 Troubleshooting

### Port Conflicts
If ports are busy:
```bash
# Find what's using a port
lsof -i :8082

# Kill process on port
./start-local.sh stop  # This handles cleanup
```

### Service Won't Start
```bash
# Check logs
./start-local.sh logs service-name

# Manual start to see errors
cd service-directory
go run .
```

### Database Connection Issues
```bash
# Check database is running
docker ps | grep postgres

# Test connection
psql -h localhost -p 5432 -U postgres -d icecream_store
```

### Clean Restart
```bash
# Stop everything
./start-local.sh stop

# Clean PIDs and logs
rm -rf .local-pids logs

# Restart
./start-local.sh start
```

## 💡 Performance Benefits

Running locally vs containerized:

| Aspect | Local | Containerized |
|--------|-------|---------------|
| **Memory Usage** | ~200MB | ~1GB+ |
| **CPU Usage** | Lower | Higher |
| **Build Time** | 2-5s | 30-60s |
| **Hot Reload** | Instant | Slow |
| **Debugging** | Native tools | Container overhead |

## 🔄 Migration from Docker

If you were running containerized services:

1. **Stop all containers**:
   ```bash
   make stop-all
   ```

2. **Start local development**:
   ```bash
   ./start-local.sh
   ```

3. **Database data persists** (same Docker volume)

## 📁 File Structure

```
.
├── start-local.sh              # Main script
├── .local-pids/               # Service PIDs
├── logs/                      # Service logs
├── data-service/              # Database (Docker)
├── session-service/           # Auth service (Local)
├── orders-service/            # Orders service (Local)
├── gateway-service/           # Gateway service (Local)
├── inventory-service/         # Inventory service (Local)
└── ui/                        # UI service (Local)
```

## 🎯 Next Steps

1. **Access the UI**: http://localhost:3000
2. **Check Gateway API**: http://localhost:8082/api/v1/gateway/p/health
3. **Start developing**: Edit any service and restart as needed

Happy coding! 🚀 