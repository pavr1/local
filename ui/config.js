// Ice Cream Store UI Configuration
// This file contains service URLs and API configuration

const CONFIG = {
    // Gateway URL - single entry point for all API calls
    GATEWAY_URL: 'http://localhost:8082',
    
    // API endpoints (all go through gateway)
    API: {
        LOGIN: '/api/v1/auth/login',
        LOGOUT: '/api/v1/auth/logout',
        VALIDATE: '/api/v1/auth/validate',
        PROFILE: '/api/v1/auth/profile',
        HEALTH: '/api/health'
    },
    
    // Authentication settings
    AUTH: {
        TOKEN_KEY: 'icecream_auth_token',
        USER_KEY: 'icecream_user_data', 
        REMEMBER_KEY: 'icecream_remember_me'
    },
    
    // Service health check URLs (for status indicators)
    SERVICES: {
        gateway: {
            name: 'Gateway Service',
            url: 'http://localhost:8082/api/health',
            element: 'gateway-status'
        },
        auth: {
            name: 'Session Service', 
            url: 'http://localhost:8081/api/v1/auth/health',
            element: 'auth-status'
        },
        orders: {
            name: 'Orders Service',
            url: 'http://localhost:8083/api/v1/orders/health', 
            element: 'orders-status'
        },
        inventory: {
            name: 'Inventory Service',
            url: 'http://localhost:8084/api/v1/inventory/health',
            element: 'inventory-status'
        }
    }
};

// Make config available globally
window.CONFIG = CONFIG; 