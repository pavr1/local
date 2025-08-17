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
                SESSION_ID_KEY: 'icecream_session_id',
                USER_KEY: 'icecream_user_data',
                REMEMBER_KEY: 'icecream_remember_me'
            };
        }
        
        // Use the gateway URL for authentication (gateway handles CORS and routing)
        this.baseURL = CONFIG.GATEWAY_URL;
        this.sessionIdKey = CONFIG.AUTH.SESSION_ID_KEY || CONFIG.AUTH.sessionIdKey;
        this.userKey = CONFIG.AUTH.USER_KEY || CONFIG.AUTH.userKey;
        this.rememberKey = CONFIG.AUTH.REMEMBER_KEY || CONFIG.AUTH.rememberKey;
        
        console.log('🔧 AuthService initialized with:', {
            baseURL: this.baseURL,
            sessionIdKey: this.sessionIdKey,
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
            console.log('✅ Login successful:', { user: data.user?.username, hasSessionId: !!data.session_id });
            
            // Store authentication data
            this.setSessionId(data.session_id, rememberMe);
            this.setUserData(data.user, data.role, data.permissions || []);
            
            return {
                success: true,
                user: data.user,
                role: data.role,
                permissions: data.permissions || []
            };
            
        } catch (error) {
            console.error('❌ Login error:', error);
            throw error;
        }
    }

    // === LOGOUT METHOD ===
    
    async logout() {
        try {
            console.log('🚪 Attempting logout...');
            
            const sessionId = this.getSessionId();
            if (!sessionId) {
                console.log('⚠️ No session ID found, clearing local data only');
                this.clearAuthData();
                return { success: true };
            }

            const response = await makeAuthenticatedRequest(`${this.baseURL}${CONFIG.API.LOGOUT}`, {
                method: 'POST'
            });

            console.log('📡 Logout response status:', response.status);

            // Always clear local data regardless of server response
            this.clearAuthData();
            
            if (response.ok) {
                console.log('✅ Logout successful');
                return { success: true };
            } else {
                console.warn('⚠️ Server logout failed, but local data cleared');
                return { success: true, warning: 'Server logout failed' };
            }
            
        } catch (error) {
            console.error('❌ Logout error:', error);
            // Always clear local data even if request fails
            this.clearAuthData();
            throw error;
        }
    }

    // === SESSION MANAGEMENT ===
    
    setSessionId(sessionId, rememberMe = false) {
        const storage = rememberMe ? localStorage : sessionStorage;
        storage.setItem(this.sessionIdKey, sessionId);
        
        // Store remember me preference
        if (rememberMe) {
            localStorage.setItem(this.rememberKey, 'true');
        } else {
            localStorage.removeItem(this.rememberKey);
        }
        
        console.log('💾 Session ID stored:', { sessionId: sessionId ? '***' : null, rememberMe });
    }

    getSessionId() {
        // Check sessionStorage first, then localStorage
        let sessionId = sessionStorage.getItem(this.sessionIdKey);
        if (!sessionId) {
            sessionId = localStorage.getItem(this.sessionIdKey);
        }
        return sessionId;
    }

    setUserData(user, role, permissions) {
        const userData = { user, role, permissions };
        const storage = this.isRememberMe() ? localStorage : sessionStorage;
        storage.setItem(this.userKey, JSON.stringify(userData));
        console.log('💾 User data stored:', { username: user?.username, role: role?.name });
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
        sessionStorage.removeItem(this.sessionIdKey);
        sessionStorage.removeItem(this.userKey);
        localStorage.removeItem(this.sessionIdKey);
        localStorage.removeItem(this.userKey);
        localStorage.removeItem(this.rememberKey);
        console.log('🧹 Auth data cleared');
    }

    isAuthenticated() {
        const sessionId = this.getSessionId();
        if (!sessionId) {
            return false;
        }
        
        // Basic validation - check if session ID exists and has a reasonable format
        // Note: Full validation should be done with a server call
        return sessionId.length > 0 && sessionId !== 'null' && sessionId !== 'undefined';
    }

    getCurrentUser() {
        const userData = this.getUserData();
        return userData?.user || null;
    }

    getCurrentRole() {
        const userData = this.getUserData();
        return userData?.role || null;
    }

    getPermissions() {
        const userData = this.getUserData();
        return userData?.permissions || [];
    }

    hasPermission(permission) {
        const permissions = this.getPermissions();
        return permissions.includes(permission);
    }
}

// === GLOBAL AUTH SERVICE INITIALIZATION ===

function initializeAuthService() {
    try {
        console.log('🔧 Initializing global AuthService...');
        
        if (window.authService) {
            console.log('⚠️ AuthService already exists, reusing...');
            return window.authService;
        }
        
        const authService = new AuthService();
        window.authService = authService;
        
        console.log('✅ Global AuthService initialized');
        return authService;
        
    } catch (error) {
        console.error('❌ Failed to initialize AuthService:', error);
        throw error;
    }
}

// === AUTHENTICATED REQUEST HELPER ===

async function makeAuthenticatedRequest(url, options = {}) {
    const authService = window.authService;
    
    if (!authService) {
        console.error('❌ AuthService not available');
        throw new Error('Authentication service not available');
    }
    
    if (!authService.isAuthenticated()) {
        console.error('❌ User not authenticated');
        redirectToLogin();
        throw new Error('User not authenticated');
    }
    
    const sessionId = authService.getSessionId();
    
    // Set up headers
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${sessionId}`,
        ...options.headers
    };
    
    // Make the request
    const response = await fetch(url, {
        ...options,
        headers
    });
    
    // Handle authentication errors
    if (response.status === 401) {
        console.warn('⚠️ Session expired, redirecting to login');
        authService.clearAuthData();
        redirectToLogin();
        throw new Error('Session expired');
    }
    
    return response;
}

// === UTILITY FUNCTIONS ===

function redirectToLogin() {
    console.log('🔄 Redirecting to login page...');
    window.location.href = 'login.html';
}

// === CONVENIENCE METHODS ===

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

// === AUTO-INITIALIZATION ===

// Initialize AuthService when this script loads
if (typeof CONFIG !== 'undefined') {
    initializeAuthService();
} else {
    console.log('⏳ Waiting for CONFIG to load before initializing AuthService...');
    window.addEventListener('load', () => {
        if (typeof CONFIG !== 'undefined') {
            initializeAuthService();
        }
    });
} 