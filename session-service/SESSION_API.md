# Session Management API Documentation

This document describes the REST API endpoints for session management that the **Gateway Service** will use to manage user sessions.

## Base URL
```
http://localhost:8081/api/v1/sessions
```

## Authentication
- **Public Endpoints**: No authentication required
- **Internal Endpoints**: For gateway use (could add API key protection later)
- **Protected Endpoints**: Require valid JWT token in Authorization header

---

## 📍 **API Endpoints**

### **Public Endpoints**

#### 1. Health Check
```http
GET /api/v1/sessions/health
```

**Description**: Check if the session service is operational.

**Response**:
```json
{
  "success": true,
  "service": "session-service",
  "status": "healthy",
  "message": "Session service is operational"
}
```

#### 2. Validate Session
```http
POST /api/v1/sessions/validate
```

**Description**: Validate a JWT token against stored sessions.

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (Valid Session)**:
```json
{
  "is_valid": true,
  "session": {
    "session_id": "abc123...",
    "user_id": "user-uuid",
    "username": "john_doe",
    "role_name": "admin",
    "permissions": ["read", "write", "admin"],
    "created_at": "2024-01-15T10:00:00Z",
    "expires_at": "2024-01-15T10:30:00Z",
    "last_activity": "2024-01-15T10:15:00Z"
  },
  "should_refresh": false,
  "new_token": "eyJ..." // Only present if should_refresh is true
}
```

**Response (Invalid Session)**:
```json
{
  "is_valid": false,
  "error_code": "session_expired",
  "error_message": "Session has expired"
}
```

---

### **Internal Endpoints (For Gateway Use)**

#### 3. Create Session
```http
POST /api/v1/sessions
```

**Description**: Create a new session after successful user login.

**Request Body**:
```json
{
  "user_id": "user-uuid",
  "username": "john_doe",
  "role_name": "admin",
  "permissions": ["read", "write", "admin"],
  "remember_me": false,
  "expires_at": "2024-01-15T10:30:00Z" // Optional, will use default if not provided
}
```

**Response**:
```json
{
  "success": true,
  "message": "Session created successfully",
  "session_id": "abc123...",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-15T10:30:00Z",
  "user": {
    "id": "user-uuid",
    "username": "john_doe",
    "role": "admin"
  }
}
```

#### 4. Refresh Session
```http
POST /api/v1/sessions/refresh
Authorization: Bearer <jwt_token>
```

**Description**: Refresh a JWT token if it's within the refresh threshold.

**Response**:
```json
{
  "success": true,
  "message": "Session refreshed successfully",
  "token": "eyJ...", // New token if refreshed
  "refreshed": true
}
```

#### 5. Logout (Revoke by Token)
```http
POST /api/v1/sessions/logout
```

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Alternative**: Send token in Authorization header:
```http
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "success": true,
  "message": "Session revoked successfully"
}
```

#### 6. Session Statistics
```http
GET /api/v1/sessions/stats
```

**Description**: Get basic session analytics.

**Response**:
```json
{
  "success": true,
  "stats": {
    "total_sessions": 150,
    "active_sessions": 45,
    "expired_sessions": 0
  }
}
```

---

### **Protected Endpoints (Require Authentication)**

#### 7. Get User Sessions
```http
GET /api/v1/sessions/user/{userID}
Authorization: Bearer <jwt_token>
```

**Description**: Get all active sessions for a specific user.

**Response**:
```json
{
  "success": true,
  "user_id": "user-uuid",
  "sessions": [
    {
      "session_id": "abc123...",
      "created_at": "2024-01-15T10:00:00Z",
      "last_activity": "2024-01-15T10:15:00Z",
      "is_active": true,
      "is_current": true
    },
    {
      "session_id": "def456...",
      "created_at": "2024-01-15T09:00:00Z",
      "last_activity": "2024-01-15T09:30:00Z",
      "is_active": true,
      "is_current": false
    }
  ],
  "count": 2
}
```

#### 8. Revoke Specific Session
```http
DELETE /api/v1/sessions/{sessionID}
Authorization: Bearer <jwt_token>
```

**Description**: Revoke a specific session by session ID.

**Response**:
```json
{
  "success": true,
  "message": "Session revoked successfully",
  "session_id": "abc123..."
}
```

#### 9. Revoke All User Sessions
```http
DELETE /api/v1/sessions/user/{userID}
Authorization: Bearer <jwt_token>
```

**Query Parameters**:
- `exclude_current=true` - Exclude the current session from revocation

**Description**: Revoke all sessions for a user (optionally excluding current session).

**Response**:
```json
{
  "success": true,
  "message": "User sessions revoked successfully",
  "user_id": "user-uuid",
  "revoked_count": 3,
  "total_sessions": 4
}
```

---

## 🔄 **Gateway Integration Examples**

### Login Flow
```javascript
// 1. Authenticate user credentials (gateway handles this)
// 2. Create session in session service
const sessionResponse = await fetch('http://localhost:8081/api/v1/sessions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    user_id: userProfile.id,
    username: userProfile.username,
    role_name: userProfile.role,
    permissions: userProfile.permissions,
    remember_me: false
  })
});

const sessionData = await sessionResponse.json();
// Return sessionData.token to client
```

### Request Validation Flow
```javascript
// For every protected request
const validationResponse = await fetch('http://localhost:8081/api/v1/sessions/validate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    token: requestToken
  })
});

const validation = await validationResponse.json();

if (validation.is_valid) {
  // Allow request to proceed
  // If validation.new_token exists, return it to client for token refresh
  const userContext = validation.session;
} else {
  // Reject request with 401 Unauthorized
}
```

### Logout Flow
```javascript
// When user logs out
await fetch('http://localhost:8081/api/v1/sessions/logout', {
  method: 'POST',
  headers: { 
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${userToken}`
  }
});
```

---

## 🔧 **Configuration**

Session behavior can be configured via environment variables:

```bash
# Session timing
SESSION_DEFAULT_EXPIRATION=30m
SESSION_REMEMBER_ME_EXPIRATION=168h  # 7 days
JWT_REFRESH_THRESHOLD=5m
SESSION_CLEANUP_INTERVAL=10m

# Session limits
SESSION_MAX_CONCURRENT=5

# Storage
SESSION_STORAGE_TYPE=memory  # memory, redis, database
```

---

## 📊 **Error Codes**

| Error Code | Description |
|------------|-------------|
| `invalid_request` | Malformed request body |
| `missing_fields` | Required fields missing |
| `missing_token` | JWT token not provided |
| `session_not_found` | Session doesn't exist |
| `session_expired` | Session has expired |
| `session_inactive` | Session is not active |
| `validation_error` | Internal validation error |
| `session_creation_failed` | Failed to create session |

---

## 🚀 **Benefits for Gateway**

1. **Enhanced Security**: Server-side session validation prevents token replay attacks
2. **Automatic Token Refresh**: Seamless user experience with background token renewal
3. **Session Management**: Users can view and revoke their sessions
4. **Centralized Control**: Single point for session lifecycle management
5. **Analytics**: Session statistics for monitoring and optimization

---

## 🔗 **Legacy Compatibility**

The service maintains backward compatibility with existing auth endpoints at `/api/v1/auth/*` for gradual migration. 