/**
 * Lazy Loading Utility
 * Provides efficient lazy loading for merchant portfolio and dashboard components
 * Implements intersection observer for performance optimization
 */

class LazyLoader {
    constructor(options = {}) {
        this.options = {
            root: null,
            rootMargin: '50px',
            threshold: 0.1,
            loadingClass: 'lazy-loading',
            loadedClass: 'lazy-loaded',
            errorClass: 'lazy-error',
            ...options
        };
        
        this.observer = null;
        this.elements = new Map();
        this.loadingPromises = new Map();
        
        this.init();
    }

    /**
     * Initialize the lazy loader
     */
    init() {
        if ('IntersectionObserver' in window) {
            this.observer = new IntersectionObserver(
                this.handleIntersection.bind(this),
                this.options
            );
        } else {
            // Fallback for browsers without IntersectionObserver
            this.loadAllElements();
        }
    }

    /**
     * Register an element for lazy loading
     * @param {HTMLElement} element - Element to lazy load
     * @param {Function} loader - Function that returns a Promise for loading content
     * @param {Object} options - Element-specific options
     */
    register(element, loader, options = {}) {
        if (!element || typeof loader !== 'function') {
            console.warn('LazyLoader: Invalid element or loader function');
            return;
        }

        const elementOptions = {
            loader,
            loaded: false,
            loading: false,
            error: false,
            ...options
        };

        this.elements.set(element, elementOptions);
        
        // Add loading class
        element.classList.add(this.options.loadingClass);
        
        // Observe element if IntersectionObserver is available
        if (this.observer) {
            this.observer.observe(element);
        } else {
            // Fallback: load immediately
            this.loadElement(element);
        }
    }

    /**
     * Handle intersection observer callback
     * @param {IntersectionObserverEntry[]} entries
     */
    handleIntersection(entries) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                this.loadElement(entry.target);
                this.observer.unobserve(entry.target);
            }
        });
    }

    /**
     * Load a specific element
     * @param {HTMLElement} element
     */
    async loadElement(element) {
        const elementOptions = this.elements.get(element);
        if (!elementOptions || elementOptions.loading || elementOptions.loaded) {
            return;
        }

        elementOptions.loading = true;
        element.classList.remove(this.options.loadingClass);
        element.classList.add('lazy-loading-active');

        try {
            // Create loading promise if it doesn't exist
            if (!this.loadingPromises.has(element)) {
                const loadingPromise = elementOptions.loader();
                this.loadingPromises.set(element, loadingPromise);
            }

            const result = await this.loadingPromises.get(element);
            
            // Mark as loaded
            elementOptions.loaded = true;
            elementOptions.loading = false;
            element.classList.remove('lazy-loading-active');
            element.classList.add(this.options.loadedClass);
            
            // Clean up
            this.loadingPromises.delete(element);
            
            // Trigger loaded event
            element.dispatchEvent(new CustomEvent('lazyLoaded', {
                detail: { element, result }
            }));

        } catch (error) {
            console.error('LazyLoader: Failed to load element', error);
            elementOptions.error = true;
            elementOptions.loading = false;
            element.classList.remove('lazy-loading-active');
            element.classList.add(this.options.errorClass);
            
            // Trigger error event
            element.dispatchEvent(new CustomEvent('lazyError', {
                detail: { element, error }
            }));
        }
    }

    /**
     * Load all elements (fallback for browsers without IntersectionObserver)
     */
    loadAllElements() {
        this.elements.forEach((options, element) => {
            this.loadElement(element);
        });
    }

    /**
     * Preload an element
     * @param {HTMLElement} element
     */
    preload(element) {
        if (this.elements.has(element)) {
            this.loadElement(element);
        }
    }

    /**
     * Unregister an element
     * @param {HTMLElement} element
     */
    unregister(element) {
        if (this.observer) {
            this.observer.unobserve(element);
        }
        this.elements.delete(element);
        this.loadingPromises.delete(element);
    }

    /**
     * Destroy the lazy loader
     */
    destroy() {
        if (this.observer) {
            this.observer.disconnect();
        }
        this.elements.clear();
        this.loadingPromises.clear();
    }

    /**
     * Get loading statistics
     */
    getStats() {
        const total = this.elements.size;
        const loaded = Array.from(this.elements.values()).filter(opt => opt.loaded).length;
        const loading = Array.from(this.elements.values()).filter(opt => opt.loading).length;
        const errors = Array.from(this.elements.values()).filter(opt => opt.error).length;
        
        return {
            total,
            loaded,
            loading,
            errors,
            pending: total - loaded - loading - errors
        };
    }
}

/**
 * Lazy Loading Component Manager
 * Manages lazy loading for specific merchant components
 */
class MerchantLazyLoader {
    constructor() {
        this.lazyLoader = new LazyLoader({
            rootMargin: '100px', // Load earlier for better UX
            threshold: 0.05
        });
        
        this.componentLoaders = new Map();
        this.init();
    }

    /**
     * Initialize component loaders
     */
    init() {
        // Register component loaders
        this.componentLoaders.set('merchant-card', this.loadMerchantCard.bind(this));
        this.componentLoaders.set('merchant-chart', this.loadMerchantChart.bind(this));
        this.componentLoaders.set('merchant-details', this.loadMerchantDetails.bind(this));
        this.componentLoaders.set('bulk-operations', this.loadBulkOperations.bind(this));
        this.componentLoaders.set('comparison-view', this.loadComparisonView.bind(this));
    }

    /**
     * Register a merchant card for lazy loading
     * @param {HTMLElement} element
     * @param {Object} merchantData
     */
    registerMerchantCard(element, merchantData) {
        this.lazyLoader.register(element, () => this.componentLoaders.get('merchant-card')(merchantData));
    }

    /**
     * Register a merchant chart for lazy loading
     * @param {HTMLElement} element
     * @param {Object} chartConfig
     */
    registerMerchantChart(element, chartConfig) {
        this.lazyLoader.register(element, () => this.componentLoaders.get('merchant-chart')(chartConfig));
    }

    /**
     * Register merchant details for lazy loading
     * @param {HTMLElement} element
     * @param {string} merchantId
     */
    registerMerchantDetails(element, merchantId) {
        this.lazyLoader.register(element, () => this.componentLoaders.get('merchant-details')(merchantId));
    }

    /**
     * Register bulk operations for lazy loading
     * @param {HTMLElement} element
     * @param {Array} selectedMerchants
     */
    registerBulkOperations(element, selectedMerchants) {
        this.lazyLoader.register(element, () => this.componentLoaders.get('bulk-operations')(selectedMerchants));
    }

    /**
     * Register comparison view for lazy loading
     * @param {HTMLElement} element
     * @param {Array} merchants
     */
    registerComparisonView(element, merchants) {
        this.lazyLoader.register(element, () => this.componentLoaders.get('comparison-view')(merchants));
    }

    /**
     * Load merchant card component
     * @param {Object} merchantData
     */
    async loadMerchantCard(merchantData) {
        // Simulate loading time for demonstration
        await new Promise(resolve => setTimeout(resolve, 100));
        
        return {
            type: 'merchant-card',
            data: merchantData,
            template: this.generateMerchantCardTemplate(merchantData)
        };
    }

    /**
     * Load merchant chart component
     * @param {Object} chartConfig
     */
    async loadMerchantChart(chartConfig) {
        // Simulate chart loading
        await new Promise(resolve => setTimeout(resolve, 200));
        
        return {
            type: 'merchant-chart',
            config: chartConfig,
            chart: this.createChart(chartConfig)
        };
    }

    /**
     * Load merchant details component
     * @param {string} merchantId
     */
    async loadMerchantDetails(merchantId) {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 150));
        
        return {
            type: 'merchant-details',
            merchantId,
            details: await this.fetchMerchantDetails(merchantId)
        };
    }

    /**
     * Load bulk operations component
     * @param {Array} selectedMerchants
     */
    async loadBulkOperations(selectedMerchants) {
        // Simulate bulk operations setup
        await new Promise(resolve => setTimeout(resolve, 100));
        
        return {
            type: 'bulk-operations',
            selectedMerchants,
            operations: this.getAvailableBulkOperations(selectedMerchants)
        };
    }

    /**
     * Load comparison view component
     * @param {Array} merchants
     */
    async loadComparisonView(merchants) {
        // Simulate comparison data loading
        await new Promise(resolve => setTimeout(resolve, 250));
        
        return {
            type: 'comparison-view',
            merchants,
            comparison: this.generateComparisonData(merchants)
        };
    }

    /**
     * Generate merchant card template
     * @param {Object} merchantData
     */
    generateMerchantCardTemplate(merchantData) {
        return `
            <div class="merchant-card" data-merchant-id="${merchantData.id}">
                <div class="merchant-header">
                    <h3>${merchantData.name}</h3>
                    <span class="risk-level ${merchantData.riskLevel}">${merchantData.riskLevel}</span>
                </div>
                <div class="merchant-info">
                    <p><strong>Industry:</strong> ${merchantData.industry}</p>
                    <p><strong>Portfolio Type:</strong> ${merchantData.portfolioType}</p>
                    <p><strong>Location:</strong> ${merchantData.location}</p>
                </div>
            </div>
        `;
    }

    /**
     * Create chart (placeholder)
     * @param {Object} chartConfig
     */
    createChart(chartConfig) {
        // Placeholder for chart creation
        return {
            type: chartConfig.type,
            data: chartConfig.data,
            options: chartConfig.options
        };
    }

    /**
     * Fetch merchant details (placeholder)
     * @param {string} merchantId
     */
    async fetchMerchantDetails(merchantId) {
        // Placeholder for API call
        return {
            id: merchantId,
            name: `Merchant ${merchantId}`,
            details: 'Detailed merchant information...'
        };
    }

    /**
     * Get available bulk operations
     * @param {Array} selectedMerchants
     */
    getAvailableBulkOperations(selectedMerchants) {
        return [
            'update-portfolio-type',
            'update-risk-level',
            'export-data',
            'send-notifications'
        ];
    }

    /**
     * Generate comparison data
     * @param {Array} merchants
     */
    generateComparisonData(merchants) {
        return {
            merchants: merchants,
            metrics: ['risk-level', 'portfolio-type', 'industry', 'location'],
            comparison: merchants.map(m => ({
                id: m.id,
                name: m.name,
                metrics: {
                    'risk-level': m.riskLevel,
                    'portfolio-type': m.portfolioType,
                    'industry': m.industry,
                    'location': m.location
                }
            }))
        };
    }

    /**
     * Get loading statistics
     */
    getStats() {
        return this.lazyLoader.getStats();
    }

    /**
     * Destroy the lazy loader
     */
    destroy() {
        this.lazyLoader.destroy();
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { LazyLoader, MerchantLazyLoader };
}
