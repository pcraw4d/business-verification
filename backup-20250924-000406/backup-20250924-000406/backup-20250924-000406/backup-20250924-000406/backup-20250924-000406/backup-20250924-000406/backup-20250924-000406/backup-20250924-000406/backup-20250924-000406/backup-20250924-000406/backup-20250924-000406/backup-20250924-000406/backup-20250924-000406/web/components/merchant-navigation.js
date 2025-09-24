/**
 * Merchant Navigation Component
 * Provides merchant-specific navigation with breadcrumb navigation and quick merchant switching
 * Integrates with SessionManager for single merchant session management
 */
class MerchantNavigation {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.sessionManager = options.sessionManager || window.sessionManager;
        this.maxBreadcrumbItems = options.maxBreadcrumbItems || 5;
        this.quickSwitchLimit = options.quickSwitchLimit || 8;
        this.currentMerchant = null;
        this.navigationHistory = [];
        this.quickSwitchMerchants = [];
        this.isInitialized = false;
        
        // Event callbacks
        this.onMerchantSwitch = options.onMerchantSwitch || null;
        this.onNavigationChange = options.onNavigationChange || null;
        this.onBreadcrumbClick = options.onBreadcrumbClick || null;
        
        this.init();
    }

    init() {
        this.createNavigationInterface();
        this.bindEvents();
        this.loadQuickSwitchMerchants();
        this.updateNavigationState();
        this.isInitialized = true;
    }

    createNavigationInterface() {
        const navigationHTML = `
            <div class="merchant-navigation-container" id="merchantNavigationContainer">
                <!-- Breadcrumb Navigation -->
                <div class="breadcrumb-navigation" id="breadcrumbNavigation">
                    <div class="breadcrumb-header">
                        <h3 class="breadcrumb-title">
                            <i class="fas fa-route"></i>
                            Navigation
                        </h3>
                        <div class="breadcrumb-actions">
                            <button class="btn btn-outline btn-sm" id="refreshBreadcrumbBtn" title="Refresh Navigation">
                                <i class="fas fa-sync-alt"></i>
                            </button>
                            <button class="btn btn-outline btn-sm" id="clearBreadcrumbBtn" title="Clear History">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    <div class="breadcrumb-trail" id="breadcrumbTrail">
                        <div class="breadcrumb-item home" data-page="home">
                            <i class="fas fa-home"></i>
                            <span>Home</span>
                        </div>
                        <div class="breadcrumb-separator">
                            <i class="fas fa-chevron-right"></i>
                        </div>
                        <div class="breadcrumb-item portfolio" data-page="portfolio">
                            <i class="fas fa-th-large"></i>
                            <span>Merchant Portfolio</span>
                        </div>
                    </div>
                </div>

                <!-- Quick Merchant Switch -->
                <div class="quick-switch-navigation" id="quickSwitchNavigation">
                    <div class="quick-switch-header">
                        <h3 class="quick-switch-title">
                            <i class="fas fa-exchange-alt"></i>
                            Quick Switch
                        </h3>
                        <div class="quick-switch-actions">
                            <button class="btn btn-outline btn-sm" id="refreshQuickSwitchBtn" title="Refresh List">
                                <i class="fas fa-sync-alt"></i>
                            </button>
                            <button class="btn btn-outline btn-sm" id="manageQuickSwitchBtn" title="Manage List">
                                <i class="fas fa-cog"></i>
                            </button>
                        </div>
                    </div>
                    <div class="quick-switch-list" id="quickSwitchList">
                        <div class="no-merchants">
                            <i class="fas fa-users"></i>
                            <p>No merchants available for quick switch</p>
                            <button class="btn btn-primary btn-sm" id="loadMerchantsBtn">
                                <i class="fas fa-download"></i>
                                Load Merchants
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Navigation History -->
                <div class="navigation-history" id="navigationHistory">
                    <div class="history-header">
                        <h3 class="history-title">
                            <i class="fas fa-history"></i>
                            Navigation History
                        </h3>
                        <div class="history-actions">
                            <button class="btn btn-outline btn-sm" id="clearHistoryBtn" title="Clear History">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    <div class="history-list" id="historyList">
                        <div class="no-history">
                            <i class="fas fa-history"></i>
                            <p>No navigation history</p>
                        </div>
                    </div>
                </div>

                <!-- Quick Switch Management Modal -->
                <div class="quick-switch-modal" id="quickSwitchModal" style="display: none;">
                    <div class="modal-overlay">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h4>Manage Quick Switch Merchants</h4>
                                <button class="btn btn-outline btn-sm" id="closeQuickSwitchModalBtn">
                                    <i class="fas fa-times"></i>
                                </button>
                            </div>
                            <div class="modal-body">
                                <div class="merchant-search">
                                    <input type="text" id="merchantSearchInput" placeholder="Search merchants..." class="form-input">
                                    <button class="btn btn-primary btn-sm" id="searchMerchantsBtn">
                                        <i class="fas fa-search"></i>
                                    </button>
                                </div>
                                <div class="available-merchants" id="availableMerchants">
                                    <!-- Available merchants will be populated here -->
                                </div>
                                <div class="selected-merchants" id="selectedMerchants">
                                    <h5>Selected for Quick Switch</h5>
                                    <div class="selected-list" id="selectedList">
                                        <!-- Selected merchants will be populated here -->
                                    </div>
                                </div>
                            </div>
                            <div class="modal-footer">
                                <button class="btn btn-secondary" id="cancelQuickSwitchBtn">Cancel</button>
                                <button class="btn btn-primary" id="saveQuickSwitchBtn">Save Changes</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.container.innerHTML = navigationHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .merchant-navigation-container {
                    background: white;
                    border-radius: 12px;
                    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
                    padding: 20px;
                    margin-bottom: 20px;
                    border-left: 4px solid #27ae60;
                }

                /* Breadcrumb Navigation */
                .breadcrumb-navigation {
                    margin-bottom: 24px;
                }

                .breadcrumb-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                }

                .breadcrumb-title {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.1rem;
                    font-weight: 600;
                }

                .breadcrumb-title i {
                    color: #27ae60;
                }

                .breadcrumb-actions {
                    display: flex;
                    gap: 8px;
                }

                .breadcrumb-trail {
                    display: flex;
                    align-items: center;
                    flex-wrap: wrap;
                    gap: 8px;
                    padding: 12px 16px;
                    background: #f8f9fa;
                    border-radius: 8px;
                    border: 1px solid #e9ecef;
                }

                .breadcrumb-item {
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    padding: 6px 12px;
                    background: white;
                    border-radius: 6px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    font-size: 0.9rem;
                    font-weight: 500;
                    color: #5a6c7d;
                    border: 1px solid transparent;
                }

                .breadcrumb-item:hover {
                    background: #e3f2fd;
                    color: #3498db;
                    transform: translateY(-1px);
                    box-shadow: 0 2px 8px rgba(52, 152, 219, 0.2);
                }

                .breadcrumb-item.active {
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    color: white;
                    box-shadow: 0 2px 8px rgba(52, 152, 219, 0.3);
                }

                .breadcrumb-item.home {
                    background: linear-gradient(135deg, #27ae60, #2ecc71);
                    color: white;
                }

                .breadcrumb-separator {
                    color: #95a5a6;
                    font-size: 0.8rem;
                }

                /* Quick Switch Navigation */
                .quick-switch-navigation {
                    margin-bottom: 24px;
                }

                .quick-switch-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                }

                .quick-switch-title {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.1rem;
                    font-weight: 600;
                }

                .quick-switch-title i {
                    color: #f39c12;
                }

                .quick-switch-actions {
                    display: flex;
                    gap: 8px;
                }

                .quick-switch-list {
                    display: grid;
                    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
                    gap: 12px;
                    max-height: 200px;
                    overflow-y: auto;
                    padding: 4px;
                }

                .quick-switch-item {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    padding: 12px;
                    background: #f8f9fa;
                    border-radius: 8px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    border: 1px solid #e9ecef;
                }

                .quick-switch-item:hover {
                    background: #e3f2fd;
                    transform: translateY(-2px);
                    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.2);
                }

                .quick-switch-avatar {
                    width: 35px;
                    height: 35px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-weight: 600;
                    font-size: 0.9rem;
                    flex-shrink: 0;
                }

                .quick-switch-details {
                    flex: 1;
                    min-width: 0;
                }

                .quick-switch-name {
                    font-weight: 600;
                    color: #2c3e50;
                    font-size: 0.9rem;
                    margin-bottom: 2px;
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }

                .quick-switch-meta {
                    font-size: 0.8rem;
                    color: #6c757d;
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }

                .no-merchants {
                    grid-column: 1 / -1;
                    text-align: center;
                    padding: 40px 20px;
                    color: #6c757d;
                }

                .no-merchants i {
                    font-size: 2rem;
                    color: #dee2e6;
                    margin-bottom: 12px;
                }

                .no-merchants p {
                    margin: 0 0 16px 0;
                    font-size: 0.9rem;
                }

                /* Navigation History */
                .navigation-history {
                    margin-bottom: 0;
                }

                .history-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                }

                .history-title {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.1rem;
                    font-weight: 600;
                }

                .history-title i {
                    color: #9b59b6;
                }

                .history-actions {
                    display: flex;
                    gap: 8px;
                }

                .history-list {
                    max-height: 150px;
                    overflow-y: auto;
                }

                .history-item {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    padding: 10px 12px;
                    background: #f8f9fa;
                    border-radius: 6px;
                    margin-bottom: 8px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    border: 1px solid #e9ecef;
                }

                .history-item:hover {
                    background: #e3f2fd;
                    transform: translateX(4px);
                }

                .history-icon {
                    width: 30px;
                    height: 30px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #9b59b6, #8e44ad);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-size: 0.8rem;
                    flex-shrink: 0;
                }

                .history-details {
                    flex: 1;
                }

                .history-name {
                    font-weight: 600;
                    color: #2c3e50;
                    font-size: 0.9rem;
                    margin-bottom: 2px;
                }

                .history-meta {
                    font-size: 0.8rem;
                    color: #6c757d;
                }

                .no-history {
                    text-align: center;
                    padding: 30px 20px;
                    color: #6c757d;
                }

                .no-history i {
                    font-size: 1.5rem;
                    color: #dee2e6;
                    margin-bottom: 8px;
                }

                .no-history p {
                    margin: 0;
                    font-size: 0.9rem;
                }

                /* Modal Styles */
                .quick-switch-modal {
                    position: fixed;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    z-index: 10000;
                }

                .modal-overlay {
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(0, 0, 0, 0.5);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    padding: 20px;
                }

                .modal-content {
                    background: white;
                    border-radius: 12px;
                    max-width: 600px;
                    width: 100%;
                    max-height: 90vh;
                    overflow-y: auto;
                    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
                }

                .modal-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    padding: 20px;
                    border-bottom: 2px solid #f8f9fa;
                }

                .modal-header h4 {
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.2rem;
                }

                .modal-body {
                    padding: 20px;
                }

                .merchant-search {
                    display: flex;
                    gap: 12px;
                    margin-bottom: 20px;
                }

                .form-input {
                    flex: 1;
                    padding: 10px 12px;
                    border: 2px solid #e9ecef;
                    border-radius: 6px;
                    font-size: 0.9rem;
                    transition: border-color 0.3s ease;
                }

                .form-input:focus {
                    outline: none;
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .available-merchants {
                    margin-bottom: 20px;
                }

                .available-merchants h5 {
                    margin: 0 0 12px 0;
                    color: #2c3e50;
                    font-size: 1rem;
                }

                .merchant-item {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    padding: 12px;
                    background: #f8f9fa;
                    border-radius: 6px;
                    margin-bottom: 8px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    border: 1px solid #e9ecef;
                }

                .merchant-item:hover {
                    background: #e3f2fd;
                }

                .merchant-item.selected {
                    background: #d4edda;
                    border-color: #27ae60;
                }

                .merchant-avatar {
                    width: 35px;
                    height: 35px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-weight: 600;
                    font-size: 0.9rem;
                    flex-shrink: 0;
                }

                .merchant-details {
                    flex: 1;
                }

                .merchant-name {
                    font-weight: 600;
                    color: #2c3e50;
                    font-size: 0.9rem;
                    margin-bottom: 2px;
                }

                .merchant-meta {
                    font-size: 0.8rem;
                    color: #6c757d;
                }

                .selected-merchants h5 {
                    margin: 0 0 12px 0;
                    color: #2c3e50;
                    font-size: 1rem;
                }

                .selected-list {
                    max-height: 200px;
                    overflow-y: auto;
                }

                .modal-footer {
                    display: flex;
                    justify-content: flex-end;
                    gap: 12px;
                    padding: 20px;
                    border-top: 2px solid #f8f9fa;
                }

                /* Button Styles */
                .btn {
                    padding: 8px 16px;
                    border: none;
                    border-radius: 6px;
                    font-size: 0.85rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    text-decoration: none;
                }

                .btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }

                .btn-sm {
                    padding: 6px 12px;
                    font-size: 0.8rem;
                }

                .btn-outline {
                    background: transparent;
                    color: #6c757d;
                    border: 2px solid #e9ecef;
                }

                .btn-outline:hover:not(:disabled) {
                    background: #6c757d;
                    color: white;
                    border-color: #6c757d;
                }

                .btn-primary {
                    background: #3498db;
                    color: white;
                }

                .btn-primary:hover:not(:disabled) {
                    background: #2980b9;
                    transform: translateY(-1px);
                }

                .btn-secondary {
                    background: #6c757d;
                    color: white;
                }

                .btn-secondary:hover:not(:disabled) {
                    background: #5a6268;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .merchant-navigation-container {
                        padding: 16px;
                    }

                    .breadcrumb-trail {
                        flex-direction: column;
                        align-items: flex-start;
                        gap: 8px;
                    }

                    .breadcrumb-separator {
                        display: none;
                    }

                    .quick-switch-list {
                        grid-template-columns: 1fr;
                    }

                    .modal-content {
                        margin: 10px;
                        max-width: none;
                    }

                    .modal-footer {
                        flex-direction: column;
                    }

                    .merchant-search {
                        flex-direction: column;
                    }
                }

                /* Loading states */
                .loading {
                    opacity: 0.6;
                    pointer-events: none;
                }

                .loading::after {
                    content: '';
                    position: absolute;
                    top: 50%;
                    left: 50%;
                    width: 20px;
                    height: 20px;
                    margin: -10px 0 0 -10px;
                    border: 2px solid #f3f3f3;
                    border-top: 2px solid #3498db;
                    border-radius: 50%;
                    animation: spin 1s linear infinite;
                }

                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#merchant-navigation-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'merchant-navigation-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        // Breadcrumb events
        const refreshBreadcrumbBtn = document.getElementById('refreshBreadcrumbBtn');
        const clearBreadcrumbBtn = document.getElementById('clearBreadcrumbBtn');
        const breadcrumbTrail = document.getElementById('breadcrumbTrail');

        refreshBreadcrumbBtn.addEventListener('click', () => {
            this.refreshBreadcrumb();
        });

        clearBreadcrumbBtn.addEventListener('click', () => {
            this.clearBreadcrumb();
        });

        // Breadcrumb item clicks
        breadcrumbTrail.addEventListener('click', (e) => {
            const breadcrumbItem = e.target.closest('.breadcrumb-item');
            if (breadcrumbItem) {
                this.handleBreadcrumbClick(breadcrumbItem);
            }
        });

        // Quick switch events
        const refreshQuickSwitchBtn = document.getElementById('refreshQuickSwitchBtn');
        const manageQuickSwitchBtn = document.getElementById('manageQuickSwitchBtn');
        const loadMerchantsBtn = document.getElementById('loadMerchantsBtn');
        const quickSwitchList = document.getElementById('quickSwitchList');

        refreshQuickSwitchBtn.addEventListener('click', () => {
            this.loadQuickSwitchMerchants();
        });

        manageQuickSwitchBtn.addEventListener('click', () => {
            this.showQuickSwitchModal();
        });

        loadMerchantsBtn.addEventListener('click', () => {
            this.loadQuickSwitchMerchants();
        });

        // Quick switch item clicks
        quickSwitchList.addEventListener('click', (e) => {
            const quickSwitchItem = e.target.closest('.quick-switch-item');
            if (quickSwitchItem) {
                const merchantId = quickSwitchItem.dataset.merchantId;
                this.switchToMerchant(merchantId);
            }
        });

        // History events
        const clearHistoryBtn = document.getElementById('clearHistoryBtn');
        const historyList = document.getElementById('historyList');

        clearHistoryBtn.addEventListener('click', () => {
            this.clearNavigationHistory();
        });

        // History item clicks
        historyList.addEventListener('click', (e) => {
            const historyItem = e.target.closest('.history-item');
            if (historyItem) {
                const merchantId = historyItem.dataset.merchantId;
                this.switchToMerchant(merchantId);
            }
        });

        // Modal events
        const closeQuickSwitchModalBtn = document.getElementById('closeQuickSwitchModalBtn');
        const cancelQuickSwitchBtn = document.getElementById('cancelQuickSwitchBtn');
        const saveQuickSwitchBtn = document.getElementById('saveQuickSwitchBtn');
        const searchMerchantsBtn = document.getElementById('searchMerchantsBtn');
        const merchantSearchInput = document.getElementById('merchantSearchInput');

        closeQuickSwitchModalBtn.addEventListener('click', () => {
            this.hideQuickSwitchModal();
        });

        cancelQuickSwitchBtn.addEventListener('click', () => {
            this.hideQuickSwitchModal();
        });

        saveQuickSwitchBtn.addEventListener('click', () => {
            this.saveQuickSwitchSelection();
        });

        searchMerchantsBtn.addEventListener('click', () => {
            this.searchMerchants();
        });

        merchantSearchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchMerchants();
            }
        });

        // Session manager integration
        if (this.sessionManager) {
            this.sessionManager.onSessionStart = (session) => {
                this.updateCurrentMerchant(session.merchant);
            };

            this.sessionManager.onSessionEnd = (session) => {
                this.updateCurrentMerchant(null);
            };

            this.sessionManager.onSessionSwitch = (merchant) => {
                this.updateCurrentMerchant(merchant);
            };
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'b':
                        e.preventDefault();
                        this.toggleBreadcrumb();
                        break;
                    case 'q':
                        e.preventDefault();
                        this.toggleQuickSwitch();
                        break;
                    case 'h':
                        e.preventDefault();
                        this.toggleHistory();
                        break;
                }
            }
        });
    }

    // Breadcrumb Management
    addBreadcrumbItem(page, title, icon = 'fas fa-file', merchantId = null) {
        const breadcrumbTrail = document.getElementById('breadcrumbTrail');
        
        // Remove existing active item
        const activeItem = breadcrumbTrail.querySelector('.breadcrumb-item.active');
        if (activeItem) {
            activeItem.classList.remove('active');
        }

        // Create new breadcrumb item
        const breadcrumbItem = document.createElement('div');
        breadcrumbItem.className = 'breadcrumb-item active';
        breadcrumbItem.dataset.page = page;
        if (merchantId) {
            breadcrumbItem.dataset.merchantId = merchantId;
        }

        breadcrumbItem.innerHTML = `
            <i class="${icon}"></i>
            <span>${title}</span>
        `;

        // Add separator if not the first item
        if (breadcrumbTrail.children.length > 0) {
            const separator = document.createElement('div');
            separator.className = 'breadcrumb-separator';
            separator.innerHTML = '<i class="fas fa-chevron-right"></i>';
            breadcrumbTrail.appendChild(separator);
        }

        breadcrumbTrail.appendChild(breadcrumbItem);

        // Limit breadcrumb items
        this.limitBreadcrumbItems();

        // Add to navigation history
        this.addToNavigationHistory(page, title, merchantId);
    }

    limitBreadcrumbItems() {
        const breadcrumbTrail = document.getElementById('breadcrumbTrail');
        const items = breadcrumbTrail.querySelectorAll('.breadcrumb-item');
        const separators = breadcrumbTrail.querySelectorAll('.breadcrumb-separator');

        if (items.length > this.maxBreadcrumbItems) {
            // Remove oldest items (keep home and portfolio)
            const itemsToRemove = items.length - this.maxBreadcrumbItems;
            for (let i = 0; i < itemsToRemove; i++) {
                if (items[i] && !items[i].classList.contains('home') && !items[i].classList.contains('portfolio')) {
                    items[i].remove();
                    if (separators[i]) {
                        separators[i].remove();
                    }
                }
            }
        }
    }

    handleBreadcrumbClick(breadcrumbItem) {
        const page = breadcrumbItem.dataset.page;
        const merchantId = breadcrumbItem.dataset.merchantId;

        // Call callback
        if (this.onBreadcrumbClick) {
            this.onBreadcrumbClick(page, merchantId);
        }

        // Update active state
        breadcrumbTrail.querySelectorAll('.breadcrumb-item').forEach(item => {
            item.classList.remove('active');
        });
        breadcrumbItem.classList.add('active');

        console.log(`Breadcrumb clicked: ${page}${merchantId ? ` (Merchant: ${merchantId})` : ''}`);
    }

    refreshBreadcrumb() {
        this.updateNavigationState();
        console.log('Breadcrumb refreshed');
    }

    clearBreadcrumb() {
        const breadcrumbTrail = document.getElementById('breadcrumbTrail');
        breadcrumbTrail.innerHTML = `
            <div class="breadcrumb-item home active" data-page="home">
                <i class="fas fa-home"></i>
                <span>Home</span>
            </div>
            <div class="breadcrumb-separator">
                <i class="fas fa-chevron-right"></i>
            </div>
            <div class="breadcrumb-item portfolio" data-page="portfolio">
                <i class="fas fa-th-large"></i>
                <span>Merchant Portfolio</span>
            </div>
        `;
        console.log('Breadcrumb cleared');
    }

    // Quick Switch Management
    async loadQuickSwitchMerchants() {
        const quickSwitchList = document.getElementById('quickSwitchList');
        quickSwitchList.classList.add('loading');

        try {
            // Load from localStorage first
            const savedMerchants = localStorage.getItem('quickSwitchMerchants');
            if (savedMerchants) {
                this.quickSwitchMerchants = JSON.parse(savedMerchants);
            }

            // If no saved merchants, load from API
            if (this.quickSwitchMerchants.length === 0) {
                const response = await fetch(`${this.apiBaseUrl}/merchants?limit=${this.quickSwitchLimit}`);
                if (response.ok) {
                    const data = await response.json();
                    this.quickSwitchMerchants = data.merchants || [];
                }
            }

            this.renderQuickSwitchList();
        } catch (error) {
            console.error('Failed to load quick switch merchants:', error);
            this.renderQuickSwitchList();
        } finally {
            quickSwitchList.classList.remove('loading');
        }
    }

    renderQuickSwitchList() {
        const quickSwitchList = document.getElementById('quickSwitchList');

        if (this.quickSwitchMerchants.length === 0) {
            quickSwitchList.innerHTML = `
                <div class="no-merchants">
                    <i class="fas fa-users"></i>
                    <p>No merchants available for quick switch</p>
                    <button class="btn btn-primary btn-sm" id="loadMerchantsBtn">
                        <i class="fas fa-download"></i>
                        Load Merchants
                    </button>
                </div>
            `;
        } else {
            quickSwitchList.innerHTML = this.quickSwitchMerchants.map(merchant => `
                <div class="quick-switch-item" data-merchant-id="${merchant.id}">
                    <div class="quick-switch-avatar">
                        ${merchant.name.charAt(0).toUpperCase()}
                    </div>
                    <div class="quick-switch-details">
                        <div class="quick-switch-name">${merchant.name}</div>
                        <div class="quick-switch-meta">${merchant.industry || 'N/A'}</div>
                    </div>
                </div>
            `).join('');
        }
    }

    switchToMerchant(merchantId) {
        const merchant = this.quickSwitchMerchants.find(m => m.id === merchantId);
        if (!merchant) return;

        // Use session manager if available
        if (this.sessionManager) {
            this.sessionManager.switchSession(merchant);
        }

        // Call callback
        if (this.onMerchantSwitch) {
            this.onMerchantSwitch(merchant);
        }

        // Update current merchant
        this.updateCurrentMerchant(merchant);

        console.log(`Switched to merchant: ${merchant.name}`);
    }

    showQuickSwitchModal() {
        const modal = document.getElementById('quickSwitchModal');
        modal.style.display = 'block';
        this.loadAvailableMerchants();
    }

    hideQuickSwitchModal() {
        const modal = document.getElementById('quickSwitchModal');
        modal.style.display = 'none';
    }

    async loadAvailableMerchants() {
        const availableMerchants = document.getElementById('availableMerchants');
        availableMerchants.innerHTML = '<p>Loading merchants...</p>';

        try {
            const response = await fetch(`${this.apiBaseUrl}/merchants?limit=50`);
            if (response.ok) {
                const data = await response.json();
                this.renderAvailableMerchants(data.merchants || []);
            }
        } catch (error) {
            console.error('Failed to load available merchants:', error);
            availableMerchants.innerHTML = '<p>Failed to load merchants</p>';
        }
    }

    renderAvailableMerchants(merchants) {
        const availableMerchants = document.getElementById('availableMerchants');
        const selectedList = document.getElementById('selectedList');

        availableMerchants.innerHTML = `
            <h5>Available Merchants</h5>
            ${merchants.map(merchant => `
                <div class="merchant-item" data-merchant-id="${merchant.id}">
                    <div class="merchant-avatar">
                        ${merchant.name.charAt(0).toUpperCase()}
                    </div>
                    <div class="merchant-details">
                        <div class="merchant-name">${merchant.name}</div>
                        <div class="merchant-meta">${merchant.industry || 'N/A'} • ${merchant.portfolioType || 'N/A'}</div>
                    </div>
                </div>
            `).join('')}
        `;

        // Add click handlers
        availableMerchants.querySelectorAll('.merchant-item').forEach(item => {
            item.addEventListener('click', () => {
                this.toggleMerchantSelection(item);
            });
        });

        // Render selected merchants
        this.renderSelectedMerchants();
    }

    toggleMerchantSelection(merchantItem) {
        const merchantId = merchantItem.dataset.merchantId;
        const isSelected = merchantItem.classList.contains('selected');

        if (isSelected) {
            merchantItem.classList.remove('selected');
            this.quickSwitchMerchants = this.quickSwitchMerchants.filter(m => m.id !== merchantId);
        } else {
            if (this.quickSwitchMerchants.length >= this.quickSwitchLimit) {
                alert(`Maximum ${this.quickSwitchLimit} merchants allowed for quick switch`);
                return;
            }
            merchantItem.classList.add('selected');
            // Add merchant to selection (you'd need to get the full merchant object)
            const merchantName = merchantItem.querySelector('.merchant-name').textContent;
            const merchantMeta = merchantItem.querySelector('.merchant-meta').textContent;
            this.quickSwitchMerchants.push({
                id: merchantId,
                name: merchantName,
                industry: merchantMeta.split(' • ')[0]
            });
        }

        this.renderSelectedMerchants();
    }

    renderSelectedMerchants() {
        const selectedList = document.getElementById('selectedList');
        
        if (this.quickSwitchMerchants.length === 0) {
            selectedList.innerHTML = '<p>No merchants selected</p>';
        } else {
            selectedList.innerHTML = this.quickSwitchMerchants.map(merchant => `
                <div class="merchant-item selected">
                    <div class="merchant-avatar">
                        ${merchant.name.charAt(0).toUpperCase()}
                    </div>
                    <div class="merchant-details">
                        <div class="merchant-name">${merchant.name}</div>
                        <div class="merchant-meta">${merchant.industry || 'N/A'}</div>
                    </div>
                </div>
            `).join('');
        }
    }

    saveQuickSwitchSelection() {
        localStorage.setItem('quickSwitchMerchants', JSON.stringify(this.quickSwitchMerchants));
        this.renderQuickSwitchList();
        this.hideQuickSwitchModal();
        console.log('Quick switch selection saved');
    }

    searchMerchants() {
        const searchInput = document.getElementById('merchantSearchInput');
        const query = searchInput.value.trim();
        
        if (query) {
            // Filter available merchants by search query
            const merchantItems = document.querySelectorAll('#availableMerchants .merchant-item');
            merchantItems.forEach(item => {
                const name = item.querySelector('.merchant-name').textContent.toLowerCase();
                const meta = item.querySelector('.merchant-meta').textContent.toLowerCase();
                
                if (name.includes(query.toLowerCase()) || meta.includes(query.toLowerCase())) {
                    item.style.display = 'flex';
                } else {
                    item.style.display = 'none';
                }
            });
        } else {
            // Show all merchants
            const merchantItems = document.querySelectorAll('#availableMerchants .merchant-item');
            merchantItems.forEach(item => {
                item.style.display = 'flex';
            });
        }
    }

    // Navigation History Management
    addToNavigationHistory(page, title, merchantId = null) {
        const historyItem = {
            page,
            title,
            merchantId,
            timestamp: new Date(),
            id: this.generateHistoryId()
        };

        this.navigationHistory.unshift(historyItem);

        // Limit history size
        if (this.navigationHistory.length > 20) {
            this.navigationHistory = this.navigationHistory.slice(0, 20);
        }

        this.renderNavigationHistory();
    }

    renderNavigationHistory() {
        const historyList = document.getElementById('historyList');

        if (this.navigationHistory.length === 0) {
            historyList.innerHTML = `
                <div class="no-history">
                    <i class="fas fa-history"></i>
                    <p>No navigation history</p>
                </div>
            `;
        } else {
            historyList.innerHTML = this.navigationHistory.slice(0, 10).map(item => `
                <div class="history-item" data-merchant-id="${item.merchantId || ''}">
                    <div class="history-icon">
                        <i class="fas fa-${this.getPageIcon(item.page)}"></i>
                    </div>
                    <div class="history-details">
                        <div class="history-name">${item.title}</div>
                        <div class="history-meta">${this.formatTimestamp(item.timestamp)}</div>
                    </div>
                </div>
            `).join('');
        }
    }

    clearNavigationHistory() {
        this.navigationHistory = [];
        this.renderNavigationHistory();
        console.log('Navigation history cleared');
    }

    // Utility Methods
    updateCurrentMerchant(merchant) {
        this.currentMerchant = merchant;
        this.updateNavigationState();
    }

    updateNavigationState() {
        if (this.currentMerchant) {
            // Update breadcrumb to show current merchant
            this.addBreadcrumbItem('merchant-detail', this.currentMerchant.name, 'fas fa-building', this.currentMerchant.id);
        }
    }

    getPageIcon(page) {
        const iconMap = {
            'home': 'home',
            'portfolio': 'th-large',
            'merchant-detail': 'building',
            'risk-assessment': 'exclamation-triangle',
            'compliance': 'clipboard-check',
            'analytics': 'chart-line'
        };
        return iconMap[page] || 'file';
    }

    generateHistoryId() {
        return 'history_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    formatTimestamp(timestamp) {
        const now = new Date();
        const time = new Date(timestamp);
        const diff = now - time;

        if (diff < 60000) { // Less than 1 minute
            return 'Just now';
        } else if (diff < 3600000) { // Less than 1 hour
            const minutes = Math.floor(diff / 60000);
            return `${minutes}m ago`;
        } else if (diff < 86400000) { // Less than 1 day
            const hours = Math.floor(diff / 3600000);
            return `${hours}h ago`;
        } else {
            return time.toLocaleDateString();
        }
    }

    // Public API Methods
    getCurrentMerchant() {
        return this.currentMerchant;
    }

    getNavigationHistory() {
        return [...this.navigationHistory];
    }

    getQuickSwitchMerchants() {
        return [...this.quickSwitchMerchants];
    }

    setQuickSwitchMerchants(merchants) {
        this.quickSwitchMerchants = merchants;
        this.renderQuickSwitchList();
    }

    // Toggle methods for keyboard shortcuts
    toggleBreadcrumb() {
        const breadcrumb = document.getElementById('breadcrumbNavigation');
        breadcrumb.style.display = breadcrumb.style.display === 'none' ? 'block' : 'none';
    }

    toggleQuickSwitch() {
        const quickSwitch = document.getElementById('quickSwitchNavigation');
        quickSwitch.style.display = quickSwitch.style.display === 'none' ? 'block' : 'none';
    }

    toggleHistory() {
        const history = document.getElementById('navigationHistory');
        history.style.display = history.style.display === 'none' ? 'block' : 'none';
    }

    destroy() {
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantNavigation;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const navigationContainer = document.getElementById('merchantNavigationContainer');
    if (navigationContainer && !window.merchantNavigation) {
        window.merchantNavigation = new MerchantNavigation({
            container: navigationContainer,
            sessionManager: window.sessionManager
        });
    }
});
