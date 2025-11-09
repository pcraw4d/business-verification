/**
 * Merchant Search Component
 * Provides real-time search functionality with filtering by portfolio type and risk level
 * Includes debouncing for performance optimization
 */
class MerchantSearch {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.debounceDelay = options.debounceDelay || 300;
        this.pageSize = options.pageSize || 20;
        this.currentPage = 1;
        this.currentFilters = {
            searchQuery: '',
            portfolioType: null,
            riskLevel: null,
            industry: '',
            status: ''
        };
        this.searchTimeout = null;
        this.isLoading = false;
        this.merchants = [];
        this.totalCount = 0;
        this.totalPages = 0;
        
        // Event callbacks
        this.onMerchantSelect = options.onMerchantSelect || null;
        this.onSearchResults = options.onSearchResults || null;
        this.onFilterChange = options.onFilterChange || null;
        
        this.init();
    }

    init() {
        this.createSearchInterface();
        this.bindEvents();
        this.loadInitialData();
    }

    createSearchInterface() {
        const searchHTML = `
            <div class="merchant-search-container">
                <div class="search-header">
                    <h2 class="search-title">
                        <i class="fas fa-search"></i>
                        Merchant Search
                    </h2>
                    <div class="search-stats" id="searchStats">
                        <span class="total-count">0 merchants found</span>
                    </div>
                </div>

                <div class="search-controls">
                    <div class="search-input-group">
                        <div class="search-input-wrapper">
                            <input 
                                type="text" 
                                id="merchantSearchInput" 
                                class="search-input" 
                                placeholder="Search by name, industry, or business type..."
                                autocomplete="off"
                            >
                            <button class="search-clear-btn" id="searchClearBtn" style="display: none;">
                                <i class="fas fa-times"></i>
                            </button>
                            <div class="search-loading" id="searchLoading" style="display: none;">
                                <i class="fas fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>

                    <div class="filter-controls">
                        <div class="filter-group">
                            <label for="portfolioTypeFilter" class="filter-label">
                                <i class="fas fa-folder"></i>
                                Portfolio Type
                            </label>
                            <select id="portfolioTypeFilter" class="filter-select">
                                <option value="">All Types</option>
                                <option value="onboarded">Onboarded</option>
                                <option value="deactivated">Deactivated</option>
                                <option value="prospective">Prospective</option>
                                <option value="pending">Pending</option>
                            </select>
                        </div>

                        <div class="filter-group">
                            <label for="riskLevelFilter" class="filter-label">
                                <i class="fas fa-exclamation-triangle"></i>
                                Risk Level
                            </label>
                            <select id="riskLevelFilter" class="filter-select">
                                <option value="">All Levels</option>
                                <option value="low">Low Risk</option>
                                <option value="medium">Medium Risk</option>
                                <option value="high">High Risk</option>
                            </select>
                        </div>

                        <div class="filter-group">
                            <label for="industryFilter" class="filter-label">
                                <i class="fas fa-industry"></i>
                                Industry
                            </label>
                            <select id="industryFilter" class="filter-select">
                                <option value="">All Industries</option>
                                <option value="technology">Technology</option>
                                <option value="retail">Retail</option>
                                <option value="finance">Finance</option>
                                <option value="healthcare">Healthcare</option>
                                <option value="manufacturing">Manufacturing</option>
                                <option value="services">Services</option>
                                <option value="other">Other</option>
                            </select>
                        </div>

                        <div class="filter-group">
                            <label for="statusFilter" class="filter-label">
                                <i class="fas fa-check-circle"></i>
                                Status
                            </label>
                            <select id="statusFilter" class="filter-select">
                                <option value="">All Statuses</option>
                                <option value="active">Active</option>
                                <option value="inactive">Inactive</option>
                                <option value="suspended">Suspended</option>
                                <option value="pending">Pending</option>
                            </select>
                        </div>

                        <div class="filter-actions">
                            <button class="btn btn-secondary" id="clearFiltersBtn">
                                <i class="fas fa-eraser"></i>
                                Clear Filters
                            </button>
                            <button class="btn btn-primary" id="applyFiltersBtn">
                                <i class="fas fa-filter"></i>
                                Apply Filters
                            </button>
                        </div>
                    </div>
                </div>

                <div class="search-results" id="searchResults">
                    <div class="results-header">
                        <div class="results-info">
                            <span class="results-count" id="resultsCount">No results</span>
                            <span class="results-pagination" id="resultsPagination"></span>
                        </div>
                        <div class="results-actions">
                            <button class="btn btn-outline" id="exportResultsBtn" disabled>
                                <i class="fas fa-download"></i>
                                Export Results
                            </button>
                        </div>
                    </div>

                    <div class="merchants-list" id="merchantsList">
                        <div class="no-results" id="noResults">
                            <div class="no-results-icon">
                                <i class="fas fa-search"></i>
                            </div>
                            <h3>No merchants found</h3>
                            <p>Try adjusting your search criteria or filters</p>
                        </div>
                    </div>

                    <div class="pagination" id="pagination" style="display: none;">
                        <button class="btn btn-outline" id="prevPageBtn" disabled>
                            <i class="fas fa-chevron-left"></i>
                            Previous
                        </button>
                        <div class="pagination-info">
                            <span id="paginationInfo">Page 1 of 1</span>
                        </div>
                        <button class="btn btn-outline" id="nextPageBtn" disabled>
                            Next
                            <i class="fas fa-chevron-right"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;

        this.container.innerHTML = searchHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .merchant-search-container {
                    background: white;
                    border-radius: 12px;
                    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
                    padding: 24px;
                    margin-bottom: 24px;
                }

                .search-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 24px;
                    padding-bottom: 16px;
                    border-bottom: 2px solid #f8f9fa;
                }

                .search-title {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.5rem;
                    font-weight: 600;
                }

                .search-title i {
                    color: #3498db;
                    font-size: 1.3rem;
                }

                .search-stats {
                    display: flex;
                    align-items: center;
                    gap: 16px;
                }

                .total-count {
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    color: white;
                    padding: 8px 16px;
                    border-radius: 20px;
                    font-size: 0.9rem;
                    font-weight: 600;
                }

                .search-controls {
                    margin-bottom: 24px;
                }

                .search-input-group {
                    margin-bottom: 20px;
                }

                .search-input-wrapper {
                    position: relative;
                    max-width: 600px;
                }

                .search-input {
                    width: 100%;
                    padding: 16px 50px 16px 20px;
                    border: 2px solid #e9ecef;
                    border-radius: 12px;
                    font-size: 1rem;
                    transition: all 0.3s ease;
                    background: #f8f9fa;
                }

                .search-input:focus {
                    outline: none;
                    border-color: #3498db;
                    background: white;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .search-clear-btn {
                    position: absolute;
                    right: 12px;
                    top: 50%;
                    transform: translateY(-50%);
                    background: none;
                    border: none;
                    color: #6c757d;
                    cursor: pointer;
                    padding: 8px;
                    border-radius: 50%;
                    transition: all 0.3s ease;
                }

                .search-clear-btn:hover {
                    background: #e9ecef;
                    color: #495057;
                }

                .search-loading {
                    position: absolute;
                    right: 12px;
                    top: 50%;
                    transform: translateY(-50%);
                    color: #3498db;
                }

                .filter-controls {
                    display: grid;
                    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                    gap: 16px;
                    align-items: end;
                }

                .filter-group {
                    display: flex;
                    flex-direction: column;
                    gap: 8px;
                }

                .filter-label {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    font-weight: 600;
                    color: #495057;
                    font-size: 0.9rem;
                }

                .filter-label i {
                    color: #6c757d;
                    font-size: 0.8rem;
                }

                .filter-select {
                    padding: 12px 16px;
                    border: 2px solid #e9ecef;
                    border-radius: 8px;
                    font-size: 0.9rem;
                    background: white;
                    transition: all 0.3s ease;
                }

                .filter-select:focus {
                    outline: none;
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .filter-actions {
                    display: flex;
                    gap: 12px;
                    align-items: end;
                }

                .btn {
                    padding: 12px 20px;
                    border: none;
                    border-radius: 8px;
                    font-size: 0.9rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    text-decoration: none;
                }

                .btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }

                .btn-primary {
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    color: white;
                }

                .btn-primary:hover:not(:disabled) {
                    background: linear-gradient(135deg, #2980b9, #1f5f8b);
                    transform: translateY(-2px);
                    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
                }

                .btn-secondary {
                    background: #6c757d;
                    color: white;
                }

                .btn-secondary:hover:not(:disabled) {
                    background: #5a6268;
                    transform: translateY(-2px);
                }

                .btn-outline {
                    background: transparent;
                    color: #3498db;
                    border: 2px solid #3498db;
                }

                .btn-outline:hover:not(:disabled) {
                    background: #3498db;
                    color: white;
                }

                .search-results {
                    margin-top: 24px;
                }

                .results-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                    padding: 16px;
                    background: #f8f9fa;
                    border-radius: 8px;
                }

                .results-info {
                    display: flex;
                    align-items: center;
                    gap: 16px;
                }

                .results-count {
                    font-weight: 600;
                    color: #495057;
                }

                .results-pagination {
                    font-size: 0.9rem;
                    color: #6c757d;
                }

                .merchants-list {
                    min-height: 200px;
                }

                .merchant-item {
                    display: flex;
                    align-items: center;
                    padding: 20px;
                    border: 2px solid #f8f9fa;
                    border-radius: 12px;
                    margin-bottom: 12px;
                    transition: all 0.3s ease;
                    cursor: pointer;
                    background: white;
                }

                .merchant-item:hover {
                    border-color: #3498db;
                    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.1);
                    transform: translateY(-2px);
                }

                .merchant-item.selected {
                    border-color: #3498db;
                    background: linear-gradient(135deg, rgba(52, 152, 219, 0.1), rgba(41, 128, 185, 0.1));
                }

                .merchant-avatar {
                    width: 50px;
                    height: 50px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-weight: 600;
                    font-size: 1.2rem;
                    margin-right: 16px;
                    flex-shrink: 0;
                }

                .merchant-info {
                    flex: 1;
                }

                .merchant-name {
                    font-size: 1.1rem;
                    font-weight: 600;
                    color: #2c3e50;
                    margin-bottom: 4px;
                }

                .merchant-details {
                    display: flex;
                    gap: 16px;
                    font-size: 0.9rem;
                    color: #6c757d;
                }

                .merchant-badges {
                    display: flex;
                    gap: 8px;
                    align-items: center;
                }

                .badge {
                    padding: 4px 12px;
                    border-radius: 20px;
                    font-size: 0.8rem;
                    font-weight: 600;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                }

                .badge-portfolio {
                    background: #e3f2fd;
                    color: #1976d2;
                }

                .badge-risk-low {
                    background: #e8f5e8;
                    color: #2e7d32;
                }

                .badge-risk-medium {
                    background: #fff3e0;
                    color: #f57c00;
                }

                .badge-risk-high {
                    background: #ffebee;
                    color: #d32f2f;
                }

                .no-results {
                    text-align: center;
                    padding: 60px 20px;
                    color: #6c757d;
                }

                .no-results-icon {
                    font-size: 3rem;
                    color: #dee2e6;
                    margin-bottom: 16px;
                }

                .no-results h3 {
                    margin: 0 0 8px 0;
                    color: #495057;
                }

                .no-results p {
                    margin: 0;
                    font-size: 0.9rem;
                }

                .pagination {
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    gap: 16px;
                    margin-top: 24px;
                    padding: 16px;
                }

                .pagination-info {
                    font-size: 0.9rem;
                    color: #6c757d;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .merchant-search-container {
                        padding: 16px;
                    }

                    .search-header {
                        flex-direction: column;
                        gap: 12px;
                        align-items: flex-start;
                    }

                    .filter-controls {
                        grid-template-columns: 1fr;
                    }

                    .filter-actions {
                        flex-direction: column;
                    }

                    .results-header {
                        flex-direction: column;
                        gap: 12px;
                        align-items: flex-start;
                    }

                    .merchant-item {
                        flex-direction: column;
                        align-items: flex-start;
                        gap: 12px;
                    }

                    .merchant-avatar {
                        margin-right: 0;
                    }

                    .merchant-details {
                        flex-direction: column;
                        gap: 4px;
                    }

                    .pagination {
                        flex-direction: column;
                        gap: 12px;
                    }
                }

                /* Loading States */
                .loading {
                    opacity: 0.6;
                    pointer-events: none;
                }

                .loading::after {
                    content: '';
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(255, 255, 255, 0.8);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                }

                /* Animation for search results */
                .merchant-item {
                    animation: slideInUp 0.3s ease-out;
                }

                @keyframes slideInUp {
                    from {
                        opacity: 0;
                        transform: translateY(20px);
                    }
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#merchant-search-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'merchant-search-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        const searchInput = document.getElementById('merchantSearchInput');
        const searchClearBtn = document.getElementById('searchClearBtn');
        const portfolioTypeFilter = document.getElementById('portfolioTypeFilter');
        const riskLevelFilter = document.getElementById('riskLevelFilter');
        const industryFilter = document.getElementById('industryFilter');
        const statusFilter = document.getElementById('statusFilter');
        const clearFiltersBtn = document.getElementById('clearFiltersBtn');
        const applyFiltersBtn = document.getElementById('applyFiltersBtn');
        const exportResultsBtn = document.getElementById('exportResultsBtn');
        const prevPageBtn = document.getElementById('prevPageBtn');
        const nextPageBtn = document.getElementById('nextPageBtn');

        // Search input with debouncing
        searchInput.addEventListener('input', (e) => {
            this.currentFilters.searchQuery = e.target.value.trim();
            this.toggleClearButton();
            this.debouncedSearch();
        });

        // Clear search button
        searchClearBtn.addEventListener('click', () => {
            searchInput.value = '';
            this.currentFilters.searchQuery = '';
            this.toggleClearButton();
            this.performSearch();
        });

        // Filter change events
        portfolioTypeFilter.addEventListener('change', (e) => {
            this.currentFilters.portfolioType = e.target.value || null;
            this.onFilterChange && this.onFilterChange(this.currentFilters);
        });

        riskLevelFilter.addEventListener('change', (e) => {
            this.currentFilters.riskLevel = e.target.value || null;
            this.onFilterChange && this.onFilterChange(this.currentFilters);
        });

        industryFilter.addEventListener('change', (e) => {
            this.currentFilters.industry = e.target.value || '';
            this.onFilterChange && this.onFilterChange(this.currentFilters);
        });

        statusFilter.addEventListener('change', (e) => {
            this.currentFilters.status = e.target.value || '';
            this.onFilterChange && this.onFilterChange(this.currentFilters);
        });

        // Filter action buttons
        clearFiltersBtn.addEventListener('click', () => {
            this.clearAllFilters();
        });

        applyFiltersBtn.addEventListener('click', () => {
            this.performSearch();
        });

        // Export results
        exportResultsBtn.addEventListener('click', () => {
            this.exportResults();
        });

        // Pagination
        prevPageBtn.addEventListener('click', () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.performSearch();
            }
        });

        nextPageBtn.addEventListener('click', () => {
            if (this.currentPage < this.totalPages) {
                this.currentPage++;
                this.performSearch();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'k':
                        e.preventDefault();
                        searchInput.focus();
                        break;
                    case 'f':
                        e.preventDefault();
                        this.clearAllFilters();
                        break;
                }
            }
        });
    }

    debouncedSearch() {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.performSearch();
        }, this.debounceDelay);
    }

    async performSearch() {
        if (this.isLoading) return;

        this.setLoading(true);
        this.currentPage = 1; // Reset to first page on new search

        try {
            const response = await this.searchMerchants();
            this.handleSearchResults(response);
        } catch (error) {
            console.error('Search error:', error);
            this.showError('Failed to search merchants. Please try again.');
        } finally {
            this.setLoading(false);
        }
    }

    async searchMerchants() {
        const params = new URLSearchParams({
            page: this.currentPage.toString(),
            page_size: this.pageSize.toString(),
            sort_by: 'created_at',
            sort_order: 'desc'
        });

        // Add search query
        if (this.currentFilters.searchQuery) {
            params.append('query', this.currentFilters.searchQuery);
        }

        // Add filters
        if (this.currentFilters.portfolioType) {
            params.append('portfolio_type', this.currentFilters.portfolioType);
        }
        if (this.currentFilters.riskLevel) {
            params.append('risk_level', this.currentFilters.riskLevel);
        }
        if (this.currentFilters.industry) {
            params.append('industry', this.currentFilters.industry);
        }
        if (this.currentFilters.status) {
            params.append('status', this.currentFilters.status);
        }

        const response = await fetch(`${this.apiBaseUrl}/merchants?${params}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.getAuthToken()}`
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await response.json();
    }

    handleSearchResults(data) {
        this.merchants = data.merchants || [];
        this.totalCount = data.total || 0;
        this.totalPages = data.total_pages || 0;

        this.updateSearchStats();
        this.renderMerchants();
        this.updatePagination();
        this.updateExportButton();

        // Call callback if provided
        if (this.onSearchResults) {
            this.onSearchResults({
                merchants: this.merchants,
                total: this.totalCount,
                page: this.currentPage,
                totalPages: this.totalPages
            });
        }
    }

    renderMerchants() {
        const merchantsList = document.getElementById('merchantsList');
        const noResults = document.getElementById('noResults');

        if (this.merchants.length === 0) {
            merchantsList.innerHTML = '';
            merchantsList.appendChild(noResults);
            noResults.style.display = 'block';
            return;
        }

        noResults.style.display = 'none';
        merchantsList.innerHTML = '';

        this.merchants.forEach(merchant => {
            const merchantElement = this.createMerchantElement(merchant);
            merchantsList.appendChild(merchantElement);
        });
    }

    createMerchantElement(merchant) {
        const element = document.createElement('div');
        element.className = 'merchant-item';
        element.dataset.merchantId = merchant.id;

        const avatar = merchant.name.charAt(0).toUpperCase();
        const portfolioType = this.formatPortfolioType(merchant.portfolio_type);
        const riskLevel = this.formatRiskLevel(merchant.risk_level);

        element.innerHTML = `
            <div class="merchant-avatar">${avatar}</div>
            <div class="merchant-info">
                <div class="merchant-name">${merchant.name}</div>
                <div class="merchant-details">
                    <span><i class="fas fa-industry"></i> ${merchant.industry || 'N/A'}</span>
                    <span><i class="fas fa-map-marker-alt"></i> ${merchant.address.city || 'N/A'}, ${merchant.address.state || 'N/A'}</span>
                    <span><i class="fas fa-calendar"></i> ${this.formatDate(merchant.created_at)}</span>
                </div>
            </div>
            <div class="merchant-badges">
                <span class="badge badge-portfolio">${portfolioType}</span>
                <span class="badge badge-risk-${merchant.risk_level}">${riskLevel}</span>
            </div>
        `;

        // Add click handler
        element.addEventListener('click', () => {
            this.selectMerchant(merchant);
        });

        return element;
    }

    selectMerchant(merchant) {
        // Remove previous selection
        document.querySelectorAll('.merchant-item').forEach(item => {
            item.classList.remove('selected');
        });

        // Add selection to clicked item
        const selectedElement = document.querySelector(`[data-merchant-id="${merchant.id}"]`);
        if (selectedElement) {
            selectedElement.classList.add('selected');
        }

        // Call callback if provided
        if (this.onMerchantSelect) {
            this.onMerchantSelect(merchant);
        }
    }

    updateSearchStats() {
        const searchStats = document.getElementById('searchStats');
        const resultsCount = document.getElementById('resultsCount');
        
        searchStats.innerHTML = `
            <span class="total-count">${this.totalCount} merchants found</span>
        `;
        
        resultsCount.textContent = `${this.totalCount} results`;
    }

    updatePagination() {
        const pagination = document.getElementById('pagination');
        const paginationInfo = document.getElementById('paginationInfo');
        const prevPageBtn = document.getElementById('prevPageBtn');
        const nextPageBtn = document.getElementById('nextPageBtn');

        if (this.totalPages <= 1) {
            pagination.style.display = 'none';
            return;
        }

        pagination.style.display = 'flex';
        paginationInfo.textContent = `Page ${this.currentPage} of ${this.totalPages}`;
        
        prevPageBtn.disabled = this.currentPage <= 1;
        nextPageBtn.disabled = this.currentPage >= this.totalPages;
    }

    updateExportButton() {
        const exportBtn = document.getElementById('exportResultsBtn');
        exportBtn.disabled = this.merchants.length === 0;
    }

    clearAllFilters() {
        // Clear search input
        const searchInput = document.getElementById('merchantSearchInput');
        searchInput.value = '';
        this.currentFilters.searchQuery = '';
        this.toggleClearButton();

        // Clear filter selects
        document.getElementById('portfolioTypeFilter').value = '';
        document.getElementById('riskLevelFilter').value = '';
        document.getElementById('industryFilter').value = '';
        document.getElementById('statusFilter').value = '';

        // Reset filters
        this.currentFilters = {
            searchQuery: '',
            portfolioType: null,
            riskLevel: null,
            industry: '',
            status: ''
        };

        // Perform search
        this.performSearch();
    }

    toggleClearButton() {
        const searchInput = document.getElementById('merchantSearchInput');
        const clearBtn = document.getElementById('searchClearBtn');
        
        if (searchInput.value.trim()) {
            clearBtn.style.display = 'block';
        } else {
            clearBtn.style.display = 'none';
        }
    }

    setLoading(loading) {
        this.isLoading = loading;
        const searchLoading = document.getElementById('searchLoading');
        const searchInput = document.getElementById('merchantSearchInput');
        
        if (loading) {
            searchLoading.style.display = 'block';
            searchInput.disabled = true;
        } else {
            searchLoading.style.display = 'none';
            searchInput.disabled = false;
        }
    }

    async loadInitialData() {
        await this.performSearch();
    }

    exportResults() {
        if (this.merchants.length === 0) return;

        const csvContent = this.generateCSV(this.merchants);
        const blob = new Blob([csvContent], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = `merchants-export-${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    }

    generateCSV(merchants) {
        const headers = ['Name', 'Industry', 'Portfolio Type', 'Risk Level', 'City', 'State', 'Created Date'];
        const rows = merchants.map(merchant => [
            merchant.name,
            merchant.industry || 'N/A',
            this.formatPortfolioType(merchant.portfolio_type),
            this.formatRiskLevel(merchant.risk_level),
            merchant.address.city || 'N/A',
            merchant.address.state || 'N/A',
            this.formatDate(merchant.created_at)
        ]);

        return [headers, ...rows].map(row => 
            row.map(field => `"${field}"`).join(',')
        ).join('\n');
    }

    showError(message) {
        // Simple error display - could be enhanced with a proper notification system
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #e74c3c;
            color: white;
            padding: 16px 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(231, 76, 60, 0.3);
            z-index: 10000;
            animation: slideInRight 0.3s ease-out;
        `;
        errorDiv.textContent = message;
        
        document.body.appendChild(errorDiv);
        
        setTimeout(() => {
            errorDiv.remove();
        }, 5000);
    }

    formatPortfolioType(type) {
        const types = {
            'onboarded': 'Onboarded',
            'deactivated': 'Deactivated',
            'prospective': 'Prospective',
            'pending': 'Pending'
        };
        return types[type] || type;
    }

    formatRiskLevel(level) {
        const levels = {
            'low': 'Low Risk',
            'medium': 'Medium Risk',
            'high': 'High Risk'
        };
        return levels[level] || level;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    }

    getAuthToken() {
        // Get auth token from localStorage or cookie
        return localStorage.getItem('auth_token') || 
               document.cookie.split('; ').find(row => row.startsWith('auth_token='))?.split('=')[1] || 
               '';
    }

    // Public methods for external control
    setFilters(filters) {
        this.currentFilters = { ...this.currentFilters, ...filters };
        this.updateFilterUI();
        this.performSearch();
    }

    updateFilterUI() {
        document.getElementById('portfolioTypeFilter').value = this.currentFilters.portfolioType || '';
        document.getElementById('riskLevelFilter').value = this.currentFilters.riskLevel || '';
        document.getElementById('industryFilter').value = this.currentFilters.industry || '';
        document.getElementById('statusFilter').value = this.currentFilters.status || '';
    }

    refresh() {
        this.performSearch();
    }

    getSelectedMerchant() {
        const selectedElement = document.querySelector('.merchant-item.selected');
        if (selectedElement) {
            const merchantId = selectedElement.dataset.merchantId;
            return this.merchants.find(m => m.id === merchantId);
        }
        return null;
    }

    destroy() {
        // Clean up event listeners and DOM elements
        clearTimeout(this.searchTimeout);
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantSearch;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const searchContainer = document.getElementById('merchantSearchContainer');
    if (searchContainer && !window.merchantSearch) {
        window.merchantSearch = new MerchantSearch({
            container: searchContainer
        });
    }
});
