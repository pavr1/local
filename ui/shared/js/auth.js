// === ICE CREAM STORE AUTHENTICATION SERVICE ===

class AuthService {
    constructor() {
        this.baseURL = CONFIG.GATEWAY_URL;
        this.tokenKey = CONFIG.AUTH.TOKEN_KEY;
        this.userKey = CONFIG.AUTH.USER_KEY;
        this.rememberKey = CONFIG.AUTH.REMEMBER_KEY;
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
                this.clearAuthData();
                return false;
            }

            console.log('✅ Token is valid');
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
            return response.ok;
        } catch (error) {
            console.warn(`Service health check failed for ${url}:`, error.message);
            return false;
        }
    }

    async checkAllServices() {
        console.log('🏥 Checking system health...');
        
        const results = {};
        
        for (const [serviceKey, service] of Object.entries(this.services)) {
            const isHealthy = await this.checkServiceHealth(service.url);
            results[serviceKey] = isHealthy;
            
            const indicator = document.getElementById(service.element);
            if (indicator) {
                this.updateStatusIndicator(indicator, isHealthy ? 'online' : 'offline');
            }
        }

        console.log('🏥 Health check results:', results);
        return results;
    }

    updateStatusIndicator(element, status) {
        if (!element) return;
        
        element.classList.remove('online', 'offline', 'loading');
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

// Create global instances
window.authService = new AuthService();
window.statusService = new StatusService();

console.log('🔧 Authentication and Status services initialized'); 