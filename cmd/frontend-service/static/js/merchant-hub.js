// Merchant Hub Class
// This class manages the merchant hub functionality

class MerchantHub {
    constructor() {
        this.merchants = [];
        this.currentMerchant = null;
        this.apiBaseUrl = 'https://api-gateway-service-production-21fd.up.railway.app/api/v1';
        this.init();
    }

    init() {
        console.log('Merchant Hub initializing...');
        this.loadMerchants();
        this.initializeComponents();
        this.bindEvents();
    }

    async loadMerchants() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/merchants`);
            if (response.ok) {
                const data = await response.json();
                this.merchants = data.merchants || [];
                this.updateMerchantStats();
                this.renderMerchants();
            } else {
                console.error('Failed to load merchants:', response.statusText);
                this.showError('Failed to load merchants');
            }
        } catch (error) {
            console.error('Error loading merchants:', error);
            this.showError('Error loading merchants');
        }
    }

    updateMerchantStats() {
        const totalMerchants = document.getElementById('totalMerchants');
        const activeMerchants = document.getElementById('activeMerchants');
        const riskAlerts = document.getElementById('riskAlerts');

        if (totalMerchants) totalMerchants.textContent = this.merchants.length;
        if (activeMerchants) activeMerchants.textContent = this.merchants.filter(m => m.status === 'active').length;
        if (riskAlerts) riskAlerts.textContent = this.merchants.filter(m => m.riskLevel === 'high').length;
    }

    renderMerchants() {
        const merchantsGrid = document.getElementById('merchantsGrid');
        if (!merchantsGrid) return;

        if (this.merchants.length === 0) {
            merchantsGrid.innerHTML = `
                <div class="no-merchants">
                    <i class="fas fa-store"></i>
                    <h3>No Merchants Found</h3>
                    <p>No merchants are currently available.</p>
                </div>
            `;
            return;
        }

        merchantsGrid.innerHTML = this.merchants.map(merchant => `
            <div class="merchant-card" data-merchant-id="${merchant.id}">
                <div class="merchant-header">
                    <div class="merchant-avatar">
                        ${merchant.name.charAt(0).toUpperCase()}
                    </div>
                    <div class="merchant-info">
                        <h3>${merchant.name}</h3>
                        <p>${merchant.industry || 'N/A'}</p>
                    </div>
                    <div class="merchant-status">
                        <span class="status-badge ${merchant.status}">${merchant.status}</span>
                    </div>
                </div>
                <div class="merchant-details">
                    <div class="detail-item">
                        <i class="fas fa-map-marker-alt"></i>
                        <span>${merchant.address || 'N/A'}</span>
                    </div>
                    <div class="detail-item">
                        <i class="fas fa-phone"></i>
                        <span>${merchant.phone || 'N/A'}</span>
                    </div>
                    <div class="detail-item">
                        <i class="fas fa-envelope"></i>
                        <span>${merchant.email || 'N/A'}</span>
                    </div>
                </div>
                <div class="merchant-actions">
                    <button class="btn btn-primary" onclick="merchantHub.viewMerchant('${merchant.id}')">
                        <i class="fas fa-eye"></i> View
                    </button>
                    <button class="btn btn-secondary" onclick="merchantHub.editMerchant('${merchant.id}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                </div>
            </div>
        `).join('');
    }

    initializeComponents() {
        // Initialize components if they exist
        try {
            // Initialize session manager
            if (typeof SessionManager !== 'undefined') {
                const sessionManager = new SessionManager();
                sessionManager.initialize();
            }
            
            // Initialize merchant search
            if (typeof MerchantSearch !== 'undefined') {
                const merchantSearch = new MerchantSearch();
                merchantSearch.initialize();
            }
            
            // Initialize coming soon banner
            if (typeof ComingSoonBanner !== 'undefined') {
                const comingSoonBanner = new ComingSoonBanner();
                comingSoonBanner.initialize();
            }
            
            // Initialize mock data warning
            if (typeof MockDataWarning !== 'undefined') {
                const mockDataWarning = new MockDataWarning();
                mockDataWarning.initialize();
            }
            
            console.log('Merchant Hub components initialized successfully');
            
        } catch (error) {
            console.error('Error initializing Merchant Hub components:', error);
        }
    }

    bindEvents() {
        // Bind any additional events here
        console.log('Merchant Hub events bound');
    }

    viewMerchant(merchantId) {
        const merchant = this.merchants.find(m => m.id === merchantId);
        if (merchant) {
            console.log('Viewing merchant:', merchant.name);
            // Navigate to merchant detail page or show modal
            alert(`Viewing merchant: ${merchant.name}`);
        }
    }

    editMerchant(merchantId) {
        const merchant = this.merchants.find(m => m.id === merchantId);
        if (merchant) {
            console.log('Editing merchant:', merchant.name);
            // Navigate to edit page or show edit modal
            alert(`Editing merchant: ${merchant.name}`);
        }
    }

    showError(message) {
        console.error(message);
        // You could show a toast notification or error modal here
    }
}

// Initialize when DOM is loaded
let merchantHub;
document.addEventListener('DOMContentLoaded', () => {
    merchantHub = new MerchantHub();
});
