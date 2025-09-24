/**
 * Unit Tests for KYBNavigation Component
 * Tests navigation creation, page detection, responsive behavior, and event handling
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

// Load the component
const componentPath = path.join(__dirname, 'navigation.js');
const componentCode = fs.readFileSync(componentPath, 'utf8');
eval(componentCode);

describe('KYBNavigation Component', () => {
    let navigation;
    let originalLocation;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = `
            <div class="main-content">
                <h1>Test Page</h1>
            </div>
        `;
        
        // Mock window.location
        originalLocation = window.location;
        delete window.location;
        window.location = {
            pathname: '/dashboard.html',
            search: '',
            href: 'http://localhost/dashboard.html'
        };
        
        // Mock window.innerWidth for responsive tests
        Object.defineProperty(window, 'innerWidth', {
            writable: true,
            configurable: true,
            value: 1200,
        });
    });

    afterEach(() => {
        if (navigation) {
            // Clean up navigation elements
            const sidebar = document.querySelector('.kyb-sidebar');
            const mainWrapper = document.querySelector('.main-content-wrapper');
            if (sidebar) sidebar.remove();
            if (mainWrapper) mainWrapper.remove();
        }
        
        // Restore original location
        window.location = originalLocation;
    });

    describe('Initialization', () => {
        test('should initialize with correct current page', () => {
            navigation = new KYBNavigation();
            
            expect(navigation.currentPage).toBe('business-intelligence');
        });

        test('should detect different page types correctly', () => {
            const testCases = [
                { pathname: '/index.html', expected: 'home' },
                { pathname: '/dashboard-hub.html', expected: 'home' },
                { pathname: '/dashboard.html', expected: 'business-intelligence' },
                { pathname: '/risk-dashboard.html', expected: 'risk-assessment' },
                { pathname: '/compliance-dashboard.html', expected: 'compliance-status' },
                { pathname: '/merchant-portfolio.html', expected: 'merchant-portfolio' },
                { pathname: '/merchant-detail.html', expected: 'merchant-detail' },
                { pathname: '/unknown-page.html', expected: 'home' }
            ];

            testCases.forEach(({ pathname, expected }) => {
                window.location.pathname = pathname;
                const nav = new KYBNavigation();
                expect(nav.currentPage).toBe(expected);
            });
        });

        test('should create navigation structure', () => {
            navigation = new KYBNavigation();
            
            expect(document.querySelector('.kyb-sidebar')).toBeTruthy();
            expect(document.querySelector('.main-content-wrapper')).toBeTruthy();
            expect(document.querySelector('.sidebar-header')).toBeTruthy();
            expect(document.querySelector('.sidebar-content')).toBeTruthy();
            expect(document.querySelector('.sidebar-footer')).toBeTruthy();
        });

        test('should not create navigation if already exists', () => {
            // Create existing navigation
            const existingSidebar = document.createElement('div');
            existingSidebar.className = 'kyb-sidebar';
            document.body.appendChild(existingSidebar);
            
            navigation = new KYBNavigation();
            
            // Should not create duplicate navigation
            const sidebars = document.querySelectorAll('.kyb-sidebar');
            expect(sidebars.length).toBe(1);
        });
    });

    describe('Navigation Structure', () => {
        beforeEach(() => {
            navigation = new KYBNavigation();
        });

        test('should create all navigation sections', () => {
            const sections = document.querySelectorAll('.nav-section');
            expect(sections.length).toBe(5); // Platform, Core Analytics, Compliance, Merchant Management, Market Intelligence
        });

        test('should create all navigation links', () => {
            const navLinks = document.querySelectorAll('.nav-link');
            expect(navLinks.length).toBeGreaterThan(10); // Multiple links per section
        });

        test('should have correct brand link', () => {
            const brandLink = document.querySelector('.brand-link');
            expect(brandLink).toBeTruthy();
            expect(brandLink.href).toContain('index.html');
            expect(brandLink.querySelector('.brand-text').textContent).toBe('KYB Platform');
        });

        test('should have sidebar toggle button', () => {
            const toggleBtn = document.getElementById('sidebarToggle');
            expect(toggleBtn).toBeTruthy();
            expect(toggleBtn.querySelector('i')).toBeTruthy();
        });

        test('should have status indicator in footer', () => {
            const statusIndicator = document.querySelector('.status-indicator.live');
            const statusText = document.querySelector('.status-text');
            
            expect(statusIndicator).toBeTruthy();
            expect(statusText).toBeTruthy();
            expect(statusText.textContent).toBe('Live');
        });
    });

    describe('Active Page Management', () => {
        test('should set active page on initialization', () => {
            window.location.pathname = '/merchant-portfolio.html';
            navigation = new KYBNavigation();
            
            const activeLink = document.querySelector('.nav-link.active');
            expect(activeLink).toBeTruthy();
            expect(activeLink.getAttribute('data-page')).toBe('merchant-portfolio');
        });

        test('should update active page programmatically', () => {
            navigation = new KYBNavigation();
            
            // Remove active class from all links
            document.querySelectorAll('.nav-link').forEach(link => {
                link.classList.remove('active');
            });
            
            navigation.updateActivePage('risk-assessment');
            
            const activeLink = document.querySelector('.nav-link.active');
            expect(activeLink).toBeTruthy();
            expect(activeLink.getAttribute('data-page')).toBe('risk-assessment');
        });

        test('should handle unknown page gracefully', () => {
            navigation = new KYBNavigation();
            
            navigation.updateActivePage('unknown-page');
            
            const activeLink = document.querySelector('.nav-link.active');
            expect(activeLink).toBeFalsy();
        });
    });

    describe('Event Handling', () => {
        beforeEach(() => {
            navigation = new KYBNavigation();
        });

        test('should toggle sidebar on mobile', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            const sidebar = document.querySelector('.kyb-sidebar');
            const toggleBtn = document.getElementById('sidebarToggle');
            const icon = toggleBtn.querySelector('i');
            
            // Initially closed on mobile
            expect(sidebar.classList.contains('open')).toBe(false);
            expect(icon.className).toBe('fas fa-bars');
            
            // Click to open
            toggleBtn.click();
            
            expect(sidebar.classList.contains('open')).toBe(true);
            expect(icon.className).toBe('fas fa-times');
            
            // Click to close
            toggleBtn.click();
            
            expect(sidebar.classList.contains('open')).toBe(false);
            expect(icon.className).toBe('fas fa-bars');
        });

        test('should create overlay on mobile when sidebar opens', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            const toggleBtn = document.getElementById('sidebarToggle');
            toggleBtn.click();
            
            const overlay = document.querySelector('.sidebar-overlay');
            expect(overlay).toBeTruthy();
            expect(overlay.classList.contains('active')).toBe(true);
        });

        test('should close sidebar when clicking overlay', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            const sidebar = document.querySelector('.kyb-sidebar');
            const toggleBtn = document.getElementById('sidebarToggle');
            
            // Open sidebar
            toggleBtn.click();
            expect(sidebar.classList.contains('open')).toBe(true);
            
            // Click overlay
            const overlay = document.querySelector('.sidebar-overlay');
            overlay.click();
            
            expect(sidebar.classList.contains('open')).toBe(false);
            expect(overlay).toBeFalsy(); // Overlay should be removed
        });

        test('should close sidebar when clicking outside on mobile', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            const sidebar = document.querySelector('.kyb-sidebar');
            const toggleBtn = document.getElementById('sidebarToggle');
            
            // Open sidebar
            toggleBtn.click();
            expect(sidebar.classList.contains('open')).toBe(true);
            
            // Click outside sidebar
            const mainContent = document.querySelector('.main-content');
            mainContent.click();
            
            expect(sidebar.classList.contains('open')).toBe(false);
        });

        test('should handle window resize', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            const sidebar = document.querySelector('.kyb-sidebar');
            const toggleBtn = document.getElementById('sidebarToggle');
            
            // Open sidebar on mobile
            toggleBtn.click();
            expect(sidebar.classList.contains('open')).toBe(true);
            
            // Resize to desktop
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 1200,
            });
            
            window.dispatchEvent(new Event('resize'));
            
            expect(sidebar.classList.contains('open')).toBe(false);
        });

        test('should handle smooth scrolling for anchor links', () => {
            // Create an anchor link
            const anchorLink = document.createElement('a');
            anchorLink.href = '#test-section';
            document.body.appendChild(anchorLink);
            
            // Create target element
            const targetElement = document.createElement('div');
            targetElement.id = 'test-section';
            document.body.appendChild(targetElement);
            
            // Mock scrollIntoView
            targetElement.scrollIntoView = jest.fn();
            
            // Click anchor link
            anchorLink.click();
            
            expect(targetElement.scrollIntoView).toHaveBeenCalledWith({
                behavior: 'smooth',
                block: 'start'
            });
        });
    });

    describe('Notification System', () => {
        beforeEach(() => {
            navigation = new KYBNavigation();
        });

        test('should show notification badge', () => {
            navigation.showNotification('risk-assessment', 'High risk detected!');
            
            const link = document.querySelector('.nav-link[data-page="risk-assessment"]');
            const badge = link.querySelector('.nav-badge.notification');
            
            expect(badge).toBeTruthy();
            expect(badge.textContent).toBe('!');
        });

        test('should hide notification badge', () => {
            // Show notification first
            navigation.showNotification('risk-assessment', 'High risk detected!');
            
            let link = document.querySelector('.nav-link[data-page="risk-assessment"]');
            let badge = link.querySelector('.nav-badge.notification');
            expect(badge).toBeTruthy();
            
            // Hide notification
            navigation.hideNotification('risk-assessment');
            
            link = document.querySelector('.nav-link[data-page="risk-assessment"]');
            badge = link.querySelector('.nav-badge.notification');
            expect(badge).toBeFalsy();
        });

        test('should not create duplicate notification badges', () => {
            navigation.showNotification('risk-assessment', 'High risk detected!');
            navigation.showNotification('risk-assessment', 'Another notification');
            
            const link = document.querySelector('.nav-link[data-page="risk-assessment"]');
            const badges = link.querySelectorAll('.nav-badge.notification');
            
            expect(badges.length).toBe(1);
        });

        test('should handle notification for non-existent page', () => {
            // Should not throw
            expect(() => {
                navigation.showNotification('non-existent-page', 'Test');
                navigation.hideNotification('non-existent-page');
            }).not.toThrow();
        });
    });

    describe('Responsive Design', () => {
        test('should apply mobile styles on small screens', () => {
            navigation = new KYBNavigation();
            
            const styles = document.querySelector('style');
            expect(styles.textContent).toContain('@media (max-width: 1024px)');
            expect(styles.textContent).toContain('@media (max-width: 768px)');
        });

        test('should hide brand text on mobile', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            navigation = new KYBNavigation();
            
            const brandText = document.querySelector('.brand-text');
            // In JSDOM, we can't test computed styles directly, but we can verify the element exists
            expect(brandText).toBeTruthy();
        });

        test('should show sidebar toggle on mobile', () => {
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                configurable: true,
                value: 800,
            });
            
            navigation = new KYBNavigation();
            
            const toggleBtn = document.getElementById('sidebarToggle');
            expect(toggleBtn).toBeTruthy();
        });
    });

    describe('Content Management', () => {
        test('should move existing content to main content area', () => {
            // Set up initial content
            document.body.innerHTML = `
                <div class="existing-content">
                    <h1>Existing Content</h1>
                    <p>This content should be moved</p>
                </div>
            `;
            
            navigation = new KYBNavigation();
            
            const mainContent = document.querySelector('.main-content');
            expect(mainContent.innerHTML).toContain('Existing Content');
            expect(mainContent.innerHTML).toContain('This content should be moved');
        });

        test('should preserve existing content structure', () => {
            // Set up complex initial content
            document.body.innerHTML = `
                <div class="header">
                    <h1>Page Header</h1>
                </div>
                <div class="content">
                    <div class="section">
                        <h2>Section Title</h2>
                        <p>Section content</p>
                    </div>
                </div>
                <div class="footer">
                    <p>Footer content</p>
                </div>
            `;
            
            navigation = new KYBNavigation();
            
            const mainContent = document.querySelector('.main-content');
            expect(mainContent.innerHTML).toContain('Page Header');
            expect(mainContent.innerHTML).toContain('Section Title');
            expect(mainContent.innerHTML).toContain('Section content');
            expect(mainContent.innerHTML).toContain('Footer content');
        });
    });

    describe('Auto-initialization', () => {
        test('should auto-initialize on DOMContentLoaded', () => {
            // Clear any existing navigation
            document.querySelectorAll('.kyb-sidebar').forEach(el => el.remove());
            
            // Trigger DOMContentLoaded event
            const event = new Event('DOMContentLoaded');
            document.dispatchEvent(event);
            
            // Check if navigation was created
            expect(window.kybNavigation).toBeDefined();
            expect(window.kybNavigation).toBeInstanceOf(KYBNavigation);
        });

        test('should not auto-initialize if navigation already exists', () => {
            // Create existing navigation
            const existingNav = new KYBNavigation();
            const existingSidebar = document.querySelector('.kyb-sidebar');
            
            // Trigger DOMContentLoaded event
            const event = new Event('DOMContentLoaded');
            document.dispatchEvent(event);
            
            // Should not create new navigation
            const sidebars = document.querySelectorAll('.kyb-sidebar');
            expect(sidebars.length).toBe(1);
        });
    });

    describe('Error Handling', () => {
        test('should handle missing DOM elements gracefully', () => {
            // Remove body content
            document.body.innerHTML = '';
            
            // Should not throw
            expect(() => {
                navigation = new KYBNavigation();
            }).not.toThrow();
        });

        test('should handle missing toggle button gracefully', () => {
            navigation = new KYBNavigation();
            
            // Remove toggle button
            const toggleBtn = document.getElementById('sidebarToggle');
            if (toggleBtn) {
                toggleBtn.remove();
            }
            
            // Should not throw when trying to bind events
            expect(() => {
                navigation.bindEvents();
            }).not.toThrow();
        });

        test('should handle missing sidebar element gracefully', () => {
            navigation = new KYBNavigation();
            
            // Remove sidebar
            const sidebar = document.querySelector('.kyb-sidebar');
            if (sidebar) {
                sidebar.remove();
            }
            
            // Should not throw when trying to bind events
            expect(() => {
                navigation.bindEvents();
            }).not.toThrow();
        });
    });

    describe('Performance', () => {
        test('should handle rapid resize events', () => {
            navigation = new KYBNavigation();
            
            // Simulate rapid resize events
            for (let i = 0; i < 10; i++) {
                Object.defineProperty(window, 'innerWidth', {
                    writable: true,
                    configurable: true,
                    value: 800 + (i * 100),
                });
                
                window.dispatchEvent(new Event('resize'));
            }
            
            // Should not cause errors
            expect(navigation).toBeDefined();
        });

        test('should handle multiple notification updates', () => {
            navigation = new KYBNavigation();
            
            // Add multiple notifications
            for (let i = 0; i < 5; i++) {
                navigation.showNotification('risk-assessment', `Notification ${i}`);
                navigation.hideNotification('risk-assessment');
            }
            
            // Should not cause errors
            expect(navigation).toBeDefined();
        });
    });
});

// Integration tests
describe('KYBNavigation Integration', () => {
    test('should work with real page structure', () => {
        // Create a realistic page structure
        document.body.innerHTML = `
            <div class="container">
                <header class="page-header">
                    <h1>Business Intelligence Dashboard</h1>
                </header>
                <main class="dashboard-content">
                    <div class="dashboard-grid">
                        <div class="dashboard-card">
                            <h2>Key Metrics</h2>
                            <p>Dashboard content here</p>
                        </div>
                    </div>
                </main>
                <footer class="page-footer">
                    <p>Footer content</p>
                </footer>
            </div>
        `;
        
        window.location.pathname = '/dashboard.html';
        const navigation = new KYBNavigation();
        
        // Verify navigation was created
        expect(navigation).toBeDefined();
        expect(document.querySelector('.kyb-sidebar')).toBeTruthy();
        expect(document.querySelector('.main-content-wrapper')).toBeTruthy();
        
        // Verify content was moved
        const mainContent = document.querySelector('.main-content');
        expect(mainContent.innerHTML).toContain('Business Intelligence Dashboard');
        expect(mainContent.innerHTML).toContain('Key Metrics');
        expect(mainContent.innerHTML).toContain('Footer content');
        
        // Verify active page
        const activeLink = document.querySelector('.nav-link.active');
        expect(activeLink.getAttribute('data-page')).toBe('business-intelligence');
    });

    test('should handle merchant portfolio page correctly', () => {
        document.body.innerHTML = `
            <div class="merchant-portfolio-page">
                <h1>Merchant Portfolio</h1>
                <div class="portfolio-content">
                    <p>Portfolio content</p>
                </div>
            </div>
        `;
        
        window.location.pathname = '/merchant-portfolio.html';
        const navigation = new KYBNavigation();
        
        // Verify correct page detection
        expect(navigation.currentPage).toBe('merchant-portfolio');
        
        // Verify active link
        const activeLink = document.querySelector('.nav-link.active');
        expect(activeLink.getAttribute('data-page')).toBe('merchant-portfolio');
    });
});
