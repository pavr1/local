# Ice Cream Store Auth Service

This service provides JWT-based authentication and authorization for the Ice Cream Store management system. It handles user login, token generation, validation, and role-based access control.

## 🚀 Features

- **JWT Authentication**: Secure token-based authentication with configurable expiration
- **Role-Based Access Control (RBAC)**: Granular permissions system
- **Password Security**: bcrypt hashing with configurable cost
- **Token Refresh**: Automatic token refresh within threshold
- **Middleware Support**: Easy integration with other services
- **Comprehensive Logging**: Structured logging with logrus
- **Health Checks**: Built-in health monitoring
- **CORS Support**: Cross-origin resource sharing
- **Docker Support**: Full containerization with Docker Compose

## 🐳 Quick Start with Docker

### Prerequisites

1. **Docker & Docker Compose**: Ensure Docker is running
2. **Data Service**: The database must be running first with schema initialized

```bash
# Start the database first (this creates the sessions table)
cd ../data-service
make start

# Then start the session service
cd ../session-service
make install
```

### Database Schema

The session service uses a `sessions` table for database-backed session storage. **This table is automatically created by the data-service** - you don't need to run any additional schema migrations.

The sessions table is defined in `data-service/docker/init/01-init-database.sql` and includes:
- Session ID and user information
- JWT token hashes for validation
- Session expiration and activity tracking
- Proper indexes for performance

### Using Make Commands

```bash
# Complete fresh setup
make fresh                    # Clean, build, start, test everything

# Basic operations
make start                    # Start auth service
make stop                     # Stop auth service
make restart                  # Restart auth service
make reset                    # Rebuild and restart

# Monitoring
make logs                     # View logs
make logs-follow              # Follow logs in real-time
make status                   # Show container status
make health                   # Check service health

# Testing
make test                     # Run comprehensive API tests
make test-basic               # Basic connectivity test
make test-login               # Test admin login

# Information
make info                     # Show service URLs and API docs
make help                     # Show all available commands
```

### Manual Docker Commands

If you prefer to use Docker directly:

```bash
# Build and start
cd docker
docker-compose build
docker-compose up -d

# View logs
docker-compose logs -f session-service

# Stop
docker-compose down
```

## 📁 Project Structure

```
session-service/
├── docker/
│   ├── Dockerfile             # Container build configuration
│   └── docker-compose.yml     # Service orchestration
├── scripts/
│   ├── start.sh              # Start service script
│   ├── stop.sh               # Stop service script
│   ├── reset.sh              # Reset service script
│   ├── logs.sh               # View logs script
│   └── test.sh               # API testing script
├── handler/
│   ├── middleware/
│   │   └── auth.go           # Authentication middleware
│   └── auth_handler.go       # HTTP handlers and business logic
├── models/
│   └── models.go             # Data models and structures
├── sql/
│   ├── scripts/              # SQL query files (auto-discovered)
│   │   ├── get_user_profile_by_username.sql
│   │   ├── get_user_profile_by_id.sql
│   │   ├── get_user_permissions.sql
│   │   └── update_last_login.sql
│   └── queries.go            # Dynamic SQL query loader
├── utils/
│   ├── jwt.go                # JWT token management
│   └── password.go           # Password hashing utilities
├── config/
│   └── config.go             # Configuration management
├── main.go                   # Application entry point
├── Makefile                  # Development automation
├── go.mod                    # Go module definition
├── config.env.example       # Environment variables template
├── docker.env.example       # Docker environment template
└── README.md                 # This file
```

## 🔧 Configuration

### Environment Variables

The service uses environment variables for configuration. Available options:

```bash
# Server Configuration
AUTH_SERVER_HOST=0.0.0.0        # Server bind address
AUTH_SERVER_PORT=8081            # Server port

# JWT Configuration
JWT_SECRET=your-secret-key       # JWT signing secret (CHANGE IN PRODUCTION!)
JWT_EXPIRATION_TIME=10m          # Token expiration time
JWT_REFRESH_THRESHOLD=2m         # When to allow token refresh

# Database Configuration
DB_HOST=postgres                 # Database host (container name)
DB_PORT=5432                     # Database port
DB_USER=postgres                 # Database user
DB_PASSWORD=postgres123          # Database password
DB_NAME=icecream_store          # Database name
DB_SSLMODE=disable              # SSL mode

# Security Configuration
BCRYPT_COST=12                   # bcrypt hashing cost
MAX_LOGIN_ATTEMPTS=5             # Maximum login attempts
LOGIN_COOLDOWN_TIME=15m          # Cooldown after max attempts

# Logging Configuration
LOG_LEVEL=info                   # Log level (debug, info, warn, error)
```

### Docker Environment

For Docker deployments, copy `docker.env.example` to `docker/.env`:

```bash
cp docker.env.example docker/.env
# Edit docker/.env with your settings
```

## 🏃‍♂️ Running the Service

### Docker (Recommended)

```bash
# Quick start (ensures database is running)
make install

# Or step by step
make deps                        # Install Go dependencies  
make build                       # Build Docker image
make start                       # Start service
make test                        # Run tests
```

### Development Mode (Without Docker)

```bash
# Prerequisites: Ensure data-service is running
cd ../data-service && make start

# Install dependencies
go mod tidy

# Copy environment configuration
cp config.env.example .env
# Edit .env with your settings

# Run the service
make dev
# or
go run main.go
```

### Production

```bash
# Build production image
docker build -f docker/Dockerfile -t icecream-auth:latest .

# Run with production configuration
docker run -d \
  --name icecream-auth \
  --network icecream_network \
  -p 8081:8081 \
  -e JWT_SECRET=your-production-secret \
  -e LOG_LEVEL=warn \
  icecream-auth:latest
```

## 📡 API Endpoints

### Public Endpoints (No Authentication Required)

#### Login
```http
POST /api/v1/sessions/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "username": "admin",
    "full_name": "System Administrator",
    "role_id": "uuid",
    "is_active": true
  },
  "role": {
    "id": "uuid",
    "role_name": "super_admin",
    "description": "Full system access and control"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Health Check
```http
GET /api/v1/sessions/health
```

**Response:**
```json
{
  "success": true,
  "message": "Auth service is healthy",
  "data": {
    "service": "session-service",
    "status": "healthy",
    "time": "2025-07-27T15:30:00Z"
  }
}
```

### Protected Endpoints (Authentication Required)

Include the JWT token in the Authorization header:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Logout
```http
POST /api/v1/sessions/logout
Authorization: Bearer <token>
```

#### Token Refresh
```http
POST /api/v1/sessions/refresh
Content-Type: application/json
Authorization: Bearer <token>

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Validate Token
```http
POST /api/v1/sessions/validate
Authorization: Bearer <token>
```

#### Get Profile
```http
GET /api/v1/sessions/profile
Authorization: Bearer <token>
```

### Admin Endpoints (Requires admin-read permission)

#### Get Token Info
```http
GET /api/v1/sessions/token-info
Authorization: Bearer <token>
```

## 🔐 Authentication Flow

### 1. Login Process
1. Client sends username/password to `/api/v1/sessions/login`
2. Service validates credentials against database
3. If valid, generates JWT token with user roles/permissions
4. Returns token with user and role information

### 2. Token Usage
1. Client includes token in `Authorization: Bearer <token>` header
2. Service validates token on protected endpoints
3. Extracts user information and permissions from token
4. Allows/denies access based on required permissions

### 3. Token Refresh
1. When token is within refresh threshold, client can refresh it
2. Send current token to `/api/v1/sessions/refresh`
3. Receive new token with extended expiration

## 🧪 Testing

### Automated Tests

```bash
# Run comprehensive test suite
make test

# Basic connectivity test
make test-basic

# Test admin login specifically
make test-login
```

### Manual Testing with curl

```bash
# Health check
curl http://localhost:8081/api/v1/sessions/health

# Login
curl -X POST http://localhost:8081/api/v1/sessions/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Use the returned token
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Validate token
curl -X POST http://localhost:8081/api/v1/sessions/validate \
  -H "Authorization: Bearer $TOKEN"

# Get profile
curl -X GET http://localhost:8081/api/v1/sessions/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Container Health Monitoring

```bash
# Check container health
make health

# View detailed container status
make status

# Follow logs in real-time
make logs-follow
```

## 🛡️ Security Features

- **Password Hashing**: bcrypt with configurable cost factor
- **JWT Security**: HMAC-SHA256 signing with secure secret
- **Token Expiration**: Configurable token lifetime (default: 10 minutes)
- **Permission-Based Access**: Granular role-based permissions
- **Request Logging**: All requests are logged for audit
- **CORS Protection**: Configurable cross-origin policies
- **Container Security**: Non-root user, minimal Alpine Linux base
- **Health Checks**: Built-in container and service health monitoring

## 🔌 Integration with Other Services

### Using the Auth Middleware

Other services can validate tokens by calling the auth service:

```go
// Validate token
resp, err := http.Post("http://session-service:8081/api/v1/sessions/validate", 
    headers: {"Authorization": "Bearer " + token})

// Check permissions in JWT claims
// The validated token contains user roles and permissions
```

### JWT Token Claims

The JWT token contains the following claims:

```json
{
  "user_id": "uuid",
  "username": "admin",
  "full_name": "System Administrator", 
  "role_id": "uuid",
  "role_name": "super_admin",
  "permissions": ["inventory-read", "inventory-write", ...],
  "iat": 1234567890,
  "exp": 1234567890,
  "sub": "uuid",
      "iss": "icecream-session-service",
  "aud": ["icecream-store"]
}
```

## 🐳 Docker Integration

### Service Dependencies

The auth service automatically connects to the existing data-service network:

```yaml
# The auth service joins the existing network
networks:
  icecream_network:
    external: true
```

### Multi-Service Setup

```bash
# Start database first
cd ../data-service
make start

# Start session service
cd ../session-service  
make start

# Both services now communicate via Docker network
curl http://localhost:8081/api/v1/sessions/health
```

### Docker Network Communication

- **Database**: `postgres:5432` (container name)
- **Auth Service**: `icecream_auth:8081` (container name)
- **External Access**: `localhost:8081` (host port mapping)

## 🔍 Monitoring and Logging

### Log Levels
- **DEBUG**: Detailed flow information
- **INFO**: General operational entries  
- **WARN**: Warning conditions
- **ERROR**: Error conditions

### Log Format
```json
{
  "level": "info",
  "msg": "User logged in successfully",
  "time": "2025-07-27T15:30:00Z",
  "user_id": "uuid",
  "username": "admin",
  "role": "super_admin"
}
```

### Container Logs

```bash
# View logs
make logs

# Follow logs in real-time
make logs-follow

# View specific time range
docker logs icecream_auth --since="1h"
```

## 🚨 Error Handling

### Error Response Format
```json
{
  "error": "invalid_credentials",
  "message": "Invalid username or password",
  "code": "invalid_credentials"
}
```

### Common Error Codes
- `invalid_request`: Malformed request
- `missing_credentials`: Username/password missing
- `invalid_credentials`: Wrong username/password
- `user_inactive`: User account disabled
- `missing_token`: Authorization token missing
- `invalid_token`: Token invalid or expired
- `insufficient_permissions`: User lacks required permission

## 📈 Performance Considerations

- **Connection Pooling**: Database connections are pooled for efficiency
- **JWT Stateless**: No server-side session storage required
- **Container Resources**: Lightweight Alpine Linux base image
- **Health Checks**: Automatic container restart on failure
- **Request Logging**: Configurable log levels to control verbosity

## 🚨 Troubleshooting

### Service Won't Start

```bash
# Check if database is running
cd ../data-service && make health

# Check Docker network
docker network ls | grep icecream

# Check auth service logs  
make logs

# Reset everything
make reset
```

### Database Connection Issues

```bash
# Test database connectivity
docker run --rm --network icecream_network postgres:15-alpine \
  pg_isready -h postgres -p 5432 -U postgres

# Check auth service can reach database
make test-basic
```

### Authentication Issues

```bash
# Test admin login
make test-login

# Check database has users
cd ../data-service && make connect
# Then: SELECT * FROM users;
```

### Container Health Issues

```bash
# Check container health
docker inspect icecream_auth --format='{{.State.Health.Status}}'

# Force health check
docker exec icecream_auth wget --spider http://localhost:8081/api/v1/sessions/health
```

## 🔄 Next Steps

1. **Rate Limiting**: Implement rate limiting for login endpoints
2. **Token Blacklisting**: Implement proper logout with token invalidation
3. **Password Policies**: Add password complexity requirements
4. **Two-Factor Authentication**: Add 2FA support
5. **Audit Logging**: Enhanced audit trail for security events
6. **Load Balancing**: Support for multiple auth service instances

## 🤝 Contributing

When making changes to the auth service:

1. Update configuration if new environment variables are added
2. Update this README for new endpoints or features
3. Ensure all new endpoints include proper error handling
4. Add appropriate logging for security events
5. Test with the data-service integration
6. Update Docker configuration as needed

---

**Ready to secure your ice cream empire!** 🍦🔐 