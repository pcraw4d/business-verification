/**
 * Unit tests for MerchantSearch component
 * Tests core functionality including search, filtering, and user interactions
 */

// Mock fetch for testing
global.fetch = jest.fn();

// Mock DOM methods
Object.defineProperty(window, 'localStorage', {
    value: {
        getItem: jest.fn(() => 'mock-token'),
        setItem: jest.fn(),
        removeItem: jest.fn(),
    },
    writable: true,
});

// Mock document methods
document.getElementById = jest.fn();
document.querySelector = jest.fn();
document.querySelectorAll = jest.fn();
document.addEventListener = jest.fn();

describe('MerchantSearch Component', () => {
    let merchantSearch;
    let mockContainer;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        fetch.mockClear();

        // Mock container element
        mockContainer = {
            innerHTML: '',
            appendChild: jest.fn(),
            removeChild: jest.fn(),
        };

        // Mock DOM elements
        const mockElement = {
            addEventListener: jest.fn(),
            style: {},
            disabled: false,
            value: '',
            textContent: '',
            innerHTML: '',
            classList: {
                add: jest.fn(),
                remove: jest.fn(),
                contains: jest.fn(),
                toggle: jest.fn(),
            },
            dataset: {},
            appendChild: jest.fn(),
            remove: jest.fn(),
        };

        document.getElementById.mockImplementation((id) => {
            const element = { ...mockElement, id };
            return element;
        });

        document.querySelector.mockImplementation(() => mockElement);
        document.querySelectorAll.mockImplementation(() => [mockElement]);

        // Create component instance
        merchantSearch = new MerchantSearch({
            container: mockContainer,
            apiBaseUrl: '/api/v1',
            debounceDelay: 100,
        });
    });

    afterEach(() => {
        if (merchantSearch) {
            merchantSearch.destroy();
        }
    });

    describe('Initialization', () => {
        test('should create search interface on init', () => {
            expect(mockContainer.innerHTML).toContain('merchant-search-container');
            expect(mockContainer.innerHTML).toContain('merchantSearchInput');
            expect(mockContainer.innerHTML).toContain('portfolioTypeFilter');
            expect(mockContainer.innerHTML).toContain('riskLevelFilter');
        });

        test('should set default values', () => {
            expect(merchantSearch.currentPage).toBe(1);
            expect(merchantSearch.pageSize).toBe(20);
            expect(merchantSearch.debounceDelay).toBe(100);
            expect(merchantSearch.isLoading).toBe(false);
        });

        test('should initialize with empty filters', () => {
            expect(merchantSearch.currentFilters).toEqual({
                searchQuery: '',
                portfolioType: null,
                riskLevel: null,
                industry: '',
                status: ''
            });
        });
    });

    describe('Search Functionality', () => {
        test('should perform search with correct parameters', async () => {
            const mockResponse = {
                merchants: [
                    {
                        id: '1',
                        name: 'Test Merchant',
                        industry: 'Technology',
                        portfolio_type: 'onboarded',
                        risk_level: 'low',
                        address: { city: 'San Francisco', state: 'CA' },
                        created_at: '2023-01-01T00:00:00Z'
                    }
                ],
                total: 1,
                total_pages: 1
            };

            fetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            });

            await merchantSearch.performSearch();

            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/v1/merchants'),
                expect.objectContaining({
                    method: 'GET',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer mock-token'
                    })
                })
            );
        });

        test('should handle search errors gracefully', async () => {
            fetch.mockRejectedValueOnce(new Error('Network error'));

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

            await merchantSearch.performSearch();

            expect(consoleSpy).toHaveBeenCalledWith(
                'Search error:',
                expect.any(Error)
            );

            consoleSpy.mockRestore();
        });

        test('should debounce search requests', () => {
            const performSearchSpy = jest.spyOn(merchantSearch, 'performSearch');
            const debouncedSearch = merchantSearch.debouncedSearch.bind(merchantSearch);

            // Call debounced search multiple times quickly
            debouncedSearch();
            debouncedSearch();
            debouncedSearch();

            // Should not call performSearch immediately
            expect(performSearchSpy).not.toHaveBeenCalled();

            // Wait for debounce delay
            setTimeout(() => {
                expect(performSearchSpy).toHaveBeenCalledTimes(1);
            }, 150);

            performSearchSpy.mockRestore();
        });
    });

    describe('Filter Management', () => {
        test('should update filters correctly', () => {
            const newFilters = {
                portfolioType: 'onboarded',
                riskLevel: 'low',
                industry: 'technology'
            };

            merchantSearch.setFilters(newFilters);

            expect(merchantSearch.currentFilters).toMatchObject({
                searchQuery: '',
                portfolioType: 'onboarded',
                riskLevel: 'low',
                industry: 'technology',
                status: ''
            });
        });

        test('should clear all filters', () => {
            // Set some filters first
            merchantSearch.currentFilters = {
                searchQuery: 'test',
                portfolioType: 'onboarded',
                riskLevel: 'low',
                industry: 'technology',
                status: 'active'
            };

            merchantSearch.clearAllFilters();

            expect(merchantSearch.currentFilters).toEqual({
                searchQuery: '',
                portfolioType: null,
                riskLevel: null,
                industry: '',
                status: ''
            });
        });
    });

    describe('Merchant Selection', () => {
        test('should select merchant correctly', () => {
            const mockMerchant = {
                id: '1',
                name: 'Test Merchant',
                industry: 'Technology'
            };

            const onMerchantSelectSpy = jest.fn();
            merchantSearch.onMerchantSelect = onMerchantSelectSpy;

            merchantSearch.selectMerchant(mockMerchant);

            expect(onMerchantSelectSpy).toHaveBeenCalledWith(mockMerchant);
        });

        test('should get selected merchant', () => {
            const mockMerchant = { id: '1', name: 'Test Merchant' };
            merchantSearch.merchants = [mockMerchant];

            // Mock selected element
            const mockSelectedElement = {
                dataset: { merchantId: '1' },
                classList: { contains: jest.fn(() => true) }
            };

            document.querySelector.mockReturnValueOnce(mockSelectedElement);

            const selected = merchantSearch.getSelectedMerchant();

            expect(selected).toEqual(mockMerchant);
        });
    });

    describe('Data Formatting', () => {
        test('should format portfolio types correctly', () => {
            expect(merchantSearch.formatPortfolioType('onboarded')).toBe('Onboarded');
            expect(merchantSearch.formatPortfolioType('deactivated')).toBe('Deactivated');
            expect(merchantSearch.formatPortfolioType('prospective')).toBe('Prospective');
            expect(merchantSearch.formatPortfolioType('pending')).toBe('Pending');
            expect(merchantSearch.formatPortfolioType('unknown')).toBe('unknown');
        });

        test('should format risk levels correctly', () => {
            expect(merchantSearch.formatRiskLevel('low')).toBe('Low Risk');
            expect(merchantSearch.formatRiskLevel('medium')).toBe('Medium Risk');
            expect(merchantSearch.formatRiskLevel('high')).toBe('High Risk');
            expect(merchantSearch.formatRiskLevel('unknown')).toBe('unknown');
        });

        test('should format dates correctly', () => {
            const dateString = '2023-01-15T10:30:00Z';
            const formatted = merchantSearch.formatDate(dateString);
            
            expect(formatted).toMatch(/Jan 15, 2023/);
        });
    });

    describe('CSV Export', () => {
        test('should generate CSV content correctly', () => {
            const mockMerchants = [
                {
                    name: 'Test Merchant',
                    industry: 'Technology',
                    portfolio_type: 'onboarded',
                    risk_level: 'low',
                    address: { city: 'San Francisco', state: 'CA' },
                    created_at: '2023-01-01T00:00:00Z'
                }
            ];

            const csv = merchantSearch.generateCSV(mockMerchants);

            expect(csv).toContain('Name,Industry,Portfolio Type,Risk Level,City,State,Created Date');
            expect(csv).toContain('"Test Merchant","Technology","Onboarded","Low Risk","San Francisco","CA"');
        });

        test('should handle empty merchant data in CSV', () => {
            const mockMerchants = [
                {
                    name: 'Test Merchant',
                    industry: null,
                    portfolio_type: 'onboarded',
                    risk_level: 'low',
                    address: { city: null, state: null },
                    created_at: '2023-01-01T00:00:00Z'
                }
            ];

            const csv = merchantSearch.generateCSV(mockMerchants);

            expect(csv).toContain('"Test Merchant","N/A","Onboarded","Low Risk","N/A","N/A"');
        });
    });

    describe('Authentication', () => {
        test('should get auth token from localStorage', () => {
            const token = merchantSearch.getAuthToken();
            expect(token).toBe('mock-token');
        });

        test('should handle missing auth token', () => {
            localStorage.getItem.mockReturnValueOnce(null);
            
            const token = merchantSearch.getAuthToken();
            expect(token).toBe('');
        });
    });

    describe('Error Handling', () => {
        test('should show error message', () => {
            const appendChildSpy = jest.spyOn(document.body, 'appendChild').mockImplementation();
            const removeSpy = jest.spyOn(document.body, 'removeChild').mockImplementation();

            merchantSearch.showError('Test error message');

            expect(appendChildSpy).toHaveBeenCalled();
            
            // Check that error element has correct properties
            const errorElement = appendChildSpy.mock.calls[0][0];
            expect(errorElement.textContent).toBe('Test error message');
            expect(errorElement.className).toBe('error-message');

            appendChildSpy.mockRestore();
            removeSpy.mockRestore();
        });
    });

    describe('Loading States', () => {
        test('should set loading state correctly', () => {
            const searchLoading = document.getElementById('searchLoading');
            const searchInput = document.getElementById('merchantSearchInput');

            merchantSearch.setLoading(true);

            expect(merchantSearch.isLoading).toBe(true);
            expect(searchLoading.style.display).toBe('block');
            expect(searchInput.disabled).toBe(true);

            merchantSearch.setLoading(false);

            expect(merchantSearch.isLoading).toBe(false);
            expect(searchLoading.style.display).toBe('none');
            expect(searchInput.disabled).toBe(false);
        });
    });

    describe('Component Lifecycle', () => {
        test('should destroy component correctly', () => {
            const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout').mockImplementation();
            
            merchantSearch.destroy();

            expect(clearTimeoutSpy).toHaveBeenCalled();
            expect(mockContainer.innerHTML).toBe('');

            clearTimeoutSpy.mockRestore();
        });

        test('should refresh search results', async () => {
            const performSearchSpy = jest.spyOn(merchantSearch, 'performSearch').mockResolvedValue();

            await merchantSearch.refresh();

            expect(performSearchSpy).toHaveBeenCalled();

            performSearchSpy.mockRestore();
        });
    });
});

// Integration tests
describe('MerchantSearch Integration', () => {
    test('should handle complete search workflow', async () => {
        const mockContainer = { innerHTML: '' };
        const mockResponse = {
            merchants: [],
            total: 0,
            total_pages: 0
        };

        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockResponse,
        });

        const merchantSearch = new MerchantSearch({
            container: mockContainer,
            apiBaseUrl: '/api/v1'
        });

        await merchantSearch.performSearch();

        expect(merchantSearch.merchants).toEqual([]);
        expect(merchantSearch.totalCount).toBe(0);
        expect(merchantSearch.totalPages).toBe(0);

        merchantSearch.destroy();
    });
});
