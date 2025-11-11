// Merchant Context Component for KYB Platform
// Provides merchant context integration for existing dashboards

class MerchantContext {
    constructor(options = {}) {
        this.currentMerchant = null;
        this.sessionManager = null;
        this.options = {
            showInHeader: true,
            showInSidebar: false,
            enableQuickSwitch: true,
            ...options
        };
        
        this.init();
    }

    init() {
        this.loadSessionManager();
        this.createContextUI();
        this.bindEvents();
        this.loadCurrentMerchant();
    }

    async loadSessionManager() {
        // Ensure shared components are loaded
        if (typeof loadSharedComponents === 'function') {
            await loadSharedComponents();
        }
        
        // Check if session manager is available
        if (typeof SessionManager !== 'undefined') {
            this.sessionManager = new SessionManager();
        }
    }

    createContextUI() {
        if (this.options.showInHeader) {
            this.createHeaderContext();
        }
        
        if (this.options.showInSidebar) {
            this.createSidebarContext();
        }
    }

    createHeaderContext() {
        // Find the main header or create one
        let header = document.querySelector('.main-header, .dashboard-header, .page-header');
        
        if (!header) {
            // Create a header if none exists
            header = document.createElement('div');
            header.className = 'merchant-context-header';
            header.style.cssText = `
                background: rgba(255, 255, 255, 0.95);
                backdrop-filter: blur(10px);
                border-radius: 15px;
                padding: 15px 25px;
                margin-bottom: 20px;
                box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
                display: flex;
                justify-content: space-between;
                align-items: center;
                flex-wrap: wrap;
                gap: 15px;
            `;
            
            // Insert at the beginning of main content
            const mainContent = document.querySelector('.main-content, .container, main');
            if (mainContent) {
                mainContent.insertBefore(header, mainContent.firstChild);
            }
        }

        // Add merchant context to header
        const contextHTML = `
            <div class="merchant-context-info">
                <div class="merchant-avatar-small">
                    <i class="fas fa-store"></i>
                </div>
                <div class="merchant-details">
                    <div class="merchant-name" id="contextMerchantName">No Merchant Selected</div>
                    <div class="merchant-status" id="contextMerchantStatus">Select a merchant to view details</div>
                </div>
            </div>
            <div class="merchant-context-actions">
                <a href="merchant-portfolio.html" class="btn btn-outline-primary">
                    <i class="fas fa-th-large"></i>
                    Portfolio
                </a>
                <button class="btn btn-primary" id="switchMerchantBtn">
                    <i class="fas fa-exchange-alt"></i>
                    Switch Merchant
                </button>
            </div>
        `;

        // Add context to header
        const contextContainer = document.createElement('div');
        contextContainer.className = 'merchant-context-container';
        contextContainer.innerHTML = contextHTML;
        header.appendChild(contextContainer);

        // Add styles
        this.addContextStyles();
    }

    createSidebarContext() {
        const sidebar = document.querySelector('.kyb-sidebar .sidebar-content');
        if (!sidebar) return;

        const contextHTML = `
            <div class="nav-section merchant-context-section">
                <h3 class="nav-section-title">Current Merchant</h3>
                <div class="merchant-context-card">
                    <div class="merchant-avatar-small">
                        <i class="fas fa-store"></i>
                    </div>
                    <div class="merchant-details">
                        <div class="merchant-name" id="sidebarMerchantName">No Merchant Selected</div>
                        <div class="merchant-status" id="sidebarMerchantStatus">Select a merchant</div>
                    </div>
                    <div class="merchant-actions">
                        <a href="merchant-detail.html" class="nav-link-small">
                            <i class="fas fa-eye"></i>
                            View Details
                        </a>
                    </div>
                </div>
            </div>
        `;

        const contextContainer = document.createElement('div');
        contextContainer.innerHTML = contextHTML;
        sidebar.insertBefore(contextContainer, sidebar.firstChild);
    }

    addContextStyles() {
        const style = document.createElement('style');
        style.textContent = `
            .merchant-context-container {
                display: flex;
                justify-content: space-between;
                align-items: center;
                width: 100%;
                flex-wrap: wrap;
                gap: 15px;
            }

            .merchant-context-info {
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .merchant-avatar-small {
                width: 40px;
                height: 40px;
                border-radius: 50%;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-size: 16px;
            }

            .merchant-details {
                display: flex;
                flex-direction: column;
            }

            .merchant-name {
                font-weight: 600;
                font-size: 16px;
                color: #333;
                margin-bottom: 2px;
            }

            .merchant-status {
                font-size: 12px;
                color: #666;
            }

            .merchant-context-actions {
                display: flex;
                gap: 10px;
                align-items: center;
            }

            .btn {
                padding: 8px 16px;
                border-radius: 8px;
                text-decoration: none;
                font-size: 14px;
                font-weight: 500;
                border: none;
                cursor: pointer;
                display: inline-flex;
                align-items: center;
                gap: 6px;
                transition: all 0.3s ease;
            }

            .btn-primary {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
            }

            .btn-primary:hover {
                transform: translateY(-2px);
                box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
            }

            .btn-outline-primary {
                background: transparent;
                color: #667eea;
                border: 2px solid #667eea;
            }

            .btn-outline-primary:hover {
                background: #667eea;
                color: white;
            }

            .merchant-context-section {
                border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                padding-bottom: 15px;
                margin-bottom: 15px;
            }

            .merchant-context-card {
                background: rgba(255, 255, 255, 0.1);
                border-radius: 10px;
                padding: 15px;
                margin-top: 10px;
            }

            .nav-link-small {
                display: flex;
                align-items: center;
                gap: 8px;
                color: rgba(255, 255, 255, 0.8);
                text-decoration: none;
                font-size: 12px;
                padding: 5px 0;
                transition: color 0.3s ease;
            }

            .nav-link-small:hover {
                color: white;
            }

            @media (max-width: 768px) {
                .merchant-context-container {
                    flex-direction: column;
                    align-items: stretch;
                }

                .merchant-context-actions {
                    justify-content: center;
                }
            }
        `;
        document.head.appendChild(style);
    }

    bindEvents() {
        // Bind switch merchant button
        const switchBtn = document.getElementById('switchMerchantBtn');
        if (switchBtn) {
            switchBtn.addEventListener('click', () => {
                this.showMerchantSelector();
            });
        }

        // Listen for merchant changes from session manager
        if (this.sessionManager) {
            document.addEventListener('merchantChanged', (event) => {
                this.updateMerchantContext(event.detail.merchant);
            });
        }
    }

    loadCurrentMerchant() {
        // Try to load current merchant from session
        if (this.sessionManager) {
            // SessionManager uses getCurrentSession() not getCurrentMerchant()
            const currentSession = this.sessionManager.getCurrentSession ? this.sessionManager.getCurrentSession() : null;
            if (currentSession && currentSession.merchant) {
                this.updateMerchantContext(currentSession.merchant);
            } else if (currentSession) {
                // If session itself is the merchant data
                this.updateMerchantContext(currentSession);
            }
        }

        // Try to load from URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const merchantId = urlParams.get('merchant_id');
        if (merchantId) {
            this.loadMerchantById(merchantId);
        }
    }

    updateMerchantContext(merchant) {
        this.currentMerchant = merchant;
        
        // Update header context
        const nameElement = document.getElementById('contextMerchantName');
        const statusElement = document.getElementById('contextMerchantStatus');
        
        if (nameElement && statusElement && merchant) {
            nameElement.textContent = merchant.name || 'Unknown Merchant';
            statusElement.textContent = `${merchant.portfolioType || 'Unknown'} • ${merchant.riskLevel || 'Unknown Risk'}`;
        }

        // Update sidebar context
        const sidebarNameElement = document.getElementById('sidebarMerchantName');
        const sidebarStatusElement = document.getElementById('sidebarMerchantStatus');
        
        if (sidebarNameElement && sidebarStatusElement && merchant) {
            sidebarNameElement.textContent = merchant.name || 'Unknown Merchant';
            sidebarStatusElement.textContent = `${merchant.portfolioType || 'Unknown'} • ${merchant.riskLevel || 'Unknown Risk'}`;
        }

        // Update page title if merchant is selected
        if (merchant) {
            const originalTitle = document.title;
            if (!originalTitle.includes(merchant.name)) {
                document.title = `${merchant.name} - ${originalTitle}`;
            }
        }
    }

    showMerchantSelector() {
        // Create a modal or redirect to portfolio for merchant selection
        if (confirm('Switch to merchant portfolio to select a different merchant?')) {
            window.location.href = 'merchant-portfolio.html';
        }
    }

    loadMerchantById(merchantId) {
        // Load merchant data by ID
        // This would typically make an API call
        fetch(`/api/merchants/${merchantId}`)
            .then(response => response.json())
            .then(merchant => {
                this.updateMerchantContext(merchant);
            })
            .catch(error => {
                console.error('Failed to load merchant:', error);
            });
    }

    // Method to add merchant context to existing dashboards
    static integrateWithDashboard(dashboardElement, options = {}) {
        const context = new MerchantContext(options);
        
        // Add merchant-specific data to dashboard
        if (context.currentMerchant) {
            context.addMerchantDataToDashboard(dashboardElement);
        }
        
        return context;
    }

    addMerchantDataToDashboard(dashboardElement) {
        if (!this.currentMerchant || !dashboardElement) return;

        // Add merchant-specific information to dashboard
        const merchantInfo = document.createElement('div');
        merchantInfo.className = 'merchant-dashboard-info';
        merchantInfo.innerHTML = `
            <div class="merchant-info-card">
                <h4>Current Merchant Context</h4>
                <p><strong>Name:</strong> ${this.currentMerchant.name}</p>
                <p><strong>Portfolio Type:</strong> ${this.currentMerchant.portfolioType}</p>
                <p><strong>Risk Level:</strong> ${this.currentMerchant.riskLevel}</p>
                <p><strong>Industry:</strong> ${this.currentMerchant.industry || 'Not specified'}</p>
            </div>
        `;

        // Insert at the beginning of dashboard content
        dashboardElement.insertBefore(merchantInfo, dashboardElement.firstChild);
    }

    // Method to get current merchant for use in other components
    getCurrentMerchant() {
        return this.currentMerchant;
    }

    // Method to set merchant context programmatically
    setMerchantContext(merchant) {
        this.updateMerchantContext(merchant);
        
        // Update session manager if available
        if (this.sessionManager) {
            this.sessionManager.setCurrentMerchant(merchant);
        }
    }
}

// Auto-initialize if on a dashboard page
document.addEventListener('DOMContentLoaded', () => {
    // Check if we're on a dashboard page that should have merchant context
    const isDashboardPage = document.querySelector('.dashboard, .main-content, .container, .max-w-7xl');
    const hasMerchantContext = document.querySelector('.merchant-context-container');
    
    if (isDashboardPage && !hasMerchantContext) {
        // Initialize merchant context for existing dashboards
        const context = new MerchantContext({
            showInHeader: true,
            showInSidebar: false,
            enableQuickSwitch: true
        });
        
        // Make context globally available for testing
        window.merchantContext = context;
    }
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MerchantContext;
}
