// Ice Cream Store UI Configuration
// This file contains service URLs and API configuration

console.log('🔧 CONFIG.JS: Starting to load...');

// Simple localhost configuration for local testing
const SERVICE_URLS = {
    gateway: 'http://localhost:8082',
    session: 'http://localhost:8081', 
    orders: 'http://localhost:8083',
    inventory: 'http://localhost:8084'
};

console.log('🔧 SERVICE_URLS: Using localhost for all services =', SERVICE_URLS);

const CONFIG = {
    // Gateway URL - single entry point for all API calls
    GATEWAY_URL: SERVICE_URLS.gateway,
    
    // API endpoints (all go through gateway)
    API: {
        LOGIN: '/api/v1/auth/login',
        LOGOUT: '/api/v1/auth/logout',
        VALIDATE: '/api/v1/sessions/validate',
        PROFILE: '/api/v1/auth/profile',
        HEALTH: '/api/health'
    },
    
    // Authentication settings
    AUTH: {
        TOKEN_KEY: 'icecream_auth_token',
        USER_KEY: 'icecream_user_data', 
        REMEMBER_KEY: 'icecream_remember_me'
    },
    
    // Service health check URLs (all go through gateway to avoid CORS)
    SERVICES: {
        gateway: {
            name: 'Gateway Service',
            url: SERVICE_URLS.gateway + '/api/health',
            element: 'gateway-status'
        },
        auth: {
            name: 'Session Service', 
            url: SERVICE_URLS.gateway + '/api/v1/auth/health',
            element: 'auth-status'
        },
        orders: {
            name: 'Orders Service',
            url: SERVICE_URLS.gateway + '/api/v1/orders/health', 
            element: 'orders-status'
        },
        inventory: {
            name: 'Inventory Service',
            url: SERVICE_URLS.gateway + '/api/v1/inventory/health',
            element: 'inventory-status'
        }
    }
};

console.log('🔧 CONFIG.JS: CONFIG object created');
console.log('🔧 CONFIG.AUTH =', CONFIG.AUTH);
console.log('🔧 CONFIG.GATEWAY_URL =', CONFIG.GATEWAY_URL);
console.log('🔧 Health check URLs:', {
    auth: CONFIG.SERVICES.auth.url,
    orders: CONFIG.SERVICES.orders.url,
    gateway: CONFIG.SERVICES.gateway.url,
    inventory: CONFIG.SERVICES.inventory.url
});

// Make config available globally
window.CONFIG = CONFIG;
console.log('🔧 CONFIG.JS: Loading complete! CONFIG.AUTH =', CONFIG.AUTH); 