// Merchant Hub Integration Script
// This file initializes all the components for the merchant hub

document.addEventListener('DOMContentLoaded', function() {
    console.log('Merchant Hub initializing...');
    
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
        
        // Initialize portfolio type filter
        if (typeof PortfolioTypeFilter !== 'undefined') {
            const portfolioFilter = new PortfolioTypeFilter();
            portfolioFilter.initialize();
        }
        
        // Initialize risk level indicator
        if (typeof RiskLevelIndicator !== 'undefined') {
            const riskIndicator = new RiskLevelIndicator();
            riskIndicator.initialize();
        }
        
        console.log('Merchant Hub initialized successfully');
        
    } catch (error) {
        console.error('Error initializing Merchant Hub:', error);
    }
});
