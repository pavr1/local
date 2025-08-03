/**
 * SPA Router for Ice Cream Store Management Dashboard
 * Handles navigation and content loading without page reloads
 */
class SPARouter {
    constructor() {
        this.routes = {};
        this.currentRoute = null;
        this.setupRoutes();
    }

    async init() {
        console.log('🧭 Initializing SPA Router...');
        this.setupRoutes();
        console.log('✅ SPA Router initialized with', Object.keys(this.routes).length, 'routes');
    }

    setupRoutes() {
        this.routes = {
            'dashboard': {
                title: 'Dashboard',
                breadcrumb: ['Dashboard'],
                contentUrl: 'partials/dashboard-content.html',
                initFunction: 'initDashboardContent'
            },
            'orders': {
                title: 'Orders Management',
                breadcrumb: ['Dashboard', 'Orders'],
                contentUrl: 'partials/orders-content.html',
                initFunction: 'initOrdersContent'
            },
            'inventory': {
                title: 'Inventory Overview',
                breadcrumb: ['Dashboard', 'Inventory', 'Overview'],
                contentUrl: 'partials/inventory-content.html',
                initFunction: 'initInventoryContent'
            },
            'inventory/suppliers': {
                title: 'Suppliers Management',
                breadcrumb: ['Dashboard', 'Inventory', 'Suppliers'],
                contentUrl: 'partials/inventory-suppliers-content.html',
                initFunction: 'initSuppliersContent'
            },
            'inventory/ingredients': {
                title: 'Ingredients Management',
                breadcrumb: ['Dashboard', 'Inventory', 'Ingredients'],
                contentUrl: 'partials/inventory-ingredients-content.html',
                initFunction: 'initIngredientsContent'
            },
            'invoice': {
                title: 'Invoice Management',
                breadcrumb: ['Dashboard', 'Invoice', 'Overview'],
                contentUrl: 'partials/invoice-content.html',
                initFunction: 'initInvoiceContent'
            },
            'invoice/invoices': {
                title: 'Invoice Details',
                breadcrumb: ['Dashboard', 'Invoice', 'Details'],
                contentUrl: 'partials/invoice-invoices-content.html',
                initFunction: 'initInvoiceDetailsContent'
            }
        };
    }

    async navigate(route, updateUrl = true) {
        console.log('🧭 Navigating to route:', route);
        
        const routeConfig = this.routes[route];
        if (!routeConfig) {
            console.error('❌ Route not found:', route);
            await this.navigate('dashboard');
            return;
        }

        try {
            // Show loading state
            this.showLoading();
            
            // Update URL hash if requested
            if (updateUrl) {
                window.location.hash = route;
            }
            
            // Update page title
            document.title = `Ice Cream Store - ${routeConfig.title}`;
            
            // Update breadcrumb
            this.updateBreadcrumb(routeConfig.breadcrumb);
            
            // Update active navigation
            this.updateActiveNavigation(route);
            
            // Load content
            await this.loadContent(routeConfig);
            
            // Store current route
            this.currentRoute = route;
            
            console.log('✅ Successfully navigated to:', route);
            
        } catch (error) {
            console.error('❌ Error navigating to route:', route, error);
            Alert.error('Navigation Error', 'Failed to load page content. Please try again.');
            this.hideLoading();
        }
    }

    async loadContent(routeConfig) {
        const contentContainer = document.getElementById('appContent');
        
        try {
            // Load the content HTML
            const response = await fetch(routeConfig.contentUrl);
            if (!response.ok) {
                throw new Error(`Failed to load content: ${response.status}`);
            }
            
            const html = await response.text();
            contentContainer.innerHTML = html;
            
            // Initialize page-specific functionality
            if (routeConfig.initFunction && typeof window[routeConfig.initFunction] === 'function') {
                console.log('🔧 Initializing page-specific functionality:', routeConfig.initFunction);
                await window[routeConfig.initFunction]();
            }
            
            // Hide loading state
            this.hideLoading();
            
        } catch (error) {
            console.error('❌ Error loading content:', error);
            contentContainer.innerHTML = this.getErrorContent(error.message);
            this.hideLoading();
        }
    }

    showLoading() {
        const contentContainer = document.getElementById('appContent');
        contentContainer.innerHTML = `
            <div class="content-loader">
                <div class="loading-spinner"></div>
                <span>Loading content...</span>
            </div>
        `;
    }

    hideLoading() {
        // Loading is hidden when content is loaded
    }

    getErrorContent(errorMessage) {
        return `
            <div class="text-center py-5">
                <i class="fas fa-exclamation-triangle fa-4x text-warning mb-3"></i>
                <h3>Content Loading Error</h3>
                <p class="text-muted mb-4">${errorMessage}</p>
                <button class="btn btn-primary" onclick="location.reload()">
                    <i class="fas fa-refresh me-2"></i>Refresh Page
                </button>
            </div>
        `;
    }

    updateBreadcrumb(breadcrumbItems) {
        const breadcrumbContainer = document.getElementById('appBreadcrumb');
        if (!breadcrumbContainer) return;

        const items = breadcrumbItems.map((item, index) => {
            const isLast = index === breadcrumbItems.length - 1;
            if (isLast) {
                return `<li class="breadcrumb-item active">${item}</li>`;
            } else {
                return `<li class="breadcrumb-item"><a href="#" class="text-decoration-none">${item}</a></li>`;
            }
        }).join('');

        breadcrumbContainer.innerHTML = items;
    }

    updateActiveNavigation(route) {
        // Remove all active classes
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.classList.remove('active');
        });

        // Add active class to current route
        const activeLink = document.querySelector(`[data-route="${route}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
            
            // If it's a submenu item, also activate parent
            const submenu = activeLink.closest('.collapse');
            if (submenu) {
                const parentToggle = document.querySelector(`[data-bs-target="#${submenu.id}"]`);
                if (parentToggle) {
                    parentToggle.classList.add('active');
                    submenu.classList.add('show');
                }
            }
        }
    }

    // Helper method to get current route
    getCurrentRoute() {
        return this.currentRoute;
    }

    // Helper method to check if a route exists
    routeExists(route) {
        return this.routes.hasOwnProperty(route);
    }
}

// Global navigation helper functions
window.navigateToRoute = function(route) {
    if (window.spaDashboard && window.spaDashboard.router) {
        window.spaDashboard.router.navigate(route);
    }
};

// Initialize global router when this script loads
window.SPARouter = SPARouter; 