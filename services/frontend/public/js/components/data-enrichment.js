/**
 * Data Enrichment Component
 * Handles external data enrichment for merchants
 */
class DataEnrichment {
    constructor(merchantId, options = {}) {
        this.merchantId = merchantId;
        this.options = {
            autoEnrich: false,
            ...options
        };
        this.enrichmentData = null;
        this.isEnriching = false;
    }

    /**
     * Initialize the enrichment component
     */
    async init() {
        await this.loadSupportedSources();
    }

    /**
     * Load supported data sources
     */
    async loadSupportedSources() {
        try {
            const response = await fetch('/api/v1/supported', {
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Check if response is JSON before parsing
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                console.warn('⚠️ API returned non-JSON response for supported sources, using empty array');
                console.warn('🔍 Response details:', {
                    status: response.status,
                    statusText: response.statusText,
                    contentType: contentType,
                    url: response.url
                });
                // Try to read response text for debugging (clone response first)
                try {
                    const clonedResponse = response.clone();
                    const text = await clonedResponse.text();
                    console.warn('🔍 Response body (first 500 chars):', text.substring(0, 500));
                } catch (e) {
                    console.warn('🔍 Could not read response body:', e.message);
                }
                this.supportedSources = [];
                return [];
            }

            this.supportedSources = await response.json();
            return this.supportedSources;
        } catch (error) {
            console.warn('⚠️ Error loading supported sources, using empty array:', error.message);
            this.supportedSources = [];
            return [];
        }
    }

    /**
     * Enrich merchant data from external sources
     * @param {string} source - Data source name (e.g., 'thomson-reuters', 'supported')
     */
    async enrichData(source = 'thomson-reuters') {
        if (this.isEnriching) {
            console.warn('Enrichment already in progress');
            return;
        }

        this.isEnriching = true;

        try {
            // Map source names to API endpoints
            const sourceEndpointMap = {
                'thomson-reuters': '/api/v1/thomson-reuters',
                'supported': '/api/v1/supported',
                'industry': '/api/v1/industry',
                'query': '/api/v1/query'
            };

            // Get endpoint from map or construct dynamically
            let endpoint = sourceEndpointMap[source];
            if (!endpoint) {
                // Fallback: construct endpoint from source name
                // Convert kebab-case or camelCase to kebab-case
                const normalizedSource = source.replace(/([A-Z])/g, '-$1').toLowerCase();
                endpoint = `/api/v1/${normalizedSource}`;
            }

            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getAuthToken()}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    merchant_id: this.merchantId,
                    source: source
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.enrichmentData = await response.json();
            return this.enrichmentData;
        } catch (error) {
            console.error('Error enriching data:', error);
            throw error;
        } finally {
            this.isEnriching = false;
        }
    }

    /**
     * Get enrichment status
     */
    getEnrichmentStatus() {
        return {
            isEnriching: this.isEnriching,
            hasData: this.enrichmentData !== null,
            data: this.enrichmentData
        };
    }

    /**
     * Get authentication token
     */
    getAuthToken() {
        const token = localStorage.getItem('auth_token') || localStorage.getItem('access_token');
        if (token) {
            return token;
        }

        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'auth_token' || name === 'access_token') {
                return value;
            }
        }

        return null;
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.DataEnrichment = DataEnrichment;
}

