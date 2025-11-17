/**
 * Unit Tests for Merchant Navigation Component
 * Tests merchant navigation functionality including breadcrumb navigation and quick merchant switching
 */

// Mock dependencies
const mockSessionManager = {
    onSessionStart: null,
    onSessionEnd: null,
    onSessionSwitch: null,
    switchSession: jest.fn(),
    getCurrentSession: jest.fn(() => null),
    isActive: jest.fn(() => false)
};

// Mock fetch
global.fetch = jest.fn();

// Mock localStorage
const localStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn()
};
global.localStorage = localStorageMock;

// Mock DOM methods
document.addEventListener = jest.fn();
document.querySelector = jest.fn();
document.querySelectorAll = jest.fn();
document.getElementById = jest.fn();
document.createElement = jest.fn();
document.head = {
    appendChild: jest.fn()
};

// Mock window
global.window = {
    sessionManager: mockSessionManager,
    merchantNavigation: null,
    location: {
        pathname: '/merchant-portfolio.html'
    },
    innerWidth: 1024
};

describe('MerchantNavigation', () => {
    let merchantNavigation;
    let mockContainer;
    let mockElement;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        localStorageMock.getItem.mockReturnValue(null);
        localStorageMock.setItem.mockImplementation(() => {});
        
        // Mock DOM elements
        mockElement = {
            innerHTML: '',
            style: { display: 'block' },
            classList: {
                add: jest.fn(),
                remove: jest.fn(),
                contains: jest.fn(() => false),
                toggle: jest.fn()
            },
            addEventListener: jest.fn(),
            querySelector: jest.fn(),
            querySelectorAll: jest.fn(() => []),
            appendChild: jest.fn(),
            remove: jest.fn(),
            textContent: '',
            dataset: {}
        };

        mockContainer = {
            innerHTML: '',
            appendChild: jest.fn(),
            querySelector: jest.fn(() => mockElement),
            querySelectorAll: jest.fn(() => [mockElement])
        };

        // Mock document methods
        document.getElementById.mockImplementation((id) => {
            if (id === 'merchantNavigationContainer') {
                return mockContainer;
            }
            return mockElement;
        });

        document.querySelector.mockReturnValue(mockElement);
        document.querySelectorAll.mockReturnValue([mockElement]);
        document.createElement.mockReturnValue(mockElement);

        // Mock fetch response
        fetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({
                merchants: [
                    { id: '1', name: 'Test Merchant 1', industry: 'Technology' },
                    { id: '2', name: 'Test Merchant 2', industry: 'Finance' }
                ]
            })
        });
    });

    afterEach(() => {
        if (merchantNavigation) {
            merchantNavigation.destroy();
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });

            expect(merchantNavigation.container).toBe(mockContainer);
            expect(merchantNavigation.apiBaseUrl).toBe('/api/v1');
            expect(merchantNavigation.maxBreadcrumbItems).toBe(5);
            expect(merchantNavigation.quickSwitchLimit).toBe(8);
            expect(merchantNavigation.isInitialized).toBe(true);
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                container: mockContainer,
                apiBaseUrl: '/api/v2',
                maxBreadcrumbItems: 10,
                quickSwitchLimit: 12,
                sessionManager: mockSessionManager
            };

            merchantNavigation = new MerchantNavigation(customOptions);

            expect(merchantNavigation.apiBaseUrl).toBe('/api/v2');
            expect(merchantNavigation.maxBreadcrumbItems).toBe(10);
            expect(merchantNavigation.quickSwitchLimit).toBe(12);
            expect(merchantNavigation.sessionManager).toBe(mockSessionManager);
        });

        test('should create navigation interface', () => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });

            expect(mockContainer.innerHTML).toContain('merchant-navigation-container');
            expect(mockContainer.innerHTML).toContain('breadcrumb-navigation');
            expect(mockContainer.innerHTML).toContain('quick-switch-navigation');
            expect(mockContainer.innerHTML).toContain('navigation-history');
        });

        test('should bind events', () => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });

            // Check that event listeners are added
            expect(mockElement.addEventListener).toHaveBeenCalled();
        });
    });

    describe('Breadcrumb Management', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should add breadcrumb item', () => {
            const breadcrumbTrail = {
                children: [],
                appendChild: jest.fn(),
                querySelectorAll: jest.fn(() => [])
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'breadcrumbTrail') {
                    return breadcrumbTrail;
                }
                return mockElement;
            });

            merchantNavigation.addBreadcrumbItem('merchant-detail', 'Test Merchant', 'fas fa-building', 'merchant-123');

            expect(breadcrumbTrail.appendChild).toHaveBeenCalled();
        });

        test('should limit breadcrumb items', () => {
            const breadcrumbTrail = {
                children: Array(10).fill(mockElement),
                querySelectorAll: jest.fn(() => Array(10).fill(mockElement)),
                appendChild: jest.fn()
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'breadcrumbTrail') {
                    return breadcrumbTrail;
                }
                return mockElement;
            });

            // Mock remove method
            mockElement.remove = jest.fn();

            merchantNavigation.limitBreadcrumbItems();

            // Should not remove home and portfolio items
            expect(mockElement.remove).not.toHaveBeenCalled();
        });

        test('should handle breadcrumb click', () => {
            const mockBreadcrumbItem = {
                dataset: { page: 'merchant-detail', merchantId: 'merchant-123' },
                classList: { remove: jest.fn(), add: jest.fn() }
            };

            const mockBreadcrumbTrail = {
                querySelectorAll: jest.fn(() => [mockBreadcrumbItem])
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'breadcrumbTrail') {
                    return mockBreadcrumbTrail;
                }
                return mockElement;
            });

            merchantNavigation.handleBreadcrumbClick(mockBreadcrumbItem);

            expect(mockBreadcrumbItem.classList.add).toHaveBeenCalledWith('active');
        });

        test('should clear breadcrumb', () => {
            const breadcrumbTrail = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'breadcrumbTrail') {
                    return breadcrumbTrail;
                }
                return mockElement;
            });

            merchantNavigation.clearBreadcrumb();

            expect(breadcrumbTrail.innerHTML).toContain('breadcrumb-item home');
            expect(breadcrumbTrail.innerHTML).toContain('breadcrumb-item portfolio');
        });
    });

    describe('Quick Switch Management', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should load quick switch merchants from localStorage', async () => {
            const savedMerchants = [
                { id: '1', name: 'Saved Merchant 1', industry: 'Technology' },
                { id: '2', name: 'Saved Merchant 2', industry: 'Finance' }
            ];

            localStorageMock.getItem.mockReturnValue(JSON.stringify(savedMerchants));

            const quickSwitchList = {
                classList: { add: jest.fn(), remove: jest.fn() },
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            await merchantNavigation.loadQuickSwitchMerchants();

            expect(localStorageMock.getItem).toHaveBeenCalledWith('quickSwitchMerchants');
            expect(merchantNavigation.quickSwitchMerchants).toEqual(savedMerchants);
        });

        test('should load quick switch merchants from API when localStorage is empty', async () => {
            localStorageMock.getItem.mockReturnValue(null);

            const quickSwitchList = {
                classList: { add: jest.fn(), remove: jest.fn() },
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            await merchantNavigation.loadQuickSwitchMerchants();

            expect(fetch).toHaveBeenCalledWith('/api/v1/merchants?limit=8');
            expect(merchantNavigation.quickSwitchMerchants).toHaveLength(2);
        });

        test('should render quick switch list with merchants', () => {
            merchantNavigation.quickSwitchMerchants = [
                { id: '1', name: 'Test Merchant 1', industry: 'Technology' },
                { id: '2', name: 'Test Merchant 2', industry: 'Finance' }
            ];

            const quickSwitchList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            merchantNavigation.renderQuickSwitchList();

            expect(quickSwitchList.innerHTML).toContain('Test Merchant 1');
            expect(quickSwitchList.innerHTML).toContain('Test Merchant 2');
        });

        test('should render empty quick switch list', () => {
            merchantNavigation.quickSwitchMerchants = [];

            const quickSwitchList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            merchantNavigation.renderQuickSwitchList();

            expect(quickSwitchList.innerHTML).toContain('No merchants available');
        });

        test('should switch to merchant', () => {
            const merchant = { id: '1', name: 'Test Merchant', industry: 'Technology' };
            merchantNavigation.quickSwitchMerchants = [merchant];

            merchantNavigation.switchToMerchant('1');

            expect(mockSessionManager.switchSession).toHaveBeenCalledWith(merchant);
            expect(merchantNavigation.currentMerchant).toBe(merchant);
        });

        test('should show quick switch modal', () => {
            const modal = {
                style: { display: 'none' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchModal') {
                    return modal;
                }
                return mockElement;
            });

            merchantNavigation.showQuickSwitchModal();

            expect(modal.style.display).toBe('block');
        });

        test('should hide quick switch modal', () => {
            const modal = {
                style: { display: 'block' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchModal') {
                    return modal;
                }
                return mockElement;
            });

            merchantNavigation.hideQuickSwitchModal();

            expect(modal.style.display).toBe('none');
        });

        test('should save quick switch selection', () => {
            merchantNavigation.quickSwitchMerchants = [
                { id: '1', name: 'Test Merchant 1', industry: 'Technology' }
            ];

            const modal = {
                style: { display: 'block' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchModal') {
                    return modal;
                }
                return mockElement;
            });

            merchantNavigation.saveQuickSwitchSelection();

            expect(localStorageMock.setItem).toHaveBeenCalledWith(
                'quickSwitchMerchants',
                JSON.stringify(merchantNavigation.quickSwitchMerchants)
            );
            expect(modal.style.display).toBe('none');
        });
    });

    describe('Navigation History Management', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should add to navigation history', () => {
            const historyList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'historyList') {
                    return historyList;
                }
                return mockElement;
            });

            merchantNavigation.addToNavigationHistory('merchant-detail', 'Test Merchant', 'merchant-123');

            expect(merchantNavigation.navigationHistory).toHaveLength(1);
            expect(merchantNavigation.navigationHistory[0].page).toBe('merchant-detail');
            expect(merchantNavigation.navigationHistory[0].title).toBe('Test Merchant');
            expect(merchantNavigation.navigationHistory[0].merchantId).toBe('merchant-123');
        });

        test('should limit navigation history size', () => {
            const historyList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'historyList') {
                    return historyList;
                }
                return mockElement;
            });

            // Add more than 20 items
            for (let i = 0; i < 25; i++) {
                merchantNavigation.addToNavigationHistory('page', `Title ${i}`);
            }

            expect(merchantNavigation.navigationHistory).toHaveLength(20);
        });

        test('should render navigation history', () => {
            merchantNavigation.navigationHistory = [
                {
                    page: 'merchant-detail',
                    title: 'Test Merchant',
                    merchantId: 'merchant-123',
                    timestamp: new Date(),
                    id: 'history-1'
                }
            ];

            const historyList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'historyList') {
                    return historyList;
                }
                return mockElement;
            });

            merchantNavigation.renderNavigationHistory();

            expect(historyList.innerHTML).toContain('Test Merchant');
        });

        test('should clear navigation history', () => {
            merchantNavigation.navigationHistory = [
                { page: 'test', title: 'Test', timestamp: new Date(), id: '1' }
            ];

            const historyList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'historyList') {
                    return historyList;
                }
                return mockElement;
            });

            merchantNavigation.clearNavigationHistory();

            expect(merchantNavigation.navigationHistory).toHaveLength(0);
        });
    });

    describe('Utility Methods', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should update current merchant', () => {
            const merchant = { id: '1', name: 'Test Merchant', industry: 'Technology' };

            merchantNavigation.updateCurrentMerchant(merchant);

            expect(merchantNavigation.currentMerchant).toBe(merchant);
        });

        test('should get page icon', () => {
            expect(merchantNavigation.getPageIcon('home')).toBe('home');
            expect(merchantNavigation.getPageIcon('portfolio')).toBe('th-large');
            expect(merchantNavigation.getPageIcon('merchant-detail')).toBe('building');
            expect(merchantNavigation.getPageIcon('unknown')).toBe('file');
        });

        test('should format timestamp', () => {
            const now = new Date();
            const oneMinuteAgo = new Date(now - 60000);
            const oneHourAgo = new Date(now - 3600000);
            const oneDayAgo = new Date(now - 86400000);

            expect(merchantNavigation.formatTimestamp(oneMinuteAgo)).toBe('1m ago');
            expect(merchantNavigation.formatTimestamp(oneHourAgo)).toBe('1h ago');
            expect(merchantNavigation.formatTimestamp(oneDayAgo)).toMatch(/\d+\/\d+\/\d+/);
        });

        test('should generate history ID', () => {
            const id1 = merchantNavigation.generateHistoryId();
            const id2 = merchantNavigation.generateHistoryId();

            expect(id1).toMatch(/^history_\d+_[a-z0-9]+$/);
            expect(id2).toMatch(/^history_\d+_[a-z0-9]+$/);
            expect(id1).not.toBe(id2);
        });
    });

    describe('Public API Methods', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should get current merchant', () => {
            const merchant = { id: '1', name: 'Test Merchant' };
            merchantNavigation.currentMerchant = merchant;

            expect(merchantNavigation.getCurrentMerchant()).toBe(merchant);
        });

        test('should get navigation history', () => {
            const history = [{ page: 'test', title: 'Test' }];
            merchantNavigation.navigationHistory = history;

            const result = merchantNavigation.getNavigationHistory();
            expect(result).toEqual(history);
            expect(result).not.toBe(history); // Should return a copy
        });

        test('should get quick switch merchants', () => {
            const merchants = [{ id: '1', name: 'Test Merchant' }];
            merchantNavigation.quickSwitchMerchants = merchants;

            const result = merchantNavigation.getQuickSwitchMerchants();
            expect(result).toEqual(merchants);
            expect(result).not.toBe(merchants); // Should return a copy
        });

        test('should set quick switch merchants', () => {
            const merchants = [{ id: '1', name: 'Test Merchant' }];

            const quickSwitchList = {
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            merchantNavigation.setQuickSwitchMerchants(merchants);

            expect(merchantNavigation.quickSwitchMerchants).toEqual(merchants);
        });

        test('should toggle breadcrumb visibility', () => {
            const breadcrumb = {
                style: { display: 'block' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'breadcrumbNavigation') {
                    return breadcrumb;
                }
                return mockElement;
            });

            merchantNavigation.toggleBreadcrumb();
            expect(breadcrumb.style.display).toBe('none');

            merchantNavigation.toggleBreadcrumb();
            expect(breadcrumb.style.display).toBe('block');
        });

        test('should toggle quick switch visibility', () => {
            const quickSwitch = {
                style: { display: 'block' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchNavigation') {
                    return quickSwitch;
                }
                return mockElement;
            });

            merchantNavigation.toggleQuickSwitch();
            expect(quickSwitch.style.display).toBe('none');

            merchantNavigation.toggleQuickSwitch();
            expect(quickSwitch.style.display).toBe('block');
        });

        test('should toggle history visibility', () => {
            const history = {
                style: { display: 'block' }
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'navigationHistory') {
                    return history;
                }
                return mockElement;
            });

            merchantNavigation.toggleHistory();
            expect(history.style.display).toBe('none');

            merchantNavigation.toggleHistory();
            expect(history.style.display).toBe('block');
        });

        test('should destroy component', () => {
            merchantNavigation.destroy();

            expect(mockContainer.innerHTML).toBe('');
        });
    });

    describe('Error Handling', () => {
        beforeEach(() => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer
            });
        });

        test('should handle fetch errors gracefully', async () => {
            fetch.mockRejectedValue(new Error('Network error'));

            const quickSwitchList = {
                classList: { add: jest.fn(), remove: jest.fn() },
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            await merchantNavigation.loadQuickSwitchMerchants();

            expect(quickSwitchList.classList.remove).toHaveBeenCalledWith('loading');
        });

        test('should handle localStorage errors gracefully', () => {
            localStorageMock.getItem.mockImplementation(() => {
                throw new Error('Storage error');
            });

            const quickSwitchList = {
                classList: { add: jest.fn(), remove: jest.fn() },
                innerHTML: ''
            };

            document.getElementById.mockImplementation((id) => {
                if (id === 'quickSwitchList') {
                    return quickSwitchList;
                }
                return mockElement;
            });

            // Should not throw error
            expect(() => {
                merchantNavigation.loadQuickSwitchMerchants();
            }).not.toThrow();
        });

        test('should handle missing merchant in switch', () => {
            merchantNavigation.quickSwitchMerchants = [];

            // Should not throw error when switching to non-existent merchant
            expect(() => {
                merchantNavigation.switchToMerchant('non-existent');
            }).not.toThrow();
        });
    });

    describe('Integration with Session Manager', () => {
        test('should integrate with session manager callbacks', () => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer,
                sessionManager: mockSessionManager
            });

            expect(mockSessionManager.onSessionStart).toBeDefined();
            expect(mockSessionManager.onSessionEnd).toBeDefined();
            expect(mockSessionManager.onSessionSwitch).toBeDefined();
        });

        test('should handle session start', () => {
            const merchant = { id: '1', name: 'Test Merchant' };
            const session = { merchant };

            merchantNavigation = new MerchantNavigation({
                container: mockContainer,
                sessionManager: mockSessionManager
            });

            // Simulate session start
            mockSessionManager.onSessionStart(session);

            expect(merchantNavigation.currentMerchant).toBe(merchant);
        });

        test('should handle session end', () => {
            merchantNavigation = new MerchantNavigation({
                container: mockContainer,
                sessionManager: mockSessionManager
            });

            merchantNavigation.currentMerchant = { id: '1', name: 'Test Merchant' };

            // Simulate session end
            mockSessionManager.onSessionEnd({});

            expect(merchantNavigation.currentMerchant).toBeNull();
        });

        test('should handle session switch', () => {
            const merchant = { id: '2', name: 'New Merchant' };

            merchantNavigation = new MerchantNavigation({
                container: mockContainer,
                sessionManager: mockSessionManager
            });

            // Simulate session switch
            mockSessionManager.onSessionSwitch(merchant);

            expect(merchantNavigation.currentMerchant).toBe(merchant);
        });
    });
});
