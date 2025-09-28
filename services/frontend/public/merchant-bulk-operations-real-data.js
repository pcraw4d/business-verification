/**
 * Merchant Bulk Operations with Real Data Integration
 * Replaces mock data with real Supabase API calls
 * Provides comprehensive bulk merchant management with live data
 */

class MerchantBulkOperationsRealData {
    constructor() {
        this.dataIntegration = new RealDataIntegration();
        this.merchants = [];
        this.selectedMerchants = new Set();
        this.filters = {
            status: 'all',
            industry: 'all',
            riskLevel: 'all',
            search: ''
        };
        this.sorting = {
            field: 'created_at',
            direction: 'desc'
        };
        this.pagination = {
            page: 1,
            limit: 50,
            total: 0
        };
        this.isLoading = false;
        this.refreshInterval = null;
        
        this.init();
    }

    /**
     * Initialize the component
     */
    async init() {
        try {
            this.showLoadingState();
            await this.loadMerchants();
            this.bindEvents();
            this.startAutoRefresh();
        } catch (error) {
            console.error('Failed to initialize bulk operations:', error);
            this.showErrorState(error.message);
        }
    }

    /**
     * Load merchants from Supabase
     */
    async loadMerchants() {
        try {
            this.isLoading = true;
            
            const response = await this.dataIntegration.getMerchants({
                page: this.pagination.page,
                limit: this.pagination.limit,
                filters: this.filters,
                sorting: this.sorting
            });
            
            this.merchants = response.merchants || [];
            this.pagination.total = response.total || 0;
            
            this.renderMerchantsTable();
            this.updatePagination();
            this.updateSelectionSummary();
            
            this.hideLoadingState();
        } catch (error) {
            console.error('Failed to load merchants:', error);
            this.showErrorState('Failed to load merchants');
        } finally {
            this.isLoading = false;
        }
    }

    /**
     * Render merchants table
     */
    renderMerchantsTable() {
        const tableBody = document.getElementById('merchantsTableBody');
        if (!tableBody) return;

        if (this.merchants.length === 0) {
            tableBody.innerHTML = `
                <tr>
                    <td colspan="8" class="text-center py-8 text-gray-500">
                        No merchants found matching your criteria
                    </td>
                </tr>
            `;
            return;
        }

        tableBody.innerHTML = this.merchants.map(merchant => `
            <tr class="merchant-row" data-merchant-id="${merchant.id}">
                <td class="px-4 py-3">
                    <input type="checkbox" 
                           class="merchant-checkbox" 
                           value="${merchant.id}"
                           ${this.selectedMerchants.has(merchant.id) ? 'checked' : ''}
                           onchange="bulkOperations.toggleMerchantSelection('${merchant.id}')">
                </td>
                <td class="px-4 py-3">
                    <div class="flex items-center">
                        <div class="merchant-avatar">
                            <i class="fas fa-building"></i>
                        </div>
                        <div class="ml-3">
                            <div class="merchant-name">${merchant.name || 'N/A'}</div>
                            <div class="merchant-id text-sm text-gray-500">${merchant.id}</div>
                        </div>
                    </div>
                </td>
                <td class="px-4 py-3">
                    <span class="status-badge ${merchant.status || 'unknown'}">
                        ${(merchant.status || 'unknown').charAt(0).toUpperCase() + (merchant.status || 'unknown').slice(1)}
                    </span>
                </td>
                <td class="px-4 py-3">
                    <div class="industry-info">
                        <div class="industry-name">${merchant.industry || 'N/A'}</div>
                        ${merchant.mcc_code ? `<div class="mcc-code text-sm text-gray-500">MCC: ${merchant.mcc_code}</div>` : ''}
                    </div>
                </td>
                <td class="px-4 py-3">
                    <span class="risk-badge ${this.getRiskLevel(merchant)}">
                        ${this.getRiskLevel(merchant).charAt(0).toUpperCase() + this.getRiskLevel(merchant).slice(1)}
                    </span>
                </td>
                <td class="px-4 py-3">
                    <div class="revenue-info">
                        ${merchant.monthly_revenue ? this.formatCurrency(merchant.monthly_revenue) : 'N/A'}
                    </div>
                </td>
                <td class="px-4 py-3">
                    <div class="date-info">
                        <div class="created-date">${this.formatDate(merchant.created_at)}</div>
                        <div class="updated-date text-sm text-gray-500">Updated: ${this.formatDate(merchant.updated_at)}</div>
                    </div>
                </td>
                <td class="px-4 py-3">
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-outline" onclick="bulkOperations.viewMerchant('${merchant.id}')">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="bulkOperations.editMerchant('${merchant.id}')">
                            <i class="fas fa-edit"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    /**
     * Update pagination controls
     */
    updatePagination() {
        const paginationContainer = document.getElementById('paginationContainer');
        if (!paginationContainer) return;

        const totalPages = Math.ceil(this.pagination.total / this.pagination.limit);
        const currentPage = this.pagination.page;

        let paginationHTML = '';

        // Previous button
        paginationHTML += `
            <button class="pagination-btn ${currentPage <= 1 ? 'disabled' : ''}" 
                    ${currentPage <= 1 ? 'disabled' : ''}
                    onclick="bulkOperations.goToPage(${currentPage - 1})">
                <i class="fas fa-chevron-left"></i>
            </button>
        `;

        // Page numbers
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, currentPage + 2);

        if (startPage > 1) {
            paginationHTML += `<button class="pagination-btn" onclick="bulkOperations.goToPage(1)">1</button>`;
            if (startPage > 2) {
                paginationHTML += `<span class="pagination-ellipsis">...</span>`;
            }
        }

        for (let i = startPage; i <= endPage; i++) {
            paginationHTML += `
                <button class="pagination-btn ${i === currentPage ? 'active' : ''}" 
                        onclick="bulkOperations.goToPage(${i})">
                    ${i}
                </button>
            `;
        }

        if (endPage < totalPages) {
            if (endPage < totalPages - 1) {
                paginationHTML += `<span class="pagination-ellipsis">...</span>`;
            }
            paginationHTML += `<button class="pagination-btn" onclick="bulkOperations.goToPage(${totalPages})">${totalPages}</button>`;
        }

        // Next button
        paginationHTML += `
            <button class="pagination-btn ${currentPage >= totalPages ? 'disabled' : ''}" 
                    ${currentPage >= totalPages ? 'disabled' : ''}
                    onclick="bulkOperations.goToPage(${currentPage + 1})">
                <i class="fas fa-chevron-right"></i>
            </button>
        `;

        paginationContainer.innerHTML = paginationHTML;

        // Update pagination info
        const paginationInfo = document.getElementById('paginationInfo');
        if (paginationInfo) {
            const start = (currentPage - 1) * this.pagination.limit + 1;
            const end = Math.min(currentPage * this.pagination.limit, this.pagination.total);
            paginationInfo.textContent = `Showing ${start}-${end} of ${this.pagination.total} merchants`;
        }
    }

    /**
     * Update selection summary
     */
    updateSelectionSummary() {
        const selectedCount = this.selectedMerchants.size;
        const summaryElement = document.getElementById('selectionSummary');
        
        if (summaryElement) {
            if (selectedCount === 0) {
                summaryElement.textContent = 'No merchants selected';
                summaryElement.className = 'selection-summary empty';
            } else {
                summaryElement.textContent = `${selectedCount} merchant${selectedCount === 1 ? '' : 's'} selected`;
                summaryElement.className = 'selection-summary active';
            }
        }

        // Update bulk action buttons
        const bulkActionButtons = document.querySelectorAll('.bulk-action-btn');
        bulkActionButtons.forEach(btn => {
            btn.disabled = selectedCount === 0;
        });
    }

    /**
     * Toggle merchant selection
     */
    toggleMerchantSelection(merchantId) {
        if (this.selectedMerchants.has(merchantId)) {
            this.selectedMerchants.delete(merchantId);
        } else {
            this.selectedMerchants.add(merchantId);
        }
        this.updateSelectionSummary();
    }

    /**
     * Select all merchants on current page
     */
    selectAllMerchants() {
        this.merchants.forEach(merchant => {
            this.selectedMerchants.add(merchant.id);
        });
        this.renderMerchantsTable();
        this.updateSelectionSummary();
    }

    /**
     * Deselect all merchants
     */
    deselectAllMerchants() {
        this.selectedMerchants.clear();
        this.renderMerchantsTable();
        this.updateSelectionSummary();
    }

    /**
     * Apply filters
     */
    async applyFilters() {
        this.pagination.page = 1; // Reset to first page
        await this.loadMerchants();
    }

    /**
     * Apply sorting
     */
    async applySorting(field, direction) {
        this.sorting.field = field;
        this.sorting.direction = direction;
        await this.loadMerchants();
    }

    /**
     * Go to specific page
     */
    async goToPage(page) {
        this.pagination.page = page;
        await this.loadMerchants();
    }

    /**
     * Change page size
     */
    async changePageSize(size) {
        this.pagination.limit = parseInt(size);
        this.pagination.page = 1;
        await this.loadMerchants();
    }

    /**
     * Search merchants
     */
    async searchMerchants(searchTerm) {
        this.filters.search = searchTerm;
        this.pagination.page = 1;
        await this.loadMerchants();
    }

    /**
     * Bulk update merchant status
     */
    async bulkUpdateStatus(newStatus) {
        if (this.selectedMerchants.size === 0) {
            this.showNotification('Please select merchants to update', 'warning');
            return;
        }

        try {
            this.showLoadingState();
            
            const merchantIds = Array.from(this.selectedMerchants);
            await this.dataIntegration.bulkUpdateMerchantStatus(merchantIds, newStatus);
            
            this.showNotification(`Successfully updated ${merchantIds.length} merchants to ${newStatus}`, 'success');
            this.deselectAllMerchants();
            await this.loadMerchants();
        } catch (error) {
            console.error('Failed to bulk update status:', error);
            this.showNotification('Failed to update merchant status', 'error');
        } finally {
            this.hideLoadingState();
        }
    }

    /**
     * Bulk export merchants
     */
    async bulkExportMerchants() {
        if (this.selectedMerchants.size === 0) {
            this.showNotification('Please select merchants to export', 'warning');
            return;
        }

        try {
            this.showLoadingState();
            
            const merchantIds = Array.from(this.selectedMerchants);
            const exportData = await this.dataIntegration.exportMerchants(merchantIds);
            
            // Create and download file
            const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `merchants-export-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            
            this.showNotification(`Successfully exported ${merchantIds.length} merchants`, 'success');
        } catch (error) {
            console.error('Failed to export merchants:', error);
            this.showNotification('Failed to export merchants', 'error');
        } finally {
            this.hideLoadingState();
        }
    }

    /**
     * Bulk delete merchants
     */
    async bulkDeleteMerchants() {
        if (this.selectedMerchants.size === 0) {
            this.showNotification('Please select merchants to delete', 'warning');
            return;
        }

        const confirmed = confirm(`Are you sure you want to delete ${this.selectedMerchants.size} merchants? This action cannot be undone.`);
        if (!confirmed) return;

        try {
            this.showLoadingState();
            
            const merchantIds = Array.from(this.selectedMerchants);
            await this.dataIntegration.bulkDeleteMerchants(merchantIds);
            
            this.showNotification(`Successfully deleted ${merchantIds.length} merchants`, 'success');
            this.deselectAllMerchants();
            await this.loadMerchants();
        } catch (error) {
            console.error('Failed to bulk delete merchants:', error);
            this.showNotification('Failed to delete merchants', 'error');
        } finally {
            this.hideLoadingState();
        }
    }

    /**
     * View merchant details
     */
    viewMerchant(merchantId) {
        window.open(`merchant-dashboard.html?id=${merchantId}`, '_blank');
    }

    /**
     * Edit merchant
     */
    editMerchant(merchantId) {
        // Open edit modal or navigate to edit page
        this.showEditModal(merchantId);
    }

    /**
     * Show edit modal
     */
    showEditModal(merchantId) {
        const merchant = this.merchants.find(m => m.id === merchantId);
        if (!merchant) return;

        // Create and show edit modal
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Edit Merchant</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    <form id="editMerchantForm">
                        <div class="form-group">
                            <label for="merchantName">Name</label>
                            <input type="text" id="merchantName" value="${merchant.name || ''}" required>
                        </div>
                        <div class="form-group">
                            <label for="merchantStatus">Status</label>
                            <select id="merchantStatus">
                                <option value="active" ${merchant.status === 'active' ? 'selected' : ''}>Active</option>
                                <option value="inactive" ${merchant.status === 'inactive' ? 'selected' : ''}>Inactive</option>
                                <option value="pending" ${merchant.status === 'pending' ? 'selected' : ''}>Pending</option>
                                <option value="suspended" ${merchant.status === 'suspended' ? 'selected' : ''}>Suspended</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label for="merchantIndustry">Industry</label>
                            <input type="text" id="merchantIndustry" value="${merchant.industry || ''}">
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
                    <button class="btn btn-primary" onclick="bulkOperations.saveMerchantEdit('${merchantId}')">Save</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    }

    /**
     * Save merchant edit
     */
    async saveMerchantEdit(merchantId) {
        try {
            const form = document.getElementById('editMerchantForm');
            const formData = new FormData(form);
            
            const updateData = {
                name: document.getElementById('merchantName').value,
                status: document.getElementById('merchantStatus').value,
                industry: document.getElementById('merchantIndustry').value
            };
            
            await this.dataIntegration.updateMerchant(merchantId, updateData);
            
            this.showNotification('Merchant updated successfully', 'success');
            document.querySelector('.modal-overlay').remove();
            await this.loadMerchants();
        } catch (error) {
            console.error('Failed to update merchant:', error);
            this.showNotification('Failed to update merchant', 'error');
        }
    }

    /**
     * Get risk level for merchant
     */
    getRiskLevel(merchant) {
        // Determine risk level based on merchant data
        if (merchant.risk_score !== undefined) {
            if (merchant.risk_score >= 0.8) return 'high';
            if (merchant.risk_score >= 0.5) return 'medium';
            return 'low';
        }
        
        // Fallback to status-based risk assessment
        if (merchant.status === 'suspended') return 'high';
        if (merchant.status === 'pending') return 'medium';
        return 'low';
    }

    /**
     * Format currency
     */
    formatCurrency(amount) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0
        }).format(amount);
    }

    /**
     * Format date
     */
    formatDate(dateString) {
        if (!dateString) return 'N/A';
        return new Date(dateString).toLocaleDateString();
    }

    /**
     * Show loading state
     */
    showLoadingState() {
        const loadingElement = document.getElementById('loadingIndicator');
        if (loadingElement) {
            loadingElement.style.display = 'block';
        }
    }

    /**
     * Hide loading state
     */
    hideLoadingState() {
        const loadingElement = document.getElementById('loadingIndicator');
        if (loadingElement) {
            loadingElement.style.display = 'none';
        }
    }

    /**
     * Show error state
     */
    showErrorState(message) {
        const errorElement = document.getElementById('errorMessage');
        if (errorElement) {
            errorElement.textContent = message;
            errorElement.style.display = 'block';
        }
        
        this.hideLoadingState();
    }

    /**
     * Show notification
     */
    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    /**
     * Bind event handlers
     */
    bindEvents() {
        // Filter controls
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.filters.status = e.target.value;
                this.applyFilters();
            });
        }

        const industryFilter = document.getElementById('industryFilter');
        if (industryFilter) {
            industryFilter.addEventListener('change', (e) => {
                this.filters.industry = e.target.value;
                this.applyFilters();
            });
        }

        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    this.searchMerchants(e.target.value);
                }, 500);
            });
        }

        // Bulk action buttons
        const bulkStatusUpdate = document.getElementById('bulkStatusUpdate');
        if (bulkStatusUpdate) {
            bulkStatusUpdate.addEventListener('change', (e) => {
                if (e.target.value) {
                    this.bulkUpdateStatus(e.target.value);
                    e.target.value = '';
                }
            });
        }

        const bulkExportBtn = document.getElementById('bulkExportBtn');
        if (bulkExportBtn) {
            bulkExportBtn.addEventListener('click', () => this.bulkExportMerchants());
        }

        const bulkDeleteBtn = document.getElementById('bulkDeleteBtn');
        if (bulkDeleteBtn) {
            bulkDeleteBtn.addEventListener('click', () => this.bulkDeleteMerchants());
        }

        // Select all checkbox
        const selectAllCheckbox = document.getElementById('selectAllCheckbox');
        if (selectAllCheckbox) {
            selectAllCheckbox.addEventListener('change', (e) => {
                if (e.target.checked) {
                    this.selectAllMerchants();
                } else {
                    this.deselectAllMerchants();
                }
            });
        }
    }

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        // Refresh every 2 minutes
        this.refreshInterval = setInterval(() => {
            if (!this.isLoading) {
                this.loadMerchants();
            }
        }, 2 * 60 * 1000);
    }

    /**
     * Stop auto-refresh
     */
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    /**
     * Cleanup
     */
    destroy() {
        this.stopAutoRefresh();
        this.dataIntegration.clearCache();
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.bulkOperations = new MerchantBulkOperationsRealData();
});

// Export for use in other components
window.MerchantBulkOperationsRealData = MerchantBulkOperationsRealData;
