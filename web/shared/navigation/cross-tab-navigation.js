/**
 * Cross-Tab Navigation
 * Provides smart navigation between tabs/pages with context
 */

class CrossTabNavigation {
    constructor() {
        this.routes = new Map();
        this.initializeRoutes();
    }
    
    /**
     * Initialize route mappings
     */
    initializeRoutes() {
        // Merchant Details page routes
        this.routes.set('merchant-details', {
            base: 'merchant-details.html',
            tabs: {
                'merchant-details': '#merchant-details-tab',
                'business-analytics': '#business-analytics-tab',
                'risk-assessment': '#risk-assessment-tab',
                'risk-indicators': '#risk-indicators-tab'
            }
        });
        
        // Compliance Dashboard routes
        this.routes.set('compliance-dashboard', {
            base: 'compliance-dashboard.html',
            sections: {
                'overview': '#compliance-overview',
                'gap-analysis': '#gap-analysis',
                'progress-tracking': '#progress-tracking'
            }
        });
        
        // Market Intelligence routes
        this.routes.set('market-analysis', {
            base: 'market-analysis-dashboard.html',
            sections: {
                'overview': '#market-overview',
                'competitive': '#competitive-analysis'
            }
        });
    }
    
    /**
     * Navigate to a specific tab/page
     * @param {string} target - Target page/tab
     * @param {Object} context - Navigation context (merchantId, filters, etc.)
     * @returns {boolean} Success
     */
    navigateTo(target, context = {}) {
        const { page, tab, section, merchantId, ...params } = context;
        
        // Build URL
        const url = this.buildNavigationURL(target, { page, tab, section, merchantId, ...params });
        
        if (url) {
            window.location.href = url;
            return true;
        }
        
        // Try to activate tab on current page
        if (this.activateTabOnPageLoad(target, context)) {
            return true;
        }
        
        return false;
    }
    
    /**
     * Get contextual links based on current context
     * @param {Object} context - Current context
     * @returns {Array} Array of contextual links
     */
    getContextualLinks(context = {}) {
        const links = [];
        const { merchantId, riskLevel, category, complianceStatus } = context;
        
        // Risk Indicators contextual links
        if (merchantId) {
            // Link to Risk Assessment for detailed analysis
            if (riskLevel === 'high' || riskLevel === 'critical') {
                links.push({
                    label: 'View Detailed Risk Analysis',
                    target: 'risk-assessment',
                    page: 'merchant-details',
                    tab: 'risk-assessment',
                    merchantId,
                    icon: 'chart-line',
                    description: 'See comprehensive risk assessment with trend analysis and SHAP explanations'
                });
            }
            
            // Link to Business Analytics for industry context
            links.push({
                label: 'See Industry Classification',
                target: 'business-analytics',
                page: 'merchant-details',
                tab: 'business-analytics',
                merchantId,
                icon: 'chart-bar',
                description: 'View MCC, NAICS, and SIC industry codes'
            });
            
            // Link to Compliance Dashboard if regulatory risk is high
            if (category === 'regulatory' && (riskLevel === 'high' || riskLevel === 'critical')) {
                links.push({
                    label: 'View Compliance Status',
                    target: 'compliance-dashboard',
                    merchantId,
                    icon: 'shield-alt',
                    description: 'Check compliance framework status and gaps'
                });
            }
        }
        
        // Risk Assessment contextual links
        if (merchantId) {
            links.push({
                label: 'View Current Risk Status',
                target: 'risk-indicators',
                page: 'merchant-details',
                tab: 'risk-indicators',
                merchantId,
                icon: 'tachometer-alt',
                description: 'See current risk indicators and alerts'
            });
        }
        
        return links;
    }
    
    /**
     * Build navigation URL
     * @param {string} target - Target identifier
     * @param {Object} params - URL parameters
     * @returns {string|null} Navigation URL
     */
    buildNavigationURL(target, params = {}) {
        const { page, tab, section, merchantId, ...queryParams } = params;
        
        // If target is a page name, navigate to that page
        const route = this.routes.get(target);
        if (route) {
            let url = route.base;
            
            // Add tab/section hash
            if (tab && route.tabs && route.tabs[tab]) {
                url += route.tabs[tab];
            } else if (section && route.sections && route.sections[section]) {
                url += route.sections[section];
            }
            
            // Add query parameters
            const query = new URLSearchParams();
            if (merchantId) {
                query.set('merchantId', merchantId);
            }
            
            Object.entries(queryParams).forEach(([key, value]) => {
                if (value !== null && value !== undefined) {
                    query.set(key, value);
                }
            });
            
            if (query.toString()) {
                url += `?${query.toString()}`;
            }
            
            return url;
        }
        
        // If target is a tab on current page, return hash
        if (tab) {
            return `#${tab}`;
        }
        
        return null;
    }
    
    /**
     * Activate tab on page load
     * @param {string} target - Target tab
     * @param {Object} context - Context
     * @returns {boolean} Success
     */
    activateTabOnPageLoad(target, context = {}) {
        const { tab, merchantId } = context;
        
        // Check if we're on the merchant details page
        if (window.location.pathname.includes('merchant-details')) {
            // Try to find and activate the tab
            const tabElement = document.querySelector(`[data-tab="${target}"]`) || 
                             document.querySelector(`#${target}-tab`) ||
                             document.querySelector(`[href="#${target}"]`);
            
            if (tabElement) {
                // Trigger click or activate tab
                if (tabElement.click) {
                    tabElement.click();
                } else if (tabElement.dispatchEvent) {
                    tabElement.dispatchEvent(new Event('click', { bubbles: true }));
                }
                
                // If merchantId is provided, initialize the tab
                if (merchantId) {
                    // Wait for tab to be visible, then initialize
                    setTimeout(() => {
                        const tabContent = document.querySelector(`#${target}`);
                        if (tabContent) {
                            // Try to find and call init method
                            const tabController = window[`${target}Controller`] || 
                                                window[`${target.charAt(0).toUpperCase() + target.slice(1)}Controller`];
                            if (tabController && tabController.init) {
                                tabController.init(merchantId);
                            }
                        }
                    }, 100);
                }
                
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * Create contextual link HTML
     * @param {Object} link - Link object
     * @returns {string} HTML string
     */
    createContextualLinkHTML(link) {
        const url = this.buildNavigationURL(link.target, {
            page: link.page,
            tab: link.tab,
            merchantId: link.merchantId
        });
        
        return `
            <a href="${url || '#'}" 
               class="contextual-link flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors"
               onclick="event.preventDefault(); window.crossTabNavigation?.navigateTo('${link.target}', { page: '${link.page || ''}', tab: '${link.tab || ''}', merchantId: '${link.merchantId || ''}' }); return false;">
                <i class="fas fa-${link.icon || 'arrow-right'} mr-3 text-blue-600"></i>
                <div class="flex-1">
                    <div class="font-medium text-gray-900">${link.label}</div>
                    ${link.description ? `<div class="text-sm text-gray-600">${link.description}</div>` : ''}
                </div>
                <i class="fas fa-chevron-right text-gray-400"></i>
            </a>
        `;
    }
    
    /**
     * Render contextual links section
     * @param {string} containerId - Container element ID
     * @param {Object} context - Context for generating links
     */
    renderContextualLinks(containerId, context = {}) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.warn(`Container not found: ${containerId}`);
            return;
        }
        
        const links = this.getContextualLinks(context);
        
        if (links.length === 0) {
            container.innerHTML = '';
            return;
        }
        
        const html = `
            <div class="contextual-links-section bg-white rounded-lg shadow-lg p-6 mb-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-4">Related Information</h3>
                <div class="space-y-2">
                    ${links.map(link => this.createContextualLinkHTML(link)).join('')}
                </div>
            </div>
        `;
        
        container.innerHTML = html;
    }
}

// Export singleton instance
let crossTabNavigationInstance = null;

export function getCrossTabNavigation() {
    if (!crossTabNavigationInstance) {
        crossTabNavigationInstance = new CrossTabNavigation();
    }
    return crossTabNavigationInstance;
}

// Make available globally for non-module environments and onclick handlers
if (typeof window !== 'undefined') {
    window.getCrossTabNavigation = getCrossTabNavigation;
    window.crossTabNavigation = getCrossTabNavigation();
    window.CrossTabNavigation = CrossTabNavigation;
}

export { CrossTabNavigation };

