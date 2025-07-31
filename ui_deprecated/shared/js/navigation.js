// === SHARED NAVIGATION UTILITIES ===

class NavigationService {
    constructor() {
        this.routes = {
            home: '../session/dashboard.html',
            session: {
                login: '../session/login.html',
                dashboard: '../session/dashboard.html'
            },
            orders: {
                index: '../orders/index.html'
            }
        };
    }

    // Navigate to a specific route
    navigateTo(service, page = 'index') {
        if (service === 'home') {
            window.location.href = this.routes.home;
            return;
        }

        if (this.routes[service] && this.routes[service][page]) {
            window.location.href = this.routes[service][page];
        } else {
            console.error(`Route not found: ${service}.${page}`);
        }
    }

    // Check if user is authenticated before navigation
    navigateSecure(service, page = 'index') {
        if (!window.authService || !authService.isAuthenticated()) {
            alert('Please login first');
            this.navigateTo('session', 'login');
            return;
        }
        this.navigateTo(service, page);
    }

    // Go back to previous page
    goBack() {
        if (document.referrer && document.referrer.includes(window.location.host)) {
            window.history.back();
        } else {
            this.navigateTo('home');
        }
    }

    // Logout and redirect
    logout() {
        if (window.authService) {
            authService.logout();
        }
        this.navigateTo('auth', 'login');
    }

    // Get current service name based on URL
    getCurrentService() {
        const path = window.location.pathname;
        if (path.includes('/auth/')) return 'auth';
        if (path.includes('/orders/')) return 'orders';
        return 'home';
    }

    // Update page title based on service
    updateTitle(subtitle = '') {
        const service = this.getCurrentService();
        const baseTitle = 'Ice Cream Store';
        
        const serviceTitles = {
            home: 'Management System',
            auth: 'Authentication',
            orders: 'Orders Management'
        };

        const serviceTitle = serviceTitles[service] || 'Service';
        document.title = subtitle ? 
            `${baseTitle} - ${serviceTitle} - ${subtitle}` : 
            `${baseTitle} - ${serviceTitle}`;
    }

    // Create breadcrumb navigation
    createBreadcrumb(items = []) {
        const service = this.getCurrentService();
        const breadcrumbItems = [
            { text: 'Home', url: this.routes.home },
            { text: this.getServiceDisplayName(service), url: null, active: items.length === 0 }
        ];

        items.forEach((item, index) => {
            breadcrumbItems.push({
                text: item.text,
                url: item.url || null,
                active: index === items.length - 1
            });
        });

        return this.renderBreadcrumb(breadcrumbItems);
    }

    renderBreadcrumb(items) {
        const breadcrumbHTML = items.map(item => {
            if (item.active) {
                return `<li class="breadcrumb-item active">${item.text}</li>`;
            } else if (item.url) {
                return `<li class="breadcrumb-item"><a href="${item.url}">${item.text}</a></li>`;
            } else {
                return `<li class="breadcrumb-item">${item.text}</li>`;
            }
        }).join('');

        return `
            <nav aria-label="breadcrumb">
                <ol class="breadcrumb">
                    ${breadcrumbHTML}
                </ol>
            </nav>
        `;
    }

    getServiceDisplayName(service) {
        const names = {
            home: 'Home',
            auth: 'Authentication',
            orders: 'Orders'
        };
        return names[service] || service;
    }
}

// Create global navigation instance
window.navigationService = new NavigationService();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = NavigationService;
} 