/**
 * Merchant Portfolio API SDK - JavaScript/Node.js
 * 
 * A comprehensive SDK for interacting with the Merchant Portfolio Management API
 * 
 * @version 1.0.0
 * @author KYB Platform Team
 */

const axios = require('axios');

/**
 * Merchant Portfolio API Client
 */
class MerchantPortfolioAPI {
  constructor(config = {}) {
    this.baseURL = config.baseURL || 'https://api.kyb-platform.com/v1';
    this.apiKey = config.apiKey;
    this.timeout = config.timeout || 30000;
    this.retryAttempts = config.retryAttempts || 3;
    this.retryDelay = config.retryDelay || 1000;
    
    // Initialize HTTP client
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: this.timeout,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        'User-Agent': 'MerchantPortfolioSDK/1.0.0'
      }
    });
    
    // Add request/response interceptors
    this.setupInterceptors();
  }
  
  /**
   * Setup request and response interceptors
   */
  setupInterceptors() {
    // Request interceptor for logging
    this.client.interceptors.request.use(
      (config) => {
        console.log(`[API] ${config.method.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('[API] Request error:', error);
        return Promise.reject(error);
      }
    );
    
    // Response interceptor for error handling and retries
    this.client.interceptors.response.use(
      (response) => {
        console.log(`[API] ${response.status} ${response.config.url}`);
        return response;
      },
      async (error) => {
        const config = error.config;
        
        // Retry logic for network errors and 5xx responses
        if (config && !config._retry && this.shouldRetry(error)) {
          config._retry = true;
          config._retryCount = (config._retryCount || 0) + 1;
          
          if (config._retryCount <= this.retryAttempts) {
            const delay = this.retryDelay * Math.pow(2, config._retryCount - 1);
            console.log(`[API] Retrying request in ${delay}ms (attempt ${config._retryCount})`);
            
            await new Promise(resolve => setTimeout(resolve, delay));
            return this.client(config);
          }
        }
        
        // Transform error response
        if (error.response) {
          const apiError = new APIError(
            error.response.data?.error?.message || 'API request failed',
            error.response.status,
            error.response.data?.error?.code || 'UNKNOWN_ERROR',
            error.response.data?.error?.details
          );
          console.error('[API] Error:', apiError);
          return Promise.reject(apiError);
        }
        
        console.error('[API] Network error:', error.message);
        return Promise.reject(new APIError(
          'Network error - unable to reach API',
          0,
          'NETWORK_ERROR',
          error.message
        ));
      }
    );
  }
  
  /**
   * Determine if a request should be retried
   */
  shouldRetry(error) {
    if (!error.response) return true; // Network error
    if (error.response.status >= 500) return true; // Server error
    if (error.response.status === 429) return true; // Rate limit
    return false;
  }
  
  // ============================================================================
  // MERCHANT CRUD OPERATIONS
  // ============================================================================
  
  /**
   * Create a new merchant
   * @param {Object} merchantData - Merchant data
   * @returns {Promise<Object>} Created merchant
   */
  async createMerchant(merchantData) {
    try {
      const response = await this.client.post('/merchants', merchantData);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create merchant: ${error.message}`);
    }
  }
  
  /**
   * Get a merchant by ID
   * @param {string} merchantId - Merchant ID
   * @returns {Promise<Object>} Merchant data
   */
  async getMerchant(merchantId) {
    try {
      const response = await this.client.get(`/merchants/${merchantId}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get merchant: ${error.message}`);
    }
  }
  
  /**
   * Update a merchant
   * @param {string} merchantId - Merchant ID
   * @param {Object} updateData - Update data
   * @returns {Promise<Object>} Updated merchant
   */
  async updateMerchant(merchantId, updateData) {
    try {
      const response = await this.client.put(`/merchants/${merchantId}`, updateData);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to update merchant: ${error.message}`);
    }
  }
  
  /**
   * Delete a merchant
   * @param {string} merchantId - Merchant ID
   * @returns {Promise<void>}
   */
  async deleteMerchant(merchantId) {
    try {
      await this.client.delete(`/merchants/${merchantId}`);
    } catch (error) {
      throw new Error(`Failed to delete merchant: ${error.message}`);
    }
  }
  
  // ============================================================================
  // SEARCH AND LISTING
  // ============================================================================
  
  /**
   * List merchants with optional filters
   * @param {Object} filters - Search filters
   * @returns {Promise<Object>} Paginated merchant list
   */
  async listMerchants(filters = {}) {
    try {
      const response = await this.client.get('/merchants', { params: filters });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to list merchants: ${error.message}`);
    }
  }
  
  /**
   * Advanced search for merchants
   * @param {Object} searchCriteria - Search criteria
   * @returns {Promise<Object>} Search results
   */
  async searchMerchants(searchCriteria) {
    try {
      const response = await this.client.post('/merchants/search', searchCriteria);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to search merchants: ${error.message}`);
    }
  }
  
  /**
   * Get all merchants with automatic pagination
   * @param {Object} filters - Search filters
   * @returns {Promise<Array>} All merchants
   */
  async getAllMerchants(filters = {}) {
    const allMerchants = [];
    let page = 1;
    let hasMore = true;
    
    while (hasMore) {
      const response = await this.listMerchants({
        ...filters,
        page,
        page_size: 100
      });
      
      allMerchants.push(...response.merchants);
      hasMore = response.has_more;
      page++;
      
      // Add small delay to respect rate limits
      if (hasMore) {
        await new Promise(resolve => setTimeout(resolve, 100));
      }
    }
    
    return allMerchants;
  }
  
  // ============================================================================
  // BULK OPERATIONS
  // ============================================================================
  
  /**
   * Bulk update portfolio type
   * @param {Array<string>} merchantIds - Array of merchant IDs
   * @param {string} portfolioType - New portfolio type
   * @returns {Promise<Object>} Bulk operation result
   */
  async bulkUpdatePortfolioType(merchantIds, portfolioType) {
    try {
      const response = await this.client.post('/merchants/bulk/portfolio-type', {
        merchant_ids: merchantIds,
        portfolio_type: portfolioType
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to bulk update portfolio type: ${error.message}`);
    }
  }
  
  /**
   * Bulk update risk level
   * @param {Array<string>} merchantIds - Array of merchant IDs
   * @param {string} riskLevel - New risk level
   * @returns {Promise<Object>} Bulk operation result
   */
  async bulkUpdateRiskLevel(merchantIds, riskLevel) {
    try {
      const response = await this.client.post('/merchants/bulk/risk-level', {
        merchant_ids: merchantIds,
        risk_level: riskLevel
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to bulk update risk level: ${error.message}`);
    }
  }
  
  /**
   * Get bulk operation status
   * @param {string} operationId - Operation ID
   * @returns {Promise<Object>} Operation status
   */
  async getBulkOperationStatus(operationId) {
    try {
      const response = await this.client.get(`/merchants/bulk/${operationId}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get bulk operation status: ${error.message}`);
    }
  }
  
  /**
   * Wait for bulk operation to complete
   * @param {string} operationId - Operation ID
   * @param {number} pollInterval - Polling interval in milliseconds
   * @returns {Promise<Object>} Final operation result
   */
  async waitForBulkOperation(operationId, pollInterval = 2000) {
    while (true) {
      const status = await this.getBulkOperationStatus(operationId);
      
      if (status.status === 'completed' || status.status === 'failed') {
        return status;
      }
      
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    }
  }
  
  // ============================================================================
  // SESSION MANAGEMENT
  // ============================================================================
  
  /**
   * Start a merchant session
   * @param {string} merchantId - Merchant ID
   * @returns {Promise<Object>} Session data
   */
  async startMerchantSession(merchantId) {
    try {
      const response = await this.client.post(`/merchants/${merchantId}/session`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to start merchant session: ${error.message}`);
    }
  }
  
  /**
   * End a merchant session
   * @param {string} merchantId - Merchant ID
   * @returns {Promise<void>}
   */
  async endMerchantSession(merchantId) {
    try {
      await this.client.delete(`/merchants/${merchantId}/session`);
    } catch (error) {
      throw new Error(`Failed to end merchant session: ${error.message}`);
    }
  }
  
  /**
   * Get active merchant session
   * @returns {Promise<Object|null>} Active session or null
   */
  async getActiveSession() {
    try {
      const response = await this.client.get('/merchants/session/active');
      return response.data;
    } catch (error) {
      if (error.status === 404) {
        return null;
      }
      throw new Error(`Failed to get active session: ${error.message}`);
    }
  }
  
  // ============================================================================
  // ANALYTICS AND REPORTING
  // ============================================================================
  
  /**
   * Get merchant analytics
   * @param {Object} filters - Analytics filters
   * @returns {Promise<Object>} Analytics data
   */
  async getAnalytics(filters = {}) {
    try {
      const response = await this.client.get('/merchants/analytics', { params: filters });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get analytics: ${error.message}`);
    }
  }
  
  /**
   * Get portfolio types
   * @returns {Promise<Object>} Portfolio types list
   */
  async getPortfolioTypes() {
    try {
      const response = await this.client.get('/merchants/portfolio-types');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get portfolio types: ${error.message}`);
    }
  }
  
  /**
   * Get risk levels
   * @returns {Promise<Object>} Risk levels list
   */
  async getRiskLevels() {
    try {
      const response = await this.client.get('/merchants/risk-levels');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get risk levels: ${error.message}`);
    }
  }
  
  /**
   * Get merchant statistics
   * @returns {Promise<Object>} Statistics data
   */
  async getStatistics() {
    try {
      const response = await this.client.get('/merchants/statistics');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get statistics: ${error.message}`);
    }
  }
  
  // ============================================================================
  // UTILITY METHODS
  // ============================================================================
  
  /**
   * Validate merchant data
   * @param {Object} data - Merchant data
   * @returns {Array<string>} Validation errors
   */
  validateMerchantData(data) {
    const errors = [];
    
    if (!data.name || data.name.trim().length === 0) {
      errors.push('Name is required');
    }
    
    if (!['onboarded', 'deactivated', 'prospective', 'pending'].includes(data.portfolio_type)) {
      errors.push('Invalid portfolio type');
    }
    
    if (!['high', 'medium', 'low'].includes(data.risk_level)) {
      errors.push('Invalid risk level');
    }
    
    if (data.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(data.email)) {
      errors.push('Invalid email format');
    }
    
    return errors;
  }
  
  /**
   * Create merchant with validation
   * @param {Object} merchantData - Merchant data
   * @returns {Promise<Object>} Created merchant
   */
  async createMerchantWithValidation(merchantData) {
    const errors = this.validateMerchantData(merchantData);
    if (errors.length > 0) {
      throw new Error(`Validation failed: ${errors.join(', ')}`);
    }
    
    return this.createMerchant(merchantData);
  }
  
  /**
   * Batch create merchants
   * @param {Array<Object>} merchants - Array of merchant data
   * @param {number} batchSize - Batch size
   * @returns {Promise<Array>} Results array
   */
  async batchCreateMerchants(merchants, batchSize = 10) {
    const results = [];
    
    for (let i = 0; i < merchants.length; i += batchSize) {
      const batch = merchants.slice(i, i + batchSize);
      
      console.log(`Processing batch ${Math.floor(i / batchSize) + 1} of ${Math.ceil(merchants.length / batchSize)}`);
      
      const batchPromises = batch.map(merchant => 
        this.createMerchant(merchant)
          .then(result => ({ success: true, merchant: merchant.name, result }))
          .catch(error => ({ success: false, merchant: merchant.name, error: error.message }))
      );
      
      const batchResults = await Promise.all(batchPromises);
      results.push(...batchResults);
      
      // Add delay between batches
      if (i + batchSize < merchants.length) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
    
    return results;
  }
}

/**
 * Custom API Error class
 */
class APIError extends Error {
  constructor(message, status, code, details) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/**
 * Session Manager for handling merchant sessions
 */
class MerchantSessionManager {
  constructor(apiClient) {
    this.api = apiClient;
    this.activeSession = null;
  }
  
  /**
   * Start a new session (ends current session if exists)
   * @param {string} merchantId - Merchant ID
   * @returns {Promise<Object>} Session data
   */
  async startSession(merchantId) {
    // End current session if exists
    if (this.activeSession) {
      await this.endCurrentSession();
    }
    
    this.activeSession = await this.api.startMerchantSession(merchantId);
    console.log(`Started session for merchant: ${this.activeSession.merchant_name}`);
    return this.activeSession;
  }
  
  /**
   * End the current session
   * @returns {Promise<void>}
   */
  async endCurrentSession() {
    if (!this.activeSession) return;
    
    await this.api.endMerchantSession(this.activeSession.merchant_id);
    console.log(`Ended session for merchant: ${this.activeSession.merchant_name}`);
    this.activeSession = null;
  }
  
  /**
   * Get the current active session
   * @returns {Object|null} Active session or null
   */
  getActiveSession() {
    return this.activeSession;
  }
  
  /**
   * Check if there's an active session
   * @returns {boolean} True if session is active
   */
  hasActiveSession() {
    return this.activeSession !== null;
  }
}

/**
 * Cache manager for API responses
 */
class APICache {
  constructor(ttl = 300000) { // 5 minutes default TTL
    this.cache = new Map();
    this.ttl = ttl;
  }
  
  /**
   * Get cached data
   * @param {string} key - Cache key
   * @returns {any} Cached data or null
   */
  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;
    
    if (Date.now() > item.expiry) {
      this.cache.delete(key);
      return null;
    }
    
    return item.data;
  }
  
  /**
   * Set cached data
   * @param {string} key - Cache key
   * @param {any} data - Data to cache
   */
  set(key, data) {
    this.cache.set(key, {
      data,
      expiry: Date.now() + this.ttl
    });
  }
  
  /**
   * Clear all cached data
   */
  clear() {
    this.cache.clear();
  }
  
  /**
   * Clear expired entries
   */
  cleanup() {
    const now = Date.now();
    for (const [key, item] of this.cache.entries()) {
      if (now > item.expiry) {
        this.cache.delete(key);
      }
    }
  }
}

// Export classes and utilities
module.exports = {
  MerchantPortfolioAPI,
  APIError,
  MerchantSessionManager,
  APICache
};

// Example usage
if (require.main === module) {
  // Example usage
  const api = new MerchantPortfolioAPI({
    baseURL: 'https://api.kyb-platform.com/v1',
    apiKey: process.env.API_TOKEN
  });
  
  const sessionManager = new MerchantSessionManager(api);
  
  async function example() {
    try {
      // Create a new merchant
      const merchant = await api.createMerchantWithValidation({
        name: 'Example Corp',
        legal_name: 'Example Corporation LLC',
        industry: 'Technology',
        portfolio_type: 'prospective',
        risk_level: 'medium',
        address: {
          street1: '123 Example St',
          city: 'San Francisco',
          state: 'CA',
          postal_code: '94105',
          country: 'United States',
          country_code: 'US'
        },
        contact_info: {
          email: 'contact@example.com',
          phone: '+1-555-123-4567'
        }
      });
      
      console.log('Created merchant:', merchant);
      
      // Start a session
      await sessionManager.startSession(merchant.id);
      
      // Get analytics
      const analytics = await api.getAnalytics();
      console.log('Analytics:', analytics);
      
      // End session
      await sessionManager.endCurrentSession();
      
    } catch (error) {
      console.error('Error:', error.message);
    }
  }
  
  example();
}
