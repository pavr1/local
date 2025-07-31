// UI Configuration
// This file contains service URLs for different environments

const CONFIG = {
    // Service URLs (these will be updated during Docker build)
    GATEWAY_URL: 'http://localhost:8082',
    
    // Individual service URLs (for health checks)
    SERVICES: {
        session: {
            name: 'Session Service',
            url: 'http://localhost:8081/api/v1/sessions/health',
            element: 'status-auth'
        },
        orders: {
            name: 'Orders Service', 
            url: 'http://localhost:8083/api/v1/orders/health',
            element: 'status-orders'
        },
        inventory: {
            name: 'Suppliers Service',
            url: 'http://localhost:8084/api/v1/inventory/health',
            element: 'status-inventory'
        },
        gateway: {
            name: 'Gateway Service',
            url: 'http://localhost:8082/api/health',
            element: 'status-gateway'
        },
        database: {
            name: 'Database',
            url: 'http://localhost:8081/api/v1/sessions/health', // Proxy through session service
            element: 'status-data'
        }
    },
    
    // API endpoints (all go through gateway)
    API: {
        LOGIN: '/api/v1/sessions/login',
        LOGOUT: '/api/v1/sessions/logout',
        ORDERS: '/api/v1/orders',
        SUPPLIERS: '/api/v1/inventory/suppliers',
        HEALTH: '/api/health'
    }
};

// Make config available globally
window.CONFIG = CONFIG; 