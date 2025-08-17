# Configuration Values Documentation

This document provides a comprehensive overview of all configuration values used across the Ice Cream Store microservices architecture.

## Table of Contents

1. [General Configuration Values](#general-configuration-values)
   - [Database Configuration](#database-configuration)
   - [Logging Configuration](#logging-configuration)
   - [JWT Configuration](#jwt-configuration)
   - [CORS Configuration](#cors-configuration)
   - [Business Configuration](#business-configuration)
2. [Service-Specific Configuration](#service-specific-configuration)
   - [Gateway Service](#gateway-service)
   - [Session Service](#session-service)
   - [Orders Service](#orders-service)
   - [Inventory Service](#inventory-service)
   - [Invoice Service](#invoice-service)
   - [Data Service](#data-service)
   - [UI Service](#ui-service)
3. [Environment Variables Summary](#environment-variables-summary)

# General Configuration Values

These configuration values are shared across multiple services and should be configured consistently.

## Database Configuration

**Used by**: All services (Session, Orders, Inventory, Invoice, Data)  
**Default Database**: PostgreSQL

| Environment Variable | Default Value | Description | Required |
|---------------------|---------------|-------------|----------|
| `DB_HOST` | `localhost` | Database host address | Yes |
| `DB_PORT` | `5432` | Database port number | Yes |
| `DB_USER` | `postgres` | Database username | Yes |
| `DB_PASSWORD` | `postgres123` | Database password | Yes |
| `DB_NAME` | `icecream_store` | Database name | Yes |
| `DB_SSL_MODE` / `DB_SSLMODE` | `disable` | Database SSL mode (disable, require, verify-ca, verify-full) | Yes |

### Database Connection Pool Settings (Data Service)
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `DB_MAX_OPEN_CONNS` | `25` | Maximum number of open connections |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum number of idle connections |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maximum lifetime of connections |
| `DB_CONN_MAX_IDLE_TIME` | `5m` | Maximum idle time of connections |

### Database Timeout Settings (Data Service)
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `DB_CONNECT_TIMEOUT` | `10s` | Database connection timeout |
| `DB_QUERY_TIMEOUT` | `30s` | Database query timeout |

### Database Retry Settings (Data Service)
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `DB_MAX_RETRIES` | `3` | Maximum number of retry attempts |
| `DB_RETRY_INTERVAL` | `1s` | Interval between retry attempts |

---

## Logging Configuration

**Used by**: All services  
**Default Logger**: Structured JSON logging

| Environment Variable | Default Value | Description | Options |
|---------------------|---------------|-------------|---------|
| `LOG_LEVEL` | `info` | Logging level | debug, info, warn, error |
| `LOG_FORMAT` | `json` | Log format (Data Service only) | json, text |

### Centralized Logging (Gateway Service)
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `FLUENTD_HOST` | `localhost` | Fluentd host for centralized logging |
| `FLUENTD_PORT` | `24224` | Fluentd port for centralized logging |

---

## JWT Configuration

**Used by**: Session Service, Orders Service  
**Purpose**: Authentication and authorization

| Environment Variable | Default Value | Description | Security Note |
|---------------------|---------------|-------------|---------------|
| `JWT_SECRET` | `icecream-super-secret-jwt-key-change-in-production-2024` | JWT signing secret | **MUST CHANGE IN PRODUCTION** |
| `JWT_EXPIRATION_TIME` | `30m` | JWT token expiration time | 30 minutes default |

---

## CORS Configuration

**Used by**: Gateway Service, UI Service  
**Purpose**: Cross-Origin Resource Sharing

| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `CORS_ALLOWED_ORIGINS` | `*` | Allowed CORS origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | `Content-Type,Authorization` | Allowed HTTP headers |

---

## Business Configuration

**Used by**: Orders Service  
**Purpose**: Business logic and tax calculations

| Environment Variable | Default Value | Description | Business Context |
|---------------------|---------------|-------------|------------------|
| `DEFAULT_TAX_RATE` | `13.0` | Default tax rate percentage | Costa Rica IVA |
| `DEFAULT_SERVICE_RATE` | `10.0` | Default service charge rate percentage | Service fee |
| `ORDER_TIMEOUT` | `30` | Order timeout in minutes | Order processing |

---

# Service-Specific Configuration

These configuration values are specific to individual services.

## Gateway Service

**Service Port**: 8082  
**Configuration File**: `gateway-service/main.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `GATEWAY_PORT` | `8082` | Port for the gateway service |
| `GATEWAY_SERVER_HOST` | `0.0.0.0` | Host binding for the gateway service |

### Service URLs (Internal Communication)
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `SESSION_SERVICE_URL` | `http://localhost:8081` | URL for session service |
| `ORDERS_SERVICE_URL` | `http://localhost:8083` | URL for orders service |
| `INVENTORY_SERVICE_URL` | `http://localhost:8084` | URL for inventory service |
| `INVOICE_SERVICE_URL` | `http://localhost:8085` | URL for invoice service |

**Note**: Gateway service uses general [Logging](#logging-configuration) and [CORS](#cors-configuration) configurations.

---

## Session Service

**Service Port**: 8081  
**Configuration File**: `session-service/config/config.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `SESSION_SERVER_PORT` | `8081` | Port for the session service |
| `SESSION_SERVER_HOST` | `0.0.0.0` | Host binding for the session service |

**Note**: Session service uses general [Database](#database-configuration), [JWT](#jwt-configuration), and [Logging](#logging-configuration) configurations.

---

## Orders Service

**Service Port**: 8083  
**Configuration File**: `orders-service/config/config.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Host binding for the orders service |
| `SERVER_PORT` | `8083` | Port for the orders service |

**Note**: Orders service uses general [Database](#database-configuration), [JWT](#jwt-configuration), [Business](#business-configuration), and [Logging](#logging-configuration) configurations.

---

## Inventory Service

**Service Port**: 8084  
**Configuration File**: `inventory-service/config/config.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `INVENTORY_SERVER_PORT` | `8084` | Port for the inventory service |
| `INVENTORY_SERVER_HOST` | `0.0.0.0` | Host binding for the inventory service |

**Note**: Inventory service uses general [Database](#database-configuration) and [Logging](#logging-configuration) configurations.

---

## Invoice Service

**Service Port**: 8085  
**Configuration File**: `invoice-service/config/config.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `INVOICE_SERVER_PORT` | `8085` | Port for the invoice service |
| `INVOICE_SERVER_HOST` | `0.0.0.0` | Host binding for the invoice service |

### Service Communication
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `INVENTORY_SERVICE_URL` | `http://localhost:8084` | URL for inventory service |

**Note**: Invoice service uses general [Database](#database-configuration) and [Logging](#logging-configuration) configurations.

---

## Data Service

**Service Port**: 8086  
**Configuration File**: `data-service/main.go`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `DATA_SERVER_PORT` | `8086` | Port for the data service |
| `DATA_SERVER_HOST` | `0.0.0.0` | Host binding for the data service |

**Note**: Data service uses general [Database](#database-configuration) and [Logging](#logging-configuration) configurations, plus additional database connection pool, timeout, and retry settings as shown in the general database configuration section.

---

## UI Service

**Service Port**: 3000  
**Configuration File**: `ui/config.js`

### Server Configuration
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `UI_PORT` | `3000` | Port for the UI service |

### Service URLs
| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `GATEWAY_URL` | `http://icecream_gateway:8082` | Gateway service URL |

### Authentication Configuration
| Configuration Key | Default Value | Description |
|------------------|---------------|-------------|
| `SESSION_ID_KEY` | `icecream_session_id` | Session ID storage key |
| `USER_KEY` | `icecream_user_data` | User data storage key |
| `REMEMBER_KEY` | `icecream_remember_me` | Remember me storage key |

**Note**: UI service uses general [CORS](#cors-configuration) configuration.


---

## Environment Variables Summary

### Development Environment
For local development, services typically use:
- **Database**: `localhost:5432`
- **Service URLs**: `http://localhost:PORT`
- **SSL Mode**: `disable`

### Production Environment
For production/Docker deployment:
- **Database**: `icecream_postgres:5432`
- **Service URLs**: `http://service_name:PORT`
- **SSL Mode**: `require` (recommended)

### Security Considerations
1. **JWT Secret**: Must be changed in production
2. **Database Password**: Use strong passwords in production
3. **SSL Mode**: Enable SSL for production databases
4. **CORS**: Restrict origins in production

### Configuration Files
- `config.env.example`: Template files for each service
- `config.env`: Actual configuration files (not in version control)
- `docker-compose.yml`: Docker environment variables
- `config.js`: UI service configuration

### Port Summary
| Service | Port | Description |
|---------|------|-------------|
| Gateway | 8082 | API Gateway and routing |
| Session | 8081 | Authentication and session management |
| Orders | 8083 | Order processing and management |
| Inventory | 8084 | Inventory and recipe management |
| Invoice | 8085 | Invoice and expense management |
| Data | 8086 | Database operations |
| UI | 3000 | Web interface |

---

## Notes

1. **Environment Detection**: The UI service automatically detects local vs production environment
2. **Service Discovery**: Services communicate through the gateway in production
3. **Configuration Inheritance**: Services inherit common database and logging configurations
4. **Validation**: All configuration values are validated at service startup
5. **Hot Reload**: Some services support configuration hot-reloading (check individual service docs)

For specific service configuration details, refer to the individual service documentation in their respective directories.
