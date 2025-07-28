// === ICE CREAM STORE AUTHENTICATION MODULE ===

class AuthService {
    constructor() {
        this.baseURL = 'http://localhost:8082/api';
        this.tokenKey = 'icecream_auth_token';
        this.userKey = 'icecream_user_data';
        this.rememberKey = 'icecream_remember_me';
    }

    // === API METHODS ===
    
    async login(username, password, rememberMe = false) {
        try {
            const response = await fetch(`${this.baseURL}/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Error de inicio de sesión');
            }

            const data = await response.json();
            
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
            console.error('Login error:', error);
            throw new Error(error.message || 'Error de red');
        }
    }

    async logout() {
        try {
            const token = this.getToken();
            if (token) {
                await fetch(`${this.baseURL}/v1/auth/logout`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    }
                });
            }
        } catch (error) {
            console.warn('Error en llamada de cierre de sesión:', error);
        } finally {
            // Clear local storage regardless of API call result
            this.clearAuthData();
        }
    }

    async validateToken() {
        try {
            const token = this.getToken();
            if (!token) return false;

            const response = await fetch(`${this.baseURL}/v1/auth/validate`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                this.clearAuthData();
                return false;
            }

            return true;
        } catch (error) {
            console.error('Token validation error:', error);
            this.clearAuthData();
            return false;
        }
    }

    async getProfile() {
        try {
            const token = this.getToken();
            if (!token) throw new Error('No hay token de autenticación');

            const response = await fetch(`${this.baseURL}/v1/auth/profile`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error('Error al obtener el perfil');
            }

            const data = await response.json();
            this.setUserData(data.user, data.role, data.permissions || []);
            
            return data;
        } catch (error) {
            console.error('Profile fetch error:', error);
            throw error;
        }
    }

    async checkSystemHealth() {
        try {
            const response = await fetch(`${this.baseURL}/health`, {
                method: 'GET',
            });

            if (!response.ok) {
                return {
                    gateway: 'offline',
                    services: {}
                };
            }

            const data = await response.json();
            return {
                gateway: data.status === 'healthy' ? 'online' : 'degraded',
                services: data.services || {}
            };
        } catch (error) {
            console.error('Health check error:', error);
            return {
                gateway: 'offline',
                services: {}
            };
        }
    }

    // === TOKEN MANAGEMENT ===
    
    setToken(token, rememberMe = false) {
        const storage = rememberMe ? localStorage : sessionStorage;
        storage.setItem(this.tokenKey, token);
        
        if (rememberMe) {
            localStorage.setItem(this.rememberKey, 'true');
        } else {
            localStorage.removeItem(this.rememberKey);
        }
    }

    getToken() {
        // Check sessionStorage first, then localStorage
        return sessionStorage.getItem(this.tokenKey) || 
               localStorage.getItem(this.tokenKey);
    }

    setUserData(user, role, permissions = []) {
        const userData = {
            user,
            role,
            permissions,
            timestamp: Date.now()
        };
        
        const storage = this.isRemembered() ? localStorage : sessionStorage;
        storage.setItem(this.userKey, JSON.stringify(userData));
    }

    getUserData() {
        try {
            const data = sessionStorage.getItem(this.userKey) || 
                        localStorage.getItem(this.userKey);
            return data ? JSON.parse(data) : null;
        } catch (error) {
            console.error('Error parsing user data:', error);
            return null;
        }
    }

    isRemembered() {
        return localStorage.getItem(this.rememberKey) === 'true';
    }

    clearAuthData() {
        // Clear both session and local storage
        sessionStorage.removeItem(this.tokenKey);
        sessionStorage.removeItem(this.userKey);
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
        localStorage.removeItem(this.rememberKey);
    }

    // === AUTHENTICATION STATE ===
    
    isAuthenticated() {
        const token = this.getToken();
        const userData = this.getUserData();
        
        if (!token || !userData) {
            return false;
        }

        // Check if data is too old (optional security measure)
        const maxAge = this.isRemembered() ? 30 * 24 * 60 * 60 * 1000 : 24 * 60 * 60 * 1000; // 30 days or 1 day
        const isExpired = Date.now() - userData.timestamp > maxAge;
        
        if (isExpired) {
            this.clearAuthData();
            return false;
        }

        return true;
    }

    getCurrentUser() {
        const userData = this.getUserData();
        return userData ? userData.user : null;
    }

    getCurrentRole() {
        const userData = this.getUserData();
        return userData ? userData.role : null;
    }

    getUserPermissions() {
        const userData = this.getUserData();
        return userData ? userData.permissions : [];
    }

    hasPermission(permission) {
        const permissions = this.getUserPermissions();
        return permissions.includes(permission);
    }

    isAdmin() {
        const role = this.getCurrentRole();
        return role && (role.role_name === 'super_admin' || role.role_name === 'admin');
    }

    // === UTILITY METHODS ===
    
    formatError(error) {
        if (error.message) {
            return error.message;
        }
        
        if (typeof error === 'string') {
            return error;
        }
        
        return 'Ocurrió un error inesperado';
    }

    // Auto-refresh token (if needed in future)
    async refreshToken() {
        try {
            const token = this.getToken();
            if (!token) throw new Error('No hay token para actualizar');

            const response = await fetch(`${this.baseURL}/v1/auth/refresh`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error('Error al actualizar el token');
            }

            const data = await response.json();
            this.setToken(data.token, this.isRemembered());
            
            return data.token;
        } catch (error) {
            console.error('Token refresh error:', error);
            this.clearAuthData();
            throw error;
        }
    }
}

// Create global auth service instance
window.authService = new AuthService();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AuthService;
} 