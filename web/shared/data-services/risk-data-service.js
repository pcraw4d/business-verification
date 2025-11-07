/**
 * Shared Risk Data Service
 * Provides unified access to risk data for all pages
 * 
 * This service consolidates risk data loading logic and eliminates duplication
 * across Risk Indicators, Risk Assessment, and other risk-related pages.
 */

// Import EventBus (will be available globally or via module system)
// For now, we'll use a getter that works with both module and global scope
function getEventBusInstance() {
    if (typeof getEventBus !== 'undefined') {
        return getEventBus();
    }
    // Fallback: create a simple event emitter if EventBus not available
    return {
        emit: () => {},
        on: () => () => {},
        off: () => {},
        once: () => {}
    };
}

class SharedRiskDataService {
    constructor(config = {}) {
        // Use provided config or fall back to global instances
        this.apiConfig = config.apiConfig || (typeof APIConfig !== 'undefined' ? APIConfig : null);
        this.cache = new Map();
        this.cacheTimeout = config.cacheTimeout || 5 * 60 * 1000; // 5 minutes
        this.eventBus = config.eventBus || getEventBusInstance();
        
        // Helpers for risk calculations (will be injected or use global)
        this.helpers = config.helpers || (typeof RiskIndicatorsHelpers !== 'undefined' ? RiskIndicatorsHelpers : null);
    }
    
    /**
     * Load comprehensive risk data for a merchant
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Loading options
     * @returns {Promise<Object>} Complete risk data
     */
    async loadRiskData(merchantId, options = {}) {
        const {
            includeHistory = false,
            includePredictions = false,
            includeBenchmarks = false,
            includeExplanations = false,
            includeRecommendations = false
        } = options;
        
        // Check cache
        const cacheKey = this.getCacheKey(merchantId, options);
        const cached = this.getCachedData(cacheKey);
        if (cached) {
            console.log('📋 Using cached risk data');
            return cached;
        }
        
        // Load data in parallel
        const dataPromises = [
            this.loadCurrentRiskAssessment(merchantId, {
                includeExplanations,
                includeRecommendations
            })
        ];
        
        if (includeHistory) {
            dataPromises.push(this.loadRiskHistory(merchantId));
        }
        
        if (includePredictions) {
            dataPromises.push(this.loadRiskPredictions(merchantId));
        }
        
        if (includeBenchmarks) {
            dataPromises.push(this.loadIndustryBenchmarks(merchantId));
        }
        
        const results = await Promise.allSettled(dataPromises);
        
        // Combine results
        const riskData = this.combineRiskData(results, merchantId);
        
        // Cache result
        this.cacheData(cacheKey, riskData);
        
        // Emit event
        this.eventBus.emit('risk-data-loaded', {
            merchantId,
            riskData
        });
        
        return riskData;
    }
    
    /**
     * Load current risk assessment
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Assessment options
     * @returns {Promise<Object>} Risk assessment data
     */
    async loadCurrentRiskAssessment(merchantId, options = {}) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.riskAssess, {
            method: 'POST',
            headers: this.apiConfig.getHeaders(),
            body: JSON.stringify({
                merchantId,
                includeTrendAnalysis: true,
                includeRecommendations: options.includeRecommendations || false,
                includeExplanations: options.includeExplanations || false
            })
        });
        
        if (!response.ok) {
            throw new Error(`Risk assessment failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load risk history
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Time range options
     * @returns {Promise<Object>} Risk history data
     */
    async loadRiskHistory(merchantId, options = {}) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const {
            timeRange = '6months',
            granularity = 'daily'
        } = options;
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(
            `${endpoints.riskHistory(merchantId)}?timeRange=${timeRange}&granularity=${granularity}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            // If endpoint doesn't exist, return empty history
            if (response.status === 404) {
                console.warn('Risk history endpoint not available, returning empty history');
                return {
                    merchantId,
                    timeRange,
                    dataPoints: [],
                    trends: []
                };
            }
            throw new Error(`Risk history failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load risk predictions
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Prediction options
     * @returns {Promise<Object>} Risk prediction data
     */
    async loadRiskPredictions(merchantId, options = {}) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const {
            horizons = [3, 6, 12], // months
            includeScenarios = true,
            includeConfidence = true
        } = options;
        
        const endpoints = this.apiConfig.getEndpoints();
        const url = `${endpoints.riskPredictions(merchantId)}?horizons=${horizons.join(',')}&includeScenarios=${includeScenarios}&includeConfidence=${includeConfidence}`;
        
        const response = await fetch(url, {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            // If endpoint doesn't exist, return empty predictions
            if (response.status === 404) {
                console.warn('Risk predictions endpoint not available, returning empty predictions');
                return {
                    merchantId,
                    horizons,
                    predictions: [],
                    scenarios: null,
                    confidence: null,
                    drivers: []
                };
            }
            throw new Error(`Risk predictions failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Load industry benchmarks
     * @param {string} merchantId - Merchant ID
     * @param {Object} industryCodes - Industry classification codes (optional, will fetch if not provided)
     * @returns {Promise<Object>} Industry benchmark data
     */
    async loadIndustryBenchmarks(merchantId, industryCodes = null) {
        // If industry codes not provided, fetch from Business Analytics
        if (!industryCodes) {
            industryCodes = await this.getIndustryCodesFromAnalytics(merchantId);
        }
        
        // If no industry codes found, return empty benchmarks
        if (!industryCodes.mcc && !industryCodes.naics && !industryCodes.sic) {
            console.warn('No industry codes found, returning empty benchmarks');
            return {
                industryCodes: {},
                industryName: 'Unknown',
                averages: {},
                percentiles: {
                    p10: {},
                    p25: {},
                    p50: {},
                    p75: {},
                    p90: {}
                },
                sampleSize: 0
            };
        }
        
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const params = new URLSearchParams({
            mcc: industryCodes.mcc || '',
            naics: industryCodes.naics || '',
            sic: industryCodes.sic || ''
        });
        
        // Check if benchmarks endpoint exists, if not use fallback
        const benchmarksUrl = endpoints.riskBenchmarks || `${this.apiConfig.getBaseURL()}/api/v1/risk/benchmarks`;
        const response = await fetch(
            `${benchmarksUrl}?${params.toString()}`,
            {
                headers: this.apiConfig.getHeaders()
            }
        );
        
        if (!response.ok) {
            // If endpoint doesn't exist, return mock benchmarks (will be replaced when backend is ready)
            if (response.status === 404) {
                console.warn('Risk benchmarks endpoint not available, using fallback benchmarks');
                return this.getFallbackBenchmarks(industryCodes);
            }
            throw new Error(`Industry benchmarks failed: ${response.status}`);
        }
        
        return await response.json();
    }
    
    /**
     * Get industry codes from Business Analytics (merchant classification data)
     * @param {string} merchantId - Merchant ID
     * @returns {Promise<Object>} Industry codes {mcc, naics, sic}
     */
    async getIndustryCodesFromAnalytics(merchantId) {
        if (!this.apiConfig) {
            throw new Error('APIConfig not available');
        }
        
        const endpoints = this.apiConfig.getEndpoints();
        const response = await fetch(endpoints.merchantById(merchantId), {
            headers: this.apiConfig.getHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`Failed to fetch merchant data: ${response.status}`);
        }
        
        const merchantData = await response.json();
        
        // Extract industry codes from classification data
        // Structure: merchantData.classification.mcc_codes[0].code
        return {
            mcc: merchantData.classification?.mcc_codes?.[0]?.code || null,
            naics: merchantData.classification?.naics_codes?.[0]?.code || null,
            sic: merchantData.classification?.sic_codes?.[0]?.code || null
        };
    }
    
    /**
     * Get fallback benchmarks when API is not available.
     * 
     * FALLBACK BEHAVIOR:
     *   - Used when risk benchmarks endpoint is unavailable or returns error
     *   - Returns industry average values as placeholder
     *   - Response includes isFallback: true flag to indicate fallback data
     * 
     * FALLBACK DATA - DO NOT USE AS PRIMARY DATA SOURCE
     * 
     * TODO: Replace with real benchmark data from backend endpoint
     * 
     * @param {Object} industryCodes - Industry codes (MCC, NAICS, SIC)
     * @returns {Object} Fallback benchmark data with isFallback flag
     */
    getFallbackBenchmarks(industryCodes) {
        // FALLBACK: Return structured fallback data when API is unavailable
        // This will be replaced with real data when backend endpoint is ready
        return {
            industryCodes,
            industryName: 'Industry Average',
            averages: {
                financial: 25,
                operational: 35,
                regulatory: 45,
                reputational: 30,
                cybersecurity: 40,
                content: 20
            },
            percentiles: {
                p10: { financial: 10, operational: 20, regulatory: 30, reputational: 15, cybersecurity: 25, content: 10 },
                p25: { financial: 15, operational: 25, regulatory: 35, reputational: 20, cybersecurity: 30, content: 15 },
                p50: { financial: 25, operational: 35, regulatory: 45, reputational: 30, cybersecurity: 40, content: 20 },
                p75: { financial: 35, operational: 45, regulatory: 55, reputational: 40, cybersecurity: 50, content: 30 },
                p90: { financial: 45, operational: 55, regulatory: 65, reputational: 50, cybersecurity: 60, content: 40 }
            },
            sampleSize: 0,
            isFallback: true // Flag to indicate this is fallback data
        };
    }
    
    /**
     * Combine risk data from multiple sources
     * @param {Array} results - Promise results
     * @param {string} merchantId - Merchant ID
     * @returns {Object} Combined risk data
     */
    combineRiskData(results, merchantId) {
        const [assessment, history, predictions, benchmarks] = results;
        
        return {
            merchantId,
            current: assessment.status === 'fulfilled' ? assessment.value : null,
            history: history?.status === 'fulfilled' ? history.value : null,
            predictions: predictions?.status === 'fulfilled' ? predictions.value : null,
            benchmarks: benchmarks?.status === 'fulfilled' ? benchmarks.value : null,
            lastUpdated: new Date().toISOString(),
            dataSources: {
                assessment: assessment.status === 'fulfilled',
                history: history?.status === 'fulfilled',
                predictions: predictions?.status === 'fulfilled',
                benchmarks: benchmarks?.status === 'fulfilled'
            }
        };
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
        
        this.eventBus.emit('risk-data-cache-cleared', { merchantId });
    }
    
    /**
     * Get cache key
     * @param {string} merchantId - Merchant ID
     * @param {Object} options - Options
     * @returns {string} Cache key
     */
    getCacheKey(merchantId, options) {
        return `risk_${merchantId}_${JSON.stringify(options)}`;
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
}

// Export singleton instance
let riskDataServiceInstance = null;

export function getRiskDataService(config) {
    if (!riskDataServiceInstance) {
        riskDataServiceInstance = new SharedRiskDataService(config);
    }
    return riskDataServiceInstance;
}

export { SharedRiskDataService };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getRiskDataService = getRiskDataService;
    window.SharedRiskDataService = SharedRiskDataService;
}

