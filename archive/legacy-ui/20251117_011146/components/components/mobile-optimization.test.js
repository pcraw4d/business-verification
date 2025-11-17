/**
 * Mobile Optimization Component Tests
 * 
 * Comprehensive test suite for the mobile optimization component
 * to ensure all mobile features work correctly across different devices.
 */

// Mock DOM environment for testing
const mockDOM = {
    createElement: (tag) => ({
        tagName: tag.toUpperCase(),
        className: '',
        classList: {
            add: jest.fn(),
            remove: jest.fn(),
            contains: jest.fn(() => false)
        },
        setAttribute: jest.fn(),
        getAttribute: jest.fn(() => null),
        querySelector: jest.fn(() => null),
        querySelectorAll: jest.fn(() => []),
        addEventListener: jest.fn(),
        appendChild: jest.fn(),
        innerHTML: '',
        textContent: ''
    }),
    head: {
        appendChild: jest.fn()
    },
    body: {
        classList: {
            add: jest.fn(),
            remove: jest.fn(),
            contains: jest.fn(() => false)
        }
    },
    addEventListener: jest.fn(),
    querySelector: jest.fn(() => null),
    querySelectorAll: jest.fn(() => [])
};

// Mock window object
const mockWindow = {
    innerWidth: 768,
    matchMedia: jest.fn(() => ({ matches: false })),
    navigator: {
        userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)'
    }
};

// Mock document
const mockDocument = {
    readyState: 'complete',
    createElement: mockDOM.createElement,
    head: mockDOM.head,
    body: mockDOM.body,
    addEventListener: jest.fn(),
    querySelector: jest.fn(() => null),
    querySelectorAll: jest.fn(() => [])
};

// Set up global mocks
global.window = mockWindow;
global.document = mockDocument;
global.navigator = mockWindow.navigator;

// Import the component (assuming it's available)
// const MobileOptimization = require('./mobile-optimization.js');

describe('Mobile Optimization Component', () => {
    let mobileOptimization;

    beforeEach(() => {
        // Reset mocks
        jest.clearAllMocks();
        
        // Create new instance
        mobileOptimization = new MobileOptimization({
            enableTouchOptimization: true,
            enableProgressiveEnhancement: true,
            enableAccessibility: true,
            enablePerformanceOptimization: true,
            touchTargetSize: 44
        });
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            const defaultMobileOptimization = new MobileOptimization();
            expect(defaultMobileOptimization.options.enableTouchOptimization).toBe(true);
            expect(defaultMobileOptimization.options.enableProgressiveEnhancement).toBe(true);
            expect(defaultMobileOptimization.options.enableAccessibility).toBe(true);
            expect(defaultMobileOptimization.options.enablePerformanceOptimization).toBe(true);
            expect(defaultMobileOptimization.options.touchTargetSize).toBe(44);
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                enableTouchOptimization: false,
                touchTargetSize: 48
            };
            const customMobileOptimization = new MobileOptimization(customOptions);
            expect(customMobileOptimization.options.enableTouchOptimization).toBe(false);
            expect(customMobileOptimization.options.touchTargetSize).toBe(48);
        });

        test('should detect mobile devices correctly', () => {
            // Test mobile user agent
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)';
            expect(mobileOptimization.detectMobile()).toBe(true);

            // Test desktop user agent
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';
            expect(mobileOptimization.detectMobile()).toBe(false);

            // Test mobile screen width
            mockWindow.innerWidth = 480;
            expect(mobileOptimization.detectMobile()).toBe(true);
        });

        test('should detect touch devices correctly', () => {
            // Test touch support
            mockWindow.ontouchstart = true;
            expect(mobileOptimization.detectTouch()).toBe(true);

            // Test no touch support
            delete mockWindow.ontouchstart;
            mockWindow.navigator.maxTouchPoints = 0;
            expect(mobileOptimization.detectTouch()).toBe(false);
        });
    });

    describe('Mobile Detection', () => {
        test('should detect iPhone as mobile', () => {
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)';
            expect(mobileOptimization.detectMobile()).toBe(true);
        });

        test('should detect Android as mobile', () => {
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36';
            expect(mobileOptimization.detectMobile()).toBe(true);
        });

        test('should detect iPad as mobile', () => {
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15';
            expect(mobileOptimization.detectMobile()).toBe(true);
        });

        test('should detect desktop as not mobile', () => {
            mockWindow.navigator.userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';
            mockWindow.innerWidth = 1920;
            expect(mobileOptimization.detectMobile()).toBe(false);
        });
    });

    describe('Touch Detection', () => {
        test('should detect touch support via ontouchstart', () => {
            mockWindow.ontouchstart = true;
            expect(mobileOptimization.detectTouch()).toBe(true);
        });

        test('should detect touch support via maxTouchPoints', () => {
            mockWindow.navigator.maxTouchPoints = 5;
            expect(mobileOptimization.detectTouch()).toBe(true);
        });

        test('should detect no touch support', () => {
            delete mockWindow.ontouchstart;
            mockWindow.navigator.maxTouchPoints = 0;
            expect(mobileOptimization.detectTouch()).toBe(false);
        });
    });

    describe('Style Injection', () => {
        test('should add mobile optimization styles', () => {
            mobileOptimization.addMobileOptimizationStyles();
            expect(mockDocument.head.appendChild).toHaveBeenCalled();
        });

        test('should include touch target size in styles', () => {
            mobileOptimization.addMobileOptimizationStyles();
            const styleCall = mockDocument.head.appendChild.mock.calls[0][0];
            expect(styleCall.textContent).toContain('44px');
        });
    });

    describe('Touch Enhancement', () => {
        test('should enhance touch interactions when touch is supported', () => {
            mockWindow.ontouchstart = true;
            mobileOptimization.isTouch = true;
            
            // Mock interactive elements
            const mockButton = mockDOM.createElement('button');
            mockDocument.querySelectorAll = jest.fn(() => [mockButton]);
            
            mobileOptimization.enhanceTouchInteractions();
            
            expect(mockButton.classList.add).toHaveBeenCalledWith('touch-target');
        });

        test('should not enhance touch interactions when touch is not supported', () => {
            mobileOptimization.isTouch = false;
            mobileOptimization.enhanceTouchInteractions();
            
            expect(mockDocument.querySelectorAll).not.toHaveBeenCalled();
        });
    });

    describe('Accessibility Enhancement', () => {
        test('should improve accessibility for elements without labels', () => {
            const mockButton = mockDOM.createElement('button');
            mockButton.querySelector = jest.fn(() => ({
                className: 'fas fa-home'
            }));
            mockButton.textContent = '';
            mockButton.getAttribute = jest.fn(() => null);
            
            mockDocument.querySelectorAll = jest.fn(() => [mockButton]);
            
            mobileOptimization.improveAccessibility();
            
            expect(mockButton.setAttribute).toHaveBeenCalledWith('aria-label', 'home');
        });

        test('should add keyboard navigation support', () => {
            mobileOptimization.improveAccessibility();
            expect(mockDocument.addEventListener).toHaveBeenCalledWith('keydown', expect.any(Function));
            expect(mockDocument.addEventListener).toHaveBeenCalledWith('mousedown', expect.any(Function));
        });
    });

    describe('Performance Optimization', () => {
        test('should add performance optimization styles', () => {
            mobileOptimization.optimizePerformance();
            expect(mockDocument.head.appendChild).toHaveBeenCalled();
        });

        test('should add reduced motion support', () => {
            mockWindow.matchMedia = jest.fn(() => ({ matches: true }));
            mobileOptimization.optimizePerformance();
            expect(mockDocument.body.classList.add).toHaveBeenCalledWith('no-animation');
        });
    });

    describe('Progressive Enhancement', () => {
        test('should add progressive enhancement classes', () => {
            mobileOptimization.addProgressiveEnhancement();
            expect(mockDocument.body.classList.add).toHaveBeenCalledWith('mobile-optimized');
        });

        test('should add mobile device class when mobile', () => {
            mobileOptimization.isMobile = true;
            mobileOptimization.addProgressiveEnhancement();
            expect(mockDocument.body.classList.add).toHaveBeenCalledWith('mobile-device');
        });

        test('should add touch device class when touch is supported', () => {
            mobileOptimization.isTouch = true;
            mobileOptimization.addProgressiveEnhancement();
            expect(mockDocument.body.classList.add).toHaveBeenCalledWith('touch-device');
        });

        test('should add viewport meta tag if not present', () => {
            mockDocument.querySelector = jest.fn(() => null);
            mobileOptimization.addProgressiveEnhancement();
            expect(mockDocument.head.appendChild).toHaveBeenCalled();
        });
    });

    describe('Element Optimization', () => {
        test('should optimize element with touch target', () => {
            const mockElement = mockDOM.createElement('div');
            mobileOptimization.isTouch = true;
            
            mobileOptimization.optimizeElement(mockElement, { addTouchTarget: true });
            
            expect(mockElement.classList.add).toHaveBeenCalledWith('touch-target');
            expect(mockElement.classList.add).toHaveBeenCalledWith('mobile-optimized');
        });

        test('should not add touch target when touch is not supported', () => {
            const mockElement = mockDOM.createElement('div');
            mobileOptimization.isTouch = false;
            
            mobileOptimization.optimizeElement(mockElement, { addTouchTarget: true });
            
            expect(mockElement.classList.add).not.toHaveBeenCalledWith('touch-target');
        });

        test('should improve element accessibility', () => {
            const mockButton = mockDOM.createElement('button');
            mockButton.querySelector = jest.fn(() => ({
                className: 'fas fa-save'
            }));
            mockButton.textContent = '';
            mockButton.getAttribute = jest.fn(() => null);
            
            mobileOptimization.improveElementAccessibility(mockButton);
            
            expect(mockButton.setAttribute).toHaveBeenCalledWith('aria-label', 'save');
        });
    });

    describe('Component Creation', () => {
        test('should create mobile-optimized component', () => {
            const component = mobileOptimization.createMobileComponent('div', {
                classes: ['test-class'],
                attributes: { 'data-test': 'value' },
                content: '<span>Test content</span>'
            });
            
            expect(component.tagName).toBe('DIV');
            expect(component.classList.add).toHaveBeenCalledWith('mobile-optimized');
            expect(component.classList.add).toHaveBeenCalledWith('test-class');
            expect(component.setAttribute).toHaveBeenCalledWith('data-test', 'value');
            expect(component.innerHTML).toBe('<span>Test content</span>');
        });
    });

    describe('Status Reporting', () => {
        test('should return correct status', () => {
            mobileOptimization.isMobile = true;
            mobileOptimization.isTouch = true;
            mobileOptimization.isInitialized = true;
            
            const status = mobileOptimization.getStatus();
            
            expect(status.isMobile).toBe(true);
            expect(status.isTouch).toBe(true);
            expect(status.isInitialized).toBe(true);
            expect(status.touchTargetSize).toBe(44);
            expect(status.features.touchOptimization).toBe(true);
            expect(status.features.progressiveEnhancement).toBe(true);
            expect(status.features.accessibility).toBe(true);
            expect(status.features.performanceOptimization).toBe(true);
        });
    });

    describe('Integration Tests', () => {
        test('should initialize completely', () => {
            // Mock all dependencies
            mockDocument.querySelectorAll = jest.fn(() => []);
            mockDocument.querySelector = jest.fn(() => null);
            
            mobileOptimization.init();
            
            expect(mobileOptimization.isInitialized).toBe(true);
            expect(mockDocument.head.appendChild).toHaveBeenCalled();
            expect(mockDocument.body.classList.add).toHaveBeenCalledWith('mobile-optimized');
        });

        test('should handle multiple initializations gracefully', () => {
            mobileOptimization.isInitialized = true;
            mobileOptimization.init();
            
            // Should not call initialization methods again
            expect(mockDocument.head.appendChild).not.toHaveBeenCalled();
        });
    });
});

// Run tests if in Node.js environment
if (typeof module !== 'undefined' && module.exports) {
    // Export test configuration
    module.exports = {
        testEnvironment: 'jsdom',
        setupFilesAfterEnv: ['<rootDir>/jest.setup.js']
    };
}
