/**
 * Unit Tests for Merchant Portfolio JavaScript
 * Tests portfolio management functionality including search, filtering, bulk operations, and pagination
 */

// Mock DOM elements and fetch for testing
const mockDOM = {
    portfolioSearchInput: { addEventListener: jest.fn(), value: '' },
    portfolioTypeFilter: { addEventListener: jest.fn(), value: '' },
    riskLevelFilter: { addEventListener: jest.fn(), value: '' },
    industryFilter: { addEventListener: jest.fn(), value: '' },
    bulkSelectBtn: { addEventListener: jest.fn(), innerHTML: '', classList: { add: jest.fn(), remove: jest.fn() } },
    exportPortfolioBtn: { addEventListener: jest.fn() },
    addMerchantBtn: { addEventListener: jest.fn() },
    refreshBtn: { addEventListener: jest.fn() },
    prevPageBtn: { addEventListener: jest.fn(), disabled: false },
    nextPageBtn: { addEventListener: jest.fn(), disabled: false },
    bulkEditBtn: { addEventListener: jest.fn() },
    bulkExportBtn: { addEventListener: jest.fn() },
    bulkCompareBtn: { addEventListener: jest.fn() },
    clearSelectionBtn: { addEventListener: jest.fn() },
    merchantsGrid: { innerHTML: '' },
    totalMerchants: { textContent: '' },
    activeMerchants: { textContent: '' },
    pendingMerchants: { textContent: '' },
    pagination: { style: { display: '' } },
    paginationInfo: { textContent: '' },
    bulkSelection: { classList: { add: jest.fn(), remove: jest.fn() } },
    bulkCount: { textContent: '' }
};

// Mock global fetch
global.fetch = jest.fn();

// Mock document methods
global.document = {
    getElementById: jest.fn((id) => mockDOM[id]),
    querySelectorAll: jest.fn(() => []),
    addEventListener: jest.fn(),
    createElement: jest.fn(() => ({
        href: '',
        download: '',
        click: jest.fn()
    })),
    body: {
        appendChild: jest.fn(),
        removeChild: jest.fn()
    }
};

// Mock window methods
global.window = {
    location: { href: '' },
    URL: {
        createObjectURL: jest.fn(() => 'mock-url'),
        revokeObjectURL: jest.fn()
    }
};

// Mock alert
global.alert = jest.fn();

// Mock setTimeout and clearTimeout
global.setTimeout = jest.fn((fn) => {
    fn();
    return 1;
});
global.clearTimeout = jest.fn();

describe('MerchantPortfolio', () => {
    let merchantPortfolio;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        
        // Mock successful API response
        global.fetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({
                merchants: [
                    {
                        id: '1',
                        name: 'Test Company',
                        industry: 'Technology',
                        portfolio_type: 'onboarded',
                        risk_level: 'low',
                        address: { city: 'San Francisco', state: 'CA' },
                        founded_date: '2020-01-01',
                        employee_count: 50,
                        annual_revenue: 1000000
                    }
                ],
                total: 1,
                total_pages: 1
            })
        });

        // Create new instance
        merchantPortfolio = new MerchantPortfolio();
    });

    describe('Initialization', () => {
        test('should initialize with default values', () => {
            expect(merchantPortfolio.apiBaseUrl).toBe('/api/v1');
            expect(merchantPortfolio.currentPage).toBe(1);
            expect(merchantPortfolio.pageSize).toBe(20);
            expect(merchantPortfolio.merchants).toEqual([]);
            expect(merchantPortfolio.selectedMerchants).toBeInstanceOf(Set);
            expect(merchantPortfolio.bulkMode).toBe(false);
        });

        test('should bind all event listeners', () => {
            expect(mockDOM.portfolioSearchInput.addEventListener).toHaveBeenCalled();
            expect(mockDOM.portfolioTypeFilter.addEventListener).toHaveBeenCalled();
            expect(mockDOM.riskLevelFilter.addEventListener).toHaveBeenCalled();
            expect(mockDOM.industryFilter.addEventListener).toHaveBeenCalled();
            expect(mockDOM.bulkSelectBtn.addEventListener).toHaveBeenCalled();
            expect(mockDOM.exportPortfolioBtn.addEventListener).toHaveBeenCalled();
            expect(mockDOM.addMerchantBtn.addEventListener).toHaveBeenCalled();
            expect(mockDOM.refreshBtn.addEventListener).toHaveBeenCalled();
        });

        test('should load merchants on initialization', () => {
            expect(global.fetch).toHaveBeenCalled();
        });
    });

    describe('Search and Filtering', () => {
        test('should handle search input changes', () => {
            const searchInput = { target: { value: 'test search' } };
            merchantPortfolio.currentFilters.searchQuery = '';
            
            // Simulate input event
            const inputHandler = mockDOM.portfolioSearchInput.addEventListener.mock.calls[0][1];
            inputHandler(searchInput);
            
            expect(merchantPortfolio.currentFilters.searchQuery).toBe('test search');
        });

        test('should handle portfolio type filter changes', () => {
            const filterInput = { target: { value: 'onboarded' } };
            merchantPortfolio.currentFilters.portfolioType = '';
            
            // Simulate change event
            const changeHandler = mockDOM.portfolioTypeFilter.addEventListener.mock.calls[0][1];
            changeHandler(filterInput);
            
            expect(merchantPortfolio.currentFilters.portfolioType).toBe('onboarded');
        });

        test('should handle risk level filter changes', () => {
            const filterInput = { target: { value: 'high' } };
            merchantPortfolio.currentFilters.riskLevel = '';
            
            // Simulate change event
            const changeHandler = mockDOM.riskLevelFilter.addEventListener.mock.calls[0][1];
            changeHandler(filterInput);
            
            expect(merchantPortfolio.currentFilters.riskLevel).toBe('high');
        });

        test('should handle industry filter changes', () => {
            const filterInput = { target: { value: 'technology' } };
            merchantPortfolio.currentFilters.industry = '';
            
            // Simulate change event
            const changeHandler = mockDOM.industryFilter.addEventListener.mock.calls[0][1];
            changeHandler(filterInput);
            
            expect(merchantPortfolio.currentFilters.industry).toBe('technology');
        });

        test('should debounce search requests', () => {
            merchantPortfolio.debouncedSearch();
            expect(global.setTimeout).toHaveBeenCalled();
        });
    });

    describe('API Integration', () => {
        test('should load merchants with correct parameters', async () => {
            await merchantPortfolio.loadMerchants();
            
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/v1/merchants')
            );
        });

        test('should handle API errors gracefully', async () => {
            global.fetch.mockRejectedValueOnce(new Error('API Error'));
            
            await merchantPortfolio.loadMerchants();
            
            expect(mockDOM.merchantsGrid.innerHTML).toContain('Error Loading Merchants');
        });

        test('should show loading state', () => {
            merchantPortfolio.showLoadingState();
            expect(mockDOM.merchantsGrid.innerHTML).toContain('Loading merchants');
        });

        test('should show error state', () => {
            merchantPortfolio.showErrorState('Test error message');
            expect(mockDOM.merchantsGrid.innerHTML).toContain('Test error message');
        });
    });

    describe('Merchant Card Rendering', () => {
        test('should create merchant card HTML', () => {
            const merchant = {
                id: '1',
                name: 'Test Company',
                industry: 'Technology',
                portfolio_type: 'onboarded',
                risk_level: 'low',
                address: { city: 'San Francisco', state: 'CA' },
                founded_date: '2020-01-01',
                employee_count: 50,
                annual_revenue: 1000000
            };

            const cardHTML = merchantPortfolio.createMerchantCard(merchant);
            
            expect(cardHTML).toContain('Test Company');
            expect(cardHTML).toContain('Technology');
            expect(cardHTML).toContain('San Francisco');
            expect(cardHTML).toContain('data-merchant-id="1"');
        });

        test('should handle missing merchant data gracefully', () => {
            const merchant = {
                id: '1',
                name: 'Test Company'
            };

            const cardHTML = merchantPortfolio.createMerchantCard(merchant);
            
            expect(cardHTML).toContain('Test Company');
            expect(cardHTML).toContain('Not specified');
            expect(cardHTML).toContain('Unknown');
        });

        test('should escape HTML in merchant data', () => {
            const merchant = {
                id: '1',
                name: '<script>alert("xss")</script>',
                industry: 'Technology'
            };

            const cardHTML = merchantPortfolio.createMerchantCard(merchant);
            
            expect(cardHTML).not.toContain('<script>');
            expect(cardHTML).toContain('&lt;script&gt;');
        });
    });

    describe('Bulk Operations', () => {
        test('should toggle bulk mode', () => {
            expect(merchantPortfolio.bulkMode).toBe(false);
            
            merchantPortfolio.toggleBulkMode();
            expect(merchantPortfolio.bulkMode).toBe(true);
            
            merchantPortfolio.toggleBulkMode();
            expect(merchantPortfolio.bulkMode).toBe(false);
        });

        test('should handle card selection in bulk mode', () => {
            merchantPortfolio.bulkMode = true;
            merchantPortfolio.selectedMerchants.clear();
            
            const mockCard = {
                dataset: { merchantId: '1' },
                classList: { remove: jest.fn(), add: jest.fn() }
            };
            
            const mockEvent = { currentTarget: mockCard };
            
            merchantPortfolio.handleCardClick(mockEvent);
            
            expect(merchantPortfolio.selectedMerchants.has('1')).toBe(true);
            expect(mockCard.classList.add).toHaveBeenCalledWith('selected');
        });

        test('should clear selection', () => {
            merchantPortfolio.selectedMerchants.add('1');
            merchantPortfolio.selectedMerchants.add('2');
            
            merchantPortfolio.clearSelection();
            
            expect(merchantPortfolio.selectedMerchants.size).toBe(0);
        });

        test('should validate bulk compare requires exactly 2 merchants', () => {
            merchantPortfolio.selectedMerchants.add('1');
            
            merchantPortfolio.bulkCompare();
            
            expect(global.alert).toHaveBeenCalledWith('Please select exactly 2 merchants to compare.');
        });

        test('should allow bulk compare with exactly 2 merchants', () => {
            merchantPortfolio.selectedMerchants.add('1');
            merchantPortfolio.selectedMerchants.add('2');
            
            merchantPortfolio.bulkCompare();
            
            expect(global.window.location.href).toContain('merchant-comparison.html');
        });
    });

    describe('Navigation', () => {
        test('should navigate to merchant detail view', () => {
            merchantPortfolio.viewMerchant('123');
            expect(global.window.location.href).toBe('merchant-detail.html?id=123');
        });

        test('should navigate to merchant edit view', () => {
            merchantPortfolio.editMerchant('123');
            expect(global.window.location.href).toBe('merchant-edit.html?id=123');
        });

        test('should navigate to merchant comparison view', () => {
            merchantPortfolio.compareMerchant('123');
            expect(global.window.location.href).toBe('merchant-comparison.html?merchant1=123');
        });

        test('should navigate to add merchant page', () => {
            merchantPortfolio.addMerchant();
            expect(global.window.location.href).toBe('merchant-add.html');
        });
    });

    describe('Export Functionality', () => {
        test('should generate CSV content', () => {
            const merchants = [
                {
                    name: 'Test Company',
                    industry: 'Technology',
                    portfolio_type: 'onboarded',
                    risk_level: 'low',
                    address: { city: 'San Francisco', state: 'CA' },
                    founded_date: '2020-01-01',
                    employee_count: 50,
                    annual_revenue: 1000000
                }
            ];

            const csvContent = merchantPortfolio.generateCSV(merchants);
            
            expect(csvContent).toContain('Test Company');
            expect(csvContent).toContain('Technology');
            expect(csvContent).toContain('Onboarded');
            expect(csvContent).toContain('Low Risk');
        });

        test('should handle missing data in CSV generation', () => {
            const merchants = [
                {
                    name: 'Test Company'
                }
            ];

            const csvContent = merchantPortfolio.generateCSV(merchants);
            
            expect(csvContent).toContain('Test Company');
            expect(csvContent).toContain('N/A');
            expect(csvContent).toContain('Unknown');
        });

        test('should download CSV file', () => {
            const content = 'test,content';
            const filename = 'test.csv';
            
            merchantPortfolio.downloadCSV(content, filename);
            
            expect(global.document.createElement).toHaveBeenCalledWith('a');
            expect(global.window.URL.createObjectURL).toHaveBeenCalled();
        });
    });

    describe('Data Formatting', () => {
        test('should format portfolio types correctly', () => {
            expect(merchantPortfolio.formatPortfolioType('onboarded')).toBe('Onboarded');
            expect(merchantPortfolio.formatPortfolioType('deactivated')).toBe('Deactivated');
            expect(merchantPortfolio.formatPortfolioType('prospective')).toBe('Prospective');
            expect(merchantPortfolio.formatPortfolioType('pending')).toBe('Pending');
            expect(merchantPortfolio.formatPortfolioType('unknown')).toBe('unknown');
        });

        test('should format risk levels correctly', () => {
            expect(merchantPortfolio.formatRiskLevel('low')).toBe('Low Risk');
            expect(merchantPortfolio.formatRiskLevel('medium')).toBe('Medium Risk');
            expect(merchantPortfolio.formatRiskLevel('high')).toBe('High Risk');
            expect(merchantPortfolio.formatRiskLevel('unknown')).toBe('unknown');
        });

        test('should escape HTML correctly', () => {
            expect(merchantPortfolio.escapeHtml('<script>alert("xss")</script>'))
                .toBe('&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;');
            expect(merchantPortfolio.escapeHtml('Normal text')).toBe('Normal text');
        });
    });

    describe('Statistics and Pagination', () => {
        test('should update statistics', () => {
            merchantPortfolio.merchants = [
                { portfolio_type: 'onboarded' },
                { portfolio_type: 'pending' },
                { portfolio_type: 'deactivated' }
            ];
            merchantPortfolio.totalCount = 3;
            
            merchantPortfolio.updateStats();
            
            expect(mockDOM.totalMerchants.textContent).toBe('3');
            expect(mockDOM.activeMerchants.textContent).toBe('1');
            expect(mockDOM.pendingMerchants.textContent).toBe('1');
        });

        test('should update pagination controls', () => {
            merchantPortfolio.totalPages = 3;
            merchantPortfolio.currentPage = 2;
            
            merchantPortfolio.updatePagination();
            
            expect(mockDOM.pagination.style.display).toBe('flex');
            expect(mockDOM.paginationInfo.textContent).toBe('Page 2 of 3');
            expect(mockDOM.prevPageBtn.disabled).toBe(false);
            expect(mockDOM.nextPageBtn.disabled).toBe(false);
        });

        test('should hide pagination for single page', () => {
            merchantPortfolio.totalPages = 1;
            
            merchantPortfolio.updatePagination();
            
            expect(mockDOM.pagination.style.display).toBe('none');
        });
    });

    describe('Utility Methods', () => {
        test('should get current filters', () => {
            const filters = merchantPortfolio.getCurrentFilters();
            expect(filters).toEqual(merchantPortfolio.currentFilters);
        });

        test('should set filters programmatically', () => {
            const newFilters = { searchQuery: 'test', portfolioType: 'onboarded' };
            merchantPortfolio.setFilters(newFilters);
            
            expect(merchantPortfolio.currentFilters.searchQuery).toBe('test');
            expect(merchantPortfolio.currentFilters.portfolioType).toBe('onboarded');
        });

        test('should get selected merchants', () => {
            merchantPortfolio.selectedMerchants.add('1');
            merchantPortfolio.selectedMerchants.add('2');
            
            const selected = merchantPortfolio.getSelectedMerchants();
            
            expect(selected).toEqual(['1', '2']);
        });

        test('should set selected merchants programmatically', () => {
            merchantPortfolio.setSelectedMerchants(['1', '2', '3']);
            
            expect(merchantPortfolio.selectedMerchants.has('1')).toBe(true);
            expect(merchantPortfolio.selectedMerchants.has('2')).toBe(true);
            expect(merchantPortfolio.selectedMerchants.has('3')).toBe(true);
        });

        test('should reset all filters and pagination', () => {
            merchantPortfolio.currentFilters.searchQuery = 'test';
            merchantPortfolio.currentPage = 5;
            merchantPortfolio.selectedMerchants.add('1');
            
            merchantPortfolio.reset();
            
            expect(merchantPortfolio.currentFilters.searchQuery).toBe('');
            expect(merchantPortfolio.currentPage).toBe(1);
            expect(merchantPortfolio.selectedMerchants.size).toBe(0);
        });
    });

    describe('Error Handling', () => {
        test('should handle missing DOM elements gracefully', () => {
            global.document.getElementById.mockReturnValue(null);
            
            // Should not throw errors
            expect(() => {
                merchantPortfolio.showLoadingState();
                merchantPortfolio.updateStats();
                merchantPortfolio.updatePagination();
            }).not.toThrow();
        });

        test('should handle API response without expected fields', async () => {
            global.fetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({})
            });
            
            await merchantPortfolio.loadMerchants();
            
            expect(merchantPortfolio.merchants).toEqual([]);
            expect(merchantPortfolio.totalCount).toBe(0);
            expect(merchantPortfolio.totalPages).toBe(0);
        });
    });
});

// Integration tests
describe('MerchantPortfolio Integration', () => {
    test('should handle complete workflow from search to export', async () => {
        const merchantPortfolio = new MerchantPortfolio();
        
        // Set search query
        merchantPortfolio.setFilters({ searchQuery: 'test' });
        
        // Add merchants to selection
        merchantPortfolio.setSelectedMerchants(['1', '2']);
        
        // Export selected merchants
        const merchants = [
            { id: '1', name: 'Test 1', industry: 'Tech' },
            { id: '2', name: 'Test 2', industry: 'Finance' }
        ];
        merchantPortfolio.merchants = merchants;
        
        const csvContent = merchantPortfolio.generateCSV(merchants);
        expect(csvContent).toContain('Test 1');
        expect(csvContent).toContain('Test 2');
    });
});
