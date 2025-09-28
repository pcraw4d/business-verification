/**
 * Unit tests for Merchant Comparison Component
 */

describe('MerchantComparison', () => {
    let merchantComparison;
    let mockMerchants;
    let mockFetch;

    beforeEach(() => {
        // Mock DOM elements
        document.body.innerHTML = `
            <div id="merchant1Select">
                <option value="">Choose a merchant...</option>
            </div>
            <div id="merchant2Select">
                <option value="">Choose a merchant...</option>
            </div>
            <div id="merchant1Search"></div>
            <div id="merchant2Search"></div>
            <div id="merchant1Preview" style="display: none;"></div>
            <div id="merchant2Preview" style="display: none;"></div>
            <div id="comparisonResults" style="display: none;"></div>
            <div id="emptyState"></div>
            <div id="loadingState" style="display: none;"></div>
            <div id="sideBySideView"></div>
            <div id="summaryView" style="display: none;"></div>
            <div id="keyDifferences"></div>
            <div id="recommendations"></div>
            <canvas id="riskComparisonChart"></canvas>
            <button id="exportReportBtn" disabled></button>
            <button id="clearComparisonBtn"></button>
            <button id="toggleViewBtn"></button>
            <button id="printReportBtn"></button>
        `;

        // Mock merchants data
        mockMerchants = [
            {
                id: 'merchant1',
                name: 'Test Merchant 1',
                portfolio_type: 'onboarded',
                risk_level: 'low',
                industry: 'Technology',
                employee_count: 50,
                annual_revenue: 1000000,
                address: {
                    street1: '123 Main St',
                    city: 'San Francisco',
                    state: 'CA',
                    country: 'USA'
                },
                contact_info: {
                    phone: '+1-555-123-4567',
                    email: 'contact@merchant1.com',
                    website: 'https://merchant1.com'
                }
            },
            {
                id: 'merchant2',
                name: 'Test Merchant 2',
                portfolio_type: 'prospective',
                risk_level: 'high',
                industry: 'Finance',
                employee_count: 200,
                annual_revenue: 5000000,
                address: {
                    street1: '456 Oak Ave',
                    city: 'New York',
                    state: 'NY',
                    country: 'USA'
                },
                contact_info: {
                    phone: '+1-555-987-6543',
                    email: 'contact@merchant2.com',
                    website: 'https://merchant2.com'
                }
            }
        ];

        // Mock fetch
        mockFetch = jest.fn();
        global.fetch = mockFetch;

        // Mock Chart.js
        global.Chart = jest.fn().mockImplementation(() => ({
            data: {
                datasets: [{ data: [0, 0] }]
            },
            update: jest.fn()
        }));

        // Create new instance
        merchantComparison = new MerchantComparison();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('Initialization', () => {
        test('should initialize with correct default values', () => {
            expect(merchantComparison.merchant1).toBeNull();
            expect(merchantComparison.merchant2).toBeNull();
            expect(merchantComparison.comparisonData).toBeNull();
            expect(merchantComparison.currentView).toBe('side-by-side');
        });

        test('should load merchant options on initialization', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => ({ merchants: mockMerchants })
            });

            await merchantComparison.loadMerchantOptions();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/merchants?page_size=1000');
        });
    });

    describe('Merchant Selection', () => {
        test('should handle merchant selection', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockMerchants[0]
            });

            await merchantComparison.handleMerchantSelection(1, 'merchant1');

            expect(merchantComparison.merchant1).toEqual(mockMerchants[0]);
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/merchants/merchant1');
        });

        test('should show merchant preview after selection', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockMerchants[0]
            });

            // Mock DOM elements for preview
            document.getElementById('merchant1Preview').innerHTML = `
                <div id="merchant1Name"></div>
                <div id="merchant1Type"></div>
                <div id="merchant1Industry"></div>
                <div id="merchant1Location"></div>
            `;

            await merchantComparison.handleMerchantSelection(1, 'merchant1');

            const preview = document.getElementById('merchant1Preview');
            expect(preview.style.display).toBe('block');
        });

        test('should handle merchant selection error', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Network error'));

            await merchantComparison.handleMerchantSelection(1, 'merchant1');

            expect(merchantComparison.merchant1).toBeNull();
        });
    });

    describe('Comparison Logic', () => {
        beforeEach(() => {
            merchantComparison.merchant1 = mockMerchants[0];
            merchantComparison.merchant2 = mockMerchants[1];
        });

        test('should find differences between merchants', () => {
            const differences = merchantComparison.findDifferences();

            expect(differences).toHaveLength(5); // portfolio_type, risk_level, industry, employee_count, annual_revenue
            expect(differences[0].category).toBe('Portfolio Type');
            expect(differences[0].merchant1).toBe('onboarded');
            expect(differences[0].merchant2).toBe('prospective');
        });

        test('should compare risk levels', () => {
            const riskComparison = merchantComparison.compareRiskLevels();

            expect(riskComparison.merchant1.level).toBe('low');
            expect(riskComparison.merchant1.value).toBe(1);
            expect(riskComparison.merchant2.level).toBe('high');
            expect(riskComparison.merchant2.value).toBe(3);
        });

        test('should generate recommendations', () => {
            const recommendations = merchantComparison.generateRecommendations();

            expect(recommendations.length).toBeGreaterThan(0);
            expect(recommendations[0]).toHaveProperty('type');
            expect(recommendations[0]).toHaveProperty('priority');
            expect(recommendations[0]).toHaveProperty('message');
            expect(recommendations[0]).toHaveProperty('action');
        });

        test('should generate comparison summary', () => {
            const summary = merchantComparison.generateSummary();

            expect(summary).toHaveProperty('totalDifferences');
            expect(summary).toHaveProperty('highImpactDifferences');
            expect(summary).toHaveProperty('riskAdvantage');
            expect(summary).toHaveProperty('portfolioAdvantage');
        });
    });

    describe('Risk Assessment', () => {
        test('should calculate risk score correctly', () => {
            expect(merchantComparison.calculateRiskScore('low')).toBe(0.2);
            expect(merchantComparison.calculateRiskScore('medium')).toBe(0.5);
            expect(merchantComparison.calculateRiskScore('high')).toBe(0.8);
            expect(merchantComparison.calculateRiskScore('invalid')).toBe(0.5);
        });
    });

    describe('Data Formatting', () => {
        test('should format address correctly', () => {
            const address = {
                street1: '123 Main St',
                city: 'San Francisco',
                state: 'CA',
                country: 'USA'
            };

            const formatted = merchantComparison.formatAddress(address);
            expect(formatted).toBe('123 Main St, San Francisco, CA, USA');
        });

        test('should handle missing address', () => {
            const formatted = merchantComparison.formatAddress(null);
            expect(formatted).toBe('Address not available');
        });

        test('should format currency correctly', () => {
            expect(merchantComparison.formatCurrency(1000000)).toBe('$1,000,000');
            expect(merchantComparison.formatCurrency(null)).toBe('Not specified');
            expect(merchantComparison.formatCurrency(undefined)).toBe('Not specified');
        });
    });

    describe('View Management', () => {
        test('should toggle between side-by-side and summary views', () => {
            merchantComparison.currentView = 'side-by-side';
            
            merchantComparison.toggleView();
            expect(merchantComparison.currentView).toBe('summary');
            
            merchantComparison.toggleView();
            expect(merchantComparison.currentView).toBe('side-by-side');
        });
    });

    describe('Export Functionality', () => {
        beforeEach(() => {
            merchantComparison.merchant1 = mockMerchants[0];
            merchantComparison.merchant2 = mockMerchants[1];
            merchantComparison.comparisonData = {
                differences: [],
                riskComparison: {},
                recommendations: [],
                summary: {}
            };
        });

        test('should generate export report', () => {
            const report = merchantComparison.generateExportReport();

            expect(report).toHaveProperty('title');
            expect(report).toHaveProperty('generatedAt');
            expect(report).toHaveProperty('merchants');
            expect(report).toHaveProperty('comparison');
            expect(report).toHaveProperty('summary');
        });

        test('should handle export without comparison data', () => {
            merchantComparison.comparisonData = null;
            
            // Mock console.error to avoid test output
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            
            merchantComparison.exportComparisonReport();
            
            expect(consoleSpy).toHaveBeenCalled();
            consoleSpy.mockRestore();
        });
    });

    describe('Clear Functionality', () => {
        test('should clear comparison data', () => {
            merchantComparison.merchant1 = mockMerchants[0];
            merchantComparison.merchant2 = mockMerchants[1];
            merchantComparison.comparisonData = {};

            merchantComparison.clearComparison();

            expect(merchantComparison.merchant1).toBeNull();
            expect(merchantComparison.merchant2).toBeNull();
            expect(merchantComparison.comparisonData).toBeNull();
        });
    });

    describe('Error Handling', () => {
        test('should show error message', () => {
            merchantComparison.showError('Test error message');

            const errorElement = document.getElementById('errorMessage');
            expect(errorElement).toBeTruthy();
            expect(errorElement.textContent).toContain('Test error message');
        });

        test('should handle API errors gracefully', async () => {
            mockFetch.mockRejectedValueOnce(new Error('API Error'));

            await merchantComparison.loadMerchantOptions();

            // Should not throw error
            expect(true).toBe(true);
        });
    });

    describe('Filtering', () => {
        test('should filter merchant options based on search term', () => {
            // Mock select element with options
            const select = document.getElementById('merchant1Select');
            select.innerHTML = `
                <option value="">Choose a merchant...</option>
                <option value="merchant1" data-merchant='${JSON.stringify(mockMerchants[0])}'>Test Merchant 1 (onboarded)</option>
                <option value="merchant2" data-merchant='${JSON.stringify(mockMerchants[1])}'>Test Merchant 2 (prospective)</option>
            `;

            merchantComparison.filterMerchantOptions(1, 'Technology');

            const options = select.querySelectorAll('option');
            expect(options[1].style.display).toBe('block'); // Technology merchant
            expect(options[2].style.display).toBe('none'); // Finance merchant
        });
    });

    describe('Print Functionality', () => {
        test('should generate print content', () => {
            merchantComparison.merchant1 = mockMerchants[0];
            merchantComparison.merchant2 = mockMerchants[1];
            merchantComparison.comparisonData = {
                differences: [],
                riskComparison: {},
                recommendations: [
                    {
                        type: 'risk',
                        priority: 'high',
                        message: 'Test recommendation',
                        action: 'Test action'
                    }
                ],
                summary: {}
            };

            const printContent = merchantComparison.generatePrintContent();

            expect(printContent).toContain('Merchant Comparison Report');
            expect(printContent).toContain('Test Merchant 1');
            expect(printContent).toContain('Test Merchant 2');
            expect(printContent).toContain('Test recommendation');
        });
    });

    describe('Recommendation Icons', () => {
        test('should return correct icons for recommendation types', () => {
            expect(merchantComparison.getRecommendationIcon('risk')).toBe('exclamation-triangle');
            expect(merchantComparison.getRecommendationIcon('onboarding')).toBe('user-plus');
            expect(merchantComparison.getRecommendationIcon('diversification')).toBe('chart-pie');
            expect(merchantComparison.getRecommendationIcon('unknown')).toBe('info-circle');
        });
    });
});
