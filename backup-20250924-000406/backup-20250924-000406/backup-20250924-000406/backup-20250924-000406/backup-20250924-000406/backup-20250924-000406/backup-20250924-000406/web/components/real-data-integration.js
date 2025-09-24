/**
 * Real Data Integration Component
 * Replaces mock data usage with real Supabase API calls
 * Integrates with all UI components for comprehensive data display
 */

class RealDataIntegration {
    constructor() {
        this.apiBaseUrl = '/api/v1';
        this.classificationUrl = '/v1/classify';
        this.cache = new Map();
        this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
        this.retryAttempts = 3;
        this.retryDelay = 1000; // 1 second
    }

    /**
     * Get merchants data from Supabase
     */
    async getMerchants(filters = {}) {
        const cacheKey = `merchants_${JSON.stringify(filters)}`;
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch merchants:', error);
            throw new Error('Unable to load merchants data');
        }
    }

    /**
     * Get merchant details by ID
     */
    async getMerchantById(merchantId) {
        const cacheKey = `merchant_${merchantId}`;
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/${merchantId}`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch merchant details:', error);
            throw new Error('Unable to load merchant details');
        }
    }

    /**
     * Get merchant analytics data
     */
    async getMerchantAnalytics() {
        const cacheKey = 'merchant_analytics';
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/analytics`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch merchant analytics:', error);
            throw new Error('Unable to load analytics data');
        }
    }

    /**
     * Get merchant statistics
     */
    async getMerchantStatistics() {
        const cacheKey = 'merchant_statistics';
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/statistics`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch merchant statistics:', error);
            throw new Error('Unable to load statistics data');
        }
    }

    /**
     * Get portfolio types
     */
    async getPortfolioTypes() {
        const cacheKey = 'portfolio_types';
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/portfolio-types`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch portfolio types:', error);
            throw new Error('Unable to load portfolio types');
        }
    }

    /**
     * Get risk levels
     */
    async getRiskLevels() {
        const cacheKey = 'risk_levels';
        
        // Check cache first
        if (this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.data;
            }
        }

        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/risk-levels`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            
            // Cache the result
            this.cache.set(cacheKey, {
                data: data,
                timestamp: Date.now()
            });

            return data;
        } catch (error) {
            console.error('Failed to fetch risk levels:', error);
            throw new Error('Unable to load risk levels');
        }
    }

    /**
     * Perform business classification
     */
    async classifyBusiness(businessName, description, websiteUrl = '') {
        try {
            const response = await this.makeRequest(this.classificationUrl, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    business_name: businessName,
                    description: description,
                    website_url: websiteUrl
                })
            });

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to classify business:', error);
            throw new Error('Unable to classify business');
        }
    }

    /**
     * Search merchants with filters
     */
    async searchMerchants(searchParams) {
        try {
            const response = await this.makeRequest(`${this.apiBaseUrl}/merchants/search`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(searchParams)
            });

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to search merchants:', error);
            throw new Error('Unable to search merchants');
        }
    }

    /**
     * Get system health status
     */
    async getSystemHealth() {
        try {
            const response = await this.makeRequest('/health', {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Failed to get system health:', error);
            throw new Error('Unable to get system health');
        }
    }

    /**
     * Make HTTP request with retry logic
     */
    async makeRequest(url, options = {}) {
        let lastError;
        
        for (let attempt = 1; attempt <= this.retryAttempts; attempt++) {
            try {
                const response = await fetch(url, options);
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                return response;
            } catch (error) {
                lastError = error;
                console.warn(`Request attempt ${attempt} failed:`, error.message);
                
                if (attempt < this.retryAttempts) {
                    await this.delay(this.retryDelay * attempt);
                }
            }
        }
        
        throw lastError;
    }

    /**
     * Delay utility for retry logic
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * Clear cache
     */
    clearCache() {
        this.cache.clear();
    }

    /**
     * Clear specific cache entry
     */
    clearCacheEntry(key) {
        this.cache.delete(key);
    }

    /**
     * Get cache statistics
     */
    getCacheStats() {
        return {
            size: this.cache.size,
            entries: Array.from(this.cache.keys())
        };
    }
}

// Export for use in other components
window.RealDataIntegration = RealDataIntegration;
