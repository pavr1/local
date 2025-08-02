/**
 * Utility to load HTML partials into pages
 * Extended with layout framework functionality
 */

class PartialLoader {
    static async loadPartial(partialPath, targetElementId) {
        try {
            const response = await fetch(partialPath);
            if (!response.ok) {
                throw new Error(`Failed to load partial: ${response.status}`);
            }
            
            const html = await response.text();
            const targetElement = document.getElementById(targetElementId);
            
            if (!targetElement) {
                throw new Error(`Target element with ID '${targetElementId}' not found`);
            }
            
            targetElement.innerHTML = html;
            console.log(`✅ Loaded partial: ${partialPath} into #${targetElementId}`);
            
            return true;
        } catch (error) {
            console.error(`❌ Failed to load partial ${partialPath}:`, error);
            return false;
        }
    }
    
    static async loadSystemStatus(targetElementId = 'system-status-container') {
        // Determine the correct path based on current location
        const currentPath = window.location.pathname;
        const isInSubdirectory = currentPath.includes('/inventory/') || currentPath.includes('/expense/');
        const partialPath = isInSubdirectory ? '../shared/partials/system-status.html' : 'shared/partials/system-status.html';
        
        return await this.loadPartial(partialPath, targetElementId);
    }
    
    /**
     * Load the complete layout framework (header + sidebar)
     * @param {Object} options - Configuration options
     * @param {string} options.headerContainer - ID of header container (default: 'layout-header-container')
     * @param {string} options.sidebarContainer - ID of sidebar container (default: 'layout-sidebar-container')  
     * @param {boolean} options.includeSidebar - Whether to include sidebar (default: true)
     * @param {string} options.pageTitle - Custom page title
     * @returns {Promise<boolean>} - Success status
     */
    static async loadLayoutFramework(options = {}) {
        const {
            headerContainer = 'layout-header-container',
            sidebarContainer = 'layout-sidebar-container',
            includeSidebar = true,
            pageTitle = null
        } = options;
        
        console.log('🎨 Loading layout framework...');
        
        try {
            // Determine correct paths
            const currentPath = window.location.pathname;
            const isInSubdirectory = currentPath.includes('/inventory/') || currentPath.includes('/expense/');
            const pathPrefix = isInSubdirectory ? '../' : '';
            
            // Load header
            const headerLoaded = await this.loadPartial(
                `${pathPrefix}shared/partials/layout-header.html`,
                headerContainer
            );
            
            if (!headerLoaded) {
                console.warn('⚠️ Failed to load header, continuing without it');
            }
            
            // Load sidebar if requested
            let sidebarLoaded = true;
            if (includeSidebar) {
                sidebarLoaded = await this.loadPartial(
                    `${pathPrefix}shared/partials/layout-sidebar.html`,
                    sidebarContainer
                );
                
                if (!sidebarLoaded) {
                    console.warn('⚠️ Failed to load sidebar, continuing without it');
                }
            }
            
            // Update page title if provided
            if (pageTitle) {
                document.title = `Ice Cream Store - ${pageTitle}`;
            }
            
            console.log(`✅ Layout framework loaded successfully (header: ${headerLoaded}, sidebar: ${sidebarLoaded})`);
            return headerLoaded || sidebarLoaded;
            
        } catch (error) {
            console.error('❌ Failed to load layout framework:', error);
            return false;
        }
    }
    
    /**
     * Initialize a standardized page layout
     * This creates the basic layout structure if it doesn't exist
     * @param {Object} options - Layout options
     * @returns {boolean} - Success status
     */
    static initializePageLayout(options = {}) {
        const {
            containerClass = 'layout-container',
            includeHeader = true,
            includeSidebar = true,
            contentClass = 'layout-content'
        } = options;
        
        // Check if layout already exists
        if (document.querySelector(`.${containerClass}`)) {
            console.log('📐 Layout already exists, skipping initialization');
            return true;
        }
        
        try {
            // Find the body or a main container
            const body = document.body;
            const existingContent = body.innerHTML;
            
            // Create the layout structure
            let layoutHTML = `<div class="${containerClass}">`;
            
            if (includeHeader) {
                layoutHTML += `<div id="layout-header-container"></div>`;
            }
            
            layoutHTML += `<div class="layout-main">`;
            
            if (includeSidebar) {
                layoutHTML += `
                    <div class="row g-0">
                        <div class="col-lg-3 col-xl-2">
                            <div id="layout-sidebar-container"></div>
                        </div>
                        <div class="col-lg-9 col-xl-10">
                            <div class="${contentClass}">${existingContent}</div>
                        </div>
                    </div>
                `;
            } else {
                layoutHTML += `<div class="${contentClass}">${existingContent}</div>`;
            }
            
            layoutHTML += `</div></div>`;
            
            body.innerHTML = layoutHTML;
            
            console.log('📐 Page layout structure initialized');
            return true;
            
        } catch (error) {
            console.error('❌ Failed to initialize page layout:', error);
            return false;
        }
    }
    
    /**
     * Complete page setup with standardized layout
     * This is the main function to call from pages that want the full layout
     * @param {Object} options - Setup options
     * @returns {Promise<boolean>} - Success status
     */
    static async setupStandardLayout(options = {}) {
        const {
            pageTitle = null,
            includeSidebar = true,
            includeHeader = true,
            initializeLayout = true
        } = options;
        
        console.log('🏗️ Setting up standardized layout...');
        
        try {
            // Initialize layout structure if needed
            if (initializeLayout) {
                this.initializePageLayout({
                    includeHeader,
                    includeSidebar
                });
            }
            
            // Load layout components
            const layoutLoaded = await this.loadLayoutFramework({
                includeSidebar,
                pageTitle
            });
            
            if (layoutLoaded) {
                // Add layout-specific styling
                this.addLayoutStyles();
                console.log('✅ Standardized layout setup complete');
                return true;
            } else {
                console.warn('⚠️ Layout setup completed with warnings');
                return false;
            }
            
        } catch (error) {
            console.error('❌ Failed to setup standardized layout:', error);
            return false;
        }
    }
    
    /**
     * Add layout-specific CSS styles
     */
    static addLayoutStyles() {
        // Check if styles already added
        if (document.getElementById('layout-framework-styles')) {
            return;
        }
        
        const styles = document.createElement('style');
        styles.id = 'layout-framework-styles';
        styles.textContent = `
            .layout-container {
                min-height: 100vh;
                display: flex;
                flex-direction: column;
            }
            
            .layout-main {
                flex: 1;
                overflow: hidden;
            }
            
            .layout-content {
                padding: 2rem;
                min-height: calc(100vh - 80px);
                background: var(--white);
            }
            
            .layout-content.with-sidebar {
                background: #f8f9fa;
            }
            
            @media (max-width: 992px) {
                .layout-content {
                    padding: 1rem;
                }
            }
            
            /* Smooth transitions */
            .layout-container * {
                transition: var(--transition, all 0.2s ease);
            }
            
            /* Mobile layout adjustments */
            @media (max-width: 768px) {
                .layout-main .row .col-lg-3 {
                    order: 2;
                }
                
                .layout-main .row .col-lg-9 {
                    order: 1;
                }
            }
        `;
        
        document.head.appendChild(styles);
        console.log('🎨 Layout framework styles added');
    }
}

// Logout function (shared across all pages)
function logout() {
    console.log('🚪 Logging out...');
    
    try {
        if (typeof AuthService !== 'undefined') {
            const auth = new AuthService();
            auth.logout();
        } else {
            // Fallback logout
            localStorage.removeItem('icecream_auth_token');
            localStorage.removeItem('icecream_user_data');
            sessionStorage.clear();
        }
        
        // Redirect to login
        window.location.href = window.location.pathname.includes('/') ? '../login.html' : 'login.html';
        
    } catch (error) {
        console.error('❌ Logout error:', error);
        // Still redirect even if logout fails
        window.location.href = 'login.html';
    }
}

// Make available globally
window.PartialLoader = PartialLoader; 