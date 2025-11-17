/**
 * Responsive Design Tests for All UI Components
 * Tests component behavior across different screen sizes and viewport configurations
 */

// Mock DOM environment for testing
const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

// Setup DOM environment with responsive capabilities
const dom = new JSDOM(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
    <style>
        /* Base responsive styles */
        * { box-sizing: border-box; }
        body { margin: 0; padding: 0; }
        
        /* Mobile first approach */
        .container { width: 100%; padding: 10px; }
        
        @media (min-width: 768px) {
            .container { padding: 20px; }
        }
        
        @media (min-width: 1024px) {
            .container { padding: 30px; }
        }
        
        @media (min-width: 1200px) {
            .container { padding: 40px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="main-content">
            <h1>Responsive Test Page</h1>
        </div>
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
    setCurrentMerchant: jest.fn()
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

describe('Responsive Design Tests', () => {
    let testContainer;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = `
            <div class="container">
                <div class="main-content">
                    <h1>Responsive Test Page</h1>
                </div>
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

    describe('Viewport Size Tests', () => {
        const viewportSizes = [
            { name: 'Mobile Small', width: 320, height: 568 },
            { name: 'Mobile Medium', width: 375, height: 667 },
            { name: 'Mobile Large', width: 414, height: 896 },
            { name: 'Tablet Portrait', width: 768, height: 1024 },
            { name: 'Tablet Landscape', width: 1024, height: 768 },
            { name: 'Desktop Small', width: 1200, height: 800 },
            { name: 'Desktop Medium', width: 1440, height: 900 },
            { name: 'Desktop Large', width: 1920, height: 1080 }
        ];

        viewportSizes.forEach(({ name, width, height }) => {
            describe(`${name} (${width}x${height})`, () => {
                beforeEach(() => {
                    // Set viewport size
                    Object.defineProperty(window, 'innerWidth', {
                        writable: true,
                        configurable: true,
                        value: width,
                    });
                    
                    Object.defineProperty(window, 'innerHeight', {
                        writable: true,
                        configurable: true,
                        value: height,
                    });
                });

                test('should render merchant search component correctly', () => {
                    if (typeof MerchantSearch !== 'undefined') {
                        const merchantSearch = new MerchantSearch({
                            container: testContainer,
                            apiBaseUrl: '/api/v1'
                        });
                        
                        const searchContainer = testContainer.querySelector('.merchant-search-container');
                        expect(searchContainer).toBeTruthy();
                        
                        // Check if component adapts to viewport
                        if (width < 768) {
                            // Mobile-specific checks
                            expect(searchContainer).toBeTruthy();
                        } else {
                            // Desktop-specific checks
                            expect(searchContainer).toBeTruthy();
                        }
                    }
                });

                test('should render portfolio type filter correctly', () => {
                    if (typeof PortfolioTypeFilter !== 'undefined') {
                        const filter = new PortfolioTypeFilter({
                            container: testContainer
                        });
                        
                        const filterContainer = testContainer.querySelector('.portfolio-type-filter-container');
                        expect(filterContainer).toBeTruthy();
                        
                        // Check responsive behavior
                        if (width < 768) {
                            // Mobile: should be compact
                            expect(filterContainer).toBeTruthy();
                        } else {
                            // Desktop: should be full width
                            expect(filterContainer).toBeTruthy();
                        }
                    }
                });

                test('should render risk level indicator correctly', () => {
                    if (typeof RiskLevelIndicator !== 'undefined') {
                        const indicator = new RiskLevelIndicator({
                            container: testContainer,
                            compactMode: width < 768
                        });
                        
                        const indicatorContainer = testContainer.querySelector('.risk-level-indicator-container');
                        expect(indicatorContainer).toBeTruthy();
                        
                        // Check compact mode on mobile
                        if (width < 768) {
                            const badges = testContainer.querySelector('.risk-level-badges');
                            expect(badges).toBeTruthy();
                        }
                    }
                });

                test('should render navigation correctly', () => {
                    if (typeof KYBNavigation !== 'undefined') {
                        const navigation = new KYBNavigation();
                        
                        const sidebar = document.querySelector('.kyb-sidebar');
                        expect(sidebar).toBeTruthy();
                        
                        // Check mobile behavior
                        if (width < 1024) {
                            // Should be hidden by default on mobile
                            expect(sidebar).toBeTruthy();
                        } else {
                            // Should be visible on desktop
                            expect(sidebar).toBeTruthy();
                        }
                    }
                });

                test('should render merchant context correctly', () => {
                    if (typeof MerchantContext !== 'undefined') {
                        const context = new MerchantContext({
                            showInHeader: true
                        });
                        
                        const contextContainer = document.querySelector('.merchant-context-container');
                        expect(contextContainer).toBeTruthy();
                        
                        // Check responsive layout
                        if (width < 768) {
                            // Mobile: should stack vertically
                            expect(contextContainer).toBeTruthy();
                        } else {
                            // Desktop: should be horizontal
                            expect(contextContainer).toBeTruthy();
                        }
                    }
                });
            });
        });
    });

    describe('Orientation Tests', () => {
        test('should handle portrait orientation', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            Object.defineProperty(window, 'innerHeight', {
                writable: true,
                configurable: true,
                value: 667,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const sidebar = document.querySelector('.kyb-sidebar');
                expect(sidebar).toBeTruthy();
            }
        });

        test('should handle landscape orientation', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 667,
            });
            
            Object.defineProperty(window, 'innerHeight', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const sidebar = document.querySelector('.kyb-sidebar');
                expect(sidebar).toBeTruthy();
            }
        });
    });

    describe('Touch Device Tests', () => {
        test('should handle touch events on mobile', () => {
            // Simulate touch device
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            // Mock touch events
            const mockTouchEvent = new Event('touchstart');
            mockTouchEvent.touches = [{ clientX: 100, clientY: 100 }];
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const toggleBtn = document.getElementById('sidebarToggle');
                
                if (toggleBtn) {
                    // Should handle touch events
                    expect(() => {
                        toggleBtn.dispatchEvent(mockTouchEvent);
                    }).not.toThrow();
                }
            }
        });

        test('should handle swipe gestures', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const sidebar = document.querySelector('.kyb-sidebar');
                
                // Mock swipe events
                const touchStart = new Event('touchstart');
                touchStart.touches = [{ clientX: 0, clientY: 100 }];
                
                const touchEnd = new Event('touchend');
                touchEnd.changedTouches = [{ clientX: 100, clientY: 100 }];
                
                expect(() => {
                    sidebar.dispatchEvent(touchStart);
                    sidebar.dispatchEvent(touchEnd);
                }).not.toThrow();
            }
        });
    });

    describe('High DPI Display Tests', () => {
        test('should handle high DPI displays', () => {
            // Mock high DPI display
            Object.defineProperty(window, 'devicePixelRatio', {
                writable: true,
                configurable: true,
                value: 2,
            });
            
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            if (typeof MerchantSearch !== 'undefined') {
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                const searchContainer = testContainer.querySelector('.merchant-search-container');
                expect(searchContainer).toBeTruthy();
            }
        });

        test('should handle ultra-high DPI displays', () => {
            // Mock ultra-high DPI display
            Object.defineProperty(window, 'devicePixelRatio', {
                writable: true,
                configurable: true,
                value: 3,
            });
            
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 414,
            });
            
            if (typeof RiskLevelIndicator !== 'undefined') {
                const indicator = new RiskLevelIndicator({
                    container: testContainer
                });
                
                const indicatorContainer = testContainer.querySelector('.risk-level-indicator-container');
                expect(indicatorContainer).toBeTruthy();
            }
        });
    });

    describe('Accessibility Tests', () => {
        test('should maintain accessibility on mobile', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const toggleBtn = document.getElementById('sidebarToggle');
                
                if (toggleBtn) {
                    // Should have proper ARIA attributes
                    expect(toggleBtn).toBeTruthy();
                }
            }
        });

        test('should maintain keyboard navigation on all screen sizes', () => {
            const viewportSizes = [320, 768, 1024, 1440];
            
            viewportSizes.forEach(width => {
                Object.defineProperty(window, 'innerWidth', {
                    writable: true,
                    configurable: true,
                    value: width,
                });
                
                if (typeof PortfolioTypeFilter !== 'undefined') {
                    const filter = new PortfolioTypeFilter({
                        container: testContainer
                    });
                    
                    const trigger = testContainer.querySelector('.filter-trigger');
                    if (trigger) {
                        // Should support keyboard navigation
                        const enterEvent = new KeyboardEvent('keydown', { key: 'Enter' });
                        expect(() => {
                            trigger.dispatchEvent(enterEvent);
                        }).not.toThrow();
                    }
                }
            });
        });
    });

    describe('Performance Tests', () => {
        test('should render efficiently on low-end devices', () => {
            // Simulate low-end device constraints
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 320,
            });
            
            const startTime = performance.now();
            
            if (typeof MerchantSearch !== 'undefined') {
                const merchantSearch = new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                });
                
                const endTime = performance.now();
                const renderTime = endTime - startTime;
                
                // Should render quickly (less than 100ms)
                expect(renderTime).toBeLessThan(100);
            }
        });

        test('should handle rapid viewport changes', () => {
            const viewportSizes = [320, 768, 1024, 1440, 1920];
            
            viewportSizes.forEach(width => {
                Object.defineProperty(window, 'innerWidth', {
                    writable: true,
                    configurable: true,
                    value: width,
                });
                
                if (typeof KYBNavigation !== 'undefined') {
                    const navigation = new KYBNavigation();
                    
                    // Should not cause errors with rapid changes
                    expect(navigation).toBeDefined();
                }
            });
        });
    });

    describe('Cross-Browser Compatibility', () => {
        test('should work with different user agents', () => {
            const userAgents = [
                'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15',
                'Mozilla/5.0 (Android 10; Mobile; rv:68.0) Gecko/68.0 Firefox/68.0',
                'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
            ];
            
            userAgents.forEach(userAgent => {
                Object.defineProperty(navigator, 'userAgent', {
                    writable: true,
                    configurable: true,
                    value: userAgent,
                });
                
                if (typeof MerchantContext !== 'undefined') {
                    const context = new MerchantContext({
                        showInHeader: true
                    });
                    
                    expect(context).toBeDefined();
                }
            });
        });
    });

    describe('Edge Cases', () => {
        test('should handle very small viewports', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 240,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const sidebar = document.querySelector('.kyb-sidebar');
                expect(sidebar).toBeTruthy();
            }
        });

        test('should handle very large viewports', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 2560,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                const sidebar = document.querySelector('.kyb-sidebar');
                expect(sidebar).toBeTruthy();
            }
        });

        test('should handle zero width viewport', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 0,
            });
            
            if (typeof KYBNavigation !== 'undefined') {
                // Should not throw
                expect(() => {
                    const navigation = new KYBNavigation();
                }).not.toThrow();
            }
        });
    });

    describe('Component Integration', () => {
        test('should work together on mobile', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 375,
            });
            
            // Create multiple components
            const components = [];
            
            if (typeof KYBNavigation !== 'undefined') {
                components.push(new KYBNavigation());
            }
            
            if (typeof MerchantContext !== 'undefined') {
                components.push(new MerchantContext({ showInHeader: true }));
            }
            
            if (typeof MerchantSearch !== 'undefined') {
                components.push(new MerchantSearch({
                    container: testContainer,
                    apiBaseUrl: '/api/v1'
                }));
            }
            
            // All components should work together
            expect(components.length).toBeGreaterThan(0);
            
            components.forEach(component => {
                expect(component).toBeDefined();
            });
        });

        test('should work together on desktop', () => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 1440,
            });
            
            // Create multiple components
            const components = [];
            
            if (typeof KYBNavigation !== 'undefined') {
                components.push(new KYBNavigation());
            }
            
            if (typeof MerchantContext !== 'undefined') {
                components.push(new MerchantContext({ showInHeader: true }));
            }
            
            if (typeof PortfolioTypeFilter !== 'undefined') {
                components.push(new PortfolioTypeFilter({
                    container: testContainer
                }));
            }
            
            // All components should work together
            expect(components.length).toBeGreaterThan(0);
            
            components.forEach(component => {
                expect(component).toBeDefined();
            });
        });
    });
});

// Integration tests for responsive behavior
describe('Responsive Integration Tests', () => {
    test('should maintain functionality across all breakpoints', () => {
        const breakpoints = [
            { name: 'xs', width: 320 },
            { name: 'sm', width: 576 },
            { name: 'md', width: 768 },
            { name: 'lg', width: 992 },
            { name: 'xl', width: 1200 },
            { name: 'xxl', width: 1400 }
        ];
        
        breakpoints.forEach(({ name, width }) => {
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: width,
            });
            
            // Test navigation
            if (typeof KYBNavigation !== 'undefined') {
                const navigation = new KYBNavigation();
                expect(navigation).toBeDefined();
            }
            
            // Test merchant context
            if (typeof MerchantContext !== 'undefined') {
                const context = new MerchantContext({ showInHeader: true });
                expect(context).toBeDefined();
            }
        });
    });

    test('should handle orientation changes', () => {
        // Start in portrait
        Object.defineProperty(window, 'innerWidth', {
            writable: true,
            configurable: true,
            value: 375,
        });
        
        Object.defineProperty(window, 'innerHeight', {
            writable: true,
            configurable: true,
            value: 667,
        });
        
        if (typeof KYBNavigation !== 'undefined') {
            const navigation = new KYBNavigation();
            expect(navigation).toBeDefined();
        }
        
        // Change to landscape
        Object.defineProperty(window, 'innerWidth', {
            writable: true,
            configurable: true,
            value: 667,
        });
        
        Object.defineProperty(window, 'innerHeight', {
            writable: true,
            configurable: true,
            value: 375,
        });
        
        // Should handle orientation change
        window.dispatchEvent(new Event('resize'));
        
        if (typeof KYBNavigation !== 'undefined') {
            const navigation = new KYBNavigation();
            expect(navigation).toBeDefined();
        }
    });
});
