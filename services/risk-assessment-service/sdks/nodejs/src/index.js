/**
 * KYB Platform Risk Assessment Service Node.js SDK
 * 
 * This SDK provides a Node.js interface to the KYB Platform Risk Assessment Service API.
 */

const axios = require('axios');
const { KYBError, ValidationError, AuthenticationError, AuthorizationError } = require('./exceptions');
const { NotFoundError, RateLimitError, ServiceUnavailableError, TimeoutError } = require('./exceptions');
const { InternalError, APIError } = require('./exceptions');

/**
 * KYB Platform Risk Assessment Service Client
 * 
 * This client provides methods to interact with the KYB Platform API for
 * risk assessment, compliance checking, and analytics.
 */
class KYBClient {
    /**
     * Initialize the KYB client.
     * 
     * @param {string} apiKey - Your KYB Platform API key
     * @param {Object} options - Configuration options
     * @param {string} options.baseUrl - Base URL for the API (default: https://api.kyb-platform.com/v1)
     * @param {number} options.timeout - Request timeout in milliseconds (default: 30000)
     * @param {number} options.maxRetries - Maximum number of retries for failed requests (default: 3)
     * @param {string} options.userAgent - User agent string for requests (default: kyb-nodejs-client/1.0.0)
     */
    constructor(apiKey, options = {}) {
        if (!apiKey) {
            throw new ValidationError('API key is required');
        }

        this.apiKey = apiKey;
        this.baseUrl = (options.baseUrl || 'https://api.kyb-platform.com/v1').replace(/\/$/, '');
        this.timeout = options.timeout || 30000;
        this.maxRetries = options.maxRetries || 3;
        this.userAgent = options.userAgent || 'kyb-nodejs-client/1.0.0';

        // Create axios instance with default configuration
        this.client = axios.create({
            baseURL: this.baseUrl,
            timeout: this.timeout,
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'Content-Type': 'application/json',
                'User-Agent': this.userAgent
            }
        });

        // Add request interceptor for retries
        this.client.interceptors.response.use(
            (response) => response,
            async (error) => {
                const config = error.config;
                
                if (!config || !config.retry) {
                    config.retry = 0;
                }

                if (config.retry >= this.maxRetries) {
                    return Promise.reject(error);
                }

                config.retry += 1;

                // Retry on specific status codes
                if (error.response && [429, 500, 502, 503, 504].includes(error.response.status)) {
                    const delay = Math.pow(2, config.retry) * 1000; // Exponential backoff
                    await new Promise(resolve => setTimeout(resolve, delay));
                    return this.client(config);
                }

                return Promise.reject(error);
            }
        );
    }

    /**
     * Perform a risk assessment for a business.
     * 
     * @param {Object} params - Risk assessment parameters
     * @param {string} params.businessName - Name of the business
     * @param {string} params.businessAddress - Address of the business
     * @param {string} params.industry - Industry of the business
     * @param {string} params.country - Country code (2-letter ISO)
     * @param {string} [params.phone] - Phone number (optional)
     * @param {string} [params.email] - Email address (optional)
     * @param {string} [params.website] - Website URL (optional)
     * @param {number} [params.predictionHorizon=3] - Prediction horizon in months (0-12)
     * @param {Object} [params.metadata] - Additional metadata (optional)
     * @returns {Promise<Object>} Risk assessment results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async assessRisk(params) {
        // Validate required fields
        if (!params.businessName) {
            throw new ValidationError('businessName is required');
        }
        if (!params.businessAddress) {
            throw new ValidationError('businessAddress is required');
        }
        if (!params.industry) {
            throw new ValidationError('industry is required');
        }
        if (!params.country) {
            throw new ValidationError('country is required');
        }
        if (params.country.length !== 2) {
            throw new ValidationError('country must be a 2-letter ISO code');
        }
        if (params.predictionHorizon !== undefined && (params.predictionHorizon < 0 || params.predictionHorizon > 24)) {
            throw new ValidationError('predictionHorizon must be between 0 and 24 months');
        }

        // Prepare request data
        const data = {
            business_name: params.businessName,
            business_address: params.businessAddress,
            industry: params.industry,
            country: params.country,
            prediction_horizon: params.predictionHorizon || 3
        };

        if (params.phone) data.phone = params.phone;
        if (params.email) data.email = params.email;
        if (params.website) data.website = params.website;
        if (params.metadata) data.metadata = params.metadata;

        return this._makeRequest('POST', '/assess', data);
    }

    /**
     * Retrieve a risk assessment by ID.
     * 
     * @param {string} assessmentId - The assessment ID
     * @returns {Promise<Object>} Risk assessment data
     * @throws {ValidationError} If the assessment ID is invalid
     * @throws {NotFoundError} If the assessment is not found
     * @throws {APIError} If the API request fails
     */
    async getRiskAssessment(assessmentId) {
        if (!assessmentId) {
            throw new ValidationError('assessmentId is required');
        }

        return this._makeRequest('GET', `/assess/${assessmentId}`);
    }

    /**
     * Perform future risk prediction for a business.
     * 
     * @param {string} assessmentId - The assessment ID
     * @param {Object} params - Prediction parameters
     * @param {number} params.horizonMonths - Prediction horizon in months (1-12)
     * @param {string[]} [params.scenarios] - List of scenarios to analyze (optional)
     * @returns {Promise<Object>} Risk prediction results
     * @throws {ValidationError} If the request data is invalid
     * @throws {NotFoundError} If the assessment is not found
     * @throws {APIError} If the API request fails
     */
    async predictRisk(assessmentId, params) {
        if (!assessmentId) {
            throw new ValidationError('assessmentId is required');
        }
        if (!params.horizonMonths || params.horizonMonths <= 0 || params.horizonMonths > 24) {
            throw new ValidationError('horizonMonths must be between 1 and 24');
        }

        const data = {
            horizon_months: params.horizonMonths
        };

        if (params.scenarios) {
            data.scenarios = params.scenarios;
        }

        return this._makeRequest('POST', `/assess/${assessmentId}/predict`, data);
    }

    /**
     * Retrieve risk assessment history for a business.
     * 
     * @param {string} assessmentId - The assessment ID
     * @returns {Promise<Object>} Risk history data
     * @throws {ValidationError} If the assessment ID is invalid
     * @throws {NotFoundError} If the assessment is not found
     * @throws {APIError} If the API request fails
     */
    async getRiskHistory(assessmentId) {
        if (!assessmentId) {
            throw new ValidationError('assessmentId is required');
        }

        return this._makeRequest('GET', `/assess/${assessmentId}/history`);
    }

    /**
     * Perform compliance checks for a business.
     * 
     * @param {Object} params - Compliance check parameters
     * @param {string} params.businessName - Name of the business
     * @param {string} params.businessAddress - Address of the business
     * @param {string} params.industry - Industry of the business
     * @param {string} params.country - Country code (2-letter ISO)
     * @param {string[]} [params.complianceTypes] - List of compliance types to check (optional)
     * @returns {Promise<Object>} Compliance check results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async checkCompliance(params) {
        if (!params.businessName) {
            throw new ValidationError('businessName is required');
        }
        if (!params.businessAddress) {
            throw new ValidationError('businessAddress is required');
        }
        if (!params.industry) {
            throw new ValidationError('industry is required');
        }
        if (!params.country) {
            throw new ValidationError('country is required');
        }

        const data = {
            business_name: params.businessName,
            business_address: params.businessAddress,
            industry: params.industry,
            country: params.country
        };

        if (params.complianceTypes) {
            data.compliance_types = params.complianceTypes;
        }

        return this._makeRequest('POST', '/compliance/check', data);
    }

    /**
     * Perform sanctions screening for a business.
     * 
     * @param {Object} params - Sanctions screening parameters
     * @param {string} params.businessName - Name of the business
     * @param {string} params.businessAddress - Address of the business
     * @param {string} params.country - Country code (2-letter ISO)
     * @returns {Promise<Object>} Sanctions screening results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async screenSanctions(params) {
        if (!params.businessName) {
            throw new ValidationError('businessName is required');
        }
        if (!params.businessAddress) {
            throw new ValidationError('businessAddress is required');
        }
        if (!params.country) {
            throw new ValidationError('country is required');
        }

        const data = {
            business_name: params.businessName,
            business_address: params.businessAddress,
            country: params.country
        };

        return this._makeRequest('POST', '/sanctions/screen', data);
    }

    /**
     * Set up adverse media monitoring for a business.
     * 
     * @param {Object} params - Media monitoring parameters
     * @param {string} params.businessName - Name of the business
     * @param {string} params.businessAddress - Address of the business
     * @param {string[]} [params.monitoringTypes] - List of monitoring types (optional)
     * @returns {Promise<Object>} Media monitoring setup results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async monitorMedia(params) {
        if (!params.businessName) {
            throw new ValidationError('businessName is required');
        }
        if (!params.businessAddress) {
            throw new ValidationError('businessAddress is required');
        }

        const data = {
            business_name: params.businessName,
            business_address: params.businessAddress
        };

        if (params.monitoringTypes) {
            data.monitoring_types = params.monitoringTypes;
        }

        return this._makeRequest('POST', '/media/monitor', data);
    }

    /**
     * Retrieve risk trends and analytics.
     * 
     * @param {Object} [params] - Query parameters
     * @param {string} [params.industry] - Filter by industry (optional)
     * @param {string} [params.country] - Filter by country (optional)
     * @param {string} [params.timeframe] - Time period (7d, 30d, 90d, 1y) (optional)
     * @param {number} [params.limit] - Number of results (optional)
     * @returns {Promise<Object>} Risk trends data
     * @throws {APIError} If the API request fails
     */
    async getRiskTrends(params = {}) {
        const queryParams = {};
        if (params.industry) queryParams.industry = params.industry;
        if (params.country) queryParams.country = params.country;
        if (params.timeframe) queryParams.timeframe = params.timeframe;
        if (params.limit) queryParams.limit = params.limit;

        return this._makeRequest('GET', '/analytics/trends', null, queryParams);
    }

    /**
     * Retrieve risk insights and recommendations.
     * 
     * @param {Object} [params] - Query parameters
     * @param {string} [params.industry] - Filter by industry (optional)
     * @param {string} [params.country] - Filter by country (optional)
     * @param {string} [params.riskLevel] - Filter by risk level (optional)
     * @returns {Promise<Object>} Risk insights data
     * @throws {APIError} If the API request fails
     */
    async getRiskInsights(params = {}) {
        const queryParams = {};
        if (params.industry) queryParams.industry = params.industry;
        if (params.country) queryParams.country = params.country;
        if (params.riskLevel) queryParams.riskLevel = params.riskLevel;

        return this._makeRequest('GET', '/analytics/insights', null, queryParams);
    }

    /**
     * Perform risk prediction with specific model selection.
     * 
     * @param {string} assessmentId - The assessment ID
     * @param {Object} params - Prediction parameters
     * @param {number} params.horizonMonths - Prediction horizon in months (1-24)
     * @param {string} [params.modelType='auto'] - Model preference ("auto", "xgboost", "lstm", "ensemble")
     * @param {boolean} [params.includeTemporalAnalysis=true] - Include temporal analysis in response
     * @returns {Promise<Object>} Risk prediction results
     * @throws {ValidationError} If the request data is invalid
     * @throws {NotFoundError} If the assessment is not found
     * @throws {APIError} If the API request fails
     */
    async predictRiskWithHorizon(assessmentId, params) {
        if (!assessmentId) {
            throw new ValidationError('assessmentId is required');
        }
        if (!params.horizonMonths || params.horizonMonths <= 0 || params.horizonMonths > 24) {
            throw new ValidationError('horizonMonths must be between 1 and 24');
        }

        const validModels = ['auto', 'xgboost', 'lstm', 'ensemble'];
        if (params.modelType && !validModels.includes(params.modelType)) {
            throw new ValidationError(`modelType must be one of: ${validModels.join(', ')}`);
        }

        const data = {
            horizon_months: params.horizonMonths,
            model_type: params.modelType || 'auto',
            include_temporal_analysis: params.includeTemporalAnalysis !== false
        };

        return this._makeRequest('POST', `/assess/${assessmentId}/predict`, data);
    }

    /**
     * Perform advanced multi-horizon risk prediction.
     * 
     * @param {Object} params - Advanced prediction parameters
     * @param {Object} params.business - Business information
     * @param {number[]} params.predictionHorizons - List of prediction horizons in months (1-24)
     * @param {string} [params.modelPreference='auto'] - Model preference ("auto", "xgboost", "lstm", "ensemble")
     * @param {boolean} [params.includeTemporalAnalysis=true] - Include temporal analysis in response
     * @param {boolean} [params.includeScenarioAnalysis=true] - Include scenario analysis in response
     * @param {boolean} [params.includeModelComparison=true] - Include model comparison analysis
     * @param {number} [params.confidenceThreshold=0.7] - Minimum confidence threshold (0-1)
     * @param {string[]} [params.customScenarios] - Custom scenarios to analyze (optional)
     * @returns {Promise<Object>} Advanced prediction results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async predictMultiHorizon(params) {
        if (!params.business) {
            throw new ValidationError('business is required');
        }
        if (!params.predictionHorizons || !Array.isArray(params.predictionHorizons) || params.predictionHorizons.length === 0) {
            throw new ValidationError('predictionHorizons is required and must be a non-empty array');
        }
        if (params.predictionHorizons.length > 5) {
            throw new ValidationError('maximum of 5 prediction horizons allowed');
        }

        for (const horizon of params.predictionHorizons) {
            if (horizon < 1 || horizon > 24) {
                throw new ValidationError('prediction horizon must be between 1 and 24 months');
            }
        }

        const validModels = ['auto', 'xgboost', 'lstm', 'ensemble'];
        if (params.modelPreference && !validModels.includes(params.modelPreference)) {
            throw new ValidationError(`modelPreference must be one of: ${validModels.join(', ')}`);
        }

        if (params.confidenceThreshold !== undefined && (params.confidenceThreshold < 0 || params.confidenceThreshold > 1)) {
            throw new ValidationError('confidenceThreshold must be between 0 and 1');
        }

        const data = {
            business: params.business,
            prediction_horizons: params.predictionHorizons,
            model_preference: params.modelPreference || 'auto',
            include_temporal_analysis: params.includeTemporalAnalysis !== false,
            include_scenario_analysis: params.includeScenarioAnalysis !== false,
            include_model_comparison: params.includeModelComparison !== false,
            confidence_threshold: params.confidenceThreshold || 0.7
        };

        if (params.customScenarios) {
            data.custom_scenarios = params.customScenarios;
        }

        return this._makeRequest('POST', '/risk/predict-advanced', data);
    }

    /**
     * Perform LSTM-specific risk prediction.
     * 
     * @param {Object} params - LSTM prediction parameters
     * @param {Object} params.business - Business information
     * @param {number} params.horizonMonths - Prediction horizon in months (1-24)
     * @param {boolean} [params.includeTemporalAnalysis=true] - Include temporal analysis in response
     * @param {boolean} [params.includeScenarioAnalysis=true] - Include scenario analysis in response
     * @returns {Promise<Object>} LSTM prediction results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async predictWithLSTM(params) {
        return this.predictMultiHorizon({
            business: params.business,
            predictionHorizons: [params.horizonMonths],
            modelPreference: 'lstm',
            includeTemporalAnalysis: params.includeTemporalAnalysis !== false,
            includeScenarioAnalysis: params.includeScenarioAnalysis !== false,
            includeModelComparison: false
        });
    }

    /**
     * Perform ensemble risk prediction.
     * 
     * @param {Object} params - Ensemble prediction parameters
     * @param {Object} params.business - Business information
     * @param {number} params.horizonMonths - Prediction horizon in months (1-24)
     * @param {boolean} [params.includeTemporalAnalysis=true] - Include temporal analysis in response
     * @param {boolean} [params.includeScenarioAnalysis=true] - Include scenario analysis in response
     * @param {boolean} [params.includeModelComparison=true] - Include model comparison analysis
     * @returns {Promise<Object>} Ensemble prediction results
     * @throws {ValidationError} If the request data is invalid
     * @throws {APIError} If the API request fails
     */
    async predictWithEnsemble(params) {
        return this.predictMultiHorizon({
            business: params.business,
            predictionHorizons: [params.horizonMonths],
            modelPreference: 'ensemble',
            includeTemporalAnalysis: params.includeTemporalAnalysis !== false,
            includeScenarioAnalysis: params.includeScenarioAnalysis !== false,
            includeModelComparison: params.includeModelComparison !== false
        });
    }

    /**
     * Retrieve information about available models.
     * 
     * @param {string} modelType - Model type ("xgboost", "lstm", "ensemble")
     * @returns {Promise<Object>} Model information
     * @throws {ValidationError} If the model type is invalid
     * @throws {APIError} If the API request fails
     */
    async getModelInfo(modelType) {
        if (!modelType) {
            throw new ValidationError('modelType is required');
        }

        const validModels = ['xgboost', 'lstm', 'ensemble'];
        if (!validModels.includes(modelType)) {
            throw new ValidationError(`modelType must be one of: ${validModels.join(', ')}`);
        }

        return this._makeRequest('GET', `/models/${modelType}/info`);
    }

    /**
     * Retrieve performance metrics for models.
     * 
     * @returns {Promise<Object>} Model performance metrics
     * @throws {APIError} If the API request fails
     */
    async getModelPerformance() {
        return this._makeRequest('GET', '/models/performance');
    }

    /**
     * Make an HTTP request to the API.
     * 
     * @param {string} method - HTTP method
     * @param {string} endpoint - API endpoint
     * @param {Object} [data] - Request data (for POST/PUT requests)
     * @param {Object} [params] - Query parameters
     * @returns {Promise<Object>} Response data
     * @throws {APIError} If the API request fails
     * @private
     */
    async _makeRequest(method, endpoint, data = null, params = null) {
        try {
            const config = {
                method: method.toLowerCase(),
                url: endpoint,
                params: params
            };

            if (data) {
                config.data = data;
            }

            const response = await this.client(config);
            return response.data;

        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Handle error responses from the API.
     * 
     * @param {Error} error - The error object
     * @throws {Appropriate exception based on error type}
     * @private
     */
    _handleError(error) {
        if (error.response) {
            // API responded with error status
            const status = error.response.status;
            const errorData = error.response.data;

            const errorCode = errorData?.error?.code || 'UNKNOWN_ERROR';
            const errorMessage = errorData?.error?.message || 'Unknown error';

            // Create appropriate exception based on error code
            switch (errorCode) {
                case 'VALIDATION_ERROR':
                    throw new ValidationError(errorMessage, errorData);
                case 'AUTHENTICATION_ERROR':
                    throw new AuthenticationError(errorMessage, errorData);
                case 'AUTHORIZATION_ERROR':
                    throw new AuthorizationError(errorMessage, errorData);
                case 'NOT_FOUND':
                    throw new NotFoundError(errorMessage, errorData);
                case 'RATE_LIMIT_EXCEEDED':
                    throw new RateLimitError(errorMessage, errorData);
                case 'SERVICE_UNAVAILABLE':
                    throw new ServiceUnavailableError(errorMessage, errorData);
                case 'REQUEST_TIMEOUT':
                    throw new TimeoutError(errorMessage, errorData);
                case 'INTERNAL_ERROR':
                    throw new InternalError(errorMessage, errorData);
                default:
                    throw new APIError(errorMessage, errorData, status);
            }
        } else if (error.request) {
            // Request was made but no response received
            if (error.code === 'ECONNABORTED') {
                throw new TimeoutError('Request timeout');
            } else {
                throw new ServiceUnavailableError('Service unavailable');
            }
        } else {
            // Something else happened
            throw new APIError(`Request failed: ${error.message}`);
        }
    }
}

module.exports = {
    KYBClient,
    KYBError,
    ValidationError,
    AuthenticationError,
    AuthorizationError,
    NotFoundError,
    RateLimitError,
    ServiceUnavailableError,
    TimeoutError,
    InternalError,
    APIError
};
