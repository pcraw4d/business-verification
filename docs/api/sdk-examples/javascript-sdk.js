/**
 * KYB Platform JavaScript SDK Example
 * 
 * This is a practical example of how to use the KYB Platform API
 * with JavaScript/Node.js. This example demonstrates:
 * - Authentication
 * - Business Classification
 * - Risk Assessment
 * - Compliance Checking
 * - Error Handling
 * - Rate Limiting
 */

class KYBPlatformSDK {
    constructor(config = {}) {
        this.baseURL = config.baseURL || 'https://api.kybplatform.com';
        this.accessToken = config.accessToken;
        this.refreshToken = config.refreshToken;
        this.logger = config.logger || console;
        
        // Rate limiting
        this.requests = [];
        this.maxRequests = 100;
        this.windowMs = 60 * 1000; // 1 minute
    }

    /**
     * Make an authenticated HTTP request
     */
    async makeRequest(endpoint, options = {}) {
        // Check rate limit
        await this.checkRateLimit();

        const url = `${this.baseURL}/v1${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (this.accessToken) {
            headers['Authorization'] = `Bearer ${this.accessToken}`;
        }

        try {
            const response = await fetch(url, {
                method: options.method || 'GET',
                headers,
                body: options.body ? JSON.stringify(options.body) : undefined,
                ...options
            });

            // Record request for rate limiting
            this.requests.push(Date.now());

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                const error = new Error(errorData.message || `HTTP ${response.status}`);
                error.status = response.status;
                error.data = errorData;
                throw error;
            }

            return await response.json();
        } catch (error) {
            this.logger.error('API request failed', {
                endpoint,
                error: error.message,
                status: error.status
            });
            throw error;
        }
    }

    /**
     * Check rate limiting
     */
    async checkRateLimit() {
        const now = Date.now();
        this.requests = this.requests.filter(time => now - time < this.windowMs);

        if (this.requests.length >= this.maxRequests) {
            const oldestRequest = this.requests[0];
            const waitTime = this.windowMs - (now - oldestRequest);
            throw new Error(`Rate limit exceeded. Wait ${waitTime}ms`);
        }
    }

    /**
     * Refresh access token
     */
    async refreshAccessToken() {
        if (!this.refreshToken) {
            throw new Error('No refresh token available');
        }

        try {
            const response = await this.makeRequest('/auth/refresh', {
                method: 'POST',
                body: {
                    refresh_token: this.refreshToken
                }
            });

            this.accessToken = response.access_token;
            this.refreshToken = response.refresh_token;

            this.logger.info('Access token refreshed successfully');
            return this.accessToken;
        } catch (error) {
            this.logger.error('Failed to refresh access token', error);
            throw error;
        }
    }

    /**
     * Authentication methods
     */
    auth = {
        /**
         * Login with email and password
         */
        login: async (credentials) => {
            const response = await this.makeRequest('/auth/login', {
                method: 'POST',
                body: credentials
            });

            this.accessToken = response.access_token;
            this.refreshToken = response.refresh_token;

            this.logger.info('Login successful');
            return response;
        },

        /**
         * Register a new user
         */
        register: async (userData) => {
            return await this.makeRequest('/auth/register', {
                method: 'POST',
                body: userData
            });
        },

        /**
         * Logout
         */
        logout: async () => {
            try {
                await this.makeRequest('/auth/logout', {
                    method: 'POST'
                });
            } finally {
                this.accessToken = null;
                this.refreshToken = null;
                this.logger.info('Logout successful');
            }
        }
    };

    /**
     * Business classification methods
     */
    classification = {
        /**
         * Classify a single business
         */
        classify: async (businessData) => {
            this.logger.info('Starting business classification', {
                business_name: businessData.business_name
            });

            try {
                const result = await this.makeRequest('/classify', {
                    method: 'POST',
                    body: businessData
                });

                this.logger.info('Classification completed', {
                    business_name: businessData.business_name,
                    classification_id: result.classification_id,
                    confidence_score: result.confidence_score
                });

                return result;
            } catch (error) {
                this.logger.error('Classification failed', {
                    business_name: businessData.business_name,
                    error: error.message
                });
                throw error;
            }
        },

        /**
         * Batch classify multiple businesses
         */
        batchClassify: async (businesses) => {
            this.logger.info('Starting batch classification', {
                count: businesses.length
            });

            try {
                const result = await this.makeRequest('/classify/batch', {
                    method: 'POST',
                    body: { businesses }
                });

                this.logger.info('Batch classification completed', {
                    count: businesses.length,
                    successful: result.results.filter(r => r.success).length,
                    failed: result.results.filter(r => !r.success).length
                });

                return result;
            } catch (error) {
                this.logger.error('Batch classification failed', {
                    count: businesses.length,
                    error: error.message
                });
                throw error;
            }
        },

        /**
         * Get classification history
         */
        getHistory: async (params = {}) => {
            const queryString = new URLSearchParams(params).toString();
            const endpoint = `/classify/history${queryString ? `?${queryString}` : ''}`;

            return await this.makeRequest(endpoint);
        },

        /**
         * Generate confidence report
         */
        generateConfidenceReport: async (params) => {
            return await this.makeRequest('/classify/confidence-report', {
                method: 'POST',
                body: params
            });
        }
    };

    /**
     * Risk assessment methods
     */
    risk = {
        /**
         * Assess business risk
         */
        assess: async (riskData) => {
            this.logger.info('Starting risk assessment', {
                business_id: riskData.business_id,
                categories: riskData.categories
            });

            try {
                const result = await this.makeRequest('/risk/assess', {
                    method: 'POST',
                    body: riskData
                });

                this.logger.info('Risk assessment completed', {
                    business_id: riskData.business_id,
                    overall_score: result.overall_score,
                    overall_level: result.overall_level
                });

                return result;
            } catch (error) {
                this.logger.error('Risk assessment failed', {
                    business_id: riskData.business_id,
                    error: error.message
                });
                throw error;
            }
        },

        /**
         * Get risk categories
         */
        getCategories: async () => {
            return await this.makeRequest('/risk/categories');
        },

        /**
         * Get risk factors
         */
        getFactors: async (category = null) => {
            const endpoint = category ? `/risk/factors?category=${category}` : '/risk/factors';
            return await this.makeRequest(endpoint);
        },

        /**
         * Get risk thresholds
         */
        getThresholds: async (category = null) => {
            const endpoint = category ? `/risk/thresholds?category=${category}` : '/risk/thresholds';
            return await this.makeRequest(endpoint);
        }
    };

    /**
     * Compliance checking methods
     */
    compliance = {
        /**
         * Check compliance status
         */
        check: async (complianceData) => {
            this.logger.info('Starting compliance check', {
                business_id: complianceData.business_id,
                frameworks: complianceData.frameworks
            });

            try {
                const result = await this.makeRequest('/compliance/check', {
                    method: 'POST',
                    body: complianceData
                });

                this.logger.info('Compliance check completed', {
                    business_id: complianceData.business_id,
                    frameworks_checked: result.frameworks.length
                });

                return result;
            } catch (error) {
                this.logger.error('Compliance check failed', {
                    business_id: complianceData.business_id,
                    error: error.message
                });
                throw error;
            }
        },

        /**
         * Get compliance status
         */
        getStatus: async (businessId) => {
            return await this.makeRequest(`/compliance/status/${businessId}`);
        },

        /**
         * Generate compliance report
         */
        generateReport: async (params) => {
            return await this.makeRequest('/compliance/report', {
                method: 'POST',
                body: params
            });
        }
    };
}

// Example usage
async function example() {
    // Initialize the SDK
    const kyb = new KYBPlatformSDK({
        baseURL: 'https://api.kybplatform.com',
        logger: console
    });

    try {
        // 1. Authenticate
        console.log('🔐 Authenticating...');
        const authResponse = await kyb.auth.login({
            email: 'your-email@example.com',
            password: 'your-password'
        });
        console.log('✅ Authentication successful');

        // 2. Classify a business
        console.log('\n🏢 Classifying business...');
        const classification = await kyb.classification.classify({
            business_name: 'Acme Corporation',
            business_address: '123 Main St, New York, NY 10001',
            business_phone: '+1-555-123-4567',
            business_website: 'https://acme.com'
        });
        console.log('✅ Classification completed:', {
            naics_code: classification.naics_code,
            confidence_score: classification.confidence_score
        });

        // 3. Assess risk
        console.log('\n⚠️ Assessing risk...');
        const riskAssessment = await kyb.risk.assess({
            business_id: 'business-123',
            business_name: 'Acme Corporation',
            categories: ['financial', 'operational']
        });
        console.log('✅ Risk assessment completed:', {
            overall_score: riskAssessment.overall_score,
            overall_level: riskAssessment.overall_level
        });

        // 4. Check compliance
        console.log('\n📋 Checking compliance...');
        const compliance = await kyb.compliance.check({
            business_id: 'business-123',
            frameworks: ['SOC2', 'PCI_DSS', 'GDPR']
        });
        console.log('✅ Compliance check completed:', {
            frameworks_checked: compliance.frameworks.length
        });

        // 5. Batch classification example
        console.log('\n📦 Performing batch classification...');
        const batchResults = await kyb.classification.batchClassify([
            {
                business_name: 'Tech Solutions LLC',
                business_address: '456 Oak Ave, San Francisco, CA'
            },
            {
                business_name: 'Global Services Inc',
                business_address: '789 Pine St, Chicago, IL'
            }
        ]);
        console.log('✅ Batch classification completed:', {
            total: batchResults.results.length,
            successful: batchResults.results.filter(r => r.success).length
        });

    } catch (error) {
        console.error('❌ Error:', error.message);
        
        if (error.status === 401) {
            console.log('🔄 Attempting to refresh token...');
            try {
                await kyb.refreshAccessToken();
                console.log('✅ Token refreshed successfully');
            } catch (refreshError) {
                console.error('❌ Token refresh failed:', refreshError.message);
            }
        }
    } finally {
        // Cleanup
        await kyb.auth.logout();
    }
}

// Error handling with retry logic
async function withRetry(fn, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            if (error.status === 429 && i < maxRetries - 1) {
                const delay = Math.pow(2, i) * 1000;
                console.log(`Rate limited. Waiting ${delay}ms before retry...`);
                await new Promise(resolve => setTimeout(resolve, delay));
                continue;
            }
            throw error;
        }
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = KYBPlatformSDK;
}

// Run example if this file is executed directly
if (typeof require !== 'undefined' && require.main === module) {
    example().catch(console.error);
}
