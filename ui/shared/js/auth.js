// === ICE CREAM STORE AUTHENTICATION SERVICE ===

// Utility functions for UTC time handling
const TimeUtils = {
    // Get current time in UTC
    nowUTC() {
        return new Date().toISOString();
    },
    
    // Convert UTC timestamp to local time for display
    utcToLocal(utcString) {
        return new Date(utcString).toLocaleString();
    },
    
    // Convert local time to UTC for API calls
    localToUTC(localDate) {
        return localDate.toISOString();
    },
    
    // Format UTC timestamp for display with timezone info
    formatUTCForDisplay(utcString) {
        const date = new Date(utcString);
        return {
            local: date.toLocaleString(),
            utc: date.toISOString(),
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
        };
    }
};

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
        
        // Use the gateway URL for authentication (gateway handles CORS and routing)
        this.baseURL = CONFIG.GATEWAY_URL;
        this.tokenKey = CONFIG.AUTH.TOKEN_KEY || CONFIG.AUTH.tokenKey;
        this.userKey = CONFIG.AUTH.USER_KEY || CONFIG.AUTH.userKey;
        this.rememberKey = CONFIG.AUTH.REMEMBER_KEY || CONFIG.AUTH.rememberKey;
        
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
            this.clearAuthData();
        }
    }

    // === TOKEN VALIDATION ===
    
    async validateToken() {
        try {
            const token = this.getToken();
            if (!token) {
                console.log('❌ No token found for validation');
                return false;
            }

            console.log('🔍 Validating token...');
            
            const response = await fetch(`${this.baseURL}${CONFIG.API.VALIDATE}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            console.log('📡 Token validation response status:', response.status);

            if (response.ok) {
                const data = await response.json();
                console.log('✅ Token is valid:', { user: data.user?.username });
                return true;
            } else {
                console.log('❌ Token validation failed:', response.status);
                return false;
            }
            
        } catch (error) {
            console.error('❌ Token validation error:', error.message);
            return false;
        }
    }

    // === TOKEN REFRESH ===
    
    async refreshToken() {
        try {
            const token = this.getToken();
            if (!token) {
                console.log('❌ No token found for refresh');
                return false;
            }

            console.log('🔄 Attempting token refresh...');
            
            const response = await fetch(`${this.baseURL}${CONFIG.API.REFRESH}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            console.log('📡 Token refresh response status:', response.status);

            if (response.ok) {
                const data = await response.json();
                console.log('✅ Token refreshed successfully');
                
                // Update stored token
                this.setToken(data.token, this.isRememberMe());
                return true;
            } else {
                console.log('❌ Token refresh failed:', response.status);
                return false;
            }
            
        } catch (error) {
            console.error('❌ Token refresh error:', error.message);
            return false;
        }
    }

    // === TOKEN STORAGE ===
    
    setToken(token, rememberMe = false) {
        if (rememberMe) {
            localStorage.setItem(this.tokenKey, token);
            localStorage.setItem(this.rememberKey, 'true');
        } else {
            sessionStorage.setItem(this.tokenKey, token);
            sessionStorage.setItem(this.rememberKey, 'false');
        }
        console.log('💾 Token stored:', { rememberMe, hasToken: !!token });
    }

    getToken() {
        const rememberMe = this.isRememberMe();
        const storage = rememberMe ? localStorage : sessionStorage;
        const token = storage.getItem(this.tokenKey);
        console.log('🔑 Token retrieved:', { rememberMe, hasToken: !!token });
        return token;
    }

    setUserData(user, role, permissions) {
        const userData = { user, role, permissions };
        const rememberMe = this.isRememberMe();
        const storage = rememberMe ? localStorage : sessionStorage;
        storage.setItem(this.userKey, JSON.stringify(userData));
        console.log('👤 User data stored:', { user: user?.username, role, permissionsCount: permissions?.length });
    }

    getUserData() {
        const rememberMe = this.isRememberMe();
        const storage = rememberMe ? localStorage : sessionStorage;
        const userData = storage.getItem(this.userKey);
        return userData ? JSON.parse(userData) : null;
    }

    isRememberMe() {
        return localStorage.getItem(this.rememberKey) === 'true';
    }

    clearAuthData() {
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
        localStorage.removeItem(this.rememberKey);
        sessionStorage.removeItem(this.tokenKey);
        sessionStorage.removeItem(this.userKey);
        sessionStorage.removeItem(this.rememberKey);
        console.log('🧹 Auth data cleared');
    }

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
// Note: StatusService is defined in shared/js/status.js - using that implementation instead

// === GLOBAL INSTANCES ===

// Function to safely create AuthService when CONFIG is available
function initializeAuthService() {
    try {
        if (typeof CONFIG === 'undefined') {
            console.warn('⚠️ CONFIG not yet available, retrying...');
            // Retry after a short delay
            setTimeout(initializeAuthService, 100);
            return;
        }
        
        // Create global AuthService instance  
        // StatusService is created in shared/js/status.js
        window.authService = new AuthService();
        
        console.log('🔧 AuthService initialized (connects to Session Service)');
    } catch (error) {
        console.error('❌ Failed to initialize AuthService:', error);
        // Retry after a delay
        setTimeout(initializeAuthService, 500);
    }
}

// Start initialization
initializeAuthService();

// === UTILITY FUNCTIONS ===

// Make authenticated API request with automatic token refresh
async function makeAuthenticatedRequest(url, options = {}) {
    try {
        // Get authentication token
        const token = window.authService ? window.authService.getToken() : null;
        
        if (!token) {
            console.warn('No authentication token available, redirecting to login');
            redirectToLogin();
            throw new Error('No authentication token available. Please log in again.');
        }
        
        // Set up headers
        const headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            ...options.headers
        };
        
        // Make the request
        const response = await fetch(url, {
            ...options,
            headers
        });
        
        // Handle authentication errors
        if (response.status === 401) {
            console.warn('Authentication failed (401), attempting token refresh...');
            
            // Try to refresh the token
            if (window.authService) {
                const refreshSuccess = await window.authService.refreshToken();
                if (refreshSuccess) {
                    // Retry the original request with the new token
                    console.log('Token refreshed, retrying request...');
                    return makeAuthenticatedRequest(url, options);
                }
            }
            
            // If refresh failed, redirect to login
            console.warn('Token refresh failed, redirecting to login');
            redirectToLogin();
            throw new Error('Authentication failed. Please log in again.');
        }
        
        return response;
        
    } catch (error) {
        console.error('Authenticated request failed:', error);
        throw error;
    }
}

// Helper function to redirect to login
function redirectToLogin() {
    if (window.authService) {
        window.authService.clearAuthData();
    }
    window.location.href = 'login.html';
}

// === CONVENIENCE FUNCTIONS ===

async function authenticatedGet(url) {
    return makeAuthenticatedRequest(url, { method: 'GET' });
}

async function authenticatedPost(url, data) {
    return makeAuthenticatedRequest(url, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

async function authenticatedPut(url, data) {
    return makeAuthenticatedRequest(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

async function authenticatedDelete(url) {
    return makeAuthenticatedRequest(url, { method: 'DELETE' });
} 