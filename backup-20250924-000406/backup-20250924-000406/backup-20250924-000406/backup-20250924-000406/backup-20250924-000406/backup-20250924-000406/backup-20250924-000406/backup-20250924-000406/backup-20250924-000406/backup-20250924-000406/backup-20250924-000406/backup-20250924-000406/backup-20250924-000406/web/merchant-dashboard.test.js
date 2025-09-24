/**
 * Unit Tests for Merchant Dashboard JavaScript
 * Tests dashboard functionality, real-time updates, and data visualization
 */

// Mock dependencies
const mockSessionManager = {
    init: jest.fn(),
    startSession: jest.fn(),
    endSession: jest.fn(),
    switchSession: jest.fn()
};

const mockRiskIndicator = {
    init: jest.fn(),
    updateRiskLevel: jest.fn()
};

const mockComingSoonBanner = {
    init: jest.fn(),
    showFeature: jest.fn(),
    hideFeature: jest.fn()
};

const mockMockDataWarning = {
    init: jest.fn(),
    show: jest.fn(),
    hide: jest.fn()
};

// Mock Chart.js
global.Chart = jest.fn().mockImplementation(() => ({
    update: jest.fn(),
    destroy: jest.fn()
}));

// Mock fetch
global.fetch = jest.fn();

// Mock URLSearchParams
global.URLSearchParams = jest.fn().mockImplementation((search) => ({
    get: jest.fn().mockReturnValue('demo-merchant-001')
}));

// Mock URL.createObjectURL and URL.revokeObjectURL
global.URL = {
    createObjectURL: jest.fn().mockReturnValue('mock-url'),
    revokeObjectURL: jest.fn()
};

describe('MerchantDashboard', () => {
    let dashboard;
    let mockContainer;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        
        // Create mock DOM elements
        mockContainer = {
            innerHTML: '',
            appendChild: jest.fn(),
            insertBefore: jest.fn(),
            querySelector: jest.fn(),
            querySelectorAll: jest.fn().mockReturnValue([]),
            getElementById: jest.fn()
        };

        // Mock document methods
        global.document = {
            getElementById: jest.fn(),
            querySelector: jest.fn(),
            querySelectorAll: jest.fn().mockReturnValue([]),
            createElement: jest.fn().mockReturnValue({
                className: '',
                innerHTML: '',
                textContent: '',
                addEventListener: jest.fn(),
                click: jest.fn(),
                appendChild: jest.fn(),
                removeChild: jest.fn()
            }),
            body: {
                appendChild: jest.fn(),
                removeChild: jest.fn()
            },
            addEventListener: jest.fn()
        };

        // Mock window methods
        global.window = {
            location: {
                href: 'merchant-detail.html?id=demo-merchant-001',
                search: '?id=demo-merchant-001'
            },
            addEventListener: jest.fn(),
            removeEventListener: jest.fn()
        };

        // Mock global objects
        global.clearInterval = jest.fn();
        global.setInterval = jest.fn().mockReturnValue(123);
    });

    afterEach(() => {
        if (dashboard) {
            dashboard.destroy();
        }
    });

    describe('Initialization', () => {
        test('should initialize with default merchant ID', () => {
            dashboard = new MerchantDashboard();
            expect(dashboard.merchantId).toBe('demo-merchant-001');
            expect(dashboard.apiBaseUrl).toBe('/api/v1');
            expect(dashboard.realTimeUpdates).toBe(true);
            expect(dashboard.updateFrequency).toBe(30000);
        });

        test('should initialize components', () => {
            dashboard = new MerchantDashboard();
            expect(dashboard.sessionManager).toBeDefined();
            expect(dashboard.riskIndicator).toBeDefined();
            expect(dashboard.comingSoonBanner).toBeDefined();
            expect(dashboard.mockDataWarning).toBeDefined();
        });

        test('should bind events', () => {
            dashboard = new MerchantDashboard();
            expect(global.window.addEventListener).toHaveBeenCalledWith('popstate', expect.any(Function));
            expect(global.window.addEventListener).toHaveBeenCalledWith('focus', expect.any(Function));
            expect(global.window.addEventListener).toHaveBeenCalledWith('blur', expect.any(Function));
        });
    });

    describe('Data Loading', () => {
        test('should load merchant data successfully', async () => {
            const mockMerchant = {
                id: 'demo-merchant-001',
                name: 'Test Merchant',
                portfolio_type: 'onboarded',
                risk_level: 'medium'
            };

            global.fetch.mockResolvedValueOnce({
                ok: true,
                json: jest.fn().mockResolvedValue(mockMerchant)
            });

            dashboard = new MerchantDashboard();
            await dashboard.loadMerchantData();

            expect(global.fetch).toHaveBeenCalledWith('/api/v1/merchants/demo-merchant-001');
            expect(dashboard.merchant).toEqual(mockMerchant);
        });

        test('should handle API errors gracefully', async () => {
            global.fetch.mockRejectedValueOnce(new Error('API Error'));

            dashboard = new MerchantDashboard();
            await dashboard.loadMerchantData();

            expect(dashboard.merchant).toBeDefined(); // Should load mock data
        });

        test('should load mock data when API fails', async () => {
            global.fetch.mockRejectedValueOnce(new Error('API Error'));

            dashboard = new MerchantDashboard();
            await dashboard.loadMerchantData();

            expect(dashboard.merchant.name).toBe('Acme Corporation');
            expect(dashboard.merchant.portfolio_type).toBe('onboarded');
        });
    });

    describe('Real-time Updates', () => {
        test('should start real-time updates', () => {
            dashboard = new MerchantDashboard();
            dashboard.startRealTimeUpdates();

            expect(global.setInterval).toHaveBeenCalledWith(
                expect.any(Function),
                30000
            );
        });

        test('should stop real-time updates', () => {
            dashboard = new MerchantDashboard();
            dashboard.refreshInterval = 123;
            dashboard.stopRealTimeUpdates();

            expect(global.clearInterval).toHaveBeenCalledWith(123);
            expect(dashboard.refreshInterval).toBeNull();
        });

        test('should toggle real-time updates', () => {
            dashboard = new MerchantDashboard();
            dashboard.toggleRealTimeUpdates(false);

            expect(dashboard.realTimeUpdates).toBe(false);
            expect(global.clearInterval).toHaveBeenCalled();
        });
    });

    describe('Data Rendering', () => {
        beforeEach(() => {
            dashboard = new MerchantDashboard();
            dashboard.merchant = {
                id: 'test-001',
                name: 'Test Merchant',
                legal_name: 'Test Merchant Inc.',
                portfolio_type: 'onboarded',
                risk_level: 'medium',
                industry: 'Technology',
                business_type: 'Corporation',
                registration_number: 'REG-001',
                tax_id: 'TAX-001',
                industry_code: 'NAICS 541511',
                annual_revenue: 1000000,
                employee_count: 25,
                founded_date: '2020-01-01',
                compliance_status: 'Compliant',
                updated_at: new Date().toISOString(),
                address: {
                    street1: '123 Test St',
                    city: 'Test City',
                    state: 'TS',
                    postal_code: '12345',
                    country: 'USA'
                },
                contact_info: {
                    phone: '+1-555-123-4567',
                    email: 'test@test.com',
                    website: 'https://test.com',
                    primary_contact: 'John Doe'
                }
            };
        });

        test('should update merchant header', () => {
            const mockElement = { textContent: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockElement);

            dashboard.updateMerchantHeader();

            expect(mockElement.textContent).toBe('Test Merchant');
        });

        test('should update business information', () => {
            const mockElement = { textContent: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockElement);

            dashboard.updateBusinessInfo();

            expect(mockElement.textContent).toBe('REG-001');
        });

        test('should update contact information', () => {
            const mockElement = { textContent: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockElement);

            dashboard.updateContactInfo();

            expect(mockElement.textContent).toBe('+1-555-123-4567');
        });

        test('should update risk assessment', () => {
            const mockElement = { textContent: '', className: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockElement);

            dashboard.updateRiskAssessment();

            expect(mockElement.textContent).toBe('medium');
            expect(mockElement.className).toBe('status-badge risk-medium');
        });
    });

    describe('Activity Timeline', () => {
        test('should render activity timeline', () => {
            dashboard = new MerchantDashboard();
            const mockTimeline = { innerHTML: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockTimeline);

            const activities = [
                {
                    timestamp: '2024-01-01T00:00:00Z',
                    title: 'Test Activity',
                    description: 'Test Description'
                }
            ];

            dashboard.renderActivityTimeline(activities);

            expect(mockTimeline.innerHTML).toContain('Test Activity');
            expect(mockTimeline.innerHTML).toContain('Test Description');
        });

        test('should render empty timeline when no activities', () => {
            dashboard = new MerchantDashboard();
            const mockTimeline = { innerHTML: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockTimeline);

            dashboard.renderActivityTimeline([]);

            expect(mockTimeline.innerHTML).toContain('No activity recorded');
        });

        test('should render mock activity timeline', () => {
            dashboard = new MerchantDashboard();
            const mockTimeline = { innerHTML: '' };
            global.document.getElementById = jest.fn().mockReturnValue(mockTimeline);

            dashboard.renderMockActivityTimeline();

            expect(mockTimeline.innerHTML).toContain('Merchant Profile Updated');
        });
    });

    describe('Charts', () => {
        test('should initialize charts', () => {
            dashboard = new MerchantDashboard();
            dashboard.initializeCharts();

            expect(global.Chart).toHaveBeenCalled();
        });

        test('should update charts', () => {
            dashboard = new MerchantDashboard();
            dashboard.charts = {
                riskTrend: { update: jest.fn() }
            };

            dashboard.updateCharts();

            expect(dashboard.charts.riskTrend.update).toHaveBeenCalled();
        });

        test('should generate mock risk data', () => {
            dashboard = new MerchantDashboard();
            const riskData = dashboard.generateMockRiskData();

            expect(riskData.labels).toHaveLength(6);
            expect(riskData.values).toHaveLength(6);
            expect(riskData.values.every(v => v >= 0.2 && v <= 0.7)).toBe(true);
        });
    });

    describe('Export Functionality', () => {
        test('should export merchant report', () => {
            dashboard = new MerchantDashboard();
            dashboard.merchant = { id: 'test-001', name: 'Test Merchant' };

            const mockLink = {
                href: '',
                download: '',
                click: jest.fn()
            };
            global.document.createElement = jest.fn().mockReturnValue(mockLink);

            dashboard.exportReport();

            expect(global.document.createElement).toHaveBeenCalledWith('a');
            expect(mockLink.download).toContain('merchant-dashboard-report');
            expect(mockLink.click).toHaveBeenCalled();
        });
    });

    describe('Session Management', () => {
        test('should handle session start', () => {
            dashboard = new MerchantDashboard();
            const mockSession = { merchantId: 'test-001' };

            dashboard.onSessionStart(mockSession);

            expect(dashboard.updateSessionIndicator).toHaveBeenCalledWith(true);
        });

        test('should handle session end', () => {
            dashboard = new MerchantDashboard();

            dashboard.onSessionEnd();

            expect(dashboard.updateSessionIndicator).toHaveBeenCalledWith(false);
        });

        test('should handle session switch', () => {
            dashboard = new MerchantDashboard();
            dashboard.loadMerchantData = jest.fn();
            const mockSession = { merchantId: 'test-002' };

            dashboard.onSessionSwitch(mockSession);

            expect(dashboard.merchantId).toBe('test-002');
            expect(dashboard.loadMerchantData).toHaveBeenCalled();
        });
    });

    describe('Utility Functions', () => {
        test('should calculate next review date', () => {
            dashboard = new MerchantDashboard();
            const lastUpdate = new Date('2024-01-01').toISOString();
            const nextReview = dashboard.calculateNextReviewDate(lastUpdate);

            expect(nextReview).toContain('2025');
        });

        test('should handle null last update date', () => {
            dashboard = new MerchantDashboard();
            const nextReview = dashboard.calculateNextReviewDate(null);

            expect(nextReview).toBe('Not scheduled');
        });
    });

    describe('Cleanup', () => {
        test('should destroy dashboard properly', () => {
            dashboard = new MerchantDashboard();
            dashboard.charts = {
                test: { destroy: jest.fn() }
            };

            dashboard.destroy();

            expect(global.clearInterval).toHaveBeenCalled();
            expect(dashboard.charts.test.destroy).toHaveBeenCalled();
            expect(dashboard.charts).toEqual({});
        });
    });
});

describe('DashboardUtils', () => {
    describe('formatCurrency', () => {
        test('should format currency correctly', () => {
            const result = DashboardUtils.formatCurrency(1000);
            expect(result).toBe('$1,000.00');
        });
    });

    describe('formatDate', () => {
        test('should format date correctly', () => {
            const result = DashboardUtils.formatDate('2024-01-01');
            expect(result).toContain('January');
            expect(result).toContain('2024');
        });
    });

    describe('formatDateTime', () => {
        test('should format date and time correctly', () => {
            const result = DashboardUtils.formatDateTime('2024-01-01T12:00:00Z');
            expect(result).toContain('Jan');
            expect(result).toContain('2024');
        });
    });

    describe('calculateRiskColor', () => {
        test('should return correct colors for risk levels', () => {
            expect(DashboardUtils.calculateRiskColor('low')).toBe('#27ae60');
            expect(DashboardUtils.calculateRiskColor('medium')).toBe('#f39c12');
            expect(DashboardUtils.calculateRiskColor('high')).toBe('#e74c3c');
            expect(DashboardUtils.calculateRiskColor('unknown')).toBe('#95a5a6');
        });
    });

    describe('calculatePortfolioTypeColor', () => {
        test('should return correct colors for portfolio types', () => {
            expect(DashboardUtils.calculatePortfolioTypeColor('onboarded')).toBe('#27ae60');
            expect(DashboardUtils.calculatePortfolioTypeColor('deactivated')).toBe('#e74c3c');
            expect(DashboardUtils.calculatePortfolioTypeColor('prospective')).toBe('#f39c12');
            expect(DashboardUtils.calculatePortfolioTypeColor('pending')).toBe('#3498db');
            expect(DashboardUtils.calculatePortfolioTypeColor('unknown')).toBe('#95a5a6');
        });
    });
});
