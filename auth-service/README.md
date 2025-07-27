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

## 📁 Project Structure

```
auth-service/
├── config/
│   └── config.go              # Configuration management
├── handler/
│   └── auth_handler.go        # HTTP handlers and business logic
├── models/
│   └── auth.go                # Data models and structures
├── middleware/                # (Future middleware implementations)
├── utils/
│   ├── jwt.go                 # JWT token management
│   └── password.go            # Password hashing utilities
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── config.env.example        # Environment variables template
└── README.md                  # This file
```

## 🔧 Configuration

The service uses environment variables for configuration. Copy `config.env.example` to `.env` and adjust values:

```bash
# Server Configuration
AUTH_SERVER_HOST=0.0.0.0        # Server bind address
AUTH_SERVER_PORT=8081            # Server port

# JWT Configuration
JWT_SECRET=your-secret-key       # JWT signing secret (CHANGE IN PRODUCTION!)
JWT_EXPIRATION_TIME=10m          # Token expiration time
JWT_REFRESH_THRESHOLD=2m         # When to allow token refresh

# Database Configuration
DB_HOST=localhost                # Database host
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

## 🏃‍♂️ Running the Service

### Prerequisites

1. **Data Service**: Ensure the PostgreSQL database is running (see data-service)
2. **Go 1.21+**: Required for building and running

### Development

```bash
# Clone and navigate to auth-service
cd auth-service

# Install dependencies
go mod tidy

# Copy environment configuration
cp config.env.example .env
# Edit .env with your settings

# Run the service
go run main.go
```

### Production

```bash
# Build the service
go build -o auth-service main.go

# Run the binary
./auth-service
```

## 📡 API Endpoints

### Public Endpoints (No Authentication Required)

#### Login
```http
POST /api/v1/auth/login
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
  "permissions": [
    {"permission_name": "inventory-read", "description": "View inventory data"},
    {"permission_name": "inventory-write", "description": "Modify inventory data"}
  ],
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-07-27T16:30:00Z",
  "refresh_at": "2025-07-27T16:28:00Z"
}
```

#### Health Check
```http
GET /api/v1/auth/health
```

**Response:**
```json
{
  "success": true,
  "message": "Auth service is healthy",
  "data": {
    "service": "auth-service",
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
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

#### Token Refresh
```http
POST /api/v1/auth/refresh
Content-Type: application/json
Authorization: Bearer <token>

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Validate Token
```http
GET /api/v1/auth/validate
Authorization: Bearer <token>
```

#### Get Profile
```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### Admin Endpoints (Requires admin-read permission)

#### Get Token Info
```http
GET /api/v1/auth/token-info
Authorization: Bearer <token>
```

## 🔐 Authentication Flow

### 1. Login Process
1. Client sends username/password to `/api/v1/auth/login`
2. Service validates credentials against database
3. If valid, generates JWT token with user roles/permissions
4. Returns token with expiration and refresh information

### 2. Token Usage
1. Client includes token in `Authorization: Bearer <token>` header
2. Service validates token on protected endpoints
3. Extracts user information and permissions from token
4. Allows/denies access based on required permissions

### 3. Token Refresh
1. When token is within refresh threshold, client can refresh it
2. Send current token to `/api/v1/auth/refresh`
3. Receive new token with extended expiration

## 🛡️ Security Features

- **Password Hashing**: bcrypt with configurable cost factor
- **JWT Security**: HMAC-SHA256 signing with secure secret
- **Token Expiration**: Configurable token lifetime
- **Permission-Based Access**: Granular role-based permissions
- **Request Logging**: All requests are logged for audit
- **CORS Protection**: Configurable cross-origin policies

## 🔌 Integration with Other Services

### Using the Auth Middleware

Other services can validate tokens by calling the auth service:

```go
// Validate token
resp, err := http.Get("http://auth-service:8081/api/v1/auth/validate", 
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
  "iss": "icecream-auth-service",
  "aud": ["icecream-store"]
}
```

## 🐳 Docker Integration

The service is designed to work with the existing Docker network from data-service:

```yaml
# In your docker-compose.yml
auth-service:
  build: ./auth-service
  ports:
    - "8081:8081"
  environment:
    DB_HOST: postgres
    JWT_SECRET: your-production-secret
  depends_on:
    - postgres
  networks:
    - icecream_network
```

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

## 🧪 Testing

### Manual Testing with curl

```bash
# Login
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Use the returned token
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Validate token
curl -X GET http://localhost:8081/api/v1/auth/validate \
  -H "Authorization: Bearer $TOKEN"

# Get profile
curl -X GET http://localhost:8081/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

## 📈 Performance Considerations

- **Connection Pooling**: Database connections are pooled for efficiency
- **JWT Stateless**: No server-side session storage required
- **Token Caching**: Consider implementing token blacklist for logout
- **Rate Limiting**: Implement rate limiting for login endpoints

## 🔄 Next Steps

1. **Implement Database Queries**: Complete the user profile database methods
2. **Add Rate Limiting**: Protect against brute force attacks  
3. **Token Blacklisting**: Implement proper logout with token invalidation
4. **Password Policies**: Add password complexity requirements
5. **Two-Factor Authentication**: Add 2FA support
6. **Audit Logging**: Enhanced audit trail for security events

## 🤝 Contributing

When making changes to the auth service:

1. Update configuration if new environment variables are added
2. Update this README for new endpoints or features
3. Ensure all new endpoints include proper error handling
4. Add appropriate logging for security events
5. Test with the data-service integration

---

**Ready to secure your ice cream empire!** 🍦🔐 