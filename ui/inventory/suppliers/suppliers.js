// === ICE CREAM STORE SUPPLIERS SERVICE ===

class SuppliersService {
    constructor() {
        this.baseURL = CONFIG.GATEWAY_URL; // Use the gateway URL
        this.apiPath = '/api/v1/inventory/suppliers';
        this.suppliers = [];
        this.filteredSuppliers = [];
        this.currentSection = 'suppliers';
        this.currentEditingId = null;
        this.init();
    }

    init() {
        this.checkAuthentication();
        this.setupEventListeners();
        this.loadSuppliers();
    }

    checkAuthentication() {
        if (!window.authService || !authService.isAuthenticated()) {
            alert('Please login first');
            window.location.href = '../../session/login.html';
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

        // Search and filters
        document.getElementById('search-input')?.addEventListener('input', () => this.applyFilters());
        document.getElementById('sort-filter')?.addEventListener('change', () => this.applyFilters());

        // Form submission
        const supplierForm = document.getElementById('supplier-form');
        if (supplierForm) {
            supplierForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.saveSupplier();
            });
        }
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
            case 'suppliers':
                this.loadSuppliers();
                break;
            case 'stats':
                this.loadStatistics();
                break;
        }
    }

    async loadSuppliers() {
        try {
            this.showLoading('suppliers-list');
            
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}${this.apiPath}`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            this.suppliers = data.data || [];
            this.filteredSuppliers = [...this.suppliers];
            this.renderSuppliers();
            this.updateStatistics();
        } catch (error) {
            console.error('Error loading suppliers:', error);
            this.renderError('suppliers-list', 'Failed to load suppliers: ' + error.message);
        }
    }

    async createSupplier(supplierData) {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}${this.apiPath}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(supplierData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to create supplier');
            }

            const result = await response.json();
            this.showSuccess('Supplier created successfully!');
            return result;
        } catch (error) {
            console.error('Error creating supplier:', error);
            this.showError('Failed to create supplier: ' + error.message);
            throw error;
        }
    }

    async updateSupplier(id, supplierData) {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}${this.apiPath}/${id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(supplierData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to update supplier');
            }

            const result = await response.json();
            this.showSuccess('Supplier updated successfully!');
            return result;
        } catch (error) {
            console.error('Error updating supplier:', error);
            this.showError('Failed to update supplier: ' + error.message);
            throw error;
        }
    }

    async deleteSupplier(id) {
        try {
            const token = authService.getToken();
            const response = await fetch(`${this.baseURL}${this.apiPath}/${id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to delete supplier');
            }

            this.showSuccess('Supplier deleted successfully!');
            this.loadSuppliers(); // Reload the list
        } catch (error) {
            console.error('Error deleting supplier:', error);
            this.showError('Failed to delete supplier: ' + error.message);
        }
    }

    renderSuppliers() {
        const container = document.getElementById('suppliers-list');
        
        if (this.filteredSuppliers.length === 0) {
            container.innerHTML = `
                <div class="text-center py-5">
                    <i class="fas fa-truck fa-4x text-muted mb-3"></i>
                    <h5>No suppliers found</h5>
                    <p class="text-muted">Start by adding your first supplier</p>
                    <button class="btn btn-primary" onclick="suppliersService.showCreateModal()">
                        <i class="fas fa-plus me-2"></i>Add First Supplier
                    </button>
                </div>
            `;
            return;
        }

        const suppliersHTML = this.filteredSuppliers.map(supplier => {
            const initials = this.getSupplierInitials(supplier.supplier_name || 'Unknown');
            const createdDate = new Date(supplier.created_at).toLocaleDateString();
            
            return `
                <div class="supplier-card">
                    <div class="row align-items-center">
                        <div class="col-auto">
                            <div class="supplier-avatar">
                                ${initials}
                            </div>
                        </div>
                        <div class="col">
                            <h5 class="mb-1">${supplier.supplier_name || 'Unnamed Supplier'}</h5>
                            <p class="text-muted mb-1">
                                <i class="fas fa-user me-2"></i>${supplier.contact_person || 'No contact person'}
                            </p>
                            <div class="row">
                                <div class="col-md-6">
                                    <small class="text-muted">
                                        <i class="fas fa-envelope me-1"></i>
                                        ${supplier.email || 'No email'}
                                    </small>
                                </div>
                                <div class="col-md-6">
                                    <small class="text-muted">
                                        <i class="fas fa-phone me-1"></i>
                                        ${supplier.phone || 'No phone'}
                                    </small>
                                </div>
                            </div>
                        </div>
                        <div class="col-auto">
                            <div class="supplier-actions">
                                <button class="btn btn-sm btn-outline-primary" 
                                        onclick="suppliersService.viewSupplier('${supplier.id}')"
                                        title="View Details">
                                    <i class="fas fa-eye"></i>
                                </button>
                                <button class="btn btn-sm btn-outline-success" 
                                        onclick="suppliersService.editSupplier('${supplier.id}')"
                                        title="Edit">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-sm btn-outline-danger" 
                                        onclick="suppliersService.confirmDelete('${supplier.id}')"
                                        title="Delete">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    ${supplier.address ? `
                        <div class="row mt-2">
                            <div class="col">
                                <small class="text-muted">
                                    <i class="fas fa-map-marker-alt me-1"></i>
                                    ${supplier.address}
                                </small>
                            </div>
                        </div>
                    ` : ''}
                    <div class="row mt-2">
                        <div class="col">
                            <small class="text-muted">
                                <i class="fas fa-calendar me-1"></i>
                                Added: ${createdDate}
                            </small>
                        </div>
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = suppliersHTML;
    }

    applyFilters() {
        const searchTerm = document.getElementById('search-input')?.value.toLowerCase() || '';
        const sortBy = document.getElementById('sort-filter')?.value || 'name';

        // Filter by search term
        this.filteredSuppliers = this.suppliers.filter(supplier => {
            const name = (supplier.supplier_name || '').toLowerCase();
            const contact = (supplier.contact_person || '').toLowerCase();
            const email = (supplier.email || '').toLowerCase();
            
            return name.includes(searchTerm) || 
                   contact.includes(searchTerm) || 
                   email.includes(searchTerm);
        });

        // Sort
        this.filteredSuppliers.sort((a, b) => {
            switch(sortBy) {
                case 'name':
                    return (a.supplier_name || '').localeCompare(b.supplier_name || '');
                case 'created_at':
                    return new Date(b.created_at) - new Date(a.created_at);
                case 'updated_at':
                    return new Date(b.updated_at) - new Date(a.updated_at);
                default:
                    return 0;
            }
        });

        this.renderSuppliers();
    }

    showCreateModal() {
        this.currentEditingId = null;
        document.getElementById('supplierModalTitle').innerHTML = 
            '<i class="fas fa-plus me-2"></i>Add New Supplier';
        this.clearForm();
        new bootstrap.Modal(document.getElementById('supplierModal')).show();
    }

    editSupplier(id) {
        const supplier = this.suppliers.find(s => s.id === id);
        if (!supplier) return;

        this.currentEditingId = id;
        document.getElementById('supplierModalTitle').innerHTML = 
            '<i class="fas fa-edit me-2"></i>Edit Supplier';
        
        // Populate form
        document.getElementById('supplier-id').value = supplier.id;
        document.getElementById('supplier-name').value = supplier.supplier_name || '';
        document.getElementById('contact-person').value = supplier.contact_person || '';
        document.getElementById('supplier-email').value = supplier.email || '';
        document.getElementById('supplier-phone').value = supplier.phone || '';
        document.getElementById('supplier-address').value = supplier.address || '';

        new bootstrap.Modal(document.getElementById('supplierModal')).show();
    }

    viewSupplier(id) {
        const supplier = this.suppliers.find(s => s.id === id);
        if (!supplier) return;

        const createdDate = new Date(supplier.created_at).toLocaleDateString();
        const updatedDate = new Date(supplier.updated_at).toLocaleDateString();

        const detailsHTML = `
            <div class="row">
                <div class="col-md-6">
                    <div class="mb-3">
                        <label class="form-label fw-bold">Supplier Name</label>
                        <p class="form-control-plaintext">${supplier.supplier_name || 'Not specified'}</p>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-bold">Contact Person</label>
                        <p class="form-control-plaintext">${supplier.contact_person || 'Not specified'}</p>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-bold">Email</label>
                        <p class="form-control-plaintext">
                            ${supplier.email ? `<a href="mailto:${supplier.email}">${supplier.email}</a>` : 'Not specified'}
                        </p>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="mb-3">
                        <label class="form-label fw-bold">Phone</label>
                        <p class="form-control-plaintext">
                            ${supplier.phone ? `<a href="tel:${supplier.phone}">${supplier.phone}</a>` : 'Not specified'}
                        </p>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-bold">Created</label>
                        <p class="form-control-plaintext">${createdDate}</p>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-bold">Last Updated</label>
                        <p class="form-control-plaintext">${updatedDate}</p>
                    </div>
                </div>
            </div>
            ${supplier.address ? `
                <div class="mb-3">
                    <label class="form-label fw-bold">Address</label>
                    <p class="form-control-plaintext">${supplier.address}</p>
                </div>
            ` : ''}
        `;

        document.getElementById('supplier-details').innerHTML = detailsHTML;
        new bootstrap.Modal(document.getElementById('viewSupplierModal')).show();
    }

    async saveSupplier() {
        const form = document.getElementById('supplier-form');
        if (!form.checkValidity()) {
            form.reportValidity();
            return;
        }

        const supplierData = {
            supplier_name: document.getElementById('supplier-name').value.trim(),
            contact_person: document.getElementById('contact-person').value.trim(),
            email: document.getElementById('supplier-email').value.trim(),
            phone: document.getElementById('supplier-phone').value.trim(),
            address: document.getElementById('supplier-address').value.trim()
        };

        try {
            if (this.currentEditingId) {
                await this.updateSupplier(this.currentEditingId, supplierData);
            } else {
                await this.createSupplier(supplierData);
            }

            // Close modal and reload
            bootstrap.Modal.getInstance(document.getElementById('supplierModal')).hide();
            this.loadSuppliers();
        } catch (error) {
            // Error already handled in the API methods
        }
    }

    confirmDelete(id) {
        const supplier = this.suppliers.find(s => s.id === id);
        if (!supplier) return;

        if (confirm(`Are you sure you want to delete supplier "${supplier.supplier_name || 'Unnamed'}"?\n\nThis action cannot be undone.`)) {
            this.deleteSupplier(id);
        }
    }

    clearForm() {
        document.getElementById('supplier-form').reset();
        document.getElementById('supplier-id').value = '';
    }

    loadStatistics() {
        const total = this.suppliers.length;
        const currentMonth = new Date().getMonth();
        const currentYear = new Date().getFullYear();
        
        const recentSuppliers = this.suppliers.filter(supplier => {
            const createdDate = new Date(supplier.created_at);
            return createdDate.getMonth() === currentMonth && 
                   createdDate.getFullYear() === currentYear;
        }).length;

        const suppliersWithEmail = this.suppliers.filter(s => s.email && s.email.trim()).length;
        const suppliersWithPhone = this.suppliers.filter(s => s.phone && s.phone.trim()).length;

        this.updateStatistics({
            total,
            recent: recentSuppliers,
            withEmail: suppliersWithEmail,
            withPhone: suppliersWithPhone
        });
    }

    updateStatistics(stats = null) {
        if (!stats) {
            const total = this.suppliers.length;
            const currentMonth = new Date().getMonth();
            const currentYear = new Date().getFullYear();
            
            const recent = this.suppliers.filter(supplier => {
                const createdDate = new Date(supplier.created_at);
                return createdDate.getMonth() === currentMonth && 
                       createdDate.getFullYear() === currentYear;
            }).length;

            const withEmail = this.suppliers.filter(s => s.email && s.email.trim()).length;
            const withPhone = this.suppliers.filter(s => s.phone && s.phone.trim()).length;

            stats = { total, recent, withEmail, withPhone };
        }

        document.getElementById('total-suppliers').textContent = stats.total;
        document.getElementById('recent-suppliers').textContent = stats.recent;
        document.getElementById('suppliers-with-email').textContent = stats.withEmail;
        document.getElementById('suppliers-with-phone').textContent = stats.withPhone;
    }

    getSupplierInitials(name) {
        if (!name || name.trim() === '') return '?';
        
        const parts = name.trim().split(' ');
        if (parts.length === 1) {
            return parts[0].charAt(0).toUpperCase();
        }
        
        return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
    }

    showLoading(containerId) {
        const container = document.getElementById(containerId);
        container.innerHTML = `
            <div class="text-center py-5">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
                <p class="mt-3 text-muted">Loading suppliers...</p>
            </div>
        `;
    }

    renderError(containerId, message) {
        const container = document.getElementById(containerId);
        container.innerHTML = `
            <div class="text-center py-5">
                <i class="fas fa-exclamation-triangle fa-4x text-warning mb-3"></i>
                <h5>Error Loading Data</h5>
                <p class="text-muted">${message}</p>
                <button class="btn btn-primary" onclick="suppliersService.loadSuppliers()">
                    <i class="fas fa-sync me-2"></i>Try Again
                </button>
            </div>
        `;
    }

    showSuccess(message) {
        this.showAlert(message, 'success');
    }

    showError(message) {
        this.showAlert(message, 'danger');
    }

    showAlert(message, type) {
        // Create alert element
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
        alertDiv.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;

        document.body.appendChild(alertDiv);

        // Auto remove after 5 seconds
        setTimeout(() => {
            if (alertDiv && alertDiv.parentNode) {
                alertDiv.parentNode.removeChild(alertDiv);
            }
        }, 5000);
    }
} 