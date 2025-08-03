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
        
        // Use the session service URL for authentication
        this.baseURL = CONFIG.API.session.replace('/api/v1/session', '');  // Get base session service URL
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

            console.log('🔍 Validating token with session service...');
            const response = await fetch(`${this.baseURL}${CONFIG.API.VALIDATE}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token: token })
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
            console.log('✅ Token validation response:', data);
            
            // Check if the session is actually valid
            if (data.is_valid === true) {
                console.log('✅ Session is valid for user:', data.session?.username);
                return true;
            } else {
                console.log('❌ Session is not valid:', data.error_code, data.error_message);
                this.clearAuthData();
                return false;
            }
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