// === ICE CREAM STORE ORDERS SERVICE ===

class OrdersService {
    constructor() {
        this.baseURL = 'http://localhost:8083/api/v1';
        this.orders = [];
        this.filteredOrders = [];
        this.currentSection = 'orders';
        this.init();
    }

    init() {
        this.checkAuthentication();
        this.setupEventListeners();
        this.loadOrders();
    }

    checkAuthentication() {
        if (!window.authService || !authService.isAuthenticated()) {
            alert('Please login first');
            window.location.href = '../auth/login.html';
            return;
        }
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = e.target.closest('.nav-link').dataset.section;
                this.showSection(section);
            });
        });

        // Create order form
        const createForm = document.getElementById('create-order-form');
        if (createForm) {
            createForm.addEventListener('submit', (e) => this.handleCreateOrder(e));
        }

        // Filters
        document.getElementById('status-filter')?.addEventListener('change', () => this.applyFilters());
        document.getElementById('search-input')?.addEventListener('input', () => this.applyFilters());
        document.getElementById('date-filter')?.addEventListener('change', () => this.applyFilters());
    }

    showSection(sectionName) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');

        // Show/hide sections
        document.querySelectorAll('.content-section').forEach(section => {
            section.classList.add('d-none');
        });
        document.getElementById(`${sectionName}-section`).classList.remove('d-none');

        this.currentSection = sectionName;

        // Load section-specific data
        switch(sectionName) {
            case 'orders':
                this.loadOrders();
                break;
            case 'stats':
                this.loadStatistics();
                break;
        }
    }

    async loadOrders() {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}/orders`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            this.orders = data.orders || [];
            this.filteredOrders = [...this.orders];
            this.renderOrders();
            this.updateStatistics();
        } catch (error) {
            console.error('Error loading orders:', error);
            this.renderError('Failed to load orders: ' + error.message);
        }
    }

    async loadStatistics() {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}/orders/summary`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            this.renderStatistics(data);
        } catch (error) {
            console.error('Error loading statistics:', error);
            document.getElementById('order-summary').innerHTML = 
                `<div class="alert alert-danger">Failed to load statistics: ${error.message}</div>`;
        }
    }

    renderOrders() {
        const container = document.getElementById('orders-list');
        
        if (this.filteredOrders.length === 0) {
            container.innerHTML = `
                <div class="text-center py-5">
                    <i class="fas fa-shopping-cart fa-3x text-muted mb-3"></i>
                    <h5 class="text-muted">No orders found</h5>
                    <p class="text-muted">Try adjusting your filters or create a new order</p>
                </div>
            `;
            return;
        }

        const ordersHTML = this.filteredOrders.map(order => `
            <div class="order-card">
                <div class="row align-items-center">
                    <div class="col-md-2">
                        <strong>Order #${order.order_id}</strong>
                        <div class="text-muted small">
                            ${this.formatDate(order.created_at)}
                        </div>
                    </div>
                    <div class="col-md-2">
                        <span class="status-badge status-${order.status}">
                            ${this.formatStatus(order.status)}
                        </span>
                    </div>
                    <div class="col-md-2">
                        <div class="text-muted small">Customer</div>
                        <strong>#${order.customer_id}</strong>
                    </div>
                    <div class="col-md-2">
                        <div class="text-muted small">Type</div>
                        <strong>${this.formatOrderType(order.order_type)}</strong>
                    </div>
                    <div class="col-md-2">
                        <div class="text-muted small">Total</div>
                        <strong>$${order.total_amount || '0.00'}</strong>
                    </div>
                    <div class="col-md-2">
                        <div class="d-grid gap-1">
                            <button class="btn btn-outline-primary btn-sm" onclick="ordersApp.viewOrder(${order.order_id})">
                                <i class="fas fa-eye"></i> View
                            </button>
                            ${order.status === 'pending' ? `
                                <button class="btn btn-outline-success btn-sm" onclick="ordersApp.updateOrderStatus(${order.order_id}, 'confirmed')">
                                    <i class="fas fa-check"></i> Confirm
                                </button>
                            ` : ''}
                        </div>
                    </div>
                </div>
                ${order.ordered_recipes && order.ordered_recipes.length > 0 ? `
                    <div class="mt-3 pt-3 border-top">
                        <div class="text-muted small mb-2">Order Items:</div>
                        <div class="row">
                            ${order.ordered_recipes.map(item => `
                                <div class="col-md-4">
                                    <span class="badge bg-light text-dark">
                                        Recipe #${item.recipe_id} × ${item.quantity}
                                    </span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}
            </div>
        `).join('');

        container.innerHTML = ordersHTML;
    }

    renderStatistics(data) {
        const container = document.getElementById('order-summary');
        
        container.innerHTML = `
            <div class="row">
                <div class="col-md-4">
                    <div class="text-center">
                        <h3>${data.total_orders || 0}</h3>
                        <p class="text-muted">Total Orders</p>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="text-center">
                        <h3>$${data.total_revenue || '0.00'}</h3>
                        <p class="text-muted">Total Revenue</p>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="text-center">
                        <h3>${data.average_order_value || '0.00'}</h3>
                        <p class="text-muted">Avg Order Value</p>
                    </div>
                </div>
            </div>
            
            ${data.status_breakdown ? `
                <hr>
                <h6>Order Status Breakdown:</h6>
                <div class="row">
                    ${Object.entries(data.status_breakdown).map(([status, count]) => `
                        <div class="col-md-3">
                            <div class="text-center">
                                <span class="status-badge status-${status}">${this.formatStatus(status)}</span>
                                <div class="mt-1"><strong>${count}</strong> orders</div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            ` : ''}
        `;
    }

    renderError(message) {
        const container = document.getElementById('orders-list');
        container.innerHTML = `
            <div class="alert alert-danger" role="alert">
                <i class="fas fa-exclamation-triangle me-2"></i>
                ${message}
                <button class="btn btn-outline-danger btn-sm ms-3" onclick="ordersApp.loadOrders()">
                    <i class="fas fa-retry me-1"></i>Retry
                </button>
            </div>
        `;
    }

    updateStatistics() {
        const stats = {
            total: this.orders.length,
            pending: this.orders.filter(o => o.status === 'pending').length,
            completed: this.orders.filter(o => ['delivered', 'completed'].includes(o.status)).length,
            revenue: this.orders.reduce((sum, o) => sum + (parseFloat(o.total_amount) || 0), 0)
        };

        document.getElementById('total-orders').textContent = stats.total;
        document.getElementById('pending-orders').textContent = stats.pending;
        document.getElementById('completed-orders').textContent = stats.completed;
        document.getElementById('total-revenue').textContent = `$${stats.revenue.toFixed(2)}`;
    }

    applyFilters() {
        const statusFilter = document.getElementById('status-filter').value;
        const searchFilter = document.getElementById('search-input').value.toLowerCase();
        const dateFilter = document.getElementById('date-filter').value;

        this.filteredOrders = this.orders.filter(order => {
            const statusMatch = !statusFilter || order.status === statusFilter;
            const searchMatch = !searchFilter || 
                order.order_id.toString().includes(searchFilter) ||
                order.customer_id.toString().includes(searchFilter);
            const dateMatch = !dateFilter || 
                order.created_at.startsWith(dateFilter);

            return statusMatch && searchMatch && dateMatch;
        });

        this.renderOrders();
    }

    async handleCreateOrder(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const customerId = formData.get('customer_id');
        const orderType = formData.get('order_type');
        
        const orderItems = [];
        document.querySelectorAll('.order-item').forEach(item => {
            const recipeId = item.querySelector('[name="recipe_id"]').value;
            const quantity = item.querySelector('[name="quantity"]').value;
            if (recipeId && quantity) {
                orderItems.push({
                    recipe_id: parseInt(recipeId),
                    quantity: parseInt(quantity)
                });
            }
        });

        if (orderItems.length === 0) {
            alert('Please add at least one order item');
            return;
        }

        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}/orders`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    customer_id: parseInt(customerId),
                    order_type: orderType,
                    ordered_recipes: orderItems
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const result = await response.json();
            alert(`Order created successfully! Order ID: ${result.order_id}`);
            
            // Reset form and go back to orders
            e.target.reset();
            this.showSection('orders');
        } catch (error) {
            console.error('Error creating order:', error);
            alert('Failed to create order: ' + error.message);
        }
    }

    async viewOrder(orderId) {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}/orders/${orderId}`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const order = await response.json();
            
            // Show order details in a modal or alert
            const details = `
Order ID: ${order.order_id}
Customer ID: ${order.customer_id}
Status: ${this.formatStatus(order.status)}
Type: ${this.formatOrderType(order.order_type)}
Created: ${this.formatDate(order.created_at)}
Total: $${order.total_amount || '0.00'}

Items:
${order.ordered_recipes ? order.ordered_recipes.map(item => 
    `- Recipe #${item.recipe_id} × ${item.quantity}`
).join('\n') : 'No items'}
            `;
            
            alert(details);
        } catch (error) {
            console.error('Error viewing order:', error);
            alert('Failed to load order details: ' + error.message);
        }
    }

    async updateOrderStatus(orderId, newStatus) {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}/orders/${orderId}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    status: newStatus
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            alert(`Order status updated to ${this.formatStatus(newStatus)}`);
            this.loadOrders(); // Refresh orders
        } catch (error) {
            console.error('Error updating order:', error);
            alert('Failed to update order: ' + error.message);
        }
    }

    async testEndpoint(type) {
        const resultsDiv = document.getElementById('test-results');
        
        try {
            let url, requiresAuth = true;
            
            switch(type) {
                case 'health':
                    url = `${this.baseURL}/orders/health`;
                    requiresAuth = false;
                    break;
                case 'orders':
                    url = `${this.baseURL}/orders`;
                    break;
                case 'summary':
                    url = `${this.baseURL}/orders/summary`;
                    break;
                default:
                    throw new Error('Unknown test type');
            }

            const options = {
                headers: {
                    'Content-Type': 'application/json'
                }
            };

            if (requiresAuth) {
                const token = authService.getToken();
                options.headers['Authorization'] = `Bearer ${token}`;
            }

            resultsDiv.innerHTML = `<div class="alert alert-info">Testing ${type} endpoint...</div>`;

            const response = await fetch(url, options);
            const data = await response.json();

            if (response.ok) {
                resultsDiv.innerHTML = `
                    <div class="alert alert-success">
                        <strong>✅ Success</strong><br>
                        Status: ${response.status}<br>
                        Endpoint: ${type}
                    </div>
                    <pre class="bg-light p-3 rounded"><code>${JSON.stringify(data, null, 2)}</code></pre>
                `;
            } else {
                resultsDiv.innerHTML = `
                    <div class="alert alert-danger">
                        <strong>❌ Error</strong><br>
                        Status: ${response.status}<br>
                        Message: ${data.message || 'Unknown error'}
                    </div>
                `;
            }
        } catch (error) {
            resultsDiv.innerHTML = `
                <div class="alert alert-danger">
                    <strong>❌ Network Error</strong><br>
                    ${error.message}
                </div>
            `;
        }
    }

    addOrderItem() {
        const container = document.getElementById('order-items');
        const itemHTML = `
            <div class="row mb-3 order-item">
                <div class="col-md-6">
                    <label class="form-label">Recipe ID</label>
                    <input type="number" class="form-control" name="recipe_id" required>
                </div>
                <div class="col-md-4">
                    <label class="form-label">Quantity</label>
                    <input type="number" class="form-control" name="quantity" min="1" value="1" required>
                </div>
                <div class="col-md-2">
                    <label class="form-label">&nbsp;</label>
                    <button type="button" class="btn btn-outline-danger w-100" onclick="removeOrderItem(this)">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `;
        container.insertAdjacentHTML('beforeend', itemHTML);
    }

    formatDate(dateString) {
        return new Date(dateString).toLocaleString();
    }

    formatStatus(status) {
        return status.charAt(0).toUpperCase() + status.slice(1).replace('_', ' ');
    }

    formatOrderType(type) {
        return type.charAt(0).toUpperCase() + type.slice(1).replace('_', ' ');
    }
}

// Global functions
function removeOrderItem(button) {
    const item = button.closest('.order-item');
    const container = document.getElementById('order-items');
    if (container.children.length > 1) {
        item.remove();
    } else {
        alert('At least one order item is required');
    }
}

function refreshOrders() {
    ordersApp.loadOrders();
}

function applyFilters() {
    ordersApp.applyFilters();
}

function goBack() {
    window.location.href = '../index.html';
}

function logout() {
    if (confirm('Are you sure you want to logout?')) {
        authService.logout();
        window.location.href = '../auth/login.html';
    }
}

// Global test functions
function testEndpoint(type) {
    ordersApp.testEndpoint(type);
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    window.ordersApp = new OrdersService();
}); 