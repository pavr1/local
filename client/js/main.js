// === ICE CREAM STORE MAIN APPLICATION ===

class LoginApp {
    constructor() {
        this.loginForm = document.getElementById('loginForm');
        this.loginBtn = document.getElementById('loginBtn');
        this.togglePasswordBtn = document.getElementById('togglePassword');
        this.alertContainer = document.getElementById('alert-container');
        this.successModal = new bootstrap.Modal(document.getElementById('successModal'));
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupPasswordToggle();
        this.checkSystemStatus();
        this.checkExistingAuth();
        
        // Set up periodic status checks
        setInterval(() => this.checkSystemStatus(), 30000); // Every 30 seconds
    }

    setupEventListeners() {
        // Form submission
        this.loginForm.addEventListener('submit', (e) => this.handleLogin(e));
        
        // Real-time validation
        const inputs = this.loginForm.querySelectorAll('input[required]');
        inputs.forEach(input => {
            input.addEventListener('blur', () => this.validateField(input));
            input.addEventListener('input', () => this.clearFieldError(input));
        });

        // Enter key handling
        this.loginForm.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !this.loginBtn.disabled) {
                this.handleLogin(e);
            }
        });
    }

    setupPasswordToggle() {
        this.togglePasswordBtn.addEventListener('click', () => {
            const passwordInput = document.getElementById('password');
            const icon = this.togglePasswordBtn.querySelector('i');
            
            if (passwordInput.type === 'password') {
                passwordInput.type = 'text';
                icon.classList.replace('fa-eye', 'fa-eye-slash');
            } else {
                passwordInput.type = 'password';
                icon.classList.replace('fa-eye-slash', 'fa-eye');
            }
        });
    }

    async handleLogin(e) {
        e.preventDefault();
        
        if (!this.validateForm()) {
            return;
        }

        const formData = new FormData(this.loginForm);
        const username = formData.get('username').trim();
        const password = formData.get('password');
        const rememberMe = formData.get('rememberMe') === 'on';

        this.setLoadingState(true);
        this.clearAlerts();

        try {
            const result = await authService.login(username, password, rememberMe);
            
            if (result.success) {
                this.showSuccess('¡Inicio de sesión exitoso! Bienvenido de nuevo, ' + result.user.full_name);
                await this.handleSuccessfulLogin(result);
            }
        } catch (error) {
            this.showError(authService.formatError(error));
            this.setLoadingState(false);
        }
    }

    async handleSuccessfulLogin(result) {
        // Show success modal with progress
        this.successModal.show();
        
        // Animate progress bar
        const progressBar = document.querySelector('.progress-bar');
        let progress = 0;
        const progressInterval = setInterval(() => {
            progress += 10;
            progressBar.style.width = progress + '%';
            
            if (progress >= 100) {
                clearInterval(progressInterval);
                setTimeout(() => {
                    this.redirectToDashboard();
                }, 500);
            }
        }, 150);
    }

    redirectToDashboard() {
        // Redirect to dashboard page
        this.successModal.hide();
        this.setLoadingState(false);
        
        setTimeout(() => {
            window.location.href = 'dashboard.html';
        }, 500);
    }

    showUserInfo() {
        const user = authService.getCurrentUser();
        const role = authService.getCurrentRole();
        
        if (user && role) {
            const message = `
                <strong>Logged in successfully!</strong><br>
                <strong>User:</strong> ${user.full_name} (${user.username})<br>
                <strong>Role:</strong> ${role.role_name}<br>
                <strong>Permissions:</strong> ${authService.getUserPermissions().length} permissions available
            `;
            this.showInfo(message);
        }
    }

    validateForm() {
        const inputs = this.loginForm.querySelectorAll('input[required]');
        let isValid = true;

        inputs.forEach(input => {
            if (!this.validateField(input)) {
                isValid = false;
            }
        });

        return isValid;
    }

    validateField(input) {
        const value = input.value.trim();
        let isValid = true;
        let message = '';

        // Clear previous state
        this.clearFieldError(input);

        // Required validation
        if (!value) {
            isValid = false;
            message = `${input.previousElementSibling.textContent.replace(/^\w+\s/, '')} es requerido`;
        }

        // Specific validation based on field
        switch (input.id) {
            case 'username':
                if (value && value.length < 3) {
                    isValid = false;
                    message = 'El usuario debe tener al menos 3 caracteres';
                }
                break;
            case 'password':
                if (value && value.length < 6) {
                    isValid = false;
                    message = 'La contraseña debe tener al menos 6 caracteres';
                }
                break;
        }

        if (!isValid) {
            this.setFieldError(input, message);
        } else {
            this.setFieldSuccess(input);
        }

        return isValid;
    }

    setFieldError(input, message) {
        input.classList.add('is-invalid');
        input.classList.remove('is-valid');
        const feedback = input.parentNode.querySelector('.invalid-feedback') || 
                        input.nextElementSibling;
        if (feedback) {
            feedback.textContent = message;
        }
    }

    setFieldSuccess(input) {
        input.classList.add('is-valid');
        input.classList.remove('is-invalid');
    }

    clearFieldError(input) {
        input.classList.remove('is-invalid', 'is-valid');
    }

    setLoadingState(loading) {
        const btnText = this.loginBtn.querySelector('.btn-text');
        const btnLoading = this.loginBtn.querySelector('.btn-loading');
        const inputs = this.loginForm.querySelectorAll('input, button');

        if (loading) {
            btnText.classList.add('d-none');
            btnLoading.classList.remove('d-none');
            this.loginBtn.disabled = true;
            inputs.forEach(input => input.disabled = true);
        } else {
            btnText.classList.remove('d-none');
            btnLoading.classList.add('d-none');
            this.loginBtn.disabled = false;
            inputs.forEach(input => input.disabled = false);
        }
    }

    // === ALERT METHODS ===
    
    showAlert(message, type = 'info', dismissible = true) {
        const alertHtml = `
            <div class="alert alert-${type} ${dismissible ? 'alert-dismissible' : ''} fade show" role="alert">
                ${message}
                ${dismissible ? '<button type="button" class="btn-close" data-bs-dismiss="alert"></button>' : ''}
            </div>
        `;
        
        this.alertContainer.innerHTML = alertHtml;
        
        // Auto-dismiss after 5 seconds
        if (dismissible) {
            setTimeout(() => {
                const alert = this.alertContainer.querySelector('.alert');
                if (alert) {
                    const bsAlert = bootstrap.Alert.getInstance(alert);
                    if (bsAlert) bsAlert.close();
                }
            }, 5000);
        }
    }

    showSuccess(message) {
        this.showAlert(message, 'success');
    }

    showError(message) {
        this.showAlert(message, 'danger');
    }

    showWarning(message) {
        this.showAlert(message, 'warning');
    }

    showInfo(message) {
        this.showAlert(message, 'info');
    }

    clearAlerts() {
        this.alertContainer.innerHTML = '';
    }

    // === SYSTEM STATUS METHODS ===
    
    async checkSystemStatus() {
        try {
            const health = await authService.checkSystemHealth();
            this.updateStatusIndicators(health);
        } catch (error) {
            console.error('Status check error:', error);
            this.updateStatusIndicators({
                gateway: 'offline',
                services: {}
            });
        }
    }

    updateStatusIndicators(health) {
        // Gateway status
        const gatewayStatus = document.getElementById('gateway-status');
        this.setStatusIndicator(gatewayStatus, health.gateway);

        // Auth service status
        const authStatus = document.getElementById('auth-status');
        const authServiceStatus = health.services['auth-service'] || 'unknown';
        this.setStatusIndicator(authStatus, authServiceStatus === 'healthy' ? 'online' : 'offline');

        // Database status (inferred from auth service status)
        const databaseStatus = document.getElementById('database-status');
        this.setStatusIndicator(databaseStatus, authServiceStatus === 'healthy' ? 'online' : 'offline');
    }

    setStatusIndicator(element, status) {
        if (!element) return;

        // Clear existing classes
        element.classList.remove('online', 'offline', 'loading');
        
        // Set new status
        switch (status) {
            case 'online':
            case 'healthy':
                element.classList.add('online');
                break;
            case 'offline':
            case 'unhealthy':
                element.classList.add('offline');
                break;
            case 'degraded':
            case 'loading':
                element.classList.add('loading');
                break;
            default:
                element.classList.add('offline');
        }
    }

    // === AUTH STATE METHODS ===
    
    checkExistingAuth() {
        if (authService.isAuthenticated()) {
            const user = authService.getCurrentUser();
            if (user) {
                this.showInfo(`¡Bienvenido de nuevo, ${user.full_name}! Ya tienes una sesión iniciada.`);
                
                // Fill the form with remembered username if available
                const usernameInput = document.getElementById('username');
                if (user.username) {
                    usernameInput.value = user.username;
                }
                
                // Check remember me if user was remembered
                const rememberInput = document.getElementById('rememberMe');
                if (authService.isRemembered()) {
                    rememberInput.checked = true;
                }
            }
        }
    }

    // === UTILITY METHODS ===
    
    formatTimestamp(timestamp) {
        return new Date(timestamp).toLocaleString();
    }

    async testAuthEndpoints() {
        if (!authService.isAuthenticated()) {
            this.showWarning('Por favor inicia sesión primero para probar los endpoints de autenticación');
            return;
        }

        try {
            const profile = await authService.getProfile();
            console.log('Profile data:', profile);
            this.showSuccess('¡Los endpoints de autenticación funcionan! Revisa la consola para ver los datos del perfil.');
        } catch (error) {
            this.showError('Error en la prueba de endpoints de autenticación: ' + authService.formatError(error));
        }
    }
}

// === INITIALIZATION ===

document.addEventListener('DOMContentLoaded', () => {
    // Initialize the application
    window.loginApp = new LoginApp();
    
    // Add some global utility functions for testing
    window.testAuth = () => loginApp.testAuthEndpoints();
    window.checkStatus = () => loginApp.checkSystemStatus();
    window.clearAuth = () => {
        authService.clearAuthData();
        loginApp.showInfo('Datos de autenticación eliminados');
    };
    
    console.log('🍦 ¡Aplicación de Login de la Heladería inicializada!');
    console.log('Funciones de prueba disponibles: testAuth(), checkStatus(), clearAuth()');
}); 