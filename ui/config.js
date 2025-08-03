// Ice Cream Store UI Configuration
// This file contains service URLs and API configuration

console.log('🔧 CONFIG.JS: Starting to load...');

// Environment detection and service URL configuration
function getServiceUrls() {
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
            invoice: 'http://localhost:8085'
        };
    } else {
        console.log('🔧 Detected production environment - using Docker service names');
        return {
            gateway: 'http://icecream_gateway:8082',
            session: 'http://icecream_session:8081',
            orders: 'http://icecream_orders:8083',
            inventory: 'http://icecream_inventory:8084',
            invoice: 'http://icecream_invoice:8085'
        };
    }
}

const SERVICE_URLS = getServiceUrls();

// Configuration object with all service URLs and health endpoints
const CONFIG = {
    GATEWAY_URL: SERVICE_URLS.gateway,
    API: {
        gateway: SERVICE_URLS.gateway + '/api/v1',
        session: SERVICE_URLS.session + '/api/v1/session',
        LOGIN: '/api/v1/sessions/p/login',
        LOGOUT: '/api/v1/sessions/logout',
        VALIDATE: '/api/v1/sessions/p/validate'
    },
    SERVICES: {
        session: SERVICE_URLS.session + '/health',
        orders: SERVICE_URLS.gateway + '/api/v1/orders/p/health',
        inventory: SERVICE_URLS.gateway + '/api/v1/inventory/p/health',
        invoices: SERVICE_URLS.gateway + '/api/v1/invoices/p/health'  // Updated to use /invoices path in gateway
    },
    AUTH: {
        login: SERVICE_URLS.gateway + '/api/v1/sessions/p/login',
        logout: SERVICE_URLS.gateway + '/api/v1/sessions/logout',
        validate: SERVICE_URLS.gateway + '/api/v1/sessions/p/validate',
        tokenKey: 'icecream_auth_token',
        userKey: 'icecream_user_data',
        rememberKey: 'icecream_remember_me',
        TOKEN_KEY: 'icecream_auth_token',
        USER_KEY: 'icecream_user_data',
        REMEMBER_KEY: 'icecream_remember_me'
    }
};

console.log('🔧 Configuration loaded:', {
    gateway: SERVICE_URLS.gateway,
    session: SERVICE_URLS.session,
    orders: SERVICE_URLS.orders,
    inventory: SERVICE_URLS.inventory,
    invoice: SERVICE_URLS.invoice
});

// Export for global access
window.CONFIG = CONFIG; 