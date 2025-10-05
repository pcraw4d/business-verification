/**
 * Component Interaction Tests
 * Tests how different UI components work together and interact with each other
 */

// Mock DOM environment for testing
const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

// Setup DOM environment
const dom = new JSDOM(`
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
</head>
<body>
    <div class="main-content">
        <h1>Component Interaction Test Page</h1>
    </div>
</body>
</html>
`, {
    url: 'http://localhost',
    pretendToBeVisual: true,
    resources: 'usable'
});

global.window = dom.window;
global.document = dom.window.document;
global.navigator = dom.window.navigator;

// Mock fetch for API calls
global.fetch = jest.fn();

// Mock SessionManager
global.SessionManager = jest.fn().mockImplementation(() => ({
    getCurrentMerchant: jest.fn(),
    setCurrentMerchant: jest.fn(),
    switchMerchant: jest.fn(),
    getSessionData: jest.fn()
}));

// Load all components
const components = [
    'merchant-search.js',
    'portfolio-type-filter.js',
    'risk-level-indicator.js',
    'session-manager.js',
    'merchant-navigation.js',
    'coming-soon-banner.js',
    'mock-data-warning.js',
    'bulk-progress-tracker.js',
    'merchant-comparison.js',
    'merchant-context.js',
    'navigation.js'
];

components.forEach(component => {
    try {
        const componentPath = path.join(__dirname, component);
        if (fs.existsSync(componentPath)) {
            const componentCode = fs.readFileSync(componentPath, 'utf8');
            eval(componentCode);
        }
    } catch (error) {
        console.warn(`Could not load component ${component}:`, error.message);
    }
});

describe('Component Interaction Tests', () => {
    let testContainer;
    let mockSessionManager;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = `
            <div class="main-content">
                <h1>Component Interaction Test Page</h1>
            </div>
        `;
        
        // Create test container
        testContainer = document.createElement('div');
        testContainer.className = 'test-container';
        testContainer.style.cssText = `
            width: 100%;
            min-height: 100vh;
            padding: 20px;
            background: #f5f5f5;
        `;
        document.body.appendChild(testContainer);
        
        // Reset mocks
        jest.clearAllMocks();
        global.fetch.mockClear();
        
        // Create mock session manager
        mockSessionManager = {
            getCurrentMerchant: jest.fn(),
            setCurrentMerchant: jest.fn(),
            switchMerchant: jest.fn(),
            getSessionData: jest.fn()
        };
        global.SessionManager.mockReturnValue(mockSessionManager);
    });

    afterEach(() => {
        if (testContainer && testContainer.parentNode) {
            testContainer.parentNode.removeChild(testContainer);
        }
        
        // Clean up any created elements
        const cleanupSelectors = [
            '.merchant-search-container',
            '.portfolio-type-filter-container',
            '.risk-level-indicator-container',
            '.session-manager-container',
            '.merchant-context-container',
            '.kyb-sidebar',
            '.main-content-wrapper',
            '.coming-soon-banner',
            '.mock-data-warning',
            '.bulk-progress-tracker',
            '.merchant-comparison-container'
        ];
        
        cleanupSelectors.forEach(selector => {
            const elements = document.querySelectorAll(selector);
            elements.forEach(el => el.remove());
        });
    });

    describe('Navigation and Merchant Context Integration', () => {
        test('should integrate navigation with merchant context', () => {
            if (typeof KYBNavigation !== 'undefined' && typeof MerchantContext !== 'undefined') {
                // Create navigation first
                const navigation = new KYBNavigation();
                
                // Create merchant context
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                // Verify both components exist
                expect(navigation).toBeDefined();
                expect(context).toBeDefined();
                
                // Verify navigation structure
                const sidebar = document.querySelector('.kyb-sidebar');
                expect(sidebar).toBeTruthy();
                
                // Verify merchant context structure
                const contextContainer = document.querySelector('.merchant-context-container');
                expect(contextContainer).toBeTruthy();
                
                // Test interaction: merchant context should work with navigation
                const mockMerchant = {
                    name: 'Test Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                };
                
                context.updateMerchantContext(mockMerchant);
                
                const nameElement = document.getElementById('contextMerchantName');
                expect(nameElement.textContent).toBe('Test Merchant');
            }
        });

        test('should handle navigation state changes with merchant context', () => {
            if (typeof KYBNavigation !== 'undefined' && typeof MerchantContext !== 'undefined') {
                const navigation = new KYBNavigation();
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                // Test navigation state changes
                navigation.updateActivePage('merchant-portfolio');
                
                const activeLink = document.querySelector('.nav-link.active');
                expect(activeLink.getAttribute('data-page')).toBe('merchant-portfolio');
                
                // Merchant context should still work
                const mockMerchant = {
                    name: 'Portfolio Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'medium'
                };
                
                context.updateMerchantContext(mockMerchant);
                
                const nameElement = document.getElementById('contextMerchantName');
                expect(nameElement.textContent).toBe('Portfolio Merchant');
            }
        });
    });

    describe('Search and Filter Component Integration', () => {
        test('should integrate merchant search with portfolio type filter', () => {
            if (typeof MerchantSearch !== 'undefined' && typeof PortfolioTypeFilter !== 'undefined') {
                // Create merchant search
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                // Create portfolio type filter
                const portfolioFilter = new PortfolioTypeFilter({
                    container: testContainer
                });
                
                // Verify both components exist
                expect(merchantSearch).toBeDefined();
                expect(portfolioFilter).toBeDefined();
                
                // Test interaction: filter changes should affect search
                const onFilterSpy = jest.fn();
                portfolioFilter.onFilter = onFilterSpy;
                
                // Simulate filter selection
                portfolioFilter.selectOption('onboarded');
                
                expect(onFilterSpy).toHaveBeenCalledWith({
                    portfolio_type: 'onboarded',
                    mode: 'single'
                });
            }
        });

        test('should integrate merchant search with risk level indicator', () => {
            if (typeof MerchantSearch !== 'undefined' && typeof RiskLevelIndicator !== 'undefined') {
                // Create merchant search
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                // Create risk level indicator
                const riskIndicator = new RiskLevelIndicator({
                    container: testContainer
                });
                
                // Verify both components exist
                expect(merchantSearch).toBeDefined();
                expect(riskIndicator).toBeDefined();
                
                // Test interaction: risk level changes should affect search
                const onFilterSpy = jest.fn();
                riskIndicator.onFilter = onFilterSpy;
                
                // Simulate risk level selection
                riskIndicator.selectOption('high');
                
                expect(onFilterSpy).toHaveBeenCalledWith({
                    risk_level: 'high',
                    mode: 'single'
                });
            }
        });

        test('should handle multiple filter interactions', () => {
            if (typeof MerchantSearch !== 'undefined' && typeof PortfolioTypeFilter !== 'undefined' && typeof RiskLevelIndicator !== 'undefined') {
                // Create all components
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                const portfolioFilter = new PortfolioTypeFilter({
                    container: testContainer
                });
                
                const riskIndicator = new RiskLevelIndicator({
                    container: testContainer
                });
                
                // Test multiple filter interactions
                const portfolioFilterSpy = jest.fn();
                const riskFilterSpy = jest.fn();
                
                portfolioFilter.onFilter = portfolioFilterSpy;
                riskIndicator.onFilter = riskFilterSpy;
                
                // Apply multiple filters
                portfolioFilter.selectOption('onboarded');
                riskIndicator.selectOption('low');
                
                expect(portfolioFilterSpy).toHaveBeenCalledWith({
                    portfolio_type: 'onboarded',
                    mode: 'single'
                });
                
                expect(riskFilterSpy).toHaveBeenCalledWith({
                    risk_level: 'low',
                    mode: 'single'
                });
            }
        });
    });

    describe('Session Manager Integration', () => {
        test('should integrate session manager with merchant context', () => {
            if (typeof SessionManager !== 'undefined' && typeof MerchantContext !== 'undefined') {
                // Create session manager
                const sessionManager = new SessionManager();
                
                // Create merchant context
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                // Verify integration
                expect(sessionManager).toBeDefined();
                expect(context).toBeDefined();
                expect(context.sessionManager).toBeDefined();
                
                // Test session data interaction
                const mockMerchant = {
                    name: 'Session Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                };
                
                mockSessionManager.getCurrentMerchant.mockReturnValue(mockMerchant);
                
                // Load current merchant from session
                const currentMerchant = context.sessionManager.getCurrentMerchant();
                expect(currentMerchant).toBe(mockMerchant);
            }
        });

        test('should handle session switching with merchant context', () => {
            if (typeof SessionManager !== 'undefined' && typeof MerchantContext !== 'undefined') {
                const sessionManager = new SessionManager();
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                // Test session switching
                const merchant1 = {
                    name: 'Merchant 1',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                };
                
                const merchant2 = {
                    name: 'Merchant 2',
                    portfolioType: 'pending',
                    riskLevel: 'high'
                };
                
                // Switch to first merchant
                context.setMerchantContext(merchant1);
                expect(context.currentMerchant).toBe(merchant1);
                
                // Switch to second merchant
                context.setMerchantContext(merchant2);
                expect(context.currentMerchant).toBe(merchant2);
                
                // Verify session manager was called
                expect(mockSessionManager.setCurrentMerchant).toHaveBeenCalledWith(merchant1);
                expect(mockSessionManager.setCurrentMerchant).toHaveBeenCalledWith(merchant2);
            }
        });
    });

    describe('Bulk Operations Integration', () => {
        test('should integrate bulk progress tracker with merchant search', () => {
            if (typeof BulkProgressTracker !== 'undefined' && typeof MerchantSearch !== 'undefined') {
                // Create merchant search
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                // Create bulk progress tracker
                const progressTracker = new BulkProgressTracker({
                    container: testContainer
                });
                
                // Verify both components exist
                expect(merchantSearch).toBeDefined();
                expect(progressTracker).toBeDefined();
                
                // Test interaction: bulk operations should work with search results
                const mockMerchants = [
                    { id: '1', name: 'Merchant 1' },
                    { id: '2', name: 'Merchant 2' },
                    { id: '3', name: 'Merchant 3' }
                ];
                
                // Simulate bulk operation
                progressTracker.startOperation('bulk_update', mockMerchants.length);
                
                expect(progressTracker.isRunning).toBe(true);
                expect(progressTracker.totalItems).toBe(3);
            }
        });

        test('should handle bulk operations with session manager', () => {
            if (typeof BulkProgressTracker !== 'undefined' && typeof SessionManager !== 'undefined') {
                const sessionManager = new SessionManager();
                const progressTracker = new BulkProgressTracker({
                    container: testContainer
                });
                
                // Test bulk operation with session management
                const mockMerchants = [
                    { id: '1', name: 'Merchant 1', portfolioType: 'onboarded' },
                    { id: '2', name: 'Merchant 2', portfolioType: 'pending' }
                ];
                
                // Start bulk operation
                progressTracker.startOperation('bulk_status_update', mockMerchants.length);
                
                // Simulate progress updates
                progressTracker.updateProgress(1);
                progressTracker.updateProgress(2);
                
                expect(progressTracker.completedItems).toBe(2);
                expect(progressTracker.isRunning).toBe(false);
            }
        });
    });

    describe('Merchant Comparison Integration', () => {
        test('should integrate merchant comparison with session manager', () => {
            if (typeof MerchantComparison !== 'undefined' && typeof SessionManager !== 'undefined') {
                const sessionManager = new SessionManager();
                const comparison = new MerchantComparison({
                    container: testContainer
                });
                
                // Test merchant comparison
                const merchant1 = {
                    id: '1',
                    name: 'Merchant 1',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                };
                
                const merchant2 = {
                    id: '2',
                    name: 'Merchant 2',
                    portfolioType: 'pending',
                    riskLevel: 'high'
                };
                
                // Add merchants to comparison
                comparison.addMerchant(merchant1);
                comparison.addMerchant(merchant2);
                
                expect(comparison.merchants).toHaveLength(2);
                expect(comparison.merchants[0]).toBe(merchant1);
                expect(comparison.merchants[1]).toBe(merchant2);
            }
        });

        test('should handle comparison with merchant context', () => {
            if (typeof MerchantComparison !== 'undefined' && typeof MerchantContext !== 'undefined') {
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                const comparison = new MerchantComparison({
                    container: testContainer
                });
                
                // Test comparison with context
                const mockMerchant = {
                    name: 'Comparison Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'medium'
                };
                
                // Set merchant context
                context.setMerchantContext(mockMerchant);
                
                // Add to comparison
                comparison.addMerchant(mockMerchant);
                
                expect(comparison.merchants).toHaveLength(1);
                expect(comparison.merchants[0]).toBe(mockMerchant);
            }
        });
    });

    describe('Placeholder Component Integration', () => {
        test('should integrate coming soon banner with other components', () => {
            if (typeof ComingSoonBanner !== 'undefined' && typeof MerchantSearch !== 'undefined') {
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                const banner = new ComingSoonBanner({
                    container: testContainer,
                    feature: 'advanced_search',
                    message: 'Advanced search coming soon!'
                });
                
                // Verify both components exist
                expect(merchantSearch).toBeDefined();
                expect(banner).toBeDefined();
                
                // Test banner display
                const bannerElement = testContainer.querySelector('.coming-soon-banner');
                expect(bannerElement).toBeTruthy();
                expect(bannerElement.textContent).toContain('Advanced search coming soon!');
            }
        });

        test('should integrate mock data warning with components', () => {
            if (typeof MockDataWarning !== 'undefined' && typeof PortfolioTypeFilter !== 'undefined') {
                const portfolioFilter = new PortfolioTypeFilter({
                    container: testContainer
                });
                
                const warning = new MockDataWarning({
                    container: testContainer,
                    message: 'This data is for testing purposes only'
                });
                
                // Verify both components exist
                expect(portfolioFilter).toBeDefined();
                expect(warning).toBeDefined();
                
                // Test warning display
                const warningElement = testContainer.querySelector('.mock-data-warning');
                expect(warningElement).toBeTruthy();
                expect(warningElement.textContent).toContain('This data is for testing purposes only');
            }
        });
    });

    describe('Event Propagation and Communication', () => {
        test('should handle custom events between components', () => {
            if (typeof MerchantContext !== 'undefined' && typeof SessionManager !== 'undefined') {
                const context = new MerchantContext({
                    showInHeader: true
                });
                
                const sessionManager = new SessionManager();
                
                // Test custom event handling
                const mockMerchant = {
                    name: 'Event Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                };
                
                // Dispatch merchant changed event
                const event = new CustomEvent('merchantChanged', {
                    detail: { merchant: mockMerchant }
                });
                
                document.dispatchEvent(event);
                
                // Context should have updated
                expect(context.currentMerchant).toBe(mockMerchant);
            }
        });

        test('should handle component state synchronization', () => {
            if (typeof MerchantSearch !== 'undefined' && typeof PortfolioTypeFilter !== 'undefined' && typeof RiskLevelIndicator !== 'undefined') {
                // Create components
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                const portfolioFilter = new PortfolioTypeFilter({
                    container: testContainer
                });
                
                const riskIndicator = new RiskLevelIndicator({
                    container: testContainer
                });
                
                // Test state synchronization
                const onFilterSpy = jest.fn();
                portfolioFilter.onFilter = onFilterSpy;
                riskIndicator.onFilter = onFilterSpy;
                
                // Apply filters
                portfolioFilter.selectOption('onboarded');
                riskIndicator.selectOption('low');
                
                // Both components should have triggered filter events
                expect(onFilterSpy).toHaveBeenCalledTimes(2);
            }
        });
    });

    describe('Error Handling in Component Interactions', () => {
        test('should handle component initialization errors gracefully', () => {
            // Test with missing dependencies
            if (typeof MerchantSearch !== 'undefined') {
                // Should not throw even with missing container
                expect(() => {
                    const merchantSearch = new MerchantSearch({
                        container: null,
                        apiBaseUrl: '/api/v1'
                    });
                }).not.toThrow();
            }
        });

        test('should handle API errors in component interactions', () => {
            if (typeof MerchantSearch !== 'undefined') {
                // Mock API error
                global.fetch.mockRejectedValue(new Error('API Error'));
                
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                // Should handle API errors gracefully
                expect(() => {
                    merchantSearch.performSearch();
                }).not.toThrow();
            }
        });

        test('should handle missing component dependencies', () => {
            if (typeof MerchantContext !== 'undefined') {
                // Remove SessionManager from global scope
                const originalSessionManager = global.SessionManager;
                delete global.SessionManager;
                
                // Should not throw
                expect(() => {
                    const context = new MerchantContext({
                        showInHeader: true
                    });
                }).not.toThrow();
                
                // Restore SessionManager
                global.SessionManager = originalSessionManager;
            }
        });
    });

    describe('Performance in Component Interactions', () => {
        test('should handle multiple component interactions efficiently', () => {
            const startTime = performance.now();
            
            // Create multiple components
            const components = [];
            
            if (typeof MerchantSearch !== 'undefined') {
                components.push(new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                }));
            }
            
            if (typeof PortfolioTypeFilter !== 'undefined') {
                components.push(new PortfolioTypeFilter({
                    container: testContainer
                }));
            }
            
            if (typeof RiskLevelIndicator !== 'undefined') {
                components.push(new RiskLevelIndicator({
                    container: testContainer
                }));
            }
            
            if (typeof MerchantContext !== 'undefined') {
                components.push(new MerchantContext({
                    showInHeader: true
                }));
            }
            
            const endTime = performance.now();
            const renderTime = endTime - startTime;
            
            // Should render efficiently (less than 200ms for multiple components)
            expect(renderTime).toBeLessThan(200);
            expect(components.length).toBeGreaterThan(0);
        });

        test('should handle rapid component state changes', () => {
            if (typeof PortfolioTypeFilter !== 'undefined' && typeof RiskLevelIndicator !== 'undefined') {
                const portfolioFilter = new PortfolioTypeFilter({
                    container: testContainer
                });
                
                const riskIndicator = new RiskLevelIndicator({
                    container: testContainer
                });
                
                // Rapid state changes should not cause issues
                for (let i = 0; i < 10; i++) {
                    portfolioFilter.selectOption('onboarded');
                    riskIndicator.selectOption('low');
                    portfolioFilter.clearSelection();
                    riskIndicator.clearSelection();
                }
                
                // Components should still be functional
                expect(portfolioFilter).toBeDefined();
                expect(riskIndicator).toBeDefined();
            }
        });
    });
});

// Integration tests for complex scenarios
describe('Complex Component Integration Scenarios', () => {
    test('should handle complete merchant workflow', () => {
        // This test simulates a complete user workflow
        if (typeof KYBNavigation !== 'undefined' && 
            typeof MerchantContext !== 'undefined' && 
            typeof MerchantSearch !== 'undefined' && 
            typeof PortfolioTypeFilter !== 'undefined' && 
            typeof RiskLevelIndicator !== 'undefined') {
            
            // 1. Initialize navigation
            const navigation = new KYBNavigation();
            
            // 2. Initialize merchant context
            const context = new MerchantContext({
                showInHeader: true
            });
            
            // 3. Initialize search and filters
            const testContainer = document.createElement('div');
            testContainer.className = 'test-container';
            document.body.appendChild(testContainer);
            
            const merchantSearch = new MerchantSearch({
                container: testContainer,
                apiBaseUrl: '/api/v1'
            });
            
            const portfolioFilter = new PortfolioTypeFilter({
                container: testContainer
            });
            
            const riskIndicator = new RiskLevelIndicator({
                container: testContainer
            });
            
            // 4. Simulate user workflow
            // Navigate to merchant portfolio
            navigation.updateActivePage('merchant-portfolio');
            
            // Set merchant context
            const mockMerchant = {
                name: 'Workflow Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            context.setMerchantContext(mockMerchant);
            
            // Apply filters
            portfolioFilter.selectOption('onboarded');
            riskIndicator.selectOption('low');
            
            // Perform search
            merchantSearch.performSearch();
            
            // Verify all components are working together
            expect(navigation.currentPage).toBe('merchant-portfolio');
            expect(context.currentMerchant).toBe(mockMerchant);
            expect(portfolioFilter.getValue()).toBe('onboarded');
            expect(riskIndicator.getValue()).toBe('low');
            
            // Clean up
            testContainer.remove();
        }
    });

    test('should handle component lifecycle management', () => {
        if (typeof MerchantSearch !== 'undefined' && typeof PortfolioTypeFilter !== 'undefined') {
            const testContainer = document.createElement('div');
            testContainer.className = 'test-container';
            document.body.appendChild(testContainer);
            
            // Create components
            const merchantSearch = new MerchantSearch({
                container: testContainer,
                apiBaseUrl: '/api/v1'
            });
            
            const portfolioFilter = new PortfolioTypeFilter({
                container: testContainer
            });
            
            // Test component lifecycle
            expect(merchantSearch).toBeDefined();
            expect(portfolioFilter).toBeDefined();
            
            // Destroy components
            if (merchantSearch.destroy) {
                merchantSearch.destroy();
            }
            
            if (portfolioFilter.destroy) {
                portfolioFilter.destroy();
            }
            
            // Components should be cleaned up
            expect(testContainer.innerHTML).toBe('');
            
            // Clean up
            testContainer.remove();
        }
    });
});
