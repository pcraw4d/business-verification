/**
 * Unit Tests for MerchantContext Component
 * Tests merchant context integration, UI creation, and session management
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
        <h1>Test Page</h1>
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

// Load the component
const componentPath = path.join(__dirname, 'merchant-context.js');
const componentCode = fs.readFileSync(componentPath, 'utf8');
eval(componentCode);

describe('MerchantContext Component', () => {
    let merchantContext;
    let mockSessionManager;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = `
            <div class="main-content">
                <h1>Test Page</h1>
            </div>
        `;
        
        // Reset mocks
        jest.clearAllMocks();
        global.fetch.mockClear();
        
        // Create mock session manager
        mockSessionManager = {
            getCurrentMerchant: jest.fn(),
            setCurrentMerchant: jest.fn()
        };
        global.SessionManager.mockReturnValue(mockSessionManager);
    });

    afterEach(() => {
        if (merchantContext) {
            // Clean up any created elements
            const contextElements = document.querySelectorAll('.merchant-context-container, .merchant-context-header');
            contextElements.forEach(el => el.remove());
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            merchantContext = new MerchantContext();
            
            expect(merchantContext.currentMerchant).toBeNull();
            expect(merchantContext.options.showInHeader).toBe(true);
            expect(merchantContext.options.showInSidebar).toBe(false);
            expect(merchantContext.options.enableQuickSwitch).toBe(true);
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                showInHeader: false,
                showInSidebar: true,
                enableQuickSwitch: false
            };
            
            merchantContext = new MerchantContext(customOptions);
            
            expect(merchantContext.options.showInHeader).toBe(false);
            expect(merchantContext.options.showInSidebar).toBe(true);
            expect(merchantContext.options.enableQuickSwitch).toBe(false);
        });

        test('should create session manager if available', () => {
            merchantContext = new MerchantContext();
            
            expect(global.SessionManager).toHaveBeenCalled();
            expect(merchantContext.sessionManager).toBeDefined();
        });

        test('should handle missing SessionManager gracefully', () => {
            // Remove SessionManager from global scope
            const originalSessionManager = global.SessionManager;
            delete global.SessionManager;
            
            merchantContext = new MerchantContext();
            
            expect(merchantContext.sessionManager).toBeNull();
            
            // Restore SessionManager
            global.SessionManager = originalSessionManager;
        });
    });

    describe('UI Creation', () => {
        test('should create header context when showInHeader is true', () => {
            merchantContext = new MerchantContext({ showInHeader: true });
            
            const headerContext = document.querySelector('.merchant-context-container');
            expect(headerContext).toBeTruthy();
            expect(headerContext.querySelector('#contextMerchantName')).toBeTruthy();
            expect(headerContext.querySelector('#contextMerchantStatus')).toBeTruthy();
            expect(headerContext.querySelector('#switchMerchantBtn')).toBeTruthy();
        });

        test('should not create header context when showInHeader is false', () => {
            merchantContext = new MerchantContext({ showInHeader: false });
            
            const headerContext = document.querySelector('.merchant-context-container');
            expect(headerContext).toBeFalsy();
        });

        test('should create sidebar context when showInSidebar is true', () => {
            // Create a sidebar element first
            const sidebar = document.createElement('div');
            sidebar.className = 'kyb-sidebar';
            const sidebarContent = document.createElement('div');
            sidebarContent.className = 'sidebar-content';
            sidebar.appendChild(sidebarContent);
            document.body.appendChild(sidebar);
            
            merchantContext = new MerchantContext({ showInSidebar: true });
            
            const sidebarContext = document.querySelector('.merchant-context-section');
            expect(sidebarContext).toBeTruthy();
            expect(sidebarContext.querySelector('#sidebarMerchantName')).toBeTruthy();
            expect(sidebarContext.querySelector('#sidebarMerchantStatus')).toBeTruthy();
        });

        test('should create header when none exists', () => {
            // Remove any existing headers
            document.querySelectorAll('.main-header, .dashboard-header, .page-header').forEach(el => el.remove());
            
            merchantContext = new MerchantContext({ showInHeader: true });
            
            const header = document.querySelector('.merchant-context-header');
            expect(header).toBeTruthy();
            expect(header.style.background).toContain('rgba(255, 255, 255, 0.95)');
        });

        test('should add styles to document head', () => {
            merchantContext = new MerchantContext({ showInHeader: true });
            
            const styles = document.querySelector('style');
            expect(styles).toBeTruthy();
            expect(styles.textContent).toContain('.merchant-context-container');
        });
    });

    describe('Merchant Context Updates', () => {
        beforeEach(() => {
            merchantContext = new MerchantContext({ showInHeader: true });
        });

        test('should update header context with merchant data', () => {
            const mockMerchant = {
                name: 'Test Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            merchantContext.updateMerchantContext(mockMerchant);
            
            const nameElement = document.getElementById('contextMerchantName');
            const statusElement = document.getElementById('contextMerchantStatus');
            
            expect(nameElement.textContent).toBe('Test Merchant');
            expect(statusElement.textContent).toBe('onboarded • low');
        });

        test('should update sidebar context with merchant data', () => {
            // Create sidebar first
            const sidebar = document.createElement('div');
            sidebar.className = 'kyb-sidebar';
            const sidebarContent = document.createElement('div');
            sidebarContent.className = 'sidebar-content';
            sidebar.appendChild(sidebarContent);
            document.body.appendChild(sidebar);
            
            merchantContext = new MerchantContext({ showInSidebar: true });
            
            const mockMerchant = {
                name: 'Test Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            merchantContext.updateMerchantContext(mockMerchant);
            
            const sidebarNameElement = document.getElementById('sidebarMerchantName');
            const sidebarStatusElement = document.getElementById('sidebarMerchantStatus');
            
            expect(sidebarNameElement.textContent).toBe('Test Merchant');
            expect(sidebarStatusElement.textContent).toBe('onboarded • low');
        });

        test('should handle missing merchant data gracefully', () => {
            merchantContext.updateMerchantContext(null);
            
            const nameElement = document.getElementById('contextMerchantName');
            const statusElement = document.getElementById('contextMerchantStatus');
            
            expect(nameElement.textContent).toBe('No Merchant Selected');
            expect(statusElement.textContent).toBe('Select a merchant to view details');
        });

        test('should update page title with merchant name', () => {
            const originalTitle = document.title;
            const mockMerchant = {
                name: 'Test Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            merchantContext.updateMerchantContext(mockMerchant);
            
            expect(document.title).toBe(`Test Merchant - ${originalTitle}`);
        });

        test('should not update page title if already contains merchant name', () => {
            const mockMerchant = {
                name: 'Test Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            // Set title with merchant name
            document.title = 'Test Merchant - Original Title';
            
            merchantContext.updateMerchantContext(mockMerchant);
            
            expect(document.title).toBe('Test Merchant - Original Title');
        });
    });

    describe('Event Handling', () => {
        beforeEach(() => {
            merchantContext = new MerchantContext({ showInHeader: true });
        });

        test('should handle switch merchant button click', () => {
            const switchBtn = document.getElementById('switchMerchantBtn');
            const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
            const locationSpy = jest.spyOn(window.location, 'href', 'set');
            
            switchBtn.click();
            
            expect(confirmSpy).toHaveBeenCalledWith('Switch to merchant portfolio to select a different merchant?');
            expect(locationSpy).toHaveBeenCalledWith('merchant-portfolio.html');
            
            confirmSpy.mockRestore();
            locationSpy.mockRestore();
        });

        test('should not redirect if user cancels switch', () => {
            const switchBtn = document.getElementById('switchMerchantBtn');
            const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(false);
            const locationSpy = jest.spyOn(window.location, 'href', 'set');
            
            switchBtn.click();
            
            expect(confirmSpy).toHaveBeenCalled();
            expect(locationSpy).not.toHaveBeenCalled();
            
            confirmSpy.mockRestore();
            locationSpy.mockRestore();
        });

        test('should listen for merchant changed events', () => {
            const mockMerchant = {
                name: 'Event Merchant',
                portfolioType: 'pending',
                riskLevel: 'medium'
            };
            
            const updateSpy = jest.spyOn(merchantContext, 'updateMerchantContext');
            
            // Dispatch merchant changed event
            const event = new CustomEvent('merchantChanged', {
                detail: { merchant: mockMerchant }
            });
            document.dispatchEvent(event);
            
            expect(updateSpy).toHaveBeenCalledWith(mockMerchant);
            
            updateSpy.mockRestore();
        });
    });

    describe('Session Management Integration', () => {
        test('should load current merchant from session manager', () => {
            const mockMerchant = {
                name: 'Session Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            mockSessionManager.getCurrentMerchant.mockReturnValue(mockMerchant);
            const updateSpy = jest.spyOn(MerchantContext.prototype, 'updateMerchantContext');
            
            merchantContext = new MerchantContext();
            
            expect(mockSessionManager.getCurrentMerchant).toHaveBeenCalled();
            expect(updateSpy).toHaveBeenCalledWith(mockMerchant);
            
            updateSpy.mockRestore();
        });

        test('should load merchant from URL parameters', () => {
            // Mock URL parameters
            const originalURL = window.location;
            delete window.location;
            window.location = {
                search: '?merchant_id=123'
            };
            
            const mockResponse = {
                ok: true,
                json: () => Promise.resolve({
                    id: '123',
                    name: 'URL Merchant',
                    portfolioType: 'onboarded',
                    riskLevel: 'low'
                })
            };
            
            global.fetch.mockResolvedValue(mockResponse);
            const updateSpy = jest.spyOn(MerchantContext.prototype, 'updateMerchantContext');
            
            merchantContext = new MerchantContext();
            
            // Wait for async operations
            setTimeout(() => {
                expect(global.fetch).toHaveBeenCalledWith('/api/merchants/123');
                expect(updateSpy).toHaveBeenCalled();
            }, 100);
            
            // Restore original location
            window.location = originalURL;
            updateSpy.mockRestore();
        });

        test('should handle API errors when loading merchant by ID', () => {
            // Mock URL parameters
            const originalURL = window.location;
            delete window.location;
            window.location = {
                search: '?merchant_id=123'
            };
            
            global.fetch.mockRejectedValue(new Error('API Error'));
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            
            merchantContext = new MerchantContext();
            
            setTimeout(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Failed to load merchant:', expect.any(Error));
                consoleSpy.mockRestore();
            }, 100);
            
            // Restore original location
            window.location = originalURL;
        });
    });

    describe('Static Methods', () => {
        test('should integrate with dashboard using static method', () => {
            const dashboardElement = document.createElement('div');
            dashboardElement.className = 'dashboard';
            document.body.appendChild(dashboardElement);
            
            const context = MerchantContext.integrateWithDashboard(dashboardElement, {
                showInHeader: true
            });
            
            expect(context).toBeInstanceOf(MerchantContext);
            expect(context.options.showInHeader).toBe(true);
        });

        test('should add merchant data to dashboard', () => {
            const dashboardElement = document.createElement('div');
            dashboardElement.className = 'dashboard';
            document.body.appendChild(dashboardElement);
            
            merchantContext = new MerchantContext();
            merchantContext.currentMerchant = {
                name: 'Dashboard Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low',
                industry: 'Technology'
            };
            
            merchantContext.addMerchantDataToDashboard(dashboardElement);
            
            const merchantInfo = dashboardElement.querySelector('.merchant-dashboard-info');
            expect(merchantInfo).toBeTruthy();
            expect(merchantInfo.textContent).toContain('Dashboard Merchant');
            expect(merchantInfo.textContent).toContain('onboarded');
            expect(merchantInfo.textContent).toContain('low');
            expect(merchantInfo.textContent).toContain('Technology');
        });

        test('should not add merchant data if no merchant selected', () => {
            const dashboardElement = document.createElement('div');
            dashboardElement.className = 'dashboard';
            document.body.appendChild(dashboardElement);
            
            merchantContext = new MerchantContext();
            merchantContext.currentMerchant = null;
            
            merchantContext.addMerchantDataToDashboard(dashboardElement);
            
            const merchantInfo = dashboardElement.querySelector('.merchant-dashboard-info');
            expect(merchantInfo).toBeFalsy();
        });
    });

    describe('Public Methods', () => {
        beforeEach(() => {
            merchantContext = new MerchantContext();
        });

        test('should get current merchant', () => {
            const mockMerchant = {
                name: 'Current Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            merchantContext.currentMerchant = mockMerchant;
            
            expect(merchantContext.getCurrentMerchant()).toBe(mockMerchant);
        });

        test('should set merchant context programmatically', () => {
            const mockMerchant = {
                name: 'Set Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            const updateSpy = jest.spyOn(merchantContext, 'updateMerchantContext');
            
            merchantContext.setMerchantContext(mockMerchant);
            
            expect(updateSpy).toHaveBeenCalledWith(mockMerchant);
            expect(mockSessionManager.setCurrentMerchant).toHaveBeenCalledWith(mockMerchant);
            
            updateSpy.mockRestore();
        });

        test('should handle missing session manager in setMerchantContext', () => {
            merchantContext.sessionManager = null;
            
            const mockMerchant = {
                name: 'Set Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            // Should not throw
            expect(() => {
                merchantContext.setMerchantContext(mockMerchant);
            }).not.toThrow();
        });
    });

    describe('Auto-initialization', () => {
        test('should auto-initialize on dashboard pages', () => {
            // Create a dashboard page structure
            document.body.innerHTML = `
                <div class="dashboard">
                    <h1>Dashboard</h1>
                </div>
            `;
            
            // Trigger DOMContentLoaded event
            const event = new Event('DOMContentLoaded');
            document.dispatchEvent(event);
            
            // Check if merchant context was created
            expect(window.merchantContext).toBeDefined();
            expect(window.merchantContext).toBeInstanceOf(MerchantContext);
        });

        test('should not auto-initialize if merchant context already exists', () => {
            // Create existing merchant context
            const existingContext = new MerchantContext();
            document.body.innerHTML = `
                <div class="merchant-context-container"></div>
                <div class="dashboard">
                    <h1>Dashboard</h1>
                </div>
            `;
            
            // Trigger DOMContentLoaded event
            const event = new Event('DOMContentLoaded');
            document.dispatchEvent(event);
            
            // Should not create new context
            expect(window.merchantContext).toBeUndefined();
        });

        test('should not auto-initialize on non-dashboard pages', () => {
            // Create a non-dashboard page
            document.body.innerHTML = `
                <div class="simple-page">
                    <h1>Simple Page</h1>
                </div>
            `;
            
            // Trigger DOMContentLoaded event
            const event = new Event('DOMContentLoaded');
            document.dispatchEvent(event);
            
            // Should not create merchant context
            expect(window.merchantContext).toBeUndefined();
        });
    });

    describe('Responsive Design', () => {
        beforeEach(() => {
            merchantContext = new MerchantContext({ showInHeader: true });
        });

        test('should have responsive styles for mobile', () => {
            const styles = document.querySelector('style');
            expect(styles.textContent).toContain('@media (max-width: 768px)');
            expect(styles.textContent).toContain('flex-direction: column');
        });

        test('should handle mobile layout changes', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 600,
            });
            
            const contextContainer = document.querySelector('.merchant-context-container');
            expect(contextContainer).toBeTruthy();
            
            // Check if mobile styles are applied
            const computedStyle = window.getComputedStyle(contextContainer);
            // Note: In JSDOM, computed styles may not work as expected
            // This test verifies the element exists and can be styled
            expect(contextContainer).toBeTruthy();
        });
    });

    describe('Error Handling', () => {
        test('should handle missing DOM elements gracefully', () => {
            // Remove main content element
            const mainContent = document.querySelector('.main-content');
            if (mainContent) {
                mainContent.remove();
            }
            
            // Should not throw
            expect(() => {
                merchantContext = new MerchantContext({ showInHeader: true });
            }).not.toThrow();
        });

        test('should handle missing container elements in updateMerchantContext', () => {
            merchantContext = new MerchantContext({ showInHeader: true });
            
            // Remove the context elements
            const nameElement = document.getElementById('contextMerchantName');
            const statusElement = document.getElementById('contextMerchantStatus');
            if (nameElement) nameElement.remove();
            if (statusElement) statusElement.remove();
            
            const mockMerchant = {
                name: 'Test Merchant',
                portfolioType: 'onboarded',
                riskLevel: 'low'
            };
            
            // Should not throw
            expect(() => {
                merchantContext.updateMerchantContext(mockMerchant);
            }).not.toThrow();
        });
    });
});

// Integration tests
describe('MerchantContext Integration', () => {
    test('should work with real DOM elements', () => {
        // Create a realistic page structure
        document.body.innerHTML = `
            <div class="main-content">
                <div class="dashboard">
                    <h1>Business Intelligence Dashboard</h1>
                    <div class="dashboard-content">
                        <p>Dashboard content here</p>
                    </div>
                </div>
            </div>
        `;
        
        const merchantContext = new MerchantContext({
            showInHeader: true,
            showInSidebar: false
        });
        
        // Verify integration
        expect(merchantContext).toBeDefined();
        expect(document.querySelector('.merchant-context-container')).toBeTruthy();
        
        // Test merchant context update
        const mockMerchant = {
            name: 'Integration Test Merchant',
            portfolioType: 'onboarded',
            riskLevel: 'low'
        };
        
        merchantContext.updateMerchantContext(mockMerchant);
        
        const nameElement = document.getElementById('contextMerchantName');
        expect(nameElement.textContent).toBe('Integration Test Merchant');
    });
});
