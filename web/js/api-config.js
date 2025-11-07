/**
 * Centralized API Configuration for KYB Platform
 * Handles environment-specific API base URLs and endpoints
 */
class APIConfig {
    /**
     * Get the appropriate API base URL based on environment
     * @returns {string} The API base URL
     */
    static getBaseURL() {
        // Check if we're in development (localhost)
        if (window.location.hostname === 'localhost' || 
            window.location.hostname === '127.0.0.1' ||
            window.location.hostname === '0.0.0.0') {
            return 'http://localhost:8080';
        }
        
        // Production Railway API Gateway
        return 'https://kyb-api-gateway-production.up.railway.app';
    }
    
    /**
     * Get all API endpoints with proper base URL
     * @returns {Object} Object containing all API endpoints
     */
    static getEndpoints() {
        const baseURL = this.getBaseURL();
        
        return {
            // Classification endpoints
            classify: `${baseURL}/v1/classify`,
            
            // Merchant management endpoints
            merchants: `${baseURL}/api/v1/merchants`,
            merchantSearch: `${baseURL}/api/v1/merchants/search`,
            merchantAnalytics: `${baseURL}/api/v1/merchants/analytics`,
            merchantPortfolioTypes: `${baseURL}/api/v1/merchants/portfolio-types`,
            merchantRiskLevels: `${baseURL}/api/v1/merchants/risk-levels`,
            merchantStatistics: `${baseURL}/api/v1/merchants/statistics`,
            merchantById: (id) => `${baseURL}/api/v1/merchants/${id}`,
            
            // Risk assessment endpoints
            riskAssess: `${baseURL}/api/v1/risk/assess`,
            riskHistory: (merchantId) => `${baseURL}/api/v1/risk/history/${merchantId}`,
            riskPredictions: (merchantId) => `${baseURL}/api/v1/risk/predictions/${merchantId}`,
            riskScenarios: `${baseURL}/api/v1/risk/scenarios`,
            riskExplain: (assessmentId) => `${baseURL}/api/v1/risk/explain/${assessmentId}`,
            riskWebSocket: `wss://${baseURL.replace('https://', '').replace('http://', '')}/api/v1/risk/ws`,
            riskBenchmarks: `${baseURL}/api/v1/risk/benchmarks`,
            
            // Risk indicators endpoints (NEW)
            riskIndicators: (merchantId) => `${baseURL}/api/v1/merchants/${merchantId}/risk-indicators`,
            websiteAnalysis: (merchantId) => `${baseURL}/api/v1/merchants/${merchantId}/website-analysis`,
            riskRecommendations: (merchantId) => `${baseURL}/api/v1/merchants/${merchantId}/risk-recommendations`,
            riskAlerts: (merchantId) => `${baseURL}/api/v1/merchants/${merchantId}/risk-alerts`,
            
            // Health and monitoring endpoints
            health: `${baseURL}/health`,
            debugWeb: `${baseURL}/debug/web`
        };
    }
    
    /**
     * Get headers for API requests
     * @returns {Object} Headers object
     */
    static getHeaders() {
        return {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }
    
    /**
     * Get the current environment
     * @returns {string} 'development' or 'production'
     */
    static getEnvironment() {
        if (window.location.hostname === 'localhost' || 
            window.location.hostname === '127.0.0.1' ||
            window.location.hostname === '0.0.0.0') {
            return 'development';
        }
        return 'production';
    }
    
    /**
     * Log current API configuration for debugging
     */
    static logConfig() {
        console.log('🌐 API Configuration:');
        console.log('  Environment:', this.getEnvironment());
        console.log('  Base URL:', this.getBaseURL());
        console.log('  Endpoints:', this.getEndpoints());
    }
}

// Auto-log configuration on load for debugging
if (typeof window !== 'undefined') {
    APIConfig.logConfig();
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = APIConfig;
}
