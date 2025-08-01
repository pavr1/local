// Ice Cream Store Status Service
// Shared component for system health checks and status monitoring

console.log('🔧 STATUS.JS: Loading status service...');

class StatusService {
    constructor() {
        this.healthCheckInterval = null;
        this.lastHealthCheck = null;
        this.init();
    }

    init() {
        console.log('🏥 StatusService: Initializing...');
        
        // Ensure CONFIG is available
        if (typeof CONFIG === 'undefined') {
            console.error('❌ CONFIG not available for StatusService');
            return;
        }
        
        console.log('✅ StatusService: Initialized successfully');
    }

    async checkAllServices() {
        console.log('🔍 StatusService: Checking all services...');
        
        const results = {};
        const services = CONFIG.SERVICES;
        
        for (const [serviceKey, serviceConfig] of Object.entries(services)) {
            try {
                console.log(`🔍 Checking ${serviceConfig.name}...`);
                const status = await this.checkServiceHealth(serviceConfig.url);
                results[serviceKey] = status;
                console.log(`${status === 'healthy' ? '✅' : '❌'} ${serviceConfig.name}: ${status}`);
            } catch (error) {
                console.error(`❌ Error checking ${serviceConfig.name}:`, error);
                results[serviceKey] = 'unhealthy';
            }
        }
        
        this.lastHealthCheck = {
            timestamp: new Date(),
            results: results
        };
        
        // Log summary
        const healthyCount = Object.values(results).filter(status => status === 'healthy').length;
        const totalCount = Object.keys(results).length;
        console.log(`🏥 Health check results: ${JSON.stringify(results)}`);
        console.log(`📊 Overall: ${healthyCount}/${totalCount} services healthy`);
        
        return results;
    }

    async checkServiceHealth(url) {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
            
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                },
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (response.ok) {
                const data = await response.json();
                return data.status === 'healthy' || data.success ? 'healthy' : 'degraded';
            } else {
                return 'unhealthy';
            }
        } catch (error) {
            if (error.name === 'AbortError') {
                console.warn(`⏰ Health check timeout for ${url}`);
            } else {
                console.warn(`⚠️ Health check failed for ${url}:`, error.message);
            }
            return 'unhealthy';
        }
    }

    startPeriodicHealthCheck(interval = 30000) {
        console.log(`🔄 StatusService: Starting periodic health checks every ${interval/1000}s`);
        
        // Clear any existing interval
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval);
        }
        
        // Do initial check
        this.checkAllServices();
        
        // Set up periodic checks
        this.healthCheckInterval = setInterval(() => {
            this.checkAllServices();
        }, interval);
    }

    stopPeriodicHealthCheck() {
        console.log('⏹️ StatusService: Stopping periodic health checks');
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval);
            this.healthCheckInterval = null;
        }
    }

    updateStatusIndicator(element, status) {
        if (!element) return;
        
        // Remove existing status classes (correct classes that exist in CSS)
        element.classList.remove('online', 'warning', 'offline', 'loading');
        
        // Add appropriate status class based on health status
        switch (status) {
            case 'healthy':
                element.classList.add('online');
                break;
            case 'degraded':
                element.classList.add('warning');
                break;
            case 'unhealthy':
                element.classList.add('offline');
                break;
            default:
                element.classList.add('loading');
        }
    }

    getLastHealthCheck() {
        return this.lastHealthCheck;
    }

    getServiceStatus(serviceKey) {
        if (!this.lastHealthCheck) return null;
        return this.lastHealthCheck.results[serviceKey] || null;
    }
}

// Create global instance
window.statusService = new StatusService();

console.log('✅ STATUS.JS: StatusService created and available globally'); 