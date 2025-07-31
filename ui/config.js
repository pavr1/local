// Ice Cream Store UI Configuration
// This file contains service URLs and API configuration

// Environment Detection
function detectEnvironment() {
    // Allow manual override via URL parameter or localStorage
    const urlParams = new URLSearchParams(window.location.search);
    const forceLocal = urlParams.get('env') === 'local' || localStorage.getItem('force_local_env') === 'true';
    const forceDocker = urlParams.get('env') === 'docker' || localStorage.getItem('force_docker_env') === 'true';
    
    if (forceLocal) {
        console.log('🔧 Environment manually set to LOCAL');
        return { isDocker: false, hostname: window.location.hostname, forced: 'local' };
    }
    
    if (forceDocker) {
        console.log('🔧 Environment manually set to DOCKER');
        return { isDocker: true, hostname: window.location.hostname, forced: 'docker' };
    }
    
    // Check if we're running in a Docker container
    const isDocker = window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1';
    
    // Check for Docker-specific indicators
    const hasDockerHostname = window.location.hostname.includes('docker') || 
                              window.location.hostname.includes('container');
    
    return {
        isDocker: isDocker || hasDockerHostname,
        hostname: window.location.hostname
    };
}

// Get environment-specific URLs
function getServiceUrls() {
    const env = detectEnvironment();
    
    if (env.isDocker) {
        console.log('🐳 Docker environment detected, using container hostnames');
        return {
            gateway: 'http://icecream_gateway:8082',
            session: 'http://icecream_session:8081',
            orders: 'http://icecream_orders:8083',
            inventory: 'http://icecream_inventory:8084'
        };
    } else {
        console.log('💻 Local environment detected, using localhost');
        return {
            gateway: 'http://localhost:8082',
            session: 'http://localhost:8081', 
            orders: 'http://localhost:8083',
            inventory: 'http://localhost:8084'
        };
    }
}

// Get the service URLs based on environment
const SERVICE_URLS = getServiceUrls();

const CONFIG = {
    // Gateway URL - single entry point for all API calls
    GATEWAY_URL: SERVICE_URLS.gateway,
    
    // Environment info
    ENVIRONMENT: detectEnvironment(),
    
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
    
    // Service health check URLs (all go through gateway to avoid CORS)
    SERVICES: {
        gateway: {
            name: 'Gateway Service',
            url: `${SERVICE_URLS.gateway}/api/health`,
            element: 'gateway-status'
        },
        auth: {
            name: 'Session Service', 
            url: `${SERVICE_URLS.gateway}/api/v1/auth/health`,
            element: 'auth-status'
        },
        orders: {
            name: 'Orders Service',
            url: `${SERVICE_URLS.gateway}/api/v1/orders/health`, 
            element: 'orders-status'
        },
        inventory: {
            name: 'Inventory Service',
            url: `${SERVICE_URLS.gateway}/api/v1/inventory/health`,
            element: 'inventory-status'
        }
    }
};

// Debug: Log the config to verify it's correct
console.log('🔧 CONFIG loaded for environment:', {
    environment: CONFIG.ENVIRONMENT,
    GATEWAY_URL: CONFIG.GATEWAY_URL,
    API_VALIDATE: CONFIG.API.VALIDATE,
    AUTH_TOKEN_KEY: CONFIG.AUTH.TOKEN_KEY,
    serviceUrls: SERVICE_URLS
});

// Helper functions for environment switching (development/testing only)
window.switchToLocal = function() {
    localStorage.setItem('force_local_env', 'true');
    localStorage.removeItem('force_docker_env');
    console.log('🔧 Environment switched to LOCAL. Refresh page to apply.');
};

window.switchToDocker = function() {
    localStorage.setItem('force_docker_env', 'true');
    localStorage.removeItem('force_local_env');
    console.log('🔧 Environment switched to DOCKER. Refresh page to apply.');
};

window.resetEnvironment = function() {
    localStorage.removeItem('force_local_env');
    localStorage.removeItem('force_docker_env');
    console.log('🔧 Environment reset to auto-detection. Refresh page to apply.');
};

// Make config available globally
window.CONFIG = CONFIG; 