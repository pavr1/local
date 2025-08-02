// Ice Cream Store UI Configuration
// This file contains service URLs and API configuration

console.log('🔧 CONFIG.JS: Starting to load...');

// Environment detection and service URL configuration
function getServiceUrls() {
    // Force localhost for development
    const isLocalDevelopment = window.location.hostname === 'localhost' || 
                              window.location.hostname === '127.0.0.1' ||
                              window.location.hostname.includes('localhost');
    
    if (isLocalDevelopment) {
        console.log('🔧 Detected local development environment - using localhost URLs');
        return {
            gateway: 'http://localhost:8082',
            session: 'http://localhost:8081', 
            orders: 'http://localhost:8083',
            inventory: 'http://localhost:8084',
            expense: 'http://localhost:8085'
        };
    } else {
        // Production/Docker environment
        console.log('🔧 Detected production environment - using Docker service names');
        return {
            gateway: 'http://icecream_gateway:8082',
            session: 'http://icecream_session:8081', 
            orders: 'http://icecream_orders:8083',
            inventory: 'http://icecream_inventory:8084',
            expense: 'http://icecream_expense:8085'
        };
    }
}

const SERVICE_URLS = getServiceUrls();
console.log('🔧 SERVICE_URLS: Using URLs for current environment =', SERVICE_URLS);

const CONFIG = {
    // Gateway URL - single entry point for all API calls
    GATEWAY_URL: SERVICE_URLS.gateway,
    
    // API endpoints (all go through gateway)
    API: {
        LOGIN: '/api/v1/sessions/p/login',
        LOGOUT: '/api/v1/sessions/p/logout',
        VALIDATE: '/api/v1/sessions/p/validate',
        PROFILE: '/api/v1/sessions/profile',
        HEALTH: '/api/v1/sessions/p/health'  // Default health endpoint through gateway
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
        session: {
            name: 'Session Service', 
            url: SERVICE_URLS.gateway + '/api/v1/sessions/p/health',
            element: 'session-status'
        },
        orders: {
            name: 'Orders Service',
            url: SERVICE_URLS.gateway + '/api/v1/orders/p/health', 
            element: 'orders-status'
        },
        inventory: {
            name: 'Inventory Service',
            url: SERVICE_URLS.gateway + '/api/v1/inventory/p/health',
            element: 'inventory-status'
        },
        expenses: {
            name: 'Expense Service',
            url: SERVICE_URLS.gateway + '/api/v1/expenses/p/health',
            element: 'expenses-status'
        }
    }
};

console.log('🔧 CONFIG.JS: CONFIG object created');
console.log('🔧 CONFIG.AUTH =', CONFIG.AUTH);
console.log('🔧 CONFIG.GATEWAY_URL =', CONFIG.GATEWAY_URL);
console.log('🔧 Health check URLs:', {
    session: CONFIG.SERVICES.session.url,
    orders: CONFIG.SERVICES.orders.url,
    gateway: CONFIG.SERVICES.gateway.url,
    inventory: CONFIG.SERVICES.inventory.url,
    expenses: CONFIG.SERVICES.expenses.url
});

// Make config available globally
window.CONFIG = CONFIG;
console.log('🔧 CONFIG.JS: Loading complete! CONFIG.AUTH =', CONFIG.AUTH); 