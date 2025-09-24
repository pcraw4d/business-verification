/**
 * Merchant Portfolio Management JavaScript
 * Provides merchant portfolio list view with pagination, search, and filtering
 * Integrates with existing components for comprehensive merchant management
 * 
 * Features:
 * - Real-time search with debouncing
 * - Advanced filtering by portfolio type, risk level, and industry
 * - Bulk selection and operations
 * - Virtual scrolling for large merchant lists (1000s)
 * - Lazy loading for performance optimization
 * - Export functionality (CSV)
 * - Merchant comparison (2-merchant limit)
 * - Responsive design with mobile support
 * - Integration with session management
 * - Mock data support for MVP testing
 * - Performance monitoring and optimization
 */

class MerchantPortfolio {
    constructor() {
        this.apiBaseUrl = '/api/v1';
        this.currentPage = 1;
        this.pageSize = 20;
        this.totalPages = 0;
        this.totalCount = 0;
        this.merchants = [];
        this.selectedMerchants = new Set();
        this.bulkMode = false;
        this.currentFilters = {
            searchQuery: '',
            portfolioType: '',
            riskLevel: '',
            industry: ''
        };
        this.searchTimeout = null;
        
        // Performance optimization components
        this.virtualScroller = null;
        this.lazyLoader = null;
        this.bundleOptimizer = null;
        this.performanceMonitor = null;
        
        this.init();
    }

    /**
     * Initialize the merchant portfolio
     */
    async init() {
        // Initialize performance monitoring first
        await this.initializePerformanceOptimizations();
        
        this.bindEvents();
        await this.loadMerchants();
        this.initializeComponents();
        this.setupVirtualScrolling();
    }

    /**
     * Initialize performance optimizations
     */
    async initializePerformanceOptimizations() {
        try {
            // Initialize performance monitor
            if (window.performanceMonitor) {
                this.performanceMonitor = window.performanceMonitor;
            }

            // Initialize bundle optimizer for dynamic imports
            if (typeof BundleOptimizer !== 'undefined') {
                this.bundleOptimizer = new BundleOptimizer({
                    baseUrl: this.apiBaseUrl,
                    cachePrefix: 'kyb-merchant-',
                    cacheVersion: '1.0.0'
                });
            }

            // Initialize lazy loader
            if (typeof MerchantLazyLoader !== 'undefined') {
                this.lazyLoader = new MerchantLazyLoader();
            }

            // Preload critical modules
            if (this.bundleOptimizer) {
                await this.bundleOptimizer.preloadModules(['high']);
            }

            // Measure initialization time
            if (this.performanceMonitor) {
                this.performanceMonitor.recordCustomMetric('portfolio-initialization-time', performance.now(), {
                    component: 'merchant-portfolio'
                });
            }
        } catch (error) {
            console.warn('Failed to initialize performance optimizations:', error);
        }
    }

    /**
     * Setup virtual scrolling for large merchant lists
     */
    setupVirtualScrolling() {
        const container = document.getElementById('merchantListContainer');
        if (!container || this.merchants.length < 100) {
            return; // Use regular rendering for small lists
        }

        try {
            if (typeof MerchantVirtualScroller !== 'undefined') {
                this.virtualScroller = new MerchantVirtualScroller({
                    container,
                    itemHeight: 100,
                    bufferSize: 10
                });

                // Set up virtual scroller with merchant data
                this.virtualScroller.setData(this.merchants);
                
                // Set up event handlers
                this.setupVirtualScrollerEvents();
                
                // Hide regular list container
                const regularContainer = document.getElementById('merchantList');
                if (regularContainer) {
                    regularContainer.style.display = 'none';
                }
            }
        } catch (error) {
            console.warn('Failed to setup virtual scrolling:', error);
        }
    }

    /**
     * Setup virtual scroller event handlers
     */
    setupVirtualScrollerEvents() {
        if (!this.virtualScroller) return;

        // Handle merchant selection
        this.virtualScroller.setItemSelectHandler((merchant, index, event) => {
            this.handleMerchantSelection(merchant, event.target.checked);
        });

        // Handle merchant clicks
        this.virtualScroller.setItemClickHandler((merchant, index, event) => {
            this.handleMerchantClick(merchant, event);
        });

        // Handle bulk mode changes
        this.virtualScroller.setBulkMode(this.bulkMode);
    }

    /**
     * Initialize external components
     */
    initializeComponents() {
        // Initialize navigation if available
        if (typeof Navigation !== 'undefined') {
            new Navigation();
        }

        // Initialize session manager if available
        if (typeof SessionManager !== 'undefined') {
            new SessionManager();
        }

        // Initialize mock data warning if available
        if (typeof MockDataWarning !== 'undefined') {
            new MockDataWarning();
        }

        // Initialize coming soon banner if available
        if (typeof ComingSoonBanner !== 'undefined') {
            new ComingSoonBanner();
        }
    }

    /**
     * Bind event listeners to DOM elements
     */
    bindEvents() {
        // Search input with debouncing
        const searchInput = document.getElementById('portfolioSearchInput');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.currentFilters.searchQuery = e.target.value;
                this.debouncedSearch();
            });
        }

        // Filter selects
        this.bindFilterEvents();
        
        // Action buttons
        this.bindActionEvents();
        
        // Pagination
        this.bindPaginationEvents();
        
        // Bulk actions
        this.bindBulkActionEvents();
    }

    /**
     * Bind filter event listeners
     */
    bindFilterEvents() {
        const portfolioTypeFilter = document.getElementById('portfolioTypeFilter');
        if (portfolioTypeFilter) {
            portfolioTypeFilter.addEventListener('change', (e) => {
                this.currentFilters.portfolioType = e.target.value;
                this.loadMerchants();
            });
        }

        const riskLevelFilter = document.getElementById('riskLevelFilter');
        if (riskLevelFilter) {
            riskLevelFilter.addEventListener('change', (e) => {
                this.currentFilters.riskLevel = e.target.value;
                this.loadMerchants();
            });
        }

        const industryFilter = document.getElementById('industryFilter');
        if (industryFilter) {
            industryFilter.addEventListener('change', (e) => {
                this.currentFilters.industry = e.target.value;
                this.loadMerchants();
            });
        }
    }

    /**
     * Bind action button event listeners
     */
    bindActionEvents() {
        const bulkSelectBtn = document.getElementById('bulkSelectBtn');
        if (bulkSelectBtn) {
            bulkSelectBtn.addEventListener('click', () => {
                this.toggleBulkMode();
            });
        }

        const exportPortfolioBtn = document.getElementById('exportPortfolioBtn');
        if (exportPortfolioBtn) {
            exportPortfolioBtn.addEventListener('click', () => {
                this.exportPortfolio();
            });
        }

        const addMerchantBtn = document.getElementById('addMerchantBtn');
        if (addMerchantBtn) {
            addMerchantBtn.addEventListener('click', () => {
                this.addMerchant();
            });
        }

        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadMerchants();
            });
        }
    }

    /**
     * Bind pagination event listeners
     */
    bindPaginationEvents() {
        const prevPageBtn = document.getElementById('prevPageBtn');
        if (prevPageBtn) {
            prevPageBtn.addEventListener('click', () => {
                if (this.currentPage > 1) {
                    this.currentPage--;
                    this.loadMerchants();
                }
            });
        }

        const nextPageBtn = document.getElementById('nextPageBtn');
        if (nextPageBtn) {
            nextPageBtn.addEventListener('click', () => {
                if (this.currentPage < this.totalPages) {
                    this.currentPage++;
                    this.loadMerchants();
                }
            });
        }
    }

    /**
     * Bind bulk action event listeners
     */
    bindBulkActionEvents() {
        const bulkEditBtn = document.getElementById('bulkEditBtn');
        if (bulkEditBtn) {
            bulkEditBtn.addEventListener('click', () => {
                this.bulkEdit();
            });
        }

        const bulkExportBtn = document.getElementById('bulkExportBtn');
        if (bulkExportBtn) {
            bulkExportBtn.addEventListener('click', () => {
                this.bulkExport();
            });
        }

        const bulkCompareBtn = document.getElementById('bulkCompareBtn');
        if (bulkCompareBtn) {
            bulkCompareBtn.addEventListener('click', () => {
                this.bulkCompare();
            });
        }

        const clearSelectionBtn = document.getElementById('clearSelectionBtn');
        if (clearSelectionBtn) {
            clearSelectionBtn.addEventListener('click', () => {
                this.clearSelection();
            });
        }
    }

    /**
     * Debounced search to prevent excessive API calls
     */
    debouncedSearch() {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.currentPage = 1;
            this.loadMerchants();
        }, 300);
    }

    /**
     * Load merchants from API with current filters and pagination
     */
    async loadMerchants() {
        const startTime = performance.now();
        
        try {
            this.showLoadingState();
            
            // Measure API call performance
            const apiStartTime = performance.now();
            
            const params = new URLSearchParams({
                page: this.currentPage.toString(),
                page_size: this.pageSize.toString(),
                sort_by: 'created_at',
                sort_order: 'desc'
            });

            // Add filters
            if (this.currentFilters.searchQuery) {
                params.append('query', this.currentFilters.searchQuery);
            }
            if (this.currentFilters.portfolioType) {
                params.append('portfolio_type', this.currentFilters.portfolioType);
            }
            if (this.currentFilters.riskLevel) {
                params.append('risk_level', this.currentFilters.riskLevel);
            }
            if (this.currentFilters.industry) {
                params.append('industry', this.currentFilters.industry);
            }

            const response = await fetch(`${this.apiBaseUrl}/merchants?${params}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.merchants = data.merchants || [];
            this.totalCount = data.total || 0;
            this.totalPages = data.total_pages || 0;

            this.renderMerchants();
            this.updateStats();
            this.updatePagination();
            
            // Update virtual scroller if available
            if (this.virtualScroller) {
                this.virtualScroller.setData(this.merchants);
            }

            // Record performance metrics
            const endTime = performance.now();
            const loadTime = endTime - startTime;
            
            if (this.performanceMonitor) {
                this.performanceMonitor.recordCustomMetric('merchant-load-time', loadTime, {
                    merchantCount: this.merchants.length,
                    page: this.currentPage,
                    filters: this.currentFilters
                });
            }

        } catch (error) {
            console.error('Error loading merchants:', error);
            this.showErrorState('Failed to load merchants. Please try again.');
            
            // Record error metrics
            const endTime = performance.now();
            const loadTime = endTime - startTime;
            
            if (this.performanceMonitor) {
                this.performanceMonitor.recordCustomMetric('merchant-load-error', loadTime, {
                    error: error.message,
                    page: this.currentPage,
                    filters: this.currentFilters
                });
            }
        }
    }

    /**
     * Show loading state in the merchants grid
     */
    showLoadingState() {
        const grid = document.getElementById('merchantsGrid');
        if (grid) {
            grid.innerHTML = `
                <div class="loading">
                    <i class="fas fa-spinner"></i>
                    Loading merchants...
                </div>
            `;
        }
    }

    /**
     * Show error state in the merchants grid
     */
    showErrorState(message) {
        const grid = document.getElementById('merchantsGrid');
        if (grid) {
            grid.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>Error Loading Merchants</h3>
                    <p>${message}</p>
                </div>
            `;
        }
    }

    /**
     * Render merchants in the grid
     */
    renderMerchants() {
        const grid = document.getElementById('merchantsGrid');
        if (!grid) return;
        
        if (this.merchants.length === 0) {
            grid.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-search"></i>
                    <h3>No merchants found</h3>
                    <p>Try adjusting your search criteria or filters</p>
                </div>
            `;
            return;
        }

        // Use virtual scrolling for large lists
        if (this.virtualScroller && this.merchants.length >= 100) {
            return; // Virtual scroller handles rendering
        }

        // Use lazy loading for regular rendering
        if (this.lazyLoader) {
            this.renderMerchantsWithLazyLoading(grid);
        } else {
            grid.innerHTML = this.merchants.map(merchant => this.createMerchantCard(merchant)).join('');
        }
    }

    /**
     * Render merchants with lazy loading
     */
    renderMerchantsWithLazyLoading(grid) {
        grid.innerHTML = '';
        
        this.merchants.forEach((merchant, index) => {
            const merchantElement = document.createElement('div');
            merchantElement.className = 'merchant-card lazy-load';
            merchantElement.dataset.merchantId = merchant.id;
            
            // Register for lazy loading
            this.lazyLoader.registerMerchantCard(merchantElement, merchant);
            
            // Add loading placeholder
            merchantElement.innerHTML = `
                <div class="merchant-card-skeleton">
                    <div class="skeleton-header"></div>
                    <div class="skeleton-content">
                        <div class="skeleton-line"></div>
                        <div class="skeleton-line"></div>
                        <div class="skeleton-line"></div>
                    </div>
                </div>
            `;
            
            grid.appendChild(merchantElement);
        });
    }

    /**
     * Create HTML for a merchant card
     */
    createMerchantCard(merchant) {
        const isSelected = this.selectedMerchants.has(merchant.id);
        const avatar = merchant.name.charAt(0).toUpperCase();
        
        return `
            <div class="merchant-card ${isSelected ? 'selected' : ''}" data-merchant-id="${merchant.id}">
                <div class="card-header">
                    <div class="merchant-avatar">${avatar}</div>
                    <div class="merchant-info">
                        <div class="merchant-name">${this.escapeHtml(merchant.name)}</div>
                        <div class="merchant-industry">
                            <i class="fas fa-industry"></i>
                            ${this.escapeHtml(merchant.industry || 'Not specified')}
                        </div>
                    </div>
                </div>
                
                <div class="card-badges">
                    <span class="badge badge-portfolio">${this.formatPortfolioType(merchant.portfolio_type)}</span>
                    <span class="badge badge-risk-${merchant.risk_level}">${this.formatRiskLevel(merchant.risk_level)}</span>
                </div>
                
                <div class="card-details">
                    <div class="detail-item">
                        <div class="detail-label">Location</div>
                        <div class="detail-value">${this.escapeHtml(merchant.address?.city || 'N/A')}, ${this.escapeHtml(merchant.address?.state || 'N/A')}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Founded</div>
                        <div class="detail-value">${merchant.founded_date ? new Date(merchant.founded_date).getFullYear() : 'Unknown'}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Employees</div>
                        <div class="detail-value">${merchant.employee_count || 'Unknown'}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Revenue</div>
                        <div class="detail-value">${merchant.annual_revenue ? `$${merchant.annual_revenue.toLocaleString()}` : 'Not disclosed'}</div>
                    </div>
                </div>
                
                <div class="card-actions">
                    <button class="btn btn-primary btn-sm" onclick="merchantPortfolio.viewMerchant('${merchant.id}')">
                        <i class="fas fa-eye"></i>
                        View
                    </button>
                    <button class="btn btn-outline btn-sm" onclick="merchantPortfolio.editMerchant('${merchant.id}')">
                        <i class="fas fa-edit"></i>
                        Edit
                    </button>
                    <button class="btn btn-outline btn-sm" onclick="merchantPortfolio.compareMerchant('${merchant.id}')">
                        <i class="fas fa-balance-scale"></i>
                        Compare
                    </button>
                </div>
            </div>
        `;
    }

    /**
     * Update statistics in the header
     */
    updateStats() {
        const totalMerchantsEl = document.getElementById('totalMerchants');
        const activeMerchantsEl = document.getElementById('activeMerchants');
        const pendingMerchantsEl = document.getElementById('pendingMerchants');
        
        if (totalMerchantsEl) {
            totalMerchantsEl.textContent = this.totalCount;
        }
        
        // Calculate active and pending counts
        const activeCount = this.merchants.filter(m => m.portfolio_type === 'onboarded').length;
        const pendingCount = this.merchants.filter(m => m.portfolio_type === 'pending').length;
        
        if (activeMerchantsEl) {
            activeMerchantsEl.textContent = activeCount;
        }
        if (pendingMerchantsEl) {
            pendingMerchantsEl.textContent = pendingCount;
        }
    }

    /**
     * Update pagination controls
     */
    updatePagination() {
        const pagination = document.getElementById('pagination');
        const paginationInfo = document.getElementById('paginationInfo');
        const prevBtn = document.getElementById('prevPageBtn');
        const nextBtn = document.getElementById('nextPageBtn');

        if (!pagination) return;

        if (this.totalPages <= 1) {
            pagination.style.display = 'none';
            return;
        }

        pagination.style.display = 'flex';
        
        if (paginationInfo) {
            paginationInfo.textContent = `Page ${this.currentPage} of ${this.totalPages}`;
        }
        
        if (prevBtn) {
            prevBtn.disabled = this.currentPage <= 1;
        }
        if (nextBtn) {
            nextBtn.disabled = this.currentPage >= this.totalPages;
        }
    }

    /**
     * Toggle bulk selection mode
     */
    toggleBulkMode() {
        this.bulkMode = !this.bulkMode;
        const bulkBtn = document.getElementById('bulkSelectBtn');
        const bulkSelection = document.getElementById('bulkSelection');
        
        if (!bulkBtn || !bulkSelection) return;
        
        if (this.bulkMode) {
            bulkBtn.innerHTML = '<i class="fas fa-times"></i> Exit Bulk Mode';
            bulkBtn.classList.add('btn-warning');
            bulkBtn.classList.remove('btn-primary');
            bulkSelection.classList.add('active');
            this.addBulkSelectionHandlers();
        } else {
            bulkBtn.innerHTML = '<i class="fas fa-check-square"></i> Bulk Select';
            bulkBtn.classList.remove('btn-warning');
            bulkBtn.classList.add('btn-primary');
            bulkSelection.classList.remove('active');
            this.clearSelection();
            this.removeBulkSelectionHandlers();
        }
    }

    /**
     * Add bulk selection event handlers to merchant cards
     */
    addBulkSelectionHandlers() {
        document.querySelectorAll('.merchant-card').forEach(card => {
            card.addEventListener('click', this.handleCardClick.bind(this));
        });
    }

    /**
     * Remove bulk selection event handlers from merchant cards
     */
    removeBulkSelectionHandlers() {
        document.querySelectorAll('.merchant-card').forEach(card => {
            card.removeEventListener('click', this.handleCardClick.bind(this));
        });
    }

    /**
     * Handle merchant card click in bulk mode
     */
    handleCardClick(event) {
        if (!this.bulkMode) return;
        
        const card = event.currentTarget;
        const merchantId = card.dataset.merchantId;
        
        if (this.selectedMerchants.has(merchantId)) {
            this.selectedMerchants.delete(merchantId);
            card.classList.remove('selected');
        } else {
            this.selectedMerchants.add(merchantId);
            card.classList.add('selected');
        }
        
        this.updateBulkSelection();
    }

    /**
     * Update bulk selection UI
     */
    updateBulkSelection() {
        const count = this.selectedMerchants.size;
        const bulkCount = document.getElementById('bulkCount');
        
        if (bulkCount) {
            bulkCount.textContent = `${count} merchant${count !== 1 ? 's' : ''} selected`;
        }
        
        // Enable/disable bulk action buttons
        const bulkActions = document.querySelectorAll('#bulkSelection .btn');
        bulkActions.forEach(btn => {
            btn.disabled = count === 0;
        });
    }

    /**
     * Clear all selected merchants
     */
    clearSelection() {
        this.selectedMerchants.clear();
        document.querySelectorAll('.merchant-card').forEach(card => {
            card.classList.remove('selected');
        });
        this.updateBulkSelection();
    }

    /**
     * Navigate to merchant detail view
     */
    viewMerchant(merchantId) {
        window.location.href = `merchant-detail.html?id=${merchantId}`;
    }

    /**
     * Navigate to merchant edit view
     */
    editMerchant(merchantId) {
        window.location.href = `merchant-edit.html?id=${merchantId}`;
    }

    /**
     * Navigate to merchant comparison view
     */
    compareMerchant(merchantId) {
        window.location.href = `merchant-comparison.html?merchant1=${merchantId}`;
    }

    /**
     * Export entire portfolio to CSV
     */
    exportPortfolio() {
        const csvContent = this.generateCSV(this.merchants);
        this.downloadCSV(csvContent, 'merchant-portfolio.csv');
    }

    /**
     * Bulk edit selected merchants
     */
    bulkEdit() {
        if (this.selectedMerchants.size === 0) return;
        // Navigate to bulk edit page
        window.location.href = `merchant-bulk-edit.html?ids=${Array.from(this.selectedMerchants).join(',')}`;
    }

    /**
     * Export selected merchants to CSV
     */
    bulkExport() {
        if (this.selectedMerchants.size === 0) return;
        const selectedMerchants = this.merchants.filter(m => this.selectedMerchants.has(m.id));
        const csvContent = this.generateCSV(selectedMerchants);
        this.downloadCSV(csvContent, 'selected-merchants.csv');
    }

    /**
     * Compare selected merchants (exactly 2 required)
     */
    bulkCompare() {
        if (this.selectedMerchants.size !== 2) {
            alert('Please select exactly 2 merchants to compare.');
            return;
        }
        const merchantIds = Array.from(this.selectedMerchants);
        window.location.href = `merchant-comparison.html?merchant1=${merchantIds[0]}&merchant2=${merchantIds[1]}`;
    }

    /**
     * Navigate to add merchant page
     */
    addMerchant() {
        window.location.href = 'merchant-add.html';
    }

    /**
     * Generate CSV content from merchant data
     */
    generateCSV(merchants) {
        const headers = ['Name', 'Industry', 'Portfolio Type', 'Risk Level', 'City', 'State', 'Founded', 'Employees', 'Revenue'];
        const rows = merchants.map(merchant => [
            merchant.name,
            merchant.industry || 'N/A',
            this.formatPortfolioType(merchant.portfolio_type),
            this.formatRiskLevel(merchant.risk_level),
            merchant.address?.city || 'N/A',
            merchant.address?.state || 'N/A',
            merchant.founded_date ? new Date(merchant.founded_date).getFullYear() : 'Unknown',
            merchant.employee_count || 'Unknown',
            merchant.annual_revenue ? `$${merchant.annual_revenue.toLocaleString()}` : 'Not disclosed'
        ]);

        return [headers, ...rows].map(row => 
            row.map(field => `"${field}"`).join(',')
        ).join('\n');
    }

    /**
     * Download CSV file
     */
    downloadCSV(content, filename) {
        const blob = new Blob([content], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    }

    /**
     * Format portfolio type for display
     */
    formatPortfolioType(type) {
        const types = {
            'onboarded': 'Onboarded',
            'deactivated': 'Deactivated',
            'prospective': 'Prospective',
            'pending': 'Pending'
        };
        return types[type] || type;
    }

    /**
     * Format risk level for display
     */
    formatRiskLevel(level) {
        const levels = {
            'low': 'Low Risk',
            'medium': 'Medium Risk',
            'high': 'High Risk'
        };
        return levels[level] || level;
    }

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Get current filter state
     */
    getCurrentFilters() {
        return { ...this.currentFilters };
    }

    /**
     * Set filters programmatically
     */
    setFilters(filters) {
        this.currentFilters = { ...this.currentFilters, ...filters };
        this.currentPage = 1;
        this.loadMerchants();
    }

    /**
     * Get selected merchant IDs
     */
    getSelectedMerchants() {
        return Array.from(this.selectedMerchants);
    }

    /**
     * Set selected merchants programmatically
     */
    setSelectedMerchants(merchantIds) {
        this.selectedMerchants.clear();
        merchantIds.forEach(id => this.selectedMerchants.add(id));
        this.updateBulkSelection();
    }

    /**
     * Refresh the current view
     */
    refresh() {
        this.loadMerchants();
    }

    /**
     * Reset all filters and pagination
     */
    reset() {
        this.currentFilters = {
            searchQuery: '',
            portfolioType: '',
            riskLevel: '',
            industry: ''
        };
        this.currentPage = 1;
        this.clearSelection();
        this.loadMerchants();
    }

    /**
     * Cleanup performance components and resources
     */
    destroy() {
        // Cleanup virtual scroller
        if (this.virtualScroller) {
            this.virtualScroller.destroy();
            this.virtualScroller = null;
        }

        // Cleanup lazy loader
        if (this.lazyLoader) {
            this.lazyLoader.destroy();
            this.lazyLoader = null;
        }

        // Cleanup bundle optimizer
        if (this.bundleOptimizer) {
            this.bundleOptimizer.destroy();
            this.bundleOptimizer = null;
        }

        // Record final performance metrics
        if (this.performanceMonitor) {
            this.performanceMonitor.recordCustomMetric('portfolio-destroy-time', performance.now(), {
                component: 'merchant-portfolio'
            });
        }
    }
}

// Initialize the portfolio when DOM is loaded
let merchantPortfolio;
document.addEventListener('DOMContentLoaded', () => {
    merchantPortfolio = new MerchantPortfolio();
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantPortfolio;
}
