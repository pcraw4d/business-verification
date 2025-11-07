/**
 * Shared Components Loader
 * 
 * Loads shared components and makes them available globally
 * Works with both module and non-module environments
 */

(function() {
    'use strict';
    
    // Check if we're in a module environment
    const isModule = typeof window !== 'undefined' && window.location.protocol === 'file:' ? false : true;
    
    // Load shared components
    async function loadSharedComponents() {
        try {
            // Try to load as modules first
            if (isModule && typeof import !== 'undefined') {
                const { getEventBus } = await import('./events/event-bus.js');
                const { getRiskDataService } = await import('./data-services/risk-data-service.js');
                const { getMerchantDataService } = await import('./data-services/merchant-data-service.js');
                const { getRiskVisualizations } = await import('./visualizations/risk-visualizations.js');
                const { getCrossTabNavigation } = await import('./navigation/cross-tab-navigation.js');
                
                // Make available globally
                window.getEventBus = getEventBus;
                window.getRiskDataService = getRiskDataService;
                window.getMerchantDataService = getMerchantDataService;
                window.getRiskVisualizations = getRiskVisualizations;
                window.getCrossTabNavigation = getCrossTabNavigation;
                
                console.log('✅ Shared component library loaded (modules)');
                return true;
            } else {
                // Fallback: Components should be loaded via script tags
                // They will register themselves globally
                console.log('📦 Shared components should be loaded via script tags');
                return false;
            }
        } catch (error) {
            console.warn('⚠️ Failed to load shared components as modules, using fallback:', error);
            return false;
        }
    }
    
    // Auto-load when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', loadSharedComponents);
    } else {
        loadSharedComponents();
    }
    
    // Also expose load function for manual loading
    window.loadSharedComponents = loadSharedComponents;
})();

