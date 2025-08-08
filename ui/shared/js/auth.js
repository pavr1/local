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
            window.location.href = 'login.html';
        }
    }

    // === TOKEN VALIDATION AND REFRESH ===
    
    async validateToken() {
        try {
            const token = this.getToken();
            if (!token) {
                return false;
            }

            const response = await fetch(`${this.baseURL}${CONFIG.API.VALIDATE}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token })
            });

            if (response.ok) {
                const data = await response.json();
                if (data.is_valid && data.new_token) {
                    // Update token if a new one is provided
                    this.setToken(data.new_token, this.isRememberMe());
                    console.log('🔄 Token refreshed successfully');
                }
                return data.is_valid;
            }
            
            return false;
        } catch (error) {
            console.error('❌ Token validation error:', error);
            return false;
        }
    }

    // === TOKEN REFRESH METHOD ===
    
    async refreshToken() {
        try {
            const token = this.getToken();
            if (!token) {
                throw new Error('No token available for refresh');
            }

            const response = await fetch(`${this.baseURL}${CONFIG.API.VALIDATE}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token })
            });

            if (response.ok) {
                const data = await response.json();
                if (data.is_valid && data.new_token) {
                    this.setToken(data.new_token, this.isRememberMe());
                    console.log('🔄 Token refreshed successfully');
                    return true;
                }
            }
            
            return false;
        } catch (error) {
            console.error('❌ Token refresh error:', error);
            return false;
        }
    }

    // === TOKEN MANAGEMENT ===
    
    setToken(token, rememberMe = false) {
        if (rememberMe) {
            localStorage.setItem(this.tokenKey, token);
            localStorage.setItem(this.rememberKey, 'true');
        } else {
            sessionStorage.setItem(this.tokenKey, token);
            localStorage.removeItem(this.rememberKey);
        }
    }

    getToken() {
        return localStorage.getItem(this.tokenKey) || sessionStorage.getItem(this.tokenKey);
    }

    setUserData(user, role, permissions) {
        const userData = { user, role, permissions };
        const storage = this.isRememberMe() ? localStorage : sessionStorage;
        storage.setItem(this.userKey, JSON.stringify(userData));
    }

    getUserData() {
        const storage = this.isRememberMe() ? localStorage : sessionStorage;
        const data = storage.getItem(this.userKey);
        return data ? JSON.parse(data) : null;
    }

    isRememberMe() {
        return localStorage.getItem(this.rememberKey) === 'true';
    }

    clearAuthData() {
        localStorage.removeItem(this.tokenKey);
        sessionStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
        sessionStorage.removeItem(this.userKey);
        localStorage.removeItem(this.rememberKey);
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

// Make authenticated GET request
async function authenticatedGet(url) {
    return makeAuthenticatedRequest(url, { method: 'GET' });
}

// Make authenticated POST request
async function authenticatedPost(url, data) {
    return makeAuthenticatedRequest(url, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

// Make authenticated PUT request
async function authenticatedPut(url, data) {
    return makeAuthenticatedRequest(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

// Make authenticated DELETE request
async function authenticatedDelete(url) {
    return makeAuthenticatedRequest(url, { method: 'DELETE' });
} 