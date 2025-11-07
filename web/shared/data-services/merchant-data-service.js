/**
 * Shared Merchant Data Service
 * Provides unified access to merchant data
 */

function getEventBusInstance() {
    if (typeof getEventBus !== 'undefined') {
        return getEventBus();
    }
    return {
        emit: () => {},
        on: () => () => {},
        off: () => {},
        once: () => {}
    };
}

class SharedMerchantDataService {
    constructor(config = {}) {
        this.apiConfig = config.apiConfig || (typeof APIConfig !== 'undefined' ? APIConfig : null);
        this.cache = new Map();
        this.cacheTimeout = config.cacheTimeout || 10 * 60 * 1000; // 10 minutes
        this.eventBus = config.eventBus || getEventBusInstance();
    }
    
    /**
     * Load merchant data
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Loading options
     * @returns {Promise<Object>} Merchant data
     */
    async loadMerchantData(merchantId, options = {}) {
        const {
            includeAnalytics = false,
            includeClassification = false,
            includeRisk = false
        } = options;
        
        // Check cache
        const cacheKey = `merchant_${merchantId}_${JSON.stringify(options)}`;
        const cached = this.getCachedData(cacheKey);
        if (cached) {
            console.log('📋 Using cached merchant data');
            return cached;
        }
        
        // Load base merchant data
        const merchant = await this.loadBaseMerchantData(merchantId);
        
        // Load additional data in parallel if requested
        const additionalPromises = [];
        
        if (includeAnalytics) {
            additionalPromises.push(this.loadMerchantAnalytics(merchantId));
        }
        
        if (includeClassification) {
            additionalPromises.push(this.loadMerchantClassification(merchantId));
        }
        
        if (includeRisk) {
            additionalPromises.push(this.loadMerchantRiskSummary(merchantId));
        }
        
        const additionalResults = await Promise.allSettled(additionalPromises);
        
        // Combine data
        const merchantData = {
            ...merchant,
            analytics: includeAnalytics && additionalResults[0]?.status === 'fulfilled' 
                ? additionalResults[0].value : null,
            classification: includeClassification && additionalResults[1]?.status === 'fulfilled'
                ? additionalResults[1].value : null,
            riskSummary: includeRisk && additionalResults[2]?.status === 'fulfilled'
                ? additionalResults[2].value : null
        };
        
        // Cache result
        this.cacheData(cacheKey, merchantData);
        
        // Emit event
        this.eventBus.emit('merchant-data-loaded', {
            merchantId,
            merchantData
        });
        
        return merchantData;
    }
    
    /**
     * Load base merchant data
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Base merchant data
     */
    async loadBaseMerchantData(merchantId) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantById(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to load merchant: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load merchant analytics
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Analytics data
     */
    async loadMerchantAnalytics(merchantId) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantAnalytics(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to load analytics: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load merchant classification
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Classification data
     */
    async loadMerchantClassification(merchantId) {
        // Classification is part of merchant data, extract it
        const merchant = await this.loadBaseMerchantData(merchantId);
        return merchant.classification || null;
    }
    
    /**
     * Load merchant risk summary
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Risk summary
     */
    async loadMerchantRiskSummary(merchantId) {
        // Import risk service dynamically to avoid circular dependencies
        if (typeof getRiskDataService === 'undefined') {
            // If risk service not available, return empty summary
            return {
                overallScore: 0,
                overallLevel: 'unknown',
                categoryScores: {}
            };
        }
        
        const riskService = getRiskDataService();
        const riskData = await riskService.loadRiskData(merchantId, {
            includeHistory: false,
            includePredictions: false,
            includeBenchmarks: false
        });
        
        return {
            overallScore: riskData.current?.overallScore || 0,
            overallLevel: riskData.current?.overallLevel || 'unknown',
            categoryScores: riskData.current?.categoryScores || {}
        };
    }
    
    /**
     * Search merchants
     * @param {Object} searchParams - Search parameters
     * @returns {Promise<Object>} Search results
     */
    async searchMerchants(searchParams) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantSearch, {
            method: 'POST',
            headers: this.apiConfig.getHeaders(),
            body: JSON.stringify(searchParams)
        });
        
        if (!response.ok) {
            throw new Error(`Search failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Get cached data
     * @param {string} key - Cache key
     * @returns {Object|null} Cached data or null
     */
    getCachedData(key) {
        const cached = this.cache.get(key);
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.data;
        }
        if (cached) {
            this.cache.delete(key);
        }
        return null;
    }
    
    /**
     * Cache data
     * @param {string} key - Cache key
     * @param {Object} data - Data to cache
     */
    cacheData(key, data) {
        this.cache.set(key, {
            data,
            timestamp: Date.now()
        });
    }
    
    /**
     * Clear cache for a merchant
     * @param {string} merchantId - Merchant ID
     */
    clearCache(merchantId) {
        const keysToDelete = [];
        for (const key of this.cache.keys()) {
            if (key.includes(merchantId)) {
                keysToDelete.push(key);
            }
        }
        keysToDelete.forEach(key => this.cache.delete(key));
        
        this.eventBus.emit('merchant-data-cache-cleared', { merchantId });
    }
}

// Export singleton instance
let merchantDataServiceInstance = null;

export function getMerchantDataService(config) {
    if (!merchantDataServiceInstance) {
        merchantDataServiceInstance = new SharedMerchantDataService(config);
    }
    return merchantDataServiceInstance;
}

export { SharedMerchantDataService };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getMerchantDataService = getMerchantDataService;
    window.SharedMerchantDataService = SharedMerchantDataService;
}

