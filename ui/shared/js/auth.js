// === ICE CREAM STORE AUTHENTICATION SERVICE ===

class AuthService {
    constructor() {
        // Ensure CONFIG is available
        if (typeof CONFIG === 'undefined') {
            console.error('❌ CONFIG not available when creating AuthService');
            throw new Error('CONFIG must be loaded before AuthService');
        }
        
        // Debug: Log the raw CONFIG values
        console.log('🔍 Raw CONFIG in AuthService:', {
            GATEWAY_URL: CONFIG.GATEWAY_URL,
            API: CONFIG.API,
            AUTH: CONFIG.AUTH
        });
        
        // Ensure CONFIG.AUTH exists with defaults
        if (!CONFIG.AUTH) {
            console.warn('⚠️ CONFIG.AUTH not found, using defaults');
            CONFIG.AUTH = {
                TOKEN_KEY: 'icecream_auth_token',
                USER_KEY: 'icecream_user_data',
                REMEMBER_KEY: 'icecream_remember_me'
            };
        }
        
        // Use the environment-aware gateway URL
        this.baseURL = CONFIG.GATEWAY_URL;
        this.tokenKey = CONFIG.AUTH.TOKEN_KEY;
        this.userKey = CONFIG.AUTH.USER_KEY;
        this.rememberKey = CONFIG.AUTH.REMEMBER_KEY;
        
        console.log('🔧 AuthService initialized with:', {
            baseURL: this.baseURL,
            tokenKey: this.tokenKey,
            userKey: this.userKey,
            rememberKey: this.rememberKey
        });
    }

    // === MAIN LOGIN METHOD ===
    
    async login(username, password, rememberMe = false) {
        try {
            console.log('🔑 Attempting login for:', username);
            
            const response = await fetch(`${this.baseURL}${CONFIG.API.LOGIN}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password })
            });

            console.log('📡 Login response status:', response.status);

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.message || errorData.error || `Login failed (${response.status})`;
                throw new Error(errorMessage);
            }

            const data = await response.json();
            console.log('✅ Login successful:', { user: data.user?.username, hasToken: !!data.token });
            
            // Store authentication data
            this.setToken(data.token, rememberMe);
            this.setUserData(data.user, data.role, data.permissions || []);
            
            return {
                success: true,
                user: data.user,
                role: data.role,
                token: data.token
            };
            
        } catch (error) {
            console.error('❌ Login error:', error.message);
            throw error;
        }
    }

    // === LOGOUT METHOD ===
    
    async logout() {
        try {
            const token = this.getToken();
            if (token) {
                console.log('🔓 Logging out...');
                await fetch(`${this.baseURL}${CONFIG.API.LOGOUT}`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    }
                });
            }
        } catch (error) {
            console.warn('⚠️ Logout API call failed:', error.message);
        } finally {
            // Clear local storage regardless of API call result
            this.clearAuthData();
            console.log('🧹 Auth data cleared');
        }
    }

    // === TOKEN VALIDATION ===
    
    async validateToken() {
        try {
            const token = this.getToken();
            if (!token) {
                console.log('❌ No token found');
                return false;
            }

            console.log('🔍 Validating token...');
            const response = await fetch(`${this.baseURL}${CONFIG.API.VALIDATE}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                console.log('❌ Token validation failed:', response.status);
                
                // Get detailed error information
                const errorData = await response.json().catch(() => ({}));
                console.log('❌ Validation error details:', errorData);
                
                this.clearAuthData();
                return false;
            }

            const data = await response.json();
            console.log('✅ Token is valid:', data);
            return true;
        } catch (error) {
            console.error('❌ Token validation error:', error.message);
            this.clearAuthData();
            return false;
        }
    }

    // === LOCAL STORAGE METHODS ===
    
    setToken(token, rememberMe = false) {
        const storage = rememberMe ? localStorage : sessionStorage;
        storage.setItem(this.tokenKey, token);
        localStorage.setItem(this.rememberKey, rememberMe.toString());
        console.log('💾 Token stored in:', rememberMe ? 'localStorage' : 'sessionStorage');
    }

    getToken() {
        // Check localStorage first (remember me), then sessionStorage
        return localStorage.getItem(this.tokenKey) || sessionStorage.getItem(this.tokenKey);
    }

    setUserData(user, role, permissions) {
        const userData = {
            user,
            role,
            permissions,
            loginTime: new Date().toISOString()
        };
        
        const storage = this.isRememberMe() ? localStorage : sessionStorage;
        storage.setItem(this.userKey, JSON.stringify(userData));
        console.log('💾 User data stored:', { username: user?.username, role });
    }

    getUserData() {
        const data = localStorage.getItem(this.userKey) || sessionStorage.getItem(this.userKey);
        return data ? JSON.parse(data) : null;
    }

    isRememberMe() {
        return localStorage.getItem(this.rememberKey) === 'true';
    }

    clearAuthData() {
        // Clear from both storages
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
        localStorage.removeItem(this.rememberKey);
        sessionStorage.removeItem(this.tokenKey);
        sessionStorage.removeItem(this.userKey);
    }

    // === AUTHENTICATION STATE ===
    
    isAuthenticated() {
        return !!this.getToken();
    }

    getCurrentUser() {
        const userData = this.getUserData();
        return userData ? userData.user : null;
    }

    getCurrentRole() {
        const userData = this.getUserData();
        return userData ? userData.role : null;
    }

    getPermissions() {
        const userData = this.getUserData();
        return userData ? userData.permissions : [];
    }

    hasPermission(permission) {
        const permissions = this.getPermissions();
        return permissions.includes(permission);
    }
}

// === STATUS CHECKER SERVICE ===

class StatusService {
    constructor() {
        this.services = CONFIG.SERVICES;
    }

    async checkServiceHealth(url) {
        try {
            const response = await fetch(url, {
                method: 'GET',
                timeout: 5000 // 5 second timeout
            });
            
            // Handle different response cases
            if (response.ok) {
                return 'healthy';
            }
            
            // For 503 responses, check if it's degraded (partial failure) vs completely down
            if (response.status === 503) {
                try {
                    const data = await response.json();
                    if (data && data.status === 'degraded') {
                        return 'degraded';
                    }
                } catch (parseError) {
                    console.warn(`Could not parse 503 response for ${url}:`, parseError.message);
                }
            }
            
            return 'unhealthy';
        } catch (error) {
            console.warn(`Service health check failed for ${url}:`, error.message);
            return 'unhealthy';
        }
    }

    async checkAllServices() {
        console.log('🏥 Checking system health...');
        
        const results = {};
        
        for (const [serviceKey, service] of Object.entries(this.services)) {
            const healthStatus = await this.checkServiceHealth(service.url);
            results[serviceKey] = healthStatus;
            
            const indicator = document.getElementById(service.element);
            if (indicator) {
                // Map health status to indicator status
                let indicatorStatus;
                switch (healthStatus) {
                    case 'healthy':
                        indicatorStatus = 'online';
                        break;
                    case 'degraded':
                        indicatorStatus = 'warning';
                        break;
                    case 'unhealthy':
                    default:
                        indicatorStatus = 'offline';
                        break;
                }
                this.updateStatusIndicator(indicator, indicatorStatus);
            }
        }

        console.log('🏥 Health check results:', results);
        return results;
    }

    updateStatusIndicator(element, status) {
        if (!element) return;
        
        element.classList.remove('online', 'offline', 'warning', 'loading');
        element.classList.add(status);
    }

    startPeriodicHealthCheck(intervalMs = 30000) {
        // Check immediately
        this.checkAllServices();
        
        // Then check every interval
        setInterval(() => {
            this.checkAllServices();
        }, intervalMs);
    }
}

// === GLOBAL INSTANCES ===

// Function to safely create services when CONFIG is available
function initializeServices() {
    try {
        if (typeof CONFIG === 'undefined') {
            console.warn('⚠️ CONFIG not yet available, retrying...');
            // Retry after a short delay
            setTimeout(initializeServices, 100);
            return;
        }
        
        // Create global instances
        window.authService = new AuthService();
        window.statusService = new StatusService();
        
        console.log('🔧 Authentication and Status services initialized');
    } catch (error) {
        console.error('❌ Failed to initialize services:', error);
        // Retry after a delay
        setTimeout(initializeServices, 500);
    }
}

// Start initialization
initializeServices(); 